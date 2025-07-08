package engine

// Values for material and checkmate
const WIN_VAL int = 1000000

var REVERSE_SQUARE = [64]int{
	56, 57, 58, 59, 60, 61, 62, 63,
	48, 49, 50, 51, 52, 53, 54, 55,
	40, 41, 42, 43, 44, 45, 46, 47,
	32, 33, 34, 35, 36, 37, 38, 39,
	24, 25, 26, 27, 28, 29, 30, 31,
	16, 17, 18, 19, 20, 21, 22, 23,
	8, 9, 10, 11, 12, 13, 14, 15,
	0, 1, 2, 3, 4, 5, 6, 7,
}

func EvaluateNNUE(b *Board) int {
	if b.IsInsufficientMaterial() {
		return 0
	}

	eval := 0
	stm := b.turn
	if stm == WHITE {
		eval = int(Forward(&GlobalNNUE, &b.accumulators.white, &b.accumulators.black))
	} else {
		eval = -int(Forward(&GlobalNNUE, &b.accumulators.black, &b.accumulators.white))
	}

	if eval < -WIN_VAL {
		eval = -WIN_VAL
	}

	if eval > WIN_VAL {
		eval = WIN_VAL
	}

	eval = materialScaling(b, eval)
	return eval
}

func materialScaling(b *Board, eval int) int {
	material := PopCount(b.pieces[W_P]|b.pieces[B_P]) * SEE_PIECE_VALUES[0]
	material += PopCount(b.pieces[W_N]|b.pieces[B_N]) * SEE_PIECE_VALUES[1]
	material += PopCount(b.pieces[W_B]|b.pieces[B_B]) * SEE_PIECE_VALUES[2]
	material += PopCount(b.pieces[W_R]|b.pieces[B_R]) * SEE_PIECE_VALUES[3]
	material += PopCount(b.pieces[W_Q]|b.pieces[B_Q]) * SEE_PIECE_VALUES[4]

	return eval * (700 + material/32) / 1024
}
