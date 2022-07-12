package engine

import (
	"fmt"
	"math"
)

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, line *[]string) int {
	pv := []string{""}

	if depth == 0 {
		return factor[c] * evaluate(b)
	}

	legalMoves := b.generateLegalMoves()

	if len(legalMoves) == 0 {
		if b.isCheck(b.turn) {
			return winVal * factor[reverseColor(b.turn)]
		} else {
			return 0
		}
	}

	var score int = 0
	var bestMove Move

	for i, move := range legalMoves {
		if depth == rd {
			fmt.Println(*line)
		}
		if i == 0 {
			b.makeMove(move)
			score = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), &pv)
			b.undo()
		} else {
			b.makeMove(move)
			score = -pvs(b, depth-1, rd, alpha-1, -alpha, reverseColor(c), &pv)
			if alpha < score && score < beta {
				score = -pvs(b, depth-1, rd, -beta, -score, reverseColor(c), &pv)
			}
			b.undo()
		}
		if score > alpha {
			alpha = score
			bestMove = move
			(*line)[0] = bestMove.toUCI()
			if len(pv) != 1 || pv[0] != "" {

				*line = append(*line, pv...)
			}

		}
		if alpha > beta {
			break
		}
	}

	return alpha
}

func Search(b *Board) {
	for i := 1; i < 100; i++ {
		line := []string{""}
		fmt.Printf("Depth %d: ", i)
		fmt.Print(pvs(b, i, i, math.MinInt, math.MaxInt, b.turn, &line))

		for i, j := 0, len(line)-1; i < j; i, j = i+1, j-1 {
			line[i], line[j] = line[j], line[i]
		}
		fmt.Print(" ")
		fmt.Print(line)
		fmt.Println()
	}
}
