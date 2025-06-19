package engine

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type moveBonus struct {
	move  Move
	bonus int
}

const PVMoveBonus = 2000000
const MVVLVABonus = 1000000
const promotionBonus = 800000
const firstKillerMoveBonus = 600000
const secondKillerMoveBonus = 590000

var nodesSearched = 0
var killerMoves = [100][2]Move{}
var historyHeuristic = [2][64][64]int{}
var flagStop = false

var MVVLVA = [7][7]int{
	{15, 13, 14, 12, 11, 10, 0}, // victim P, attacker P, B, N, R, Q, K, Empty
	{35, 33, 34, 32, 31, 30, 0}, // victim B, attacker P, B, N, R, Q, K, Empty
	{25, 23, 24, 22, 21, 20, 0}, // victim N, attacker P, B, N, R, Q, K, Empty
	{45, 43, 44, 42, 41, 40, 0}, // victim R, attacker P, B, N, R, Q, K, Empty
	{55, 53, 54, 52, 51, 50, 0}, // victim Q, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim K, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim Empty, attacker P, B, N, R, Q, K, Empty
}

// Futility margins from Blunder
var futilityMargins = [9]int{
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

var lmrTable = [101][101]int{}

// Global variables for PVS to keep track of search time
// Defaults to 10 sec/search but these values are changed in searchWithTime
var startTime time.Time = time.Now()
var allowedTime int64 = 10000

func quiesce(b *Board, limit int, alpha int, beta int, c Color, rd int) int {
	nodesSearched++

	// Check for stop signal
	select {
	case <-stopChannel:
		return 0
	default:
		// Continue search
	}

	if flagStop {
		return 0
	}

	eval := evaluate(b) * factor[c]

	inCheck := rd <= 2 && b.isCheck(c)

	if limit <= 0 && !inCheck {
		return eval
	}

	// Beta cutoff
	if eval >= beta && !inCheck {
		return beta
	}

	// Delta pruning
	delta := mgValues[queen]
	if eval < alpha-delta {
		return alpha
	}

	if alpha < eval {
		alpha = eval
	}

	allLegalMoves := b.generateLegalMoves()

	for _, move := range allLegalMoves {
		if inCheck || move.movetype == CAPTURE || move.movetype == CAPTUREANDPROMOTION || move.movetype == ENPASSANT {
			b.makeMove(move)
			score := -quiesce(b, limit-1, -beta, -alpha, reverseColor(c), rd+1)
			b.undo()

			// Check for stop signal after recursive call
			select {
			case <-stopChannel:
				return 0
			default:
				// Continue
			}

			if flagStop {
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

func orderMovesPV(b *Board, moves *[]Move, pv Move, c Color, depth int) {
	// Pre-allocate bonuses slice to avoid reallocations
	bonuses := make([]moveBonus, len(*moves))

	// First pass: handle PV move and calculate bonuses
	for i, mv := range *moves {
		if mv == pv {
			(*moves)[i], (*moves)[0] = (*moves)[0], pv
			bonuses[i] = moveBonus{move: mv, bonus: PVMoveBonus} // Highest priority for PV moves
			continue
		}

		var bonus int
		if mv.movetype == CAPTURE || mv.movetype == CAPTUREANDPROMOTION {
			score := MVVLVA[pieceToPieceType(mv.captured)][pieceToPieceType(mv.piece)]
			bonus = MVVLVABonus + score
		} else if mv.movetype == PROMOTION {
			bonus = promotionBonus // Promotions are valuable
		} else {
			// Quiet moves
			if mv == killerMoves[depth][0] {
				bonus = firstKillerMoveBonus // First killer move
			} else if mv == killerMoves[depth][1] {
				bonus = secondKillerMoveBonus // Second killer move
			} else {
				// History heuristic for remaining quiet moves
				bonus = historyHeuristic[b.turn][mv.from][mv.to]
			}
		}
		bonuses[i] = moveBonus{move: mv, bonus: bonus}
	}

	// Sort moves in-place using the bonuses
	sort.Slice(bonuses, func(i, j int) bool {
		return bonuses[i].bonus > bonuses[j].bonus
	})

	// Update moves slice in-place
	for i := range bonuses {
		(*moves)[i] = bonuses[i].move
	}
}

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move) (int, bool) {
	nodesSearched++

	// Check for stop signal
	select {
	case <-stopChannel:
		flagStop = true
		return 0, true
	default:
		// Continue search
	}

	if flagStop {
		return 0, true
	}

	if nodesSearched%2047 == 0 && time.Since(startTime).Milliseconds() > allowedTime {
		flagStop = true
		return 0, true
	}

	// Pre-allocate PV line to avoid reallocations
	if cap(*line) < 100 {
		*line = make([]Move, 0, 100)
	} else {
		*line = (*line)[:0]
	}

	isPv := beta > alpha+1
	isRoot := depth == rd
	check := b.isCheck(c)
	if check {
		depth++
	}

	if depth <= 0 {
		return quiesce(b, 4, alpha, beta, c, rd-depth), false
	}

	// Check for two-fold repetition
	if !isRoot && b.isTwoFold() {
		// Don't store repetition positions in TT since their value depends on game history
		return 0, false
	}

	legalMoves := b.generateLegalMoves()
	if len(legalMoves) == 0 {
		if b.isCheck(b.turn) {
			return -winVal, false
		} else {
			return 0, false
		}
	}

	timeout := false
	bestMove := Move{}

	res, score := probeTT(b, alpha, beta, uint8(depth), &bestMove)
	if res && !isRoot {
		*line = append(*line, bestMove)
		return score, false
	}

	// Static Null Move Pruning
	// Only do SNMP if:
	// 1. Not in check
	// 2. Not in PV node
	// 3. Not searching for mate
	if !check && !isPv && beta < winVal-100 && beta > -winVal+100 {

		// Get static evaluation with some margin
		staticEval := evaluate(b) * factor[c]

		// Margin increases with depth
		margin := 85 * depth

		// If static eval is way above beta, likely we can prune
		if staticEval-margin >= beta {
			return staticEval - margin, false
		}
	}

	// Pre-allocate PV line for child nodes
	childPV := make([]Move, 0, 100)

	// Null Move Pruning
	hasNotJustPawns := b.colors[c] ^ b.getColorPieces(pawn, c)
	if doNull && !check && !isPv && hasNotJustPawns != 0 {
		b.makeNullMove()
		R := 3 + depth/6
		score, timeout = pvs(b, depth-1-R, rd, -beta, -beta+1, reverseColor(c), false, &childPV)
		score *= -1
		if timeout {
			b.undoNullMove()
			return 0, true
		}

		childPV = make([]Move, 0, 100)
		b.undoNullMove()

		if flagStop {
			return 0, true
		}

		if score >= beta {
			return score, false
		}
	}

	// Razoring
	// Only apply razoring at shallow depths and non-PV nodes
	if depth <= 2 && !check && !isPv {
		staticEval := evaluate(b) * factor[c]

		// Razoring margins based on depth
		razorMargin := 300 + depth*60

		if staticEval+razorMargin <= alpha {
			// Try qsearch to verify if position is really bad
			qScore := quiesce(b, 4, alpha, beta, c, 4)
			if qScore < alpha {
				return qScore, false
			}
		}
	}

	// Internal Iterative Deepening
	if depth >= 4 && isPv && bestMove.from == 0 {
		_, _ = pvs(b, depth-3, rd+1, -beta, -alpha, c, false, line)
		if len(*line) > 0 {
			bestMove = (*line)[0]
		}
		*line = (*line)[:0]
		if flagStop {
			return 0, true
		}
	}

	// Order moves
	if depth > 1 {
		orderMovesPV(b, &legalMoves, bestMove, c, depth)
	}

	bestScore := -winVal - 1
	bestMove = Move{}
	ttFlag := upper
	staticEval := evaluate(b) * factor[c]

	for mvCnt, move := range legalMoves {
		b.makeMove(move)

		score := 0
		if mvCnt == 0 {
			score, timeout = pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &childPV)
			score *= -1
		} else {
			// Late move reduction
			shouldReduce := !isPv && move.movetype != CAPTURE && move.movetype != PROMOTION
			R := 0

			if depth >= 3 && shouldReduce && !check {
				R = lmrTable[depth][mvCnt+1]
			}

			checkAfterMove := b.isCheck(b.turn)
			if shouldReduce && !checkAfterMove && depth-R-1 < len(futilityMargins) && futilityMargins[depth-R-1]+staticEval < alpha {
				b.undo()
				continue
			}

			score, timeout = pvs(b, depth-1-R, rd, -alpha-1, -alpha, reverseColor(c), true, &childPV)
			score *= -1

			if score > alpha && R > 0 {
				score, timeout = pvs(b, depth-1, rd, -alpha-1, -alpha, reverseColor(c), true, &childPV)
				score *= -1
				if score > alpha {
					score, timeout = pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &childPV)
					score *= -1
				}
			} else if score > alpha && score < beta {
				score, timeout = pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &childPV)
				score *= -1
			}
		}

		b.undo()

		if score > bestScore {
			bestScore = score
			bestMove = move
		}

		if score >= beta {
			ttFlag = lower

			// Store killer moves
			if move.movetype == QUIET && killerMoves[depth][0].toUCI() != move.toUCI() {
				killerMoves[depth][1] = killerMoves[depth][0]
				killerMoves[depth][0] = move
			}

			// Increment history table
			if move.movetype == QUIET {
				historyHeuristic[b.turn][move.from][move.to] += depth * depth
			}

			if historyHeuristic[b.turn][move.from][move.to] > secondKillerMoveBonus {
				historyHeuristic[b.turn][move.from][move.to] /= 2
			}

			break
		} else {
			// Decrement history table
			if move.movetype == QUIET {
				if historyHeuristic[b.turn][move.from][move.to] > 0 {
					historyHeuristic[b.turn][move.from][move.to]--
				}
			}
		}

		if score > alpha {
			alpha = score
			ttFlag = exact

			*line = (*line)[:0]
			*line = append(*line, move)
			*line = append(*line, childPV...)

			// Increment history table
			if move.movetype == QUIET {
				historyHeuristic[b.turn][move.from][move.to] += depth * depth
			}

			if historyHeuristic[b.turn][move.from][move.to] > secondKillerMoveBonus {
				historyHeuristic[b.turn][move.from][move.to] /= 2
			}
		} else {
			// Decrement history table
			if move.movetype == QUIET {
				if historyHeuristic[b.turn][move.from][move.to] > 0 {
					historyHeuristic[b.turn][move.from][move.to]--
				}
			}
		}
	}

	if !timeout && !flagStop {
		storeEntry(b, bestScore, ttFlag, bestMove, uint8(depth))
	}

	return bestScore, timeout
}

func initializeLMRTable() {
	for depth := 1; depth <= 100; depth++ {
		for moveCnt := 1; moveCnt <= 100; moveCnt++ {
			// Formula from Ethereal
			lmrTable[depth][moveCnt] = int(0.7844 + math.Log(float64(depth))*math.Log(float64(moveCnt))/2.4696)
		}
	}
}

// searchWithTime searches for the best move given a time constraint
func searchWithTime(b *Board, movetime int64) Move {
	if useOpeningBook {
		// Initialize opening book
		book := NewOpeningBook()

		// Check if we can use a book move
		if bookMove, variation := book.LookupPosition(b, b.turn); bookMove != "" {
			// Convert UCI string to Move
			move := fromUCI(bookMove, b)
			if variation != "" {
				fmt.Printf("info string Book move: %s (%s)\n", bookMove, variation)
			} else {
				fmt.Printf("info string Book move: %s\n", bookMove)
			}
			return move
		}
	}

	// Check tablebase if we're in an endgame position
	if useTablebase && isTablebasePosition(b) {
		if score, bestMove, found := probeTablebase(b.toFEN()); found {
			if bestMove != "" {
				// Check if this is a drawing position with multiple moves
				if strings.HasPrefix(bestMove, "draw:") {
					// Get the list of drawing moves
					drawingMoves := strings.Split(strings.TrimPrefix(bestMove, "draw:"), ",")

					// Search these moves to find the best one
					bestScore := -winVal - 1
					var bestDrawingMove Move
					line := []Move{}
					strLine := ""

					searchStart := time.Now()
					// Try each drawing move
					for _, moveStr := range drawingMoves {
						move := fromUCI(moveStr, b)
						b.makeMove(move)
						score, _ := pvs(b, 6, 6, -winVal-1, winVal+1, b.turn, true, &line)
						score *= -1 // Negate score since we evaluated from opponent's perspective
						b.undo()

						if score > bestScore {
							bestScore = score
							bestDrawingMove = move
							strLine = bestDrawingMove.toUCI()
							for _, m := range line {
								strLine += " " + m.toUCI()
							}
						}
					}

					fmt.Printf("info depth 6 nodes %d time %d score cp %d pv %s\n", nodesSearched, time.Since(searchStart).Milliseconds(), bestScore*factor[b.turn], strLine)
					return bestDrawingMove
				}

				// Not a drawing position, use the tablebase move directly
				move := fromUCI(bestMove, b)
				fmt.Printf("info depth 99 nodes 1 time 1 score cp %d pv %s\n", score, bestMove)
				return move
			}
		}
	}

	fmt.Printf("searching for movetime %d\n", movetime)
	startTime = time.Now()
	allowedTime = movetime
	line := []Move{}
	legalMoves := b.generateLegalMoves()
	prevScore := 0
	flagStop = false

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
		nodesSearched = 0
		searchStart := time.Now()

		// Aspiration windows
		alpha := -winVal - 1
		beta := winVal + 1
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
			score, timeout = pvs(b, i, i, alpha, beta, b.turn, true, &line)
			if timeout {
				return prevBest
			}

			if score <= alpha {
				alpha = max(-winVal-1, alpha-windowSize*2)
				continue
			}
			if score >= beta {
				beta = min(winVal+1, beta+windowSize*2)
				continue
			}
			break
		}

		prevScore = score
		timeTaken := time.Since(searchStart).Milliseconds()
		timeTakenNanoSeconds := time.Since(searchStart).Nanoseconds()

		incrementAge()

		strLine := ""
		for _, move := range line {
			strLine += " " + move.toUCI()
		}

		nps := nodesSearched * 1000000000 / max(int(timeTakenNanoSeconds), 1)
		fmt.Printf("info depth %d nodes %d time %d score cp %d nps %d pv%s\n", i, nodesSearched, timeTaken, score, nps, strLine)

		if score == winVal || score == -winVal {
			clearTTable()
			return line[0]
		}

		prevBest = line[0]

		// Check for stop signal after completing a depth
		select {
		case <-stopChannel:
			return prevBest
		default:
			// Continue to next depth
		}
	}
	return prevBest
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
