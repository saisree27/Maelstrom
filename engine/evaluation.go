package engine

var winVal int = 1000000
var factor = map[Color]int{
	WHITE: 1, BLACK: -1,
}

var material = map[Piece]int{
	wP: 100, bP: -100,
	wN: 300, bN: -300,
	wB: 300, bB: -300,
	wR: 500, bR: -500,
	wQ: 900, bQ: -900,
	wK: 0, bK: 0,
	EMPTY: 0,
}

// Returns an evaluation of the position in cp
// 1000000 or -1000000 is designated as checkmate
// Evaluations always in the perspective of the player to move
func evaluate(b *Board) int {
	moves := b.generateLegalMoves()
	if len(moves) == 0 {
		if b.isCheck(b.turn) {
			return winVal * factor[reverseColor(b.turn)]
		} else {
			return 0
		}
	}

	eval := 0
	eval += (totalMaterial(b) * factor[b.turn])
	return eval
}

func totalMaterial(b *Board) int {
	sum := 0
	for _, piece := range b.squares {
		sum += material[piece]
	}
	return sum
}
