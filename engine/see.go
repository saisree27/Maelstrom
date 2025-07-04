package engine

var SEE_PIECE_VALUES = [6]int{
	100,
	300,
	300,
	500,
	900,
	WIN_VAL,
}

// STATIC EXCHANGE EVALUATION
//
//	We can prune some captures by doing a unary tree search to determine if the capture is ideal or not.
//	At each step, we just choose the least valuable attacker of the square we move to and check if it can
//	capture the square the move went to. Once we do this until there are no more captures left to that square,
//	we'll have a decent idea how much this capture gains/loses material.
//
// Pseudocode implementation from https://www.chessprogramming.org/SEE_-_The_Swap_Algorithm
func see(b *Board, move Move) int {
	gain := [32]int{}
	d := 0
	mayXRay := b.occupied & ^(b.pieces[W_K] | b.pieces[W_N] | b.pieces[B_K] | b.pieces[B_N]) // pawns, bishops, rooks. queens
	fromSet := SQUARE_TO_BITBOARD[move.from]
	occ := b.occupied ^ SQUARE_TO_BITBOARD[move.from] ^ SQUARE_TO_BITBOARD[move.to]
	attadef := b.AllAttackersOf(move.to, occ)

	stm := ReverseColor(b.turn)
	initial_gain := 0
	if move.captured != EMPTY {
		initial_gain += SEE_PIECE_VALUES[PieceToPieceType(move.captured)]
	}

	attacker := PieceToPieceType(move.piece)
	seen := u64(0)

	gain[d] = initial_gain

	for fromSet != 0 {
		d++
		gain[d] = SEE_PIECE_VALUES[attacker] - gain[d-1]
		attadef &= ^fromSet
		occ &= ^fromSet
		seen |= fromSet

		if fromSet&mayXRay != 0 {
			attadef |= b.XRayAttacks(move.to, occ) & ^seen
		}

		fromSet = getLeastValuablePiece(b, attadef, stm, &attacker)

		stm = ReverseColor(stm)
	}

	// Negamax-ing gains
	for d > 1 {
		d--
		gain[d-1] = -Max(-gain[d-1], gain[d])
	}

	return gain[0]
}

func getLeastValuablePiece(b *Board, attadef u64, stm Color, piece *PieceType) u64 {
	for *piece = PAWN; *piece <= KING; *piece++ {
		subset := attadef & b.GetColorPieces(*piece, stm)
		if subset != 0 {
			return subset & -subset
		}
	}

	return 0
}
