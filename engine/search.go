package engine

import (
	"fmt"
	"math"
	"strings"
)

var COLOR_SIGN = [2]int{
	WHITE: 1, BLACK: -1,
}

var LMR_TABLE = [101][101]int{}

type SearchInfo struct {
	NodesSearched    int
	PonderMove       Move
	RootDepth        int
	IsPondering      bool
	PonderingEnabled bool
}

type Searcher struct {
	Position    *Board
	KillerMoves [101][2]Move
	History     [2][64][64]int
	Info        SearchInfo
}

func (s *Searcher) QuiescenceSearch(alpha int, beta int) int {
	s.Info.NodesSearched++

	if s.Info.NodesSearched%2047 == 0 {
		Timer.CheckPVS(&s.Info)
	}

	if Timer.Stop {
		return 0
	}

	eval := -2 * WIN_VAL

	// Probe TT in QS, see if we can get a TT cutoff or just get static eval
	res, score := ProbeTT(s.Position, alpha, beta, uint8(0), &Move{}, &eval)
	if res {
		return score
	}

	// If we didn't get static eval from TT
	if eval < -WIN_VAL {
		eval = EvaluateNNUE(s.Position) * COLOR_SIGN[s.Position.turn]
	}

	// Beta cutoff
	if eval >= beta {
		return beta
	}

	if alpha < eval {
		alpha = eval
	}

	moves := s.Position.GenerateCaptures()

	for mvCnt := range moves {
		move := s.SelectMove(mvCnt, moves, Move{}, 0)
		if move.movetype != CAPTURE_AND_PROMOTION && s.SEE(move) < 0 {
			continue
		}

		s.Position.MakeMove(move)
		score := -s.QuiescenceSearch(-beta, -alpha)
		s.Position.Undo()

		if Timer.Stop {
			return 0
		}

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

func (s *Searcher) Pvs(depth int, alpha int, beta int, doNull bool, line *[]Move) int {
	s.Info.NodesSearched++

	if s.Info.NodesSearched%2047 == 0 {
		Timer.CheckPVS(&s.Info)
	}

	if Timer.Stop {
		return 0
	}

	isPv := beta > alpha+1
	isRoot := depth == s.Info.RootDepth
	stm := s.Position.turn
	check := s.Position.IsCheck(stm)

	// CHECK EXTENSION
	if check && depth < 100 {
		depth++
	}

	if depth <= 0 {
		return s.QuiescenceSearch(alpha, beta)
	}

	// Check for two-fold repetition or 50 move rule. Edge case check from Blunder:
	// ensure that mate in 1 is not possible when checking for 50-move rule.
	possibleCheckmate := check && isRoot
	if !isRoot && (s.Position.IsTwoFold() || (s.Position.plyCnt50 >= 100 && !possibleCheckmate)) {
		// Don't store repetition positions in TT since their value depends on game history
		return 0
	}

	bestMove := Move{}

	staticEval := -2 * WIN_VAL

	// Probe TT:
	// - Try to get PV move, static eval
	// - If conditions are met, we can return ttScore from TT
	res, ttScore := ProbeTT(s.Position, alpha, beta, uint8(depth), &bestMove, &staticEval)
	if res && !isRoot && !isPv {
		return ttScore
	}

	// Compute static eval to be used for pruning checks
	if staticEval < -WIN_VAL {
		staticEval = EvaluateNNUE(s.Position) * COLOR_SIGN[stm]
	}

	// Pre-allocate PV line for child nodes
	childPV := []Move{}

	// Conditions for RFP, Razoring, NMP
	if !isRoot && !isPv && !check {
		// REVERSE FUTILITY PRUNING / STATIC NULL MOVE PRUNING
		// Motivation: If the material balance minus a safety margin does not improve beta,
		// 			   then we can fail-high because the child node would not improve alpha.
		// 			   Currently the margin is a constant multiple of depth, this can be improved.
		// Conditions:
		//    1. Not in check
		//    2. Not in PV node
		//    3. TT move exists and is not a capture
		// More info: https://www.chessprogramming.org/Reverse_Futility_Pruning
		if depth <= Params.RFP_MAX_DEPTH && !bestMove.IsEmpty() && bestMove.movetype != CAPTURE && beta > -WIN_VAL-100 {
			margin := Params.RFP_MULT * depth
			if staticEval-margin >= beta {
				return beta + (staticEval-beta)/2
			}
		}

		// RAZORING
		// Motivation: we can prune branches at shallow depths if the static evaluation
		//             with a safety margin is less than alpha and quiescence confirms that
		//             we can fail low.
		// Conditions:
		//    1. Not a PV node
		// 	  2. Not searching for mate
		// More info: https://www.chessprogramming.org/Razoring
		if depth <= Params.RAZORING_MAX_DEPTH && alpha < WIN_VAL/10 && beta > -WIN_VAL/10 {
			razorMargin := Params.RAZORING_MULT * depth
			if staticEval+razorMargin <= alpha {
				// Try qsearch to verify if position is really bad
				qScore := s.QuiescenceSearch(alpha, beta)
				if qScore < alpha {
					return qScore
				}
			}
		}

		// NULL MOVE PRUNING
		// Motivation: The null move would be worse than any possible legal move, so
		// 			   if we can have a reduced search by performing a null move that still fails high
		// 			   we can be relatively sure that the best legal move would also fail high over beta,
		//             so we can avoid having to check this node. Reduced depth here is 3 + d/6, formula from Blunder
		// Conditions:
		//    1. Not in check
		//    2. Side to move does not have just king and pawns
		// 	  3. Not in PV node
		// More info: https://www.chessprogramming.org/Null_Move_Pruning
		notJustPawnsAndKing := s.Position.colors[stm] ^ (s.Position.GetColorPieces(PAWN, stm) | s.Position.GetColorPieces(KING, stm))
		if depth >= Params.NMP_MIN_DEPTH && doNull && notJustPawnsAndKing != 0 && staticEval >= beta {
			s.Position.MakeNullMove()
			R := 4 + depth/3 + Min((staticEval-beta)/200, 3)
			score := -s.Pvs(depth-1-R, -beta, -beta+1, false, &childPV)
			s.Position.UndoNullMove()

			childPV = []Move{}

			if Timer.Stop {
				return 0
			}

			if score >= beta {
				return score
			}
		}
	}

	// INTERNAL ITERATIVE REDUCTION (IIR)
	// Motivation: If there is no TT move, we hope we can safely reduce the depth of node since we
	//             did not look at this position in prior searches. We can do IIR on PV nodes and
	//             expected cut nodes.
	// Conditions:
	//    1. TT move not found
	//    2. Depth is greater than some limit
	//    3. Is in a PV node
	if bestMove.IsEmpty() && depth >= Params.IIR_MIN_DEPTH && isPv {
		depth -= Params.IIR_DEPTH_REDUCTION
	}

	pvMove := bestMove
	bestScore := -WIN_VAL - 1
	bestMove = Move{}
	ttFlag := UPPER

	moves := s.Position.GenerateLegalMoves()
	if len(moves) == 0 {
		if check {
			// Calculate mate distance
			return -WIN_VAL + (s.Info.RootDepth - depth + 1)
		} else {
			return 0
		}
	}

	var quietsSearched []Move

	for mvCnt := 0; mvCnt < len(moves); mvCnt++ {
		// BEST MOVE SELECTION (MOVE ORDERING)
		// Motivation: If we select good moves to search first, we can prune later moves.
		move := s.SelectMove(mvCnt, moves, pvMove, depth)

		isQuiet := move.movetype == QUIET || move.movetype == K_CASTLE || move.movetype == Q_CASTLE
		lmrDepth := Max(depth-LMR_TABLE[depth][mvCnt+1], 0)

		if !isRoot && !isPv && bestScore > -WIN_VAL+100 {
			// FUTILITY PRUNING
			// Motivation: We want to discard moves which have no potential of raising alpha. We use a futilityMargin to estimate
			//             the potential value of a move (based on depth).
			// Conditions:
			//    1. Quiet, seemingly unimportant move (means not a PV node)
			//    2. Ensure we are not searching for mate
			// https://www.chessprogramming.org/Futility_Pruning
			futilityMargin := Params.FUTILITY_MULT*lmrDepth + Params.FUTILITY_BASE
			if isQuiet && !check && lmrDepth <= Params.FUTILITY_MAX_DEPTH && staticEval+futilityMargin <= alpha {
				continue
			}

			// SEE QUIET PRUNING
			seeMargin := -Params.SEE_QUIET_PRUNING_MULT * lmrDepth * lmrDepth
			if isQuiet && s.SEE(move) < seeMargin {
				continue
			}

			// SEE CAPTURE PRUNING
			seeMargin = -Params.SEE_CAPTURE_PRUNING_MULT * depth
			if (move.movetype == CAPTURE || move.movetype == EN_PASSANT) && s.SEE(move) < seeMargin {
				continue
			}
		}

		if isQuiet {
			quietsSearched = append(quietsSearched, move)
		}

		s.Position.MakeMove(move)

		score := 0
		if mvCnt == 0 {
			score = -s.Pvs(depth-1, -beta, -alpha, true, &childPV)
		} else {
			// LATE MOVE REDUCTION (LMR)
			// Motivation: Because of move ordering, we can save time by reducing search depth of moves that are
			//             late in order. If these moves fail high, we need to re-search them since we probably
			//             missed something. The reduction depth is currently based on a formula used by Ethereal
			// Conditions:
			//    1. Quiet, seemingly unimportant move (means not a PV node)
			//    2. Position not in check before move
			//    3. Not a shallow depth
			// https://www.chessprogramming.org/Late_Move_Reductions
			R := 0

			if depth >= Params.LMR_MIN_DEPTH && isQuiet && !isPv && !check {
				R = Max(LMR_TABLE[depth][mvCnt+1], 0)
			}

			score = -s.Pvs(depth-1-R, -alpha-1, -alpha, true, &childPV)

			if score > alpha && R > 0 {
				score = -s.Pvs(depth-1, -alpha-1, -alpha, true, &childPV)
				if score > alpha {
					score = -s.Pvs(depth-1, -beta, -alpha, true, &childPV)
				}
			} else if score > alpha && score < beta {
				score = -s.Pvs(depth-1, -beta, -alpha, true, &childPV)
			}
		}

		s.Position.Undo()

		if Timer.Stop {
			return 0
		}

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score >= beta {
			ttFlag = LOWER
			if isQuiet {
				s.storeKillerMove(move, depth)

				// History bonus using history gravity formula
				bonus := 300*depth - 250
				s.updateHistory(move, s.Position.turn, bonus)

				// Malus for all quiet moves we searched prior
				for idx := 0; idx < len(quietsSearched)-1; idx++ {
					move = quietsSearched[idx]
					s.updateHistory(move, s.Position.turn, -bonus)
				}
			}
			break
		}

		if score > alpha {
			alpha = score
			ttFlag = EXACT

			*line = (*line)[:0]
			*line = append(*line, move)
			*line = append(*line, childPV...)
		}

		childPV = []Move{}
	}

	if !Timer.Stop {
		StoreEntry(s.Position, bestScore, ttFlag, bestMove, uint8(depth), staticEval)
	}

	return bestScore
}

func (s *Searcher) storeKillerMove(move Move, depth int) {
	if s.KillerMoves[depth][0] != move {
		s.KillerMoves[depth][1] = s.KillerMoves[depth][0]
		s.KillerMoves[depth][0] = move
	}
}

func (s *Searcher) updateHistory(move Move, color Color, bonus int) {
	if bonus > HISTORY_MAX_BONUS {
		bonus = HISTORY_MAX_BONUS
	} else if bonus < -HISTORY_MAX_BONUS {
		bonus = -HISTORY_MAX_BONUS
	}

	absBonus := bonus
	if absBonus < 0 {
		absBonus *= -1
	}

	s.History[color][move.from][move.to] += bonus - s.History[color][move.from][move.to]*absBonus/HISTORY_MAX_BONUS
}

func (s *Searcher) ClearKillers() {
	for i := 0; i < len(s.KillerMoves); i++ {
		s.KillerMoves[i][0] = Move{}
		s.KillerMoves[i][1] = Move{}
	}
}

func (s *Searcher) ClearHistory() {
	s.History = [2][64][64]int{}
}

func (s *Searcher) HalfHistory() {
	for c := 0; c < 2; c++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {
				s.History[c][from][to] /= 2
			}
		}
	}
}

func InitializeLMRTable() {
	for depth := 1; depth <= 100; depth++ {
		for moveCnt := 1; moveCnt <= 100; moveCnt++ {
			// Formula from Ethereal
			LMR_TABLE[depth][moveCnt] = int(0.7844 + math.Log(float64(depth))*math.Log(float64(moveCnt))/2.4696)
		}
	}
}

func (s *Searcher) ResetInfo() {
	s.Info.NodesSearched = 0
}

func (s *Searcher) SearchPosition() Move {
	Timer.StartSearch()
	s.ResetInfo()

	line := []Move{}
	legalMoves := s.Position.GenerateLegalMoves()
	prevScore := 0

	if len(legalMoves) == 1 {
		return legalMoves[0]
	}

	// If no legal moves, return empty move
	if len(legalMoves) == 0 {
		return Move{}
	}

	// Set prevBest to first legal move in case search is stopped immediately
	prevBest := legalMoves[0]

	// Half the history heuristic values after each search
	s.HalfHistory()

	for depth := 1; depth <= 100; depth++ {
		s.Info.RootDepth = depth

		// Aspiration windows
		alpha := -WIN_VAL - 1
		beta := WIN_VAL + 1

		alphaWindowSize := -Params.ASPIRATION_WINDOW_SIZE
		betaWindowSize := Params.ASPIRATION_WINDOW_SIZE

		if depth > 5 {
			alpha = prevScore + alphaWindowSize
			beta = prevScore + betaWindowSize
		}

		score := 0

		// Aspiration window search with exponentially-widening research on fail
		for {
			score = s.Pvs(depth, alpha, beta, true, &line)
			if Timer.Stop {
				return prevBest
			}

			if score <= alpha {
				alpha = Max(-WIN_VAL-1, alpha+alphaWindowSize*2)
				alphaWindowSize *= -alphaWindowSize
				continue
			}
			if score >= beta {
				beta = Min(WIN_VAL+1, beta+betaWindowSize*2)
				betaWindowSize *= betaWindowSize
				continue
			}
			break
		}

		prevScore = score

		delta := Max(int(Timer.Delta()), 1)
		nps := s.Info.NodesSearched * 1000 / delta

		if !s.Info.IsPondering {
			// HANDLE MATE SCORES:
			if score >= WIN_VAL-101 || score <= -WIN_VAL+101 {
				dist := 0
				if score > 0 {
					dist = WIN_VAL - score
				} else {
					dist = WIN_VAL + score
				}
				// Convert to moves from plies
				dist = dist/2 + 1
				if score < 0 {
					dist *= -1
				}

				fmt.Printf("info depth %d nodes %d time %d score mate %d nps %d pv %s\n", depth, s.Info.NodesSearched, delta, dist, nps, strings.Trim(fmt.Sprint(line), "[]"))

				if dist < 3 {
					return line[0]
				}
			} else {
				fmt.Printf("info depth %d nodes %d time %d score cp %d nps %d pv %s\n", depth, s.Info.NodesSearched, delta, score, nps, strings.Trim(fmt.Sprint(line), "[]"))
			}
		}

		prevBest = line[0]

		if s.Info.PonderingEnabled {
			if len(line) > 1 {
				s.Info.PonderMove = line[1]
			} else {
				s.Info.PonderMove = Move{}

				// Attempt to get ponder move from TT
				staticEval := 0
				s.Position.MakeMove(prevBest)
				ProbeTT(s.Position, -WIN_VAL-1, WIN_VAL+1, 1, &s.Info.PonderMove, &staticEval)
				s.Position.Undo()
			}
		}

		Timer.CheckID(&s.Info, depth)
		if Timer.Stop {
			return prevBest
		}
	}

	IncrementTTAge()
	return prevBest
}
