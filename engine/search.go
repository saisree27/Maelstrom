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

func quiesce(b *Board, limit int, alpha int, beta int, c Color) int {
	nodesSearched++
	eval := evaluate(b) * factor[c]
	if eval >= beta {
		return beta
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
				if depth >= rd-2 {
					b.makeMove(mv)
					if b.isCheck(b.turn) {
						bonus = 50
					} else {
						switch mv.piece {
						case wN:
							bonus = knightSquareTable[mv.to] - knightSquareTable[mv.from]
						case bN:
							bonus = knightSquareTable[reversePSQ[mv.to]] - knightSquareTable[reversePSQ[mv.from]]
						case wB:
							bonus = bishopSquareTable[mv.to] - bishopSquareTable[mv.from]
						case bB:
							bonus = bishopSquareTable[reversePSQ[mv.to]] - bishopSquareTable[reversePSQ[mv.from]]
						case wR:
							bonus = rookSquareTable[mv.to] - rookSquareTable[mv.from]
						case bR:
							bonus = rookSquareTable[reversePSQ[mv.to]] - rookSquareTable[reversePSQ[mv.from]]
						case wQ:
							bonus = queenSquareTable[mv.to] - queenSquareTable[mv.from]
						case bQ:
							bonus = queenSquareTable[reversePSQ[mv.to]] - queenSquareTable[reversePSQ[mv.from]]
						case wP:
							bonus = pawnSquareTable[mv.to] - pawnSquareTable[mv.from]
						case bP:
							bonus += pawnSquareTable[reversePSQ[mv.to]] - pawnSquareTable[reversePSQ[mv.from]]
						}
					}
					b.undo()
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

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move, tR int64, st time.Time) int {
	nodesSearched++
	pv := []Move{}

	if depth <= 0 {
		return quiesce(b, 4, alpha, beta, c)
	}

	legalMoves := b.generateLegalMoves()

	if len(legalMoves) == 0 {
		return evaluate(b) * factor[c]
	}

	if b.isThreeFoldRep() {
		return 0
	}

	bestScore := 0
	bestMove := Move{}
	oldAlpha := alpha

	res, found := probeTT(b, &bestScore, &alpha, &beta, depth, &bestMove)
	if res {
		*line = []Move{}
		*line = append(*line, bestMove)
		*line = append(*line, pv...)
		return found
	}

	if depth > 1 {
		orderMovesPV(b, &legalMoves, bestMove, c, depth, rd)
	}

	if depth == 1 {
		// futility pruning
		eval := evaluate(b) * factor[c]
		if eval + knightVal < alpha {
			return quiesce(b, 4, alpha, beta, c)
		}
	}

	hasNotJustPawns := b.colors[c] ^ b.getColorPieces(pawn, c)

	if doNull && hasNotJustPawns != 0 && depth >= R+1 && !b.isCheck(c) && b.plyCnt > 0 {
		b.makeNullMove()
		bestScore = -pvs(b, depth-R-1, rd, -beta, -beta+1, reverseColor(c), false, &pv, tR, st)
		b.undoNullMove()

		if bestScore >= beta {
			return beta
		}
	}

	for i, move := range legalMoves {
		duration := time.Since(st).Milliseconds()
		if duration > tR {
			*line = []Move{}
			return 0
		}
		b.makeMove(move)

		if i == 0 {
			bestScore = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &pv, tR, st)
			b.undo()
			if bestScore > alpha {
				bestMove = move
				*line = []Move{}
				*line = append(*line, move)
				*line = append(*line, pv...)

				if bestScore >= beta {
					break
				}
				alpha = bestScore
			}
		} else {
			score := alpha + 1
			check := b.isCheck(b.turn)
			// Late move reduction
			if i >= 4 && depth >= 3 && move.movetype != CAPTURE && move.movetype != CAPTUREANDPROMOTION && !check {
				score = -pvs(b, depth-2, rd, -alpha-1, -alpha, reverseColor(c), true, &pv, tR, st)
			}
			if score > alpha {
				score := -pvs(b, depth-1, rd, -alpha-1, -alpha, reverseColor(c), true, &pv, tR, st)
				if score > alpha && score < beta {
					score = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &pv, tR, st)
				}
				if score > alpha {
					bestMove = move
					alpha = score
					*line = []Move{}

					*line = append(*line, move)
					*line = append(*line, pv...)

				}
				b.undo()
				if score > bestScore {
					bestScore = score
					if score >= beta {
						break
					}
				}
			} else {
				b.undo()
			}
		}
	}

	var flag bound
	if bestScore <= oldAlpha {
		flag = upper
	} else if bestScore >= beta {
		flag = lower
	} else {
		flag = exact
	}

	storeEntry(b, bestScore, flag, bestMove, depth)

	return bestScore
}

func searchWithTime(b *Board, movetime int64) Move {
	startTime := time.Now()
	line := []Move{}
	legalMoves := b.generateLegalMoves()
	prevBest := legalMoves[0]

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
		score := pvs(b, i, i, -winVal-1, winVal+1, b.turn, true, &line, timeRemaining, time.Now()) * factor[b.turn]
		timeTaken := time.Since(searchStart).Milliseconds()

		if len(line) == 0 {
			break
		}

		if score == winVal || score == -winVal {
			return line[0]
		}

		strLine := ""
		for _, move := range line {
			strLine += " " + move.toUCI()
		}

		fmt.Printf("info depth %d nodes %d time %d score cp %d pv%s\n", i, nodesSearched, timeTaken, score, strLine)

		prevBest = line[0]
	}
	return prevBest
}
