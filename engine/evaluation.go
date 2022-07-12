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

var whitePawnSquareTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, -30, -30, -30, 10, 10, 5,
	-1, -5, -15, -30, 36, -10, -5, -1,
	-5, -15, 40, 40, 40, -10, 0, -5,
	5, 5, 10, 25, 25, 10, 5, 5,
	10, 10, 20, 30, 30, 20, 10, 10,
	50, 50, 50, 50, 50, 50, 50, 50,
	65, 65, 65, 65, 65, 65, 65, 65,
}

var blackPawnSquareTable = [64]int{
	65, 65, 65, 65, 65, 65, 65, 65,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	-5, -15, 40, 40, 40, -10, 0, -5,
	-1, -5, -15, -30, 36, -10, -5, -1,
	5, 10, -30, -30, -30, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightSquareTable = [64]int{
	-50, -10, -30, -30, -30, -30, -10, -50,
	-40, -20, 0, -5, -5, 0, -20, -40,
	-50, 0, 10, 15, 15, 10, 0, -50,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopSquareTable = [64]int{
	-20, -10, -20, -10, -10, -20, -10, -20,
	-10, 5, 0, 0, 0, 0, 30, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 0, 10, 10, 10, 15, 15, -10,
	-10, 10, 5, 10, 10, 15, 15, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 30, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookSquareTable = [64]int{
	-10, -10, 0, 5, 5, 0, -10, -10,
	-15, 0, 0, 0, 0, 0, 0, -15,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var queenSquareTable = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, -5, -5, -5, -5, -5, 0, -10,
	0, 0, 5, 5, 5, 5, 0, -5,
	-5, 0, 5, 5, 5, 5, 0, -5,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingSquareTableMiddlegame = [64]int{
	20, 30, 15, -75, -11, -75, 30, 20,
	20, 20, -50, -200, -200, -65, 20, 20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-30, -40, -40, -50, -60, -40, -40, -30,
	-30, -40, -65, -50, -50, -40, -40, -30,
	-30, -40, -50, -200, -200, -40, -40, -30,
	-30, -40, -60, -75, -75, -60, -40, -30,
}

// Returns an evaluation of the position in cp
// 1000000 or -1000000 is designated as checkmate
// Evaluations are not in the perspective of the player
func evaluate(b *Board) int {
	moves := b.generateLegalMoves()
	if len(moves) == 0 {
		if b.isCheck(b.turn) {
			return winVal * factor[b.turn]
		} else {
			return 0
		}
	}

	eval := 0
	eval += totalMaterial(b)
	eval += int(float64(piecePosition(b)))
	return eval
}

func totalMaterial(b *Board) int {
	sum := 0
	for _, piece := range b.squares {
		sum += material[piece]
	}
	return sum
}

func piecePosition(b *Board) int {
	sum := 0
	for i, piece := range b.squares {
		switch piece {
		case wK:
			sum += kingSquareTableMiddlegame[i]
		case bK:
			sum += kingSquareTableMiddlegame[i] * -1
		case wN:
			sum += knightSquareTable[i]
		case bN:
			sum += knightSquareTable[i] * -1
		case wB:
			sum += bishopSquareTable[i]
		case bB:
			sum += bishopSquareTable[i] * -1
		case wR:
			sum += rookSquareTable[i]
		case bR:
			sum += rookSquareTable[i] * -1
		case wQ:
			sum += queenSquareTable[i]
		case bQ:
			sum += queenSquareTable[i] * -1
		case wP:
			sum += whitePawnSquareTable[i]
		case bP:
			sum += blackPawnSquareTable[i] * -1
		}
	}
	return sum
}
