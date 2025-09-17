package engine

import (
	"fmt"
	"math"
	"strings"
)

// SearchInfo stores global search statistics which will be used
// to report search information via UCI.
type SearchInfo struct {
	NodesSearched    int
	PonderMove       Move
	RootDepth        int
	IsPondering      bool
	PonderingEnabled bool
	NodesPerMove     map[Move]int
}

// SearchStack stores additional information that we will keep on
// the stack and index using ply.
type SearchStack struct {
	staticEval    int
	staticEvalSet bool
	move          Move
}

// Searcher is the primary search thread. All information stored in
// this struct is specific to each individual search.
type Searcher struct {
	Position     *Board
	KillerMoves  [101][2]Move
	History      [2][64][64]int
	CounterMoves [12][64]Move
	ContHist     [12][64][12][64]int
	Info         SearchInfo
}

// This tables stores the pre-computed depth reductions based on
// depth and number of explored moves.
var LMR_TABLE = [101][101]int{}

// Max depth and ply we will search
const MAX_DEPTH = 100
const MAX_PLY = 256

// Quiescence search - utilized at leaf nodes to mitigate the horizon effect
// by calculating all possible captures and only computing a static evaluation
// when the position is quiet.
func (s *Searcher) QuiescenceSearch(alpha int, beta int) int {
	s.Info.NodesSearched++

	if s.Info.NodesSearched%2047 == 0 {
		Timer.CheckPVS(&s.Info)
	}

	if Timer.Stop {
		return 0
	}

	ttMove := Move{}

	// Probe TT in QS, see if we can get a TT cutoff or just get static eval
	probeResult, score, entry := ProbeTT(s.Position, alpha, beta, uint8(0), &ttMove)
	if probeResult == CUTOFF {
		return score
	}

	// If TT move is not a capture, we should not use it in move ordering
	if !ttMove.IsCapture() {
		ttMove = Move{}
	}

	eval := -2 * WIN_VAL
	if probeResult == NULL {
		eval = EvaluateNNUE(s.Position)
	} else {
		eval = int(entry.staticEval)
	}

	// Beta cutoff
	if eval >= beta {
		return beta
	}

	// Delta pruning
	if eval < alpha-1000 {
		return alpha
	}

	if alpha < eval {
		alpha = eval
	}

	mp := NewMovePicker(s, []SearchStack{}, ttMove, Move{}, Move{}, Move{}, 0, true)

	for {
		move := mp.NextMove()
		if move.IsEmpty() {
			break
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

// Principal Variation Search (variation of Negamax) - primary search function.
// The main difference between this and negamax is we explore the first move in
// the move ordering (the principal move) and search it in a large window in an
// attempt to be able to prune future moves. The doNull parameter refers to whether
// we can perform Null Move Pruning at this node, prevMove refers to the move played
// in the parent node (used for updating counter moves), and line refers to the tracked
// principal variation.
func (s *Searcher) Pvs(depth int, ply int, alpha int, beta int, doNull bool, ss []SearchStack, line *[]Move) int {
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
	ss[ply].staticEvalSet = false

	///////////////////////////////////////////////////////////////////////////////
	// CHECK EXTENSION
	// When we encounter a check, we should ensure that we resolve the check. This
	// can be implemented simply by searching for one more ply.
	///////////////////////////////////////////////////////////////////////////////
	if check && depth < 100 {
		depth++
	}

	if depth <= 0 || ply >= MAX_PLY {
		return s.QuiescenceSearch(alpha, beta)
	}

	// Check for two-fold repetition or 50 move rule. Edge case check from Blunder:
	// ensure that mate in 1 is not possible when checking for 50-move rule.
	possibleCheckmate := check && isRoot
	if !isRoot && (s.Position.IsTwoFold() || (s.Position.plyCnt50 >= 100 && !possibleCheckmate)) {
		// Don't store repetition positions in TT since their value depends on game history
		return 0
	}

	ttMove := Move{}

	///////////////////////////////////////////////////////////////////////////////
	// TRANSPOSITION TABLE PROBE
	// The goal is to prune the node entirely using the saved score in TT.
	// Even if this doesn't happen though, we can still utilize saved static eval.
	///////////////////////////////////////////////////////////////////////////////
	probeResult, ttScore, entry := ProbeTT(s.Position, alpha, beta, uint8(depth), &ttMove)
	if probeResult == CUTOFF && !isRoot && !isPv {
		return ttScore
	}

	///////////////////////////////////////////////////////////////////////////////
	// STATIC EVAL CALCULATION/CORRECTION
	// Need the static evaluation of the board position for certain pruning
	// techniques. When the TT probe returns an entry, we can potentially utilize
	// the saved static eval. In many cases we can also use the saved TT score as
	// the static eval, which would be a more accurate assessment of the position.
	///////////////////////////////////////////////////////////////////////////////
	staticEval := -2 * WIN_VAL
	if check || probeResult == NULL {
		staticEval = EvaluateNNUE(s.Position)
	} else {
		staticEval = int(entry.staticEval)
		if ttScore > staticEval && (entry.bd == EXACT || entry.bd == LOWER) {
			staticEval = ttScore
		}
		if ttScore < staticEval && (entry.bd == EXACT || entry.bd == UPPER) {
			staticEval = ttScore
		}
	}
	ss[ply].staticEval = staticEval
	ss[ply].staticEvalSet = true

	///////////////////////////////////////////////////////////////////////////////
	// IMPROVING HEURISTIC
	// We can use the static evaluation difference between plies to determine
	// whether this is a node worth reducing more or less. This will be useful in
	// RFP, LMP, NMP, and LMR techniques.
	///////////////////////////////////////////////////////////////////////////////
	// improving := !check && ply >= 2 && ss[ply-2].staticEvalSet && ss[ply-2].staticEval < ss[ply].staticEval

	// Pre-allocate PV line for child nodes
	childPV := []Move{}

	///////////////////////////////////////////////////////////////////////////////
	// NODE PRUNING CONDITIONS
	// In order to prune the entire node via RFP/NMP/etc we need to ensure that we
	// are not in a root node, aren't in a PV node, and that the position is not in
	// check.
	///////////////////////////////////////////////////////////////////////////////
	if !isRoot && !isPv && !check {
		///////////////////////////////////////////////////////////////////////////////
		// REVERSE FUTILITY PRUNING
		// If the material balance minus a safety margin does not improve beta,
		// then we can fail-high because the child node would not improve alpha.
		// Currently the margin is a constant multiple of depth, this can be improved.
		// More info: https://www.chessprogramming.org/Reverse_Futility_Pruning
		///////////////////////////////////////////////////////////////////////////////
		if depth <= Params.RFP_MAX_DEPTH && !ttMove.IsEmpty() && ttMove.movetype != CAPTURE && beta > -WIN_VAL-100 {
			margin := Params.RFP_MULT * depth
			if staticEval-margin >= beta {
				return beta + (staticEval-beta)/2
			}
		}

		///////////////////////////////////////////////////////////////////////////////
		// RAZORING
		// We can prune branches at shallow depths if the static evaluation with a
		// safety margin is less than alpha and quiescence confirms that we can fail
		// low. However, we need to ensure that we are not searching for mate.
		// More info: https://www.chessprogramming.org/Razoring
		///////////////////////////////////////////////////////////////////////////////
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

		///////////////////////////////////////////////////////////////////////////////
		// NULL MOVE PRUNING
		// The null move would be worse than any possible legal move, so if we can have
		// a reduced search by performing a null move that still fails high we can be
		// relatively sure that the best legal move would also fail high over beta, so
		// we can avoid having to check this node. However, we need to ensure that the
		// side to move doesn't have just a king and pawns.
		// More info: https://www.chessprogramming.org/Null_Move_Pruning
		///////////////////////////////////////////////////////////////////////////////
		notJustPawnsAndKing := s.Position.colors[stm] ^ (s.Position.GetColorPieces(PAWN, stm) | s.Position.GetColorPieces(KING, stm))
		if depth >= Params.NMP_MIN_DEPTH && doNull && notJustPawnsAndKing != 0 && staticEval >= beta {
			s.Position.MakeNullMove()
			R := 4 + depth/3 + Min((staticEval-beta)/200, 3)
			score := -s.Pvs(depth-1-R, ply+2, -beta, -beta+1, false, ss, &childPV)
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

	///////////////////////////////////////////////////////////////////////////////
	// INTERNAL ITERATIVE REDUCTION
	// If there is no TT move, we hope we can safely reduce the depth of node since
	// we did not look at this position in prior searches. We can do IIR on PV
	// nodes and expected cut nodes.
	///////////////////////////////////////////////////////////////////////////////
	if ttMove.IsEmpty() && depth >= Params.IIR_MIN_DEPTH && isPv {
		depth -= Params.IIR_DEPTH_REDUCTION
	}

	pvMove := ttMove
	bestScore := -WIN_VAL - 1
	bestMove := Move{}
	ttFlag := UPPER

	killer1 := s.KillerMoves[depth][0]
	killer2 := s.KillerMoves[depth][1]
	counter := Move{}
	if ply > 0 && !ss[ply-1].move.IsEmpty() {
		counter = s.CounterMoves[ss[ply-1].move.piece][ss[ply-1].move.to]
	}

	mp := NewMovePicker(s, ss, pvMove, killer1, killer2, counter, ply, false)

	var quietsSearched []Move

	mvCnt := 0
	for {
		///////////////////////////////////////////////////////////////////////////////
		// MOVE ORDERING
		// If we select good moves to search first, we can prune later moves.
		///////////////////////////////////////////////////////////////////////////////
		move := mp.NextMove()
		if move.IsEmpty() {
			break
		}

		mvCnt++

		isQuiet := move.IsQuiet()
		lmrDepth := Max(depth-LMR_TABLE[depth][mvCnt], 0)

		if !isRoot && !isPv && bestScore > -WIN_VAL+100 {
			///////////////////////////////////////////////////////////////////////////////
			// LATE MOVE PRUNING
			// With good move ordering, we can skip quiet moves that are late in the move
			// tree. However, at higher depths we should avoid doing as much skipping based
			// on move number, so we apply a quadratic depth-based formula to determine at
			// what point to start pruning moves.
			///////////////////////////////////////////////////////////////////////////////
			if isQuiet && depth <= Params.LMP_MAX_DEPTH && mvCnt > Params.LMP_BASE+Params.LMP_MULT*depth*depth && !check {
				mp.SkipQuiets()
				continue
			}

			///////////////////////////////////////////////////////////////////////////////
			// FUTILITY PRUNING
			// We want to discard moves which have no potential of raising alpha. We use a
			// futilityMargin to estimate the potential value of a move (based on depth).
			// We need to ensure that we futility prune only in quiet positions.
			// More info: https://www.chessprogramming.org/Futility_Pruning
			///////////////////////////////////////////////////////////////////////////////
			futilityMargin := Params.FUTILITY_MULT*lmrDepth + Params.FUTILITY_BASE
			if isQuiet && !check && lmrDepth <= Params.FUTILITY_MAX_DEPTH && staticEval+futilityMargin <= alpha {
				mp.SkipQuiets()
				continue
			}

			///////////////////////////////////////////////////////////////////////////////
			// SEE PRUNING (CAPTURES AND QUIETS)
			// Using Static Exchange Evaluation (SEE), we can determine whether a move is
			// worth exploring by whether it loses a large amount of material. At higher
			// depths we should increase the SEE margin threshold to avoid pruning moves
			// that seem to lose material but are good. As a result we can maintain depth-
			// based formulas for both quiet moves and capture moves.
			///////////////////////////////////////////////////////////////////////////////
			seeMargin := -Params.SEE_QUIET_PRUNING_MULT * lmrDepth * lmrDepth
			if isQuiet && !SEE(move, s.Position, seeMargin) {
				continue
			}

			seeMargin = -Params.SEE_CAPTURE_PRUNING_MULT * depth
			if move.IsCapture() && !SEE(move, s.Position, seeMargin) {
				continue
			}
		}

		s.Position.MakeMove(move)
		ss[ply].move = move
		prevNodes := s.Info.NodesSearched

		if isQuiet {
			quietsSearched = append(quietsSearched, move)
		}

		score := 0
		if mvCnt == 1 {
			score = -s.Pvs(depth-1, ply+1, -beta, -alpha, true, ss, &childPV)
		} else {
			///////////////////////////////////////////////////////////////////////////////
			// LATE MOVE REDUCTION
			// Because of move ordering, we can save time by reducing search depth of moves
			// that are late in order. If these moves fail high, we need to re-search them
			// since we probably missed something. The reduction depth is currently based
			// on a formula used by Ethereal. However, we reduce less when the position is
			// either in check prior or after the move, and reduce more if we are not in a
			// PV node. We also do not reduce at all when the move is a capture.
			// More info: https://www.chessprogramming.org/Late_Move_Reductions
			///////////////////////////////////////////////////////////////////////////////
			R := 0

			if depth >= Params.LMR_MIN_DEPTH && isQuiet {
				R = LMR_TABLE[depth][mvCnt] * 1024

				if s.Position.IsCheck(s.Position.turn) {
					R -= Params.LMR_CHECK
				}

				if !isPv {
					R += Params.LMR_NOT_PV
				}

				if ttMove.IsCapture() {
					R += Params.LMR_TT_CAPTURE
				}

				hist := mp.history[stm][move.from][move.to] + s.getContHist(ss, move, ply, 1)
				R -= hist * 1024 / (HISTORY_MAX_BONUS)

				R = Max(R/1024, 0)
			}

			score = -s.Pvs(depth-1-R, ply+1, -alpha-1, -alpha, true, ss, &childPV)

			if score > alpha && R > 0 {
				score = -s.Pvs(depth-1, ply+1, -alpha-1, -alpha, true, ss, &childPV)
				if score > alpha {
					score = -s.Pvs(depth-1, ply+1, -beta, -alpha, true, ss, &childPV)
				}
			} else if score > alpha && score < beta {
				score = -s.Pvs(depth-1, ply+1, -beta, -alpha, true, ss, &childPV)
			}
		}

		s.Position.Undo()
		ss[ply].move = Move{}

		if isRoot {
			s.Info.NodesPerMove[move] = s.Info.NodesSearched - prevNodes
		}

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
				///////////////////////////////////////////////////////////////////////////////
				// FAIL HIGH QUIET MOVES
				// Quiet moves that fail high can be used later on to enhance move ordering.
				///////////////////////////////////////////////////////////////////////////////
				s.storeKillerMove(move, depth)
				if ply > 0 {
					s.storeCounterMove(move, ss[ply-1].move)
				}

				// History bonus using history gravity formula
				bonus := 300*depth - 250
				s.updateHistory(move, s.Position.turn, bonus)
				s.updateContHist(ss, move, bonus, ply, 1)

				// Malus for all quiet moves we searched prior
				for idx := 0; idx < len(quietsSearched)-1; idx++ {
					move = quietsSearched[idx]
					s.updateHistory(move, s.Position.turn, -bonus)
					s.updateContHist(ss, move, -bonus, ply, 1)
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

	if mvCnt == 0 {
		if check {
			// Calculate mate distance
			return -WIN_VAL + (s.Info.RootDepth - depth + 1)
		} else {
			return 0
		}
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

func (s *Searcher) storeCounterMove(move Move, prevMove Move) {
	if !prevMove.IsEmpty() {
		s.CounterMoves[prevMove.piece][prevMove.to] = move
	}
}

func (s *Searcher) updateHistory(move Move, color Color, bonus int) {
	bonus = Clamp(bonus, -HISTORY_MAX_BONUS, HISTORY_MAX_BONUS)
	absBonus := Abs(bonus)
	s.History[color][move.from][move.to] += bonus - s.History[color][move.from][move.to]*absBonus/HISTORY_MAX_BONUS
}

func (s *Searcher) updateContHist(ss []SearchStack, move Move, bonus int, curPly int, traverse int) {
	bonus = Clamp(bonus, -HISTORY_MAX_BONUS, HISTORY_MAX_BONUS)
	absBonus := Abs(bonus)
	if curPly >= traverse {
		prevMove := ss[curPly-traverse].move
		if !prevMove.IsEmpty() {
			s.ContHist[move.piece][move.to][prevMove.piece][prevMove.to] += bonus - s.ContHist[move.piece][move.to][prevMove.piece][prevMove.to]*absBonus/HISTORY_MAX_BONUS
		}
	}
}

func (s *Searcher) getContHist(ss []SearchStack, move Move, curPly int, traverse int) int {
	if curPly >= traverse {
		prevMove := ss[curPly-traverse].move
		if !prevMove.IsEmpty() {
			return s.ContHist[move.piece][move.to][prevMove.piece][prevMove.to]
		}
	}
	return 0
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

func (s *Searcher) ClearContHist() {
	s.ContHist = [12][64][12][64]int{}
}

func (s *Searcher) ClearCounters() {
	s.CounterMoves = [12][64]Move{}
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
	s.Info.NodesPerMove = map[Move]int{}
}

func (s *Searcher) SearchPosition() Move {
	Timer.StartSearch()
	s.ResetInfo()

	line := []Move{}
	legalMoves := s.Position.GenerateLegalMoves()
	prevScore := 0

	if len(legalMoves) == 1 {
		Timer.hardLimit /= 10
	}

	// If no legal moves, return empty move
	if len(legalMoves) == 0 {
		return Move{}
	}

	// Set prevBest to first legal move in case search is stopped immediately
	prevBest := legalMoves[0]

	for depth := 1; depth <= MAX_DEPTH; depth++ {
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
			searchStack := [MAX_PLY]SearchStack{}
			score = s.Pvs(depth, 0, alpha, beta, true, searchStack[:], &line)
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

				if dist < 3 && dist > -3 {
					return line[0]
				}
			} else {
				fmt.Printf("info depth %d nodes %d time %d score cp %d nps %d pv %s\n", depth, s.Info.NodesSearched, delta, score, nps, strings.Trim(fmt.Sprint(line), "[]"))
			}
		}

		Timer.CheckID(&s.Info, depth)

		if depth > 1 {
			Timer.UpdateSoftLimit(&s.Info, line[0], prevBest, score, prevScore)
		}

		prevBest = line[0]
		prevScore = score
		s.Info.NodesPerMove = map[Move]int{}

		if s.Info.PonderingEnabled {
			if len(line) > 1 {
				s.Info.PonderMove = line[1]
			} else {
				s.Info.PonderMove = Move{}
			}
		}

		if Timer.Stop {
			return prevBest
		}
	}

	IncrementTTAge()
	return prevBest
}
