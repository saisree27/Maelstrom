package engine

import (
	"fmt"
	"math"
	"strings"
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

const RFP_DEPTH_MARGIN = 85

// Razoring margins from Carballo
var RAZORING_MARGINS = [3]int{0, 225, 230}

// Futility margins from Blunder
var FUTILITY_MARGINS = [9]int{
	0,
	100, // depth 1
	160, // depth 2
	220, // depth 3
	280, // depth 4
	340, // depth 5
	400, // depth 6
	460, // depth 7
	520, // depth 8
}

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

// Global variables for PVS to keep track of search time
// Defaults to 10 sec/search but these values are changed in searchWithTime
var SearchStartTime time.Time = time.Now()
var AllowedTime int64 = 10000
var SearchStop = false

func QuiescenceSearch(b *Board, limit int, alpha int, beta int, c Color, rd int) int {
	NodesSearched++

	// Check for stop signal
	select {
	case <-StopChannel:
		return 0
	default:
		// Continue search
	}

	if SearchStop {
		return 0
	}

	eval := EvaluateNNUE(b) * COLOR_SIGN[c]

	inCheck := rd <= 2 && b.IsCheck(c)

	if limit <= 0 && !inCheck {
		return eval
	}

	// Beta cutoff
	if eval >= beta && !inCheck {
		return beta
	}

	// Delta pruning
	delta := MG_VALUES[QUEEN]
	if eval < alpha-delta {
		return alpha
	}

	if alpha < eval {
		alpha = eval
	}

	var moves []Move
	if inCheck {
		moves = b.GenerateLegalMoves()
	} else {
		moves = b.GenerateCaptures()
	}

	for _, move := range moves {
		if inCheck || move.movetype == CAPTURE || move.movetype == CAPTURE_AND_PROMOTION || move.movetype == EN_PASSANT {
			b.MakeMove(move)
			score := -QuiescenceSearch(b, limit-1, -beta, -alpha, ReverseColor(c), rd+1)
			b.Undo()

			// Check for stop signal after recursive call
			select {
			case <-StopChannel:
				return 0
			default:
				// Continue
			}

			if SearchStop {
				return 0
			}

			if score >= beta {
				return beta
			}
			if score > alpha {
				alpha = score
			}
		}
	}

	return alpha
}

func moveScore(b *Board, mv Move, pv Move, depth int) int {
	if mv == pv {
		return PV_MOVE_BONUS
	}

	if mv.movetype == CAPTURE || mv.movetype == CAPTURE_AND_PROMOTION {
		score := MVV_LVA_TABLE[PieceToPieceType(mv.captured)][PieceToPieceType(mv.piece)]
		return MVV_LVA_BONUS + score
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

func Pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move) (int, bool) {
	NodesSearched++

	// Check for stop signal
	select {
	case <-StopChannel:
		SearchStop = true
		return 0, true
	default:
		// Continue search
	}

	if SearchStop {
		return 0, true
	}

	if NodesSearched%2047 == 0 && time.Since(SearchStartTime).Milliseconds() > AllowedTime {
		fmt.Printf("stopping search after %d\n", time.Since(SearchStartTime).Milliseconds())
		SearchStop = true
		return 0, true
	}

	isPv := beta > alpha+1
	isRoot := depth == rd
	check := b.IsCheck(c)
	if check && depth < 100 {
		depth++
	}

	if depth <= 0 {
		return QuiescenceSearch(b, 4, alpha, beta, c, rd-depth), false
	}

	// Check for two-fold repetition or 50 move rule. Edge case check from Blunder:
	// ensure that mate in 1 is not possible when checking for 50-move rule.
	possibleCheckmate := check && rd == depth
	if !isRoot && (b.IsTwoFold() || (b.plyCnt50 >= 100 && !possibleCheckmate)) {
		// Don't store repetition positions in TT since their value depends on game history
		return 0, false
	}

	legalMoves := b.GenerateLegalMoves()
	if len(legalMoves) == 0 {
		if b.IsCheck(b.turn) {
			return -WIN_VAL, false
		} else {
			return 0, false
		}
	}

	timeout := false
	bestMove := Move{}

	res, score := ProbeTT(b, alpha, beta, uint8(depth), &bestMove)
	if res && !isRoot {
		return score, false
	}

	// Compute static eval to be used for pruning checks
	staticEval := EvaluateNNUE(b) * COLOR_SIGN[c]

	// REVERSE FUTILITY PRUNING / STATIC NULL MOVE PRUNING
	// Motivation: If the material balance minus a safety margin does not improve beta,
	// 			   then we can fail-high because the child node would not improve alpha.
	// 			   Currently the margin is a constant multiple of depth, this can be improved.
	// Conditions:
	//    1. Not in check
	//    2. Not in PV node
	//    3. TT move exists and is not a capture
	// More info: https://www.chessprogramming.org/Reverse_Futility_Pruning
	if !check && !isPv && bestMove.from != 0 && bestMove.movetype != CAPTURE {
		margin := RFP_DEPTH_MARGIN * depth
		if staticEval-margin >= beta {
			return staticEval - margin, false
		}
	}

	// Pre-allocate PV line for child nodes
	childPV := []Move{}

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
	if doNull && !check && !isPv && notJustPawnsAndKing != 0 {
		b.MakeNullMove()
		R := 3 + depth/6
		score, timeout = Pvs(b, depth-1-R, rd, -beta, -beta+1, ReverseColor(c), false, &childPV)
		score *= -1
		b.UndoNullMove()

		childPV = []Move{}

		if SearchStop || timeout {
			return 0, true
		}

		if score >= beta {
			return score, false
		}
	}

	// RAZORING
	// Motivation: we can prune branches at shallow depths if the static evaluation
	//             with a safety margin is less than alpha and quiescence confirms that
	//             we can fail low.
	// Conditions:
	//    1. Not a PV node
	//    2. TT move does not exist
	// 	  3. Not searching for mate
	// More info: https://www.chessprogramming.org/Razoring
	if depth < len(RAZORING_MARGINS) && !isPv && bestMove.from == 0 && beta < WIN_VAL-100 && -beta > -(WIN_VAL-100) {
		razorMargin := RAZORING_MARGINS[depth]

		if staticEval+razorMargin <= alpha {
			// Try qsearch to verify if position is really bad
			qScore := QuiescenceSearch(b, 4, alpha, beta, c, 4)
			if qScore < alpha {
				return qScore, false
			}
		}
	}

	// INTERNAL ITERATIVE DEEPENING (IID)
	// Motivation: when we don't have a TT move but we're in a PV node, we want to find a
	//             good move to search first rather than go through all moves at high depth.
	//             We can find a good move by first searching at a reduced depth.
	// Conditions:
	//    1. Depth >= 4
	//    2. Is in a PV node
	//    3. TT move does not exist
	// https://www.chessprogramming.org/Internal_Iterative_Deepening
	if depth >= 4 && isPv && bestMove.from == 0 {
		_, _ = Pvs(b, depth-3, rd+1, -beta, -alpha, c, false, line)
		if len(*line) > 0 {
			bestMove = (*line)[0]
		}
		*line = (*line)[:0]
		if SearchStop {
			return 0, true
		}
	}

	pvMove := bestMove
	bestScore := -WIN_VAL - 1
	bestMove = Move{}
	ttFlag := UPPER

	for mvCnt := 0; mvCnt < len(legalMoves); mvCnt++ {
		// BEST MOVE SELECTION (MOVE ORDERING)
		// Motivation: If we select good moves to search first, we can prune later moves.
		move := selectMove(mvCnt, legalMoves, b, pvMove, depth)
		b.MakeMove(move)

		score := 0
		if mvCnt == 0 {
			score, timeout = Pvs(b, depth-1, rd, -beta, -alpha, ReverseColor(c), true, &childPV)
			score *= -1
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
			quietMove := !isPv && move.movetype != CAPTURE && move.movetype != PROMOTION
			R := 0

			if depth >= 3 && quietMove && !check {
				R = LMR_TABLE[depth][mvCnt+1]
			}

			// FUTILITY PRUNING
			// Motivation: We want to discard moves which have no potential of raising alpha. We use a margin to estimate
			//             the potential value of a move (based on depth).
			// Conditions:
			//    1. Quiet, seemingly unimportant move (means not a PV node)
			//    2. Ensure moves pruned don't give check
			// https://www.chessprogramming.org/Futility_Pruning
			checkAfterMove := b.IsCheck(b.turn)
			if quietMove && !checkAfterMove && depth-R-1 < len(FUTILITY_MARGINS) && FUTILITY_MARGINS[depth-R-1]+staticEval < alpha {
				b.Undo()
				continue
			}

			score, timeout = Pvs(b, depth-1-R, rd, -alpha-1, -alpha, ReverseColor(c), true, &childPV)
			score *= -1

			if score > alpha && R > 0 {
				score, timeout = Pvs(b, depth-1, rd, -alpha-1, -alpha, ReverseColor(c), true, &childPV)
				score *= -1
				if score > alpha {
					score, timeout = Pvs(b, depth-1, rd, -beta, -alpha, ReverseColor(c), true, &childPV)
					score *= -1
				}
			} else if score > alpha && score < beta {
				score, timeout = Pvs(b, depth-1, rd, -beta, -alpha, ReverseColor(c), true, &childPV)
				score *= -1
			}
		}

		b.Undo()

		if timeout || SearchStop {
			return 0, true
		}

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score >= beta {
			ttFlag = LOWER
			storeKillerMove(move, depth)
			incrementHistoryTable(move, depth, b.turn)
			break
		} else {
			decrementHistoryTable(move, b.turn)
		}

		if score > alpha {
			alpha = score
			ttFlag = EXACT

			*line = (*line)[:0]
			*line = append(*line, move)
			*line = append(*line, childPV...)

			incrementHistoryTable(move, depth, b.turn)
		} else {
			decrementHistoryTable(move, b.turn)
		}

		childPV = []Move{}
	}

	if !timeout && !SearchStop {
		StoreEntry(b, bestScore, ttFlag, bestMove, uint8(depth))
	}

	return bestScore, timeout
}

func storeKillerMove(move Move, depth int) {
	if move.movetype == QUIET && KillerMovesTable[depth][0].ToUCI() != move.ToUCI() {
		KillerMovesTable[depth][1] = KillerMovesTable[depth][0]
		KillerMovesTable[depth][0] = move
	}
}

func incrementHistoryTable(move Move, depth int, color Color) {
	if move.movetype == QUIET {
		HistoryTable[color][move.from][move.to] += depth * depth
	}

	if HistoryTable[color][move.from][move.to] > SECOND_KILLER_MOVE_BONUS {
		HistoryTable[color][move.from][move.to] /= 2
	}
}

func decrementHistoryTable(move Move, color Color) {
	if move.movetype == QUIET {
		if HistoryTable[color][move.from][move.to] > 0 {
			HistoryTable[color][move.from][move.to]--
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

// SearchWithTime searches for the best move given a time constraint
func SearchWithTime(b *Board, movetime int64) Move {
	if UseOpeningBook {
		// Initialize opening book
		book := NewOpeningBook()

		// Check if we can use a book move
		if bookMove, variation := book.LookupPosition(b, b.turn); bookMove != "" {
			// Convert UCI string to Move
			move := FromUCI(bookMove, b)
			if variation != "" {
				fmt.Printf("info string Book move: %s (%s)\n", bookMove, variation)
			} else {
				fmt.Printf("info string Book move: %s\n", bookMove)
			}
			return move
		}
	}

	// Check tablebase if we're in an endgame position
	if UseTablebase && IsTablebasePosition(b) {
		if score, bestMove, found := ProbeTablebase(b.ToFEN()); found {
			if bestMove != "" {
				// Check if this is a drawing position with multiple moves
				if strings.HasPrefix(bestMove, "draw:") {
					// Get the list of drawing moves
					drawingMoves := strings.Split(strings.TrimPrefix(bestMove, "draw:"), ",")

					// Search these moves to find the best one
					bestScore := -WIN_VAL - 1
					var bestDrawingMove Move
					line := []Move{}
					strLine := ""

					searchStart := time.Now()
					// Try each drawing move
					for _, moveStr := range drawingMoves {
						move := FromUCI(moveStr, b)
						b.MakeMove(move)
						score, _ := Pvs(b, 6, 6, -WIN_VAL-1, WIN_VAL+1, b.turn, true, &line)
						score *= -1 // Negate score since we evaluated from opponent's perspective
						b.Undo()

						if score > bestScore {
							bestScore = score
							bestDrawingMove = move
							strLine = bestDrawingMove.ToUCI()
							for _, m := range line {
								strLine += " " + m.ToUCI()
							}
						}
					}

					fmt.Printf("info depth 6 nodes %d time %d score cp %d pv %s\n", NodesSearched, time.Since(searchStart).Milliseconds(), bestScore*COLOR_SIGN[b.turn], strLine)
					return bestDrawingMove
				}

				// Not a drawing position, use the tablebase move directly
				move := FromUCI(bestMove, b)
				fmt.Printf("info depth 99 nodes 1 time 1 score cp %d pv %s\n", score, bestMove)
				return move
			}
		}
	}

	fmt.Printf("searching for movetime %d\n", movetime)

	SearchStartTime = time.Now()
	AllowedTime = movetime
	SearchStop = false

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

	for i := 1; i <= 100; i++ {
		NodesSearched = 0
		searchStart := time.Now()

		// Aspiration windows
		alpha := -WIN_VAL - 1
		beta := WIN_VAL + 1
		windowSize := 50 // Default window size

		if i > 5 { // Only use aspiration windows after depth 5
			if i > 7 {
				windowSize = 25 // Narrow window for deeper searches
			}
			alpha = prevScore - windowSize
			beta = prevScore + windowSize
		}

		score := 0
		timeout := false

		// Aspiration window search with research on fail
		for {
			score, timeout = Pvs(b, i, i, alpha, beta, b.turn, true, &line)
			if timeout || SearchStop {
				fmt.Printf("returning prev best move after %d\n", time.Since(SearchStartTime).Milliseconds())
				return prevBest
			}

			if score <= alpha {
				alpha = Max(-WIN_VAL-1, alpha-windowSize*2)
				continue
			}
			if score >= beta {
				beta = Min(WIN_VAL+1, beta+windowSize*2)
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
		fmt.Printf("info depth %d nodes %d time %d score cp %d nps %d pv%s\n", i, NodesSearched, timeTaken, score, nps, strLine)

		if score == WIN_VAL || score == -WIN_VAL {
			ClearTT()
			return line[0]
		}

		prevBest = line[0]

		// Check for stop signal after completing a depth
		select {
		case <-StopChannel:
			return prevBest
		default:
			// Continue to next depth
		}
	}
	return prevBest
}
