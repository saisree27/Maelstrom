package engine

// Piece values for SEE calculation - using smaller values than main evaluation
const (
	seePawnVal   = 100
	seeKnightVal = 325
	seeBishopVal = 325
	seeRookVal   = 500
	seeQueenVal  = 975
	seeKingVal   = 10000
)

var seePieceValues = map[Piece]int{
	wP: seePawnVal, bP: seePawnVal,
	wN: seeKnightVal, bN: seeKnightVal,
	wB: seeBishopVal, bB: seeBishopVal,
	wR: seeRookVal, bR: seeRookVal,
	wQ: seeQueenVal, bQ: seeQueenVal,
	wK: seeKingVal, bK: seeKingVal,
	EMPTY: 0,
}

// getLeastValuableAttacker finds the least valuable piece that can attack a square
func getLeastValuableAttacker(b *Board, target Square, side Color, occupied u64) (Square, Piece) {
	var attackers u64

	if side == WHITE {
		// White pawn attacks
		attackers = whitePawnAttacksSquareLookup[target] & b.pieces[wP] & occupied
		if attackers != 0 {
			return Square(bitScanForward(attackers)), wP
		}
	} else {
		// Black pawn attacks
		attackers = blackPawnAttacksSquareLookup[target] & b.pieces[bP] & occupied
		if attackers != 0 {
			return Square(bitScanForward(attackers)), bP
		}
	}

	// Knight attacks
	attackers = knightAttacks(target) & b.getColorPieces(knight, side) & occupied
	if attackers != 0 {
		return Square(bitScanForward(attackers)), getPieceFromTypeAndColor(knight, side)
	}

	// Bishop/Queen diagonal attacks
	diagonalSliders := b.getColorPieces(bishop, side) | b.getColorPieces(queen, side)
	attackers = getBishopAttacks(target, occupied) & diagonalSliders & occupied
	if attackers != 0 {
		sq := Square(bitScanForward(attackers))
		if b.squares[sq] == wB || b.squares[sq] == bB {
			return sq, getPieceFromTypeAndColor(bishop, side)
		}
		return sq, getPieceFromTypeAndColor(queen, side)
	}

	// Rook/Queen orthogonal attacks
	orthogonalSliders := b.getColorPieces(rook, side) | b.getColorPieces(queen, side)
	attackers = getRookAttacks(target, occupied) & orthogonalSliders & occupied
	if attackers != 0 {
		sq := Square(bitScanForward(attackers))
		if b.squares[sq] == wR || b.squares[sq] == bR {
			return sq, getPieceFromTypeAndColor(rook, side)
		}
		return sq, getPieceFromTypeAndColor(queen, side)
	}

	// King attacks
	attackers = kingAttacks(target) & b.getColorPieces(king, side) & occupied
	if attackers != 0 {
		return Square(bitScanForward(attackers)), getPieceFromTypeAndColor(king, side)
	}

	return Square(0), EMPTY
}

// Helper function to get the correct piece from type and color
func getPieceFromTypeAndColor(pt PieceType, c Color) Piece {
	if c == WHITE {
		switch pt {
		case pawn:
			return wP
		case knight:
			return wN
		case bishop:
			return wB
		case rook:
			return wR
		case queen:
			return wQ
		case king:
			return wK
		}
	} else {
		switch pt {
		case pawn:
			return bP
		case knight:
			return bN
		case bishop:
			return bB
		case rook:
			return bR
		case queen:
			return bQ
		case king:
			return bK
		}
	}
	return EMPTY
}

// seeCapture performs static exchange evaluation on a capture move
func seeCapture(b *Board, m Move) int {
	if m.movetype != CAPTURE && m.movetype != CAPTUREANDPROMOTION {
		return 0
	}

	target := m.to
	occupied := b.occupied
	side := reverseColor(b.turn)
	score := seePieceValues[m.captured]

	// Remove attacker from occupied squares
	occupied ^= sToBB[m.from]

	// If it's a promotion, use the promoted piece value
	attackerValue := seePieceValues[m.piece]
	if m.movetype == CAPTUREANDPROMOTION {
		attackerValue = seePieceValues[m.promote]
	}

	// First capture
	gain := make([]int, 32)
	depth := 1
	gain[0] = score

	// Continue making captures until no more are possible
	for {
		score = -score + attackerValue // Value of previous capture

		// Find next least valuable attacker
		from, piece := getLeastValuableAttacker(b, target, side, occupied)
		if piece == EMPTY {
			break
		}

		// Remove attacker from occupied squares
		occupied ^= sToBB[from]
		attackerValue = seePieceValues[piece]
		side = reverseColor(side)

		// Add to gain array
		gain[depth] = -gain[depth-1] + attackerValue
		depth++
	}

	// Back up through the gain array to find the best capture sequence
	depth--
	for i := depth - 1; i >= 0; i-- {
		if i == 0 {
			gain[i] = max(-gain[i+1], gain[i])
		} else {
			gain[i] = max(-gain[i+1], gain[i])
		}
	}

	return gain[0]
}

// seeMove performs static exchange evaluation on any move
func seeMove(b *Board, m Move) int {
	switch m.movetype {
	case CAPTURE, CAPTUREANDPROMOTION:
		return seeCapture(b, m)
	case PROMOTION:
		return seePieceValues[m.promote] - seePieceValues[wP]
	default:
		return 0
	}
}

// seeGreaterThan is an optimized version that checks if SEE value exceeds a threshold
func seeGreaterThan(b *Board, m Move, threshold int) bool {
	if m.movetype == QUIET {
		return false
	}

	target := m.to
	occupied := b.occupied
	side := reverseColor(b.turn)
	score := seePieceValues[m.captured]

	if score <= threshold {
		return false
	}

	// Remove attacker from occupied squares
	occupied ^= sToBB[m.from]

	// If it's a promotion, use the promoted piece value
	attackerValue := seePieceValues[m.piece]
	if m.movetype == CAPTUREANDPROMOTION {
		attackerValue = seePieceValues[m.promote]
	}

	score -= attackerValue
	if score >= threshold {
		return true
	}

	for {
		from, piece := getLeastValuableAttacker(b, target, side, occupied)
		if piece == EMPTY {
			break
		}

		occupied ^= sToBB[from]
		score = -score + seePieceValues[piece]
		side = reverseColor(side)

		if score >= threshold {
			return true
		}
	}

	return false
}
