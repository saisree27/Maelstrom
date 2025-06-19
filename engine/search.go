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

var COLOR_SIGN = [2]int{
	WHITE: 1, BLACK: -1,
}

const PV_MOVE_BONUS = 2000000
const MVV_LVA_BONUS = 1000000
const PROMOTION_BONUS = 800000
const FIRST_KILLER_MOVE_BONUS = 600000
const SECOND_KILLER_MOVE_BONUS = 590000

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

var KillerMovesTable = [100][2]Move{}
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
	case <-stopChannel:
		return 0
	default:
		// Continue search
	}

	if SearchStop {
		return 0
	}

	eval := Evaluate(b) * COLOR_SIGN[c]

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

	allLegalMoves := b.GenerateLegalMoves()

	for _, move := range allLegalMoves {
		if inCheck || move.movetype == CAPTURE || move.movetype == CAPTURE_AND_PROMOTION || move.movetype == EN_PASSANT {
			b.MakeMove(move)
			score := -QuiescenceSearch(b, limit-1, -beta, -alpha, ReverseColor(c), rd+1)
			b.Undo()

			// Check for stop signal after recursive call
			select {
			case <-stopChannel:
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

func orderMovesPV(b *Board, moves *[]Move, pv Move, c Color, depth int) {
	// Pre-allocate bonuses slice to avoid reallocations
	bonuses := make([]moveBonus, len(*moves))

	// First pass: handle PV move and calculate bonuses
	for i, mv := range *moves {
		if mv == pv {
			(*moves)[i], (*moves)[0] = (*moves)[0], pv
			bonuses[i] = moveBonus{move: mv, bonus: PV_MOVE_BONUS} // Highest priority for PV moves
			continue
		}

		var bonus int
		if mv.movetype == CAPTURE || mv.movetype == CAPTURE_AND_PROMOTION {
			score := MVV_LVA_TABLE[PieceToPieceType(mv.captured)][PieceToPieceType(mv.piece)]
			bonus = MVV_LVA_BONUS + score
		} else if mv.movetype == PROMOTION {
			bonus = PROMOTION_BONUS // Promotions are valuable
		} else {
			// Quiet moves
			if mv == KillerMovesTable[depth][0] {
				bonus = FIRST_KILLER_MOVE_BONUS // First killer move
			} else if mv == KillerMovesTable[depth][1] {
				bonus = SECOND_KILLER_MOVE_BONUS // Second killer move
			} else {
				// History heuristic for remaining quiet moves
				bonus = HistoryTable[b.turn][mv.from][mv.to]
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

func Pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move) (int, bool) {
	NodesSearched++

	// Check for stop signal
	select {
	case <-stopChannel:
		SearchStop = true
		return 0, true
	default:
		// Continue search
	}

	if SearchStop {
		return 0, true
	}

	if NodesSearched%2047 == 0 && time.Since(SearchStartTime).Milliseconds() > AllowedTime {
		SearchStop = true
		return 0, true
	}

	isPv := beta > alpha+1
	isRoot := depth == rd
	check := b.IsCheck(c)
	if check {
		depth++
	}

	if depth <= 0 {
		return QuiescenceSearch(b, 4, alpha, beta, c, rd-depth), false
	}

	// Check for two-fold repetition
	if !isRoot && b.IsTwoFold() {
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
		*line = append(*line, bestMove)
		return score, false
	}

	// Static Null Move Pruning
	// Only do SNMP if:
	// 1. Not in check
	// 2. Not in PV node
	// 3. Not searching for mate
	if !check && !isPv && beta < WIN_VAL-100 && beta > -WIN_VAL+100 {

		// Get static evaluation with some margin
		staticEval := Evaluate(b) * COLOR_SIGN[c]

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
	hasNotJustPawns := b.colors[c] ^ b.GetColorPieces(PAWN, c)
	if doNull && !check && !isPv && hasNotJustPawns != 0 {
		b.MakeNullMove()
		R := 3 + depth/6
		score, timeout = Pvs(b, depth-1-R, rd, -beta, -beta+1, ReverseColor(c), false, &childPV)
		score *= -1
		if timeout {
			b.UndoNullMove()
			return 0, true
		}

		childPV = make([]Move, 0, 100)
		b.UndoNullMove()

		if SearchStop {
			return 0, true
		}

		if score >= beta {
			return score, false
		}
	}

	// Razoring
	// Only apply razoring at shallow depths and non-PV nodes
	if depth <= 2 && !check && !isPv {
		staticEval := Evaluate(b) * COLOR_SIGN[c]

		// Razoring margins based on depth
		razorMargin := 300 + depth*60

		if staticEval+razorMargin <= alpha {
			// Try qsearch to verify if position is really bad
			qScore := QuiescenceSearch(b, 4, alpha, beta, c, 4)
			if qScore < alpha {
				return qScore, false
			}
		}
	}

	// Internal Iterative Deepening
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

	// Order moves
	if depth > 1 {
		orderMovesPV(b, &legalMoves, bestMove, c, depth)
	}

	bestScore := -WIN_VAL - 1
	bestMove = Move{}
	ttFlag := UPPER
	staticEval := Evaluate(b) * COLOR_SIGN[c]

	for mvCnt, move := range legalMoves {
		b.MakeMove(move)

		score := 0
		if mvCnt == 0 {
			score, timeout = Pvs(b, depth-1, rd, -beta, -alpha, ReverseColor(c), true, &childPV)
			score *= -1
		} else {
			// Late move reduction
			quietMove := !isPv && move.movetype != CAPTURE && move.movetype != PROMOTION
			R := 0

			if depth >= 3 && quietMove && !check {
				R = LMR_TABLE[depth][mvCnt+1]
			}

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
	if useOpeningBook {
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
	if useTablebase && IsTablebasePosition(b) {
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
	line := []Move{}
	legalMoves := b.GenerateLegalMoves()
	prevScore := 0
	SearchStop = false

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
			if timeout {
				return prevBest
			}

			if score <= alpha {
				alpha = max(-WIN_VAL-1, alpha-windowSize*2)
				continue
			}
			if score >= beta {
				beta = min(WIN_VAL+1, beta+windowSize*2)
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

		nps := NodesSearched * 1000000000 / max(int(timeTakenNanoSeconds), 1)
		fmt.Printf("info depth %d nodes %d time %d score cp %d nps %d pv%s\n", i, NodesSearched, timeTaken, score, nps, strLine)

		if score == WIN_VAL || score == -WIN_VAL {
			ClearTT()
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
