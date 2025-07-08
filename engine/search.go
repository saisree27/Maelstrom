package engine

import (
	"fmt"
	"math"
	"time"
)

var COLOR_SIGN = [2]int{
	WHITE: 1, BLACK: -1,
}

const PV_MOVE_BONUS = 2000000
const MVV_LVA_BONUS = 1000000
const PROMOTION_BONUS = 800000
const FIRST_KILLER_MOVE_BONUS = 600000
const SECOND_KILLER_MOVE_BONUS = 590000
const HISTORY_MAX_BONUS = 16384
const BAD_CAPTURE_BONUS = -32768

var MVV_LVA_TABLE = [7][7]int{
	{15, 13, 14, 12, 11, 10, 0}, // victim P, attacker P, B, N, R, Q, K, Empty
	{35, 33, 34, 32, 31, 30, 0}, // victim B, attacker P, B, N, R, Q, K, Empty
	{25, 23, 24, 22, 21, 20, 0}, // victim N, attacker P, B, N, R, Q, K, Empty
	{45, 43, 44, 42, 41, 40, 0}, // victim R, attacker P, B, N, R, Q, K, Empty
	{55, 53, 54, 52, 51, 50, 0}, // victim Q, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim K, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim Empty, attacker P, B, N, R, Q, K, Empty
}

var LMR_TABLE = [101][101]int{}

var KillerMovesTable = [101][2]Move{}
var HistoryTable = [2][64][64]int{}

var NodesSearched = 0

var PonderMove Move

func QuiescenceSearch(b *Board, alpha int, beta int, c Color) int {
	NodesSearched++

	if NodesSearched%2047 == 0 {
		Timer.CheckPVS()
	}

	if Timer.Stop {
		return 0
	}

	eval := EvaluateNNUE(b) * COLOR_SIGN[c]

	// Beta cutoff
	if eval >= beta {
		return beta
	}

	if alpha < eval {
		alpha = eval
	}

	moves := b.GenerateCaptures()

	for mvCnt := range moves {
		move := selectMove(mvCnt, moves, b, Move{}, 0)
		if move.movetype != CAPTURE_AND_PROMOTION && see(b, move) < 0 {
			continue
		}

		b.MakeMove(move)
		score := -QuiescenceSearch(b, -beta, -alpha, ReverseColor(c))
		b.Undo()

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

func moveScore(b *Board, mv Move, pv Move, depth int) int {
	if mv == pv {
		return PV_MOVE_BONUS
	}

	if mv.movetype == CAPTURE || mv.movetype == CAPTURE_AND_PROMOTION || mv.movetype == EN_PASSANT {
		mvvLvaScore := MVV_LVA_TABLE[PieceToPieceType(mv.captured)][PieceToPieceType(mv.piece)]
		// SEE + MVV/LVA move ordering
		// Rank bad SEE captures below quiets
		if mv.movetype != CAPTURE_AND_PROMOTION {
			if see(b, mv) < 0 {
				// Bad SEE capture
				return BAD_CAPTURE_BONUS + mvvLvaScore
			} else {
				return MVV_LVA_BONUS + mvvLvaScore
			}
		} else {
			return MVV_LVA_BONUS + mvvLvaScore
		}
	} else if mv.movetype == PROMOTION {
		return PROMOTION_BONUS // Promotions are valuable
	} else {
		// Quiet moves
		if mv == KillerMovesTable[depth][0] {
			return FIRST_KILLER_MOVE_BONUS // First killer move
		} else if mv == KillerMovesTable[depth][1] {
			return SECOND_KILLER_MOVE_BONUS // Second killer move
		} else {
			return HistoryTable[b.turn][mv.from][mv.to]
		}
	}
}

// Selection-based move ordering. This is more efficient than sorting
// all moves by score since we will likely have cutoffs early in the mvoe list.
func selectMove(idx int, moves []Move, b *Board, pv Move, depth int) Move {
	bestScore := moveScore(b, moves[idx], pv, depth)
	bestIndex := idx
	for j := idx + 1; j < len(moves); j++ {
		score := moveScore(b, moves[j], pv, depth)
		if score > bestScore {
			bestScore = score
			bestIndex = j
		}
	}

	moves[idx], moves[bestIndex] = moves[bestIndex], moves[idx]
	return moves[idx]
}

func Pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move) int {
	NodesSearched++

	if NodesSearched%2047 == 0 {
		Timer.CheckPVS()
	}

	if Timer.Stop {
		return 0
	}

	isPv := beta > alpha+1
	isRoot := depth == rd
	check := b.IsCheck(c)

	// CHECK EXTENSION
	if check && depth < 100 {
		depth++
	}

	if depth <= 0 {
		return QuiescenceSearch(b, alpha, beta, c)
	}

	// Check for two-fold repetition or 50 move rule. Edge case check from Blunder:
	// ensure that mate in 1 is not possible when checking for 50-move rule.
	possibleCheckmate := check && rd == depth
	if !isRoot && (b.IsTwoFold() || (b.plyCnt50 >= 100 && !possibleCheckmate)) {
		// Don't store repetition positions in TT since their value depends on game history
		return 0
	}

	bestMove := Move{}

	res, score := ProbeTT(b, alpha, beta, uint8(depth), &bestMove)
	if res && !isRoot {
		return score
	}

	// Compute static eval to be used for pruning checks
	staticEval := EvaluateNNUE(b) * COLOR_SIGN[c]

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
		if depth <= Params.RFP_MAX_DEPTH && bestMove.from != bestMove.to && bestMove.movetype != CAPTURE && beta > -WIN_VAL-100 {
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
				qScore := QuiescenceSearch(b, alpha, beta, c)
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
		notJustPawnsAndKing := b.colors[c] ^ (b.GetColorPieces(PAWN, c) | b.GetColorPieces(KING, c))
		if depth >= Params.NMP_MIN_DEPTH && doNull && notJustPawnsAndKing != 0 && staticEval >= beta {
			b.MakeNullMove()
			R := 4 + depth/3 + Min((staticEval-beta)/200, 3)
			score = -Pvs(b, depth-1-R, rd, -beta, -beta+1, ReverseColor(c), false, &childPV)
			b.UndoNullMove()

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
	if bestMove.from == bestMove.to && depth >= Params.IIR_MIN_DEPTH && isPv {
		depth -= Params.IIR_DEPTH_REDUCTION
	}

	pvMove := bestMove
	bestScore := -WIN_VAL - 1
	bestMove = Move{}
	ttFlag := UPPER

	legalMoves := b.GenerateLegalMoves()
	if len(legalMoves) == 0 {
		if b.IsCheck(b.turn) {
			return -WIN_VAL
		} else {
			return 0
		}
	}

	var quietsSearched []Move

	for mvCnt := 0; mvCnt < len(legalMoves); mvCnt++ {
		// BEST MOVE SELECTION (MOVE ORDERING)
		// Motivation: If we select good moves to search first, we can prune later moves.
		move := selectMove(mvCnt, legalMoves, b, pvMove, depth)

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
			if isQuiet && see(b, move) < seeMargin {
				continue
			}

			// SEE CAPTURE PRUNING
			seeMargin = -Params.SEE_CAPTURE_PRUNING_MULT * depth
			if (move.movetype == CAPTURE || move.movetype == EN_PASSANT) && see(b, move) < seeMargin {
				continue
			}
		}

		if isQuiet {
			quietsSearched = append(quietsSearched, move)
		}

		b.MakeMove(move)

		score := 0
		if mvCnt == 0 {
			score = -Pvs(b, depth-1, rd, -beta, -alpha, ReverseColor(c), true, &childPV)
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

			score = -Pvs(b, depth-1-R, rd, -alpha-1, -alpha, ReverseColor(c), true, &childPV)

			if score > alpha && R > 0 {
				score = -Pvs(b, depth-1, rd, -alpha-1, -alpha, ReverseColor(c), true, &childPV)
				if score > alpha {
					score = -Pvs(b, depth-1, rd, -beta, -alpha, ReverseColor(c), true, &childPV)
				}
			} else if score > alpha && score < beta {
				score = -Pvs(b, depth-1, rd, -beta, -alpha, ReverseColor(c), true, &childPV)
			}
		}

		b.Undo()

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
				storeKillerMove(move, depth)

				// History bonus using history gravity formula
				bonus := 300*depth - 250
				updateHistory(move, b.turn, bonus)

				// Malus for all quiet moves we searched prior
				for idx := 0; idx < len(quietsSearched)-1; idx++ {
					move = quietsSearched[idx]
					updateHistory(move, b.turn, -bonus)
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
		StoreEntry(b, bestScore, ttFlag, bestMove, uint8(depth))
	}

	return bestScore
}

func storeKillerMove(move Move, depth int) {
	if KillerMovesTable[depth][0] != move {
		KillerMovesTable[depth][1] = KillerMovesTable[depth][0]
		KillerMovesTable[depth][0] = move
	}
}

func updateHistory(move Move, color Color, bonus int) {
	if bonus > HISTORY_MAX_BONUS {
		bonus = HISTORY_MAX_BONUS
	} else if bonus < -HISTORY_MAX_BONUS {
		bonus = -HISTORY_MAX_BONUS
	}

	absBonus := bonus
	if absBonus < 0 {
		absBonus *= -1
	}

	HistoryTable[color][move.from][move.to] += bonus - HistoryTable[color][move.from][move.to]*absBonus/HISTORY_MAX_BONUS
}

func ClearKillers() {
	for i := 0; i < len(KillerMovesTable); i++ {
		KillerMovesTable[i][0] = Move{}
		KillerMovesTable[i][1] = Move{}
	}
}

func ClearHistory() {
	HistoryTable = [2][64][64]int{}
}

func HalfHistory() {
	for c := 0; c < 2; c++ {
		for from := 0; from < 64; from++ {
			for to := 0; to < 64; to++ {
				HistoryTable[c][from][to] /= 2
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

func SearchPosition(b *Board) Move {
	Timer.StartSearch()

	if IS_PONDERING {
		fmt.Println("pondering")
	}

	if !IS_PONDERING {
		Timer.PrintConditions()
	}

	line := []Move{}
	legalMoves := b.GenerateLegalMoves()
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
	HalfHistory()

	for depth := 1; depth <= 100; depth++ {
		NodesSearched = 0
		searchStart := time.Now()

		// Aspiration windows
		alpha := -WIN_VAL - 1
		beta := WIN_VAL + 1

		alphaWindowSize := -25
		betaWindowSize := 25

		if depth > 5 {
			alpha = prevScore + alphaWindowSize
			beta = prevScore + betaWindowSize
		}

		score := 0

		// Aspiration window search with exponentially-widening research on fail
		for {
			score = Pvs(b, depth, depth, alpha, beta, b.turn, true, &line)
			if Timer.Stop {
				if !IS_PONDERING {
					fmt.Printf("returning prev best move after %d\n", Timer.Delta())
				}
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
		timeTaken := time.Since(searchStart).Milliseconds()
		timeTakenNanoSeconds := time.Since(searchStart).Nanoseconds()

		IncrementTTAge()

		strLine := ""
		for _, move := range line {
			strLine += " " + move.ToUCI()
		}

		nps := NodesSearched * 1000000000 / Max(int(timeTakenNanoSeconds), 1)

		if !IS_PONDERING {
			fmt.Printf("info depth %d nodes %d time %d score cp %d nps %d pv%s\n", depth, NodesSearched, timeTaken, score, nps, strLine)
		}

		if score == WIN_VAL || score == -WIN_VAL {
			// Don't return out of search early if pondering
			if !IS_PONDERING {
				return line[0]
			}
		}

		prevBest = line[0]

		if IS_PONDERING {
			if len(line) > 1 {
				PonderMove = line[1]
			} else {
				PonderMove = Move{}

				// Attempt to get ponder move from TT
				b.MakeMove(prevBest)
				ProbeTT(b, -WIN_VAL-1, WIN_VAL+1, 1, &PonderMove)
				b.Undo()
			}
		}

		Timer.CheckID(depth)
		if Timer.Stop {
			if !IS_PONDERING {
				fmt.Printf("returning prev best move after %d\n", Timer.Delta())
			}
			return prevBest
		}
	}
	return prevBest
}
