package engine

import (
	"fmt"
	"sort"
	"time"
)

type moveBonus struct {
	move  Move
	bonus int
}

// null move pruning constant R
const R = 3

var nodesSearched = 0
var killerMoves = [100][2]Move{}
var historyHeuristic = [100][64][64]int{}

func quiesce(b *Board, limit int, alpha int, beta int, c Color) int {
	nodesSearched++
	eval := evaluate(b) * factor[c]

	// Beta cutoff
	if eval >= beta {
		return beta
	}

	// Delta pruning before move generation
	delta := queenVal + 200 // Add some margin for positional compensation
	if eval < alpha-delta {
		return alpha
	}

	if alpha < eval {
		alpha = eval
	}

	if limit == 0 {
		return eval
	}

	legalMoves := b.generateCaptures()

	for _, move := range legalMoves {
		if move.movetype == CAPTURE || move.movetype == CAPTUREANDPROMOTION || move.movetype == ENPASSANT {
			// SEE pruning - skip clearly bad captures
			if !seeGreaterThan(b, move, -50) {
				continue
			}

			b.makeMove(move)
			score := -quiesce(b, limit-1, -beta, -alpha, reverseColor(c))
			b.undo()

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

func orderMovesPV(b *Board, moves *[]Move, pv Move, c Color, depth int, rd int) {
	// Pre-allocate bonuses slice to avoid reallocations
	bonuses := make([]moveBonus, len(*moves))

	// First pass: handle PV move and calculate bonuses
	for i, mv := range *moves {
		if mv == pv {
			(*moves)[i], (*moves)[0] = (*moves)[0], pv
			bonuses[i] = moveBonus{move: mv, bonus: 2000000} // Highest priority for PV moves
			continue
		}

		var bonus int
		if mv.movetype == CAPTURE || mv.movetype == CAPTUREANDPROMOTION {
			// Use SEE score for captures, scaled up to match bonus range
			seeScore := seeMove(b, mv)
			if seeScore > 0 {
				bonus = 1000000 + seeScore // Good captures
			} else {
				bonus = 500000 + seeScore // Bad captures still above quiet moves
			}
		} else if mv.movetype == PROMOTION {
			bonus = 800000 + material[mv.promote] // Promotions are valuable
		} else {
			// Quiet moves
			if mv == killerMoves[depth][0] {
				bonus = 600000 // First killer move
			} else if mv == killerMoves[depth][1] {
				bonus = 590000 // Second killer move
			} else {
				// History heuristic for remaining quiet moves
				bonus = historyHeuristic[depth][mv.from][mv.to]
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

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move, tR int64, st time.Time) (int, bool) {
	nodesSearched++

	// Pre-allocate PV line to avoid reallocations
	if cap(*line) < 100 {
		*line = make([]Move, 0, 100)
	} else {
		*line = (*line)[:0]
	}

	if depth <= 0 {
		return quiesce(b, 4, alpha, beta, c), false
	}

	legalMoves := b.generateLegalMoves()
	if len(legalMoves) == 0 {
		return evaluate(b) * factor[c], false
	}

	if b.isThreeFoldRep() {
		return 0, false
	}

	bestScore := 0
	timeout := false
	bestMove := Move{}
	oldAlpha := alpha
	isPv := beta > alpha+1

	res, found := probeTT(b, &bestScore, &alpha, &beta, depth, rd, &bestMove)
	if res {
		*line = append(*line, bestMove)
		return found, false
	}

	// Internal Iterative Deepening
	if depth >= 4 && isPv && bestMove.from == 0 {
		_, _ = pvs(b, depth-2, rd+1, alpha, beta, c, false, line, tR, st)
		if len(*line) > 0 {
			bestMove = (*line)[0]
		}
		*line = (*line)[:0]
	}

	if depth > 1 {
		orderMovesPV(b, &legalMoves, bestMove, c, depth, rd)
	}

	check := b.isCheck(c)
	if check {
		depth++
	}

	// Razoring
	// Only apply razoring at shallow depths and non-PV nodes
	if depth <= 3 && !check && rd > 1 &&
		beta < winVal-100 && beta > -winVal+100 &&
		beta == alpha+1 { // Non-PV node check

		staticEval := evaluate(b) * factor[c]

		// Razoring margins based on depth
		razorMargin := 300 + depth*75

		if staticEval+razorMargin <= alpha {
			// Try qsearch to verify if position is really bad
			qScore := quiesce(b, 4, alpha-razorMargin, alpha-razorMargin+1, c)
			if qScore <= alpha-razorMargin {
				return qScore, false
			}
		}
	}

	// Static Null Move Pruning
	// Only do SNMP if:
	// 1. Not in check
	// 2. Not at root
	// 3. Not searching for mate
	// 4. Not too deep in the tree
	if !check && depth >= 1 && depth <= 6 && rd > 1 &&
		beta < winVal-100 && beta > -winVal+100 {

		// Get static evaluation with some margin
		staticEval := evaluate(b) * factor[c]

		// Margin increases with depth
		margin := 120 + 50*depth

		// If static eval is way above beta, likely we can prune
		if staticEval-margin >= beta {
			return staticEval - margin, false
		}
	}

	// Reverse Futility Pruning
	// Similar to SNMP but with different margins and depth conditions
	if !check && depth <= 5 && rd > 1 &&
		beta < winVal-100 && beta > -winVal+100 {

		staticEval := evaluate(b) * factor[c]

		// More aggressive pruning for very shallow depths
		// Values tuned based on piece values
		var rfpMargin int
		if depth <= 3 {
			rfpMargin = pawnVal * depth
		} else {
			rfpMargin = pawnVal*3 + (depth-3)*175
		}

		// Check material balance to ensure we're not in a tactical position
		material, pieces := totalMaterialAndPieces(b)
		if pieces >= 10 && material*factor[c] > -pawnVal { // Don't prune if down material
			if staticEval >= beta+rfpMargin {
				return beta, false
			}
		}
	}

	hasNotJustPawns := b.colors[c] ^ b.getColorPieces(pawn, c)

	if doNull && !check && hasNotJustPawns != 0 && b.plyCnt > 0 {
		b.makeNullMove()
		bestScore, timeout = pvs(b, depth-R-1, rd, -beta, -beta+1, reverseColor(c), false, line, tR, st)
		bestScore *= -1
		b.undoNullMove()

		if bestScore >= beta {
			return beta, timeout
		}
	}

	// Pre-allocate PV line for child nodes
	childPV := make([]Move, 0, 100)

	for i, move := range legalMoves {
		if timeout {
			break
		}

		if time.Since(st).Milliseconds() > tR {
			return bestScore, true
		}

		b.makeMove(move)

		if i == 0 {
			bestScore, timeout = pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &childPV, tR, st)
			bestScore *= -1
			b.undo()

			if bestScore > alpha && !timeout {
				bestMove = move
				*line = (*line)[:0]
				*line = append(*line, move)
				*line = append(*line, childPV...)

				if bestScore >= beta {
					if killerMoves[depth][0].toUCI() != move.toUCI() {
						killerMoves[depth][1] = killerMoves[depth][0]
						killerMoves[depth][0] = move
					}
					historyHeuristic[depth][move.from][move.to]++
					break
				}
				alpha = bestScore
			}
		} else {
			score := alpha + 1
			check := b.isCheck(b.turn)

			// FUTILITY PRUNING
			if depth == 1 && !check &&
				move.movetype != CAPTURE &&
				move.movetype != CAPTUREANDPROMOTION &&
				move.movetype != PROMOTION {

				staticEval := evaluate(b) * factor[c]
				b.undo()

				if staticEval+30*depth <= alpha {
					continue // prune this move
				}

				b.makeMove(move) // re-do the move for actual search
			}

			// Late move reduction
			if i >= 4 && depth >= 3 && move.movetype != CAPTURE && move.movetype != CAPTUREANDPROMOTION && !check {
				// Calculate reduction depth based on move history and position
				reduction := 1

				// Increase reduction for moves with low history score
				historyScore := historyHeuristic[depth][move.from][move.to]
				if historyScore < 100 {
					reduction++
				}

				// Increase reduction for moves that are not in check and not threatening
				if !b.isCheck(reverseColor(c)) {
					reduction++
				}

				// Cap reduction at depth-2
				if reduction > depth-2 {
					reduction = depth - 2
				}

				// Only reduce if we have enough depth
				if depth > reduction {
					score, timeout = pvs(b, depth-reduction, rd, -alpha-1, -alpha, reverseColor(c), true, &childPV, tR, st)
					score *= -1
				}
			}

			if score > alpha {
				score, timeout = pvs(b, depth-1, rd, -alpha-1, -alpha, reverseColor(c), true, &childPV, tR, st)
				score *= -1

				if timeout {
					break
				}

				if score > alpha && score < beta {
					score, timeout = pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &childPV, tR, st)
					score *= -1
				}

				if score > alpha && !timeout {
					bestMove = move
					alpha = score
					*line = (*line)[:0]
					*line = append(*line, move)
					*line = append(*line, childPV...)
				}

				b.undo()
				if score > bestScore && !timeout {
					bestScore = score
					if score >= beta {
						if killerMoves[depth][0].toUCI() != move.toUCI() {
							killerMoves[depth][1] = killerMoves[depth][0]
							killerMoves[depth][0] = move
						}
						historyHeuristic[depth][move.from][move.to]++
						break
					}
				}
			} else {
				b.undo()
			}
		}
	}

	if !timeout {
		var flag bound
		if bestScore <= oldAlpha {
			flag = upper
		} else if bestScore >= beta {
			flag = lower
		} else {
			flag = exact
		}

		storeEntry(b, bestScore, flag, bestMove, depth)
	}

	return bestScore, timeout
}

func searchWithTime(b *Board, movetime int64) Move {
	b.printFromBitBoards()
	startTime := time.Now()
	line := []Move{}
	legalMoves := b.generateLegalMoves()
	prevBest := Move{}
	prevScore := 0

	if len(legalMoves) == 1 {
		return legalMoves[0]
	}

	for i := 1; i <= 100; i++ {
		nodesSearched = 0
		duration := time.Since(startTime).Milliseconds()
		timeRemaining := movetime - duration

		if movetime > timeRemaining*2 {
			break
		}
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
			score, timeout = pvs(b, i, i, alpha, beta, b.turn, true, &line, timeRemaining, time.Now())
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

		incrementAge()

		b.makeMove(line[0])

		if b.isTwoFold() && score > 0 {
			table.entries[b.zobrist%table.count] = TTEntry{}
			b.undo()
			fmt.Println("Two-fold repetition encountered, removing TT entry")
			table.entries[b.zobrist%table.count] = TTEntry{}
		} else {
			b.undo()
		}

		strLine := ""
		for _, move := range line {
			strLine += " " + move.toUCI()
		}

		signed := score * factor[b.turn]

		fmt.Printf("info depth %d nodes %d time %d score cp %d pv%s\n", i, nodesSearched, timeTaken, signed, strLine)

		if score == winVal || score == -winVal {
			clearTTable()
			return line[0]
		}

		prevBest = line[0]
	}
	return prevBest
}

// Helper function for aspiration windows
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
