package engine

import (
	"fmt"
)

func orderMovesPV(moves *[]Move, prev *[]Move) {
	for i, mv := range *moves {
		if mv.toUCI() == (*prev)[0].toUCI() {
			(*moves)[i], (*moves)[0] = (*moves)[0], (*prev)[0]
		}
	}
}

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, line *[]Move, prev *[]Move) int {
	pv := []Move{}

	if depth <= 0 {
		return evaluate(b) * factor[c]
	}

	legalMoves := b.generateLegalMoves()

	if depth == rd && depth > 1 {
		orderMovesPV(&legalMoves, prev)
	}

	if len(legalMoves) == 0 {
		return evaluate(b) * factor[c]
	}

	if b.isThreeFoldRep() {
		return 0
	}

	bestScore := 0

	for i, move := range legalMoves {
		b.makeMove(move)
		if i == 0 {
			bestScore = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), &pv, prev)
			b.undo()
			if bestScore > alpha {

				*line = []Move{}
				*line = append(*line, move)
				*line = append(*line, pv...)

				if bestScore >= beta {
					break
				}
				alpha = bestScore
			}
		} else {
			score := -pvs(b, depth-1, rd, -alpha-1, -alpha, reverseColor(c), &pv, prev)
			if score > alpha && score < beta {
				score = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), &pv, prev)
			}
			if score > alpha {
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
				bestScore = score
			}
		}
	}

	return bestScore
}

func RunSearch(position string, depth int) {
	b := Board{}

	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}

	b.printFromBitBoards()

	prev := []Move{}
	for i := 1; i <= depth; i++ {
		line := []Move{}

		fmt.Printf("Depth %d: ", i)

		score := pvs(&b, i, i, -winVal-1, winVal+1, b.turn, &line, &prev) * factor[b.turn]

		if score == winVal || score == -winVal {
			fmt.Println("Found mate.")
			break
		}

		fmt.Print(score)
		fmt.Print(" ")

		strLine := []string{}
		for i, _ := range line {
			strLine = append(strLine, line[i].toUCI())
		}

		fmt.Print(strLine)
		fmt.Println()

		prev = make([]Move, len(line))
		copy(prev, line)
	}
}

func RunSelfPlay(position string, depth int) {
	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}

	movesPlayed := []string{}

	legals := b.generateLegalMoves()
	for {
		if len(legals) == 0 {
			break
		}

		b.printFromBitBoards()

		line := []Move{}
		prev := []Move{}

		score := 0

		newDepth := depth
		// var inputDepth string
		// fmt.Scanln(&inputDepth)

		// newDepth, _ := strconv.Atoi(inputDepth)

		for i := 1; i <= newDepth; i++ {
			line = []Move{}

			fmt.Printf("Depth %d: ", i)

			score = pvs(&b, i, i, -winVal-1, winVal+1, b.turn, &line, &prev)

			if score == winVal || score == -winVal {
				fmt.Println("Found mate.")
				break
			}

			fmt.Print(score * factor[b.turn])
			fmt.Print(" ")

			strLine := []string{}
			for i, _ := range line {
				strLine = append(strLine, line[i].toUCI())
			}

			fmt.Print(strLine)
			fmt.Println()

			prev = make([]Move, len(line))
			copy(prev, line)
		}
		bestMove := line[0]

		fmt.Printf("SCORE: %d\n", score*factor[b.turn])

		movesPlayed = append(movesPlayed, bestMove.toUCI())
		b.makeMove(bestMove)

		fmt.Println(movesPlayed)
	}
}
