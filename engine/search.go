package engine

import (
	"fmt"
	"math"
)

func pvs(b *Board, depth int, rd int, alpha int, beta int, c Color, line *[]string) int {
	pv := []string{}

	if depth <= 0 {
		return evaluate(b) * factor[c]
	}

	legalMoves := b.generateLegalMoves()

	if len(legalMoves) == 0 {
		return evaluate(b) * factor[c]
	}

	bestScore := 0

	for i, move := range legalMoves {
		b.makeMove(move)
		if i == 0 {
			bestScore = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), &pv)
			b.undo()
			if bestScore > alpha {

				*line = []string{}

				*line = append(*line, move.toUCI())
				*line = append(*line, pv...)
				if bestScore >= beta {
					break
				}
				alpha = bestScore
			}
		} else {
			score := -pvs(b, depth-1, rd, -alpha-1, -alpha, reverseColor(c), &pv)
			if score > alpha && score < beta {
				score = -pvs(b, depth-1, rd, -beta, -alpha, reverseColor(c), &pv)
			}
			if score > alpha {
				alpha = score
				*line = []string{}

				*line = append(*line, move.toUCI())
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

	// if depth == 5 {
	// 	fmt.Println(*line)
	// }
	return bestScore
}

func Search(position string) {
	for i := 1; i < 100; i++ {
		b := Board{}
		if position == "startpos" {
			b.InitStartPos()
		} else {
			b.InitFEN(position)
		}
		line := []string{""}
		fmt.Printf("Depth %d: ", i)
		fmt.Print(pvs(&b, i, i, math.MinInt, math.MaxInt, b.turn, &line))

		fmt.Print(" ")
		fmt.Print(line)
		fmt.Println()
	}
}
