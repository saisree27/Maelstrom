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
	if eval >= beta {
		return beta
	}

	delta := queenVal
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
	bonuses := []moveBonus{}
	for i, mv := range *moves {
		if mv == pv {
			(*moves)[i], (*moves)[0] = (*moves)[0], pv
			bonuses = append(bonuses, moveBonus{move: mv, bonus: 30000})
		} else {
			if mv.movetype == CAPTUREANDPROMOTION {
				bonus := material[mv.promote] - material[mv.captured] - material[mv.piece]
				bonuses = append(bonuses, moveBonus{move: mv, bonus: bonus * factor[c]})
			} else if mv.movetype == CAPTURE {
				bonus := -material[mv.captured] - material[mv.piece]
				bonuses = append(bonuses, moveBonus{move: mv, bonus: bonus * factor[c]})
			} else if mv.movetype == PROMOTION {
				bonus := material[mv.promote] - material[mv.piece]
				bonuses = append(bonuses, moveBonus{move: mv, bonus: bonus * factor[c]})
			} else {
				bonus := 0
				if mv.toUCI() == killerMoves[depth][0].toUCI() {
					bonus = 150
				} else if mv.toUCI() == killerMoves[depth][1].toUCI() {
					bonus = 140
				}
				if depth >= rd {
					switch mv.piece {
					case wN:
						bonus += knightSquareTable[mv.to] - knightSquareTable[mv.from]
					case bN:
						bonus += knightSquareTable[reversePSQ[mv.to]] - knightSquareTable[reversePSQ[mv.from]]
					case wB:
						bonus += bishopSquareTable[mv.to] - bishopSquareTable[mv.from]
					case bB:
						bonus += bishopSquareTable[reversePSQ[mv.to]] - bishopSquareTable[reversePSQ[mv.from]]
					case wR:
						bonus += rookSquareTable[mv.to] - rookSquareTable[mv.from]
					case bR:
						bonus += rookSquareTable[reversePSQ[mv.to]] - rookSquareTable[reversePSQ[mv.from]]
					case wQ:
						bonus += queenSquareTable[mv.to] - queenSquareTable[mv.from]
					case bQ:
						bonus += queenSquareTable[reversePSQ[mv.to]] - queenSquareTable[reversePSQ[mv.from]]
					case wP:
						bonus += pawnSquareTable[mv.to] - pawnSquareTable[mv.from]
					case bP:
						bonus += pawnSquareTable[reversePSQ[mv.to]] - pawnSquareTable[reversePSQ[mv.from]]
					}
				}
				bonuses = append(bonuses, moveBonus{move: mv, bonus: bonus})
			}
		}
	}

	sort.Slice(bonuses, func(i, j int) bool {
		return bonuses[i].bonus > bonuses[j].bonus
	})

	for i, b := range bonuses {
		(*moves)[i] = b.move
	}
}

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move, tR int64, st time.Time) (int, bool) {
	nodesSearched++
	pv := []Move{}

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

	res, found := probeTT(b, &bestScore, &alpha, &beta, depth, rd, &bestMove)
	if res {
		*line = []Move{}
		*line = append(*line, bestMove)
		*line = append(*line, pv...)
		return found, false
	}

	if depth > 1 {
		orderMovesPV(b, &legalMoves, bestMove, c, depth, rd)
	}

	check := b.isCheck(c)

	if check {
		depth++
	}

	hasNotJustPawns := b.colors[c] ^ b.getColorPieces(pawn, c)

	if doNull && !check && hasNotJustPawns != 0 && b.plyCnt > 0 {
		b.makeNullMove()
		bestScore, timeout = pvs(b, depth-R-1, rd, -beta, -beta+1, reverseColor(c), false, &pv, tR, st)
		bestScore *= -1
		b.undoNullMove()

		if bestScore >= beta {
			return beta, timeout
		}
	}

	for i, move := range legalMoves {
		if timeout {
			break
		}

		duration := time.Since(st).Milliseconds()
		if duration > tR {
			*line = []Move{}
			return bestScore, true
		}
		b.makeMove(move)

		if i == 0 {
			bestScore, timeout = pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &pv, tR, st)
			bestScore *= -1
			b.undo()
			if bestScore > alpha && !timeout {
				bestMove = move
				*line = []Move{}
				*line = append(*line, move)
				*line = append(*line, pv...)

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
			// Late move reduction
			if i >= 4 && depth >= 3 && move.movetype != CAPTURE && move.movetype != CAPTUREANDPROMOTION && !check {
				score, timeout = pvs(b, depth-2, rd, -alpha-1, -alpha, reverseColor(c), true, &pv, tR, st)
				score *= -1
			}
			if score > alpha {
				score, timeout = pvs(b, depth-1, rd, -alpha-1, -alpha, reverseColor(c), true, &pv, tR, st)
				score *= -1

				if timeout {
					break
				}

				if score > alpha && score < beta {
					score, timeout = pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &pv, tR, st)
					score *= -1
				}
				if score > alpha && !timeout {
					bestMove = move
					alpha = score
					*line = []Move{}

					*line = append(*line, move)
					*line = append(*line, pv...)

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

	if len(legalMoves) == 1 {
		return legalMoves[0]
	}

	for i := 1; i <= 100; i++ {
		nodesSearched = 0
		duration := time.Since(startTime).Milliseconds()
		timeRemaining := movetime - duration

		if movetime > timeRemaining*2 {
			// We will probably take twice as much time in the new iteration
			// than the previous one, so probably wise not to go into the new iteration
			break
		}
		searchStart := time.Now()
		score, timeout := pvs(b, i, i, -winVal-1, winVal+1, b.turn, true, &line, timeRemaining, time.Now())
		timeTaken := time.Since(searchStart).Milliseconds()

		if timeout {
			return prevBest
		}

		b.makeMove(line[0])

		if b.isTwoFold() && score > 0 {
			table.entries[b.zobrist%table.count] = TTEntry{}
			b.undo()
			// TT three-fold issue
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
