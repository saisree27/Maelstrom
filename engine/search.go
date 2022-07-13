package engine

// null move pruning
var R = 3

func quiesce(b *Board, limit int, alpha int, beta int, c Color) int {
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

	// TODO: optimize so we only calculate captures
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

func orderMovesPV(moves *[]Move, pv Move) {
	for i, mv := range *moves {
		if mv == pv {
			(*moves)[i], (*moves)[0] = (*moves)[0], pv
		}
	}
}

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, doNull bool, line *[]Move) int {
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
		orderMovesPV(&legalMoves, bestMove)
	}

	hasNotJustPawns := b.getColorPieces(queen, c) | b.getColorPieces(rook, c) | b.getColorPieces(bishop, c) | b.getColorPieces(knight, c)

	if doNull && b.plyCnt > 0 && hasNotJustPawns != 0 && depth >= R+1 && !b.isCheck(c) {
		b.makeNullMove()
		bestScore = -pvs(b, depth-R-1, rd, -beta, -beta+1, reverseColor(c), false, &pv)
		b.undoNullMove()

		if bestScore >= beta {
			return beta
		}
	}

	for i, move := range legalMoves {
		b.makeMove(move)
		if i == 0 {
			bestScore = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &pv)
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
			score := -pvs(b, depth-1, rd, -alpha-1, -alpha, reverseColor(c), true, &pv)
			if score > alpha && score < beta {
				score = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), true, &pv)
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
