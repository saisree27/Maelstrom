package engine

var SEE_PIECE_VALUES = [6]int{
	Params.SEE_PAWN_VALUE,
	Params.SEE_KNIGHT_VALUE,
	Params.SEE_BISHOP_VALUE,
	Params.SEE_ROOK_VALUE,
	Params.SEE_QUEEN_VALUE,
	WIN_VAL,
}

// STATIC EXCHANGE EVALUATION
//
//	We can prune some captures by doing a unary tree search to determine if the capture is ideal or not.
//	At each step, we just choose the least valuable attacker of the square we move to and check if it can
//	capture the square the move went to. Once we do this until there are no more captures left to that square,
//	we'll have a decent idea how much this capture gains/loses material.
//
// Implementation inspired from Ethereal/Weiss/Stormphrax's implementation of SEE thresholding
func SEE(move Move, b *Board, threshold int) bool {
	stm := b.turn
	gain := 0

	// Determine gain of the first move
	if move.movetype == EN_PASSANT {
		gain += SEE_PIECE_VALUES[PAWN]
	} else if move.movetype == CAPTURE || move.movetype == CAPTURE_AND_PROMOTION {
		gain += SEE_PIECE_VALUES[PieceToPieceType(move.captured)]
	}
	if move.movetype == PROMOTION || move.movetype == CAPTURE_AND_PROMOTION {
		gain += SEE_PIECE_VALUES[PieceToPieceType(move.promote)] - SEE_PIECE_VALUES[PAWN]
	}

	// Subtract threshold to keep our logic in terms of greater and less than 0
	gain -= threshold

	// If our gain is already negative after making the move, it will not be able to increase
	if gain < 0 {
		return false
	}

	// If we were to lose the moving piece (or promoted piece), are we still above threshold?
	// If so, no need to continue with SEE, just return true
	nextLoss := SEE_PIECE_VALUES[PieceToPieceType(move.piece)]
	if move.movetype == PROMOTION || move.movetype == CAPTURE_AND_PROMOTION {
		nextLoss = SEE_PIECE_VALUES[PieceToPieceType(move.promote)]
	}
	gain -= nextLoss
	if gain >= 0 {
		return true
	}

	sq := move.to
	occ := b.occupied ^ SQUARE_TO_BITBOARD[move.from] ^ SQUARE_TO_BITBOARD[move.to]
	queens := b.pieces[B_Q] | b.pieces[W_Q]
	bishops := b.pieces[B_B] | b.pieces[W_B] | queens
	rooks := b.pieces[B_R] | b.pieces[W_R] | queens

	var attackerPiece PieceType

	attackers := b.AllAttackersOf(sq, occ)
	nextTm := ReverseColor(stm)

	for {
		ourAttackers := attackers & b.colors[nextTm]

		if ourAttackers == 0 {
			break
		}

		next := getLeastValuablePiece(b, ourAttackers, nextTm, &attackerPiece)
		occ ^= next

		if attackerPiece == PAWN || attackerPiece == BISHOP || attackerPiece == QUEEN {
			attackers |= BishopAttacks(sq, occ) & bishops
		}

		if attackerPiece == ROOK || attackerPiece == QUEEN {
			attackers |= RookAttacks(sq, occ) & rooks
		}

		attackers &= occ
		gain = -gain - 1 - SEE_PIECE_VALUES[attackerPiece]
		nextTm = ReverseColor(nextTm)

		if gain >= 0 {
			if attackerPiece == KING && attackers&b.colors[nextTm] != 0 {
				nextTm = ReverseColor(nextTm)
			}
			break
		}
	}
	return stm != nextTm
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
