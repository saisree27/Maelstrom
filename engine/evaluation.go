package engine

var winVal int = 1000000
var factor = map[Color]int{
	WHITE: 1, BLACK: -1,
}

var material = map[Piece]int{
	wP: 120, bP: -120,
	wN: 300, bN: -300,
	wB: 300, bB: -300,
	wR: 500, bR: -500,
	wQ: 900, bQ: -900,
	wK: 0, bK: 0,
	EMPTY: 0,
}

var reversePSQ = [64]int{
	56, 57, 58, 59, 60, 61, 62, 63,
	48, 49, 50, 51, 52, 53, 54, 55,
	40, 41, 42, 43, 44, 45, 46, 47,
	32, 33, 34, 35, 36, 37, 38, 39,
	24, 25, 26, 27, 28, 29, 30, 31,
	16, 17, 18, 19, 20, 21, 22, 23,
	8, 9, 10, 11, 12, 13, 14, 15,
	0, 1, 2, 3, 4, 5, 6, 7,
}

var pawnSquareTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	10, 10, 0, -10, -10, 0, 10, 10,
	5, 0, 0, 5, 5, 0, 0, 5,
	0, 0, 10, 20, 20, 10, 0, 0,
	5, 5, 5, 10, 10, 5, 5, 5,
	10, 10, 10, 20, 20, 10, 10, 10,
	20, 20, 20, 30, 30, 20, 20, 20,
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
	-10, 5, 0, -10, -10, 0, 30, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 0, 10, 10, 10, 15, 15, -10,
	-10, 10, 5, 10, 10, 15, 15, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 30, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookSquareTable = [64]int{
	-10, -10, 3, 5, 5, 3, -10, -10,
	-15, 0, 0, 5, 5, 0, 0, -15,
	-5, 0, 0, 5, 5, 0, 0, -5,
	-5, 0, 0, 5, 5, 0, 0, -5,
	-5, 0, 0, 5, 5, 0, 0, -5,
	-5, 0, 0, 5, 5, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var queenSquareTable = [64]int{
	-20, -10, -10, 5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, -5, -5, -5, -5, -5, 0, -10,
	0, 0, 5, 5, 5, 5, 0, -5,
	-5, 0, 5, 5, 5, 5, 0, -5,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingSquareTableMiddlegame = [64]int{
	10, 30, 25, -10, -10, -15, 30, 10,
	0, -5, -50, -75, -75, -65, -5, 0,
	-10, -20, -20, -20, -20, -20, -20, -10,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-30, -40, -40, -50, -60, -40, -40, -30,
	-30, -40, -65, -50, -50, -40, -40, -30,
	-30, -40, -50, -75, -75, -40, -40, -30,
	-30, -40, -60, -75, -75, -60, -40, -30,
}

var kingSquareTableEndgame = [64]int{
	-50, -10, 0, 0, 0, 0, -10, -50,
	-10, 0, 10, 10, 10, 10, 0, -10,
	0, 10, 15, 15, 15, 15, 10, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 10, 15, 15, 15, 15, 10, 0,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-50, -10, 0, 0, 0, 0, -10, -50,
}

var minorPieceDevelopment = 30
var kingAir = 30
var noCastlingRights = 80
var pawnsBlocked = 20
var mobility = 10
var centerControl = 25

// Returns an evaluation of the position in cp
// 1000000 or -1000000 is designated as checkmate
// Evaluations are not in the perspective of the player
func evaluate(b *Board) int {
	moves := b.generateLegalMoves()
	if len(moves) == 0 {
		if b.isCheck(b.turn) {
			// if white is in check, then val should be -1000000
			return winVal * factor[reverseColor(b.turn)]
		} else {
			return 0
		}
	}

	eval := 0

	material, total := totalMaterialAndPieces(b)

	eval += int(float64(material) * 1.2)
	eval += piecePosition(b, total)

	whiteAttacks := b.getAllAttacks(WHITE, b.occupied, b.getColorPieces(rook, WHITE), b.getColorPieces(bishop, WHITE))

	eval += popCount(whiteAttacks) * mobility

	blackAttacks := b.getAllAttacks(BLACK, b.occupied, b.getColorPieces(rook, BLACK), b.getColorPieces(bishop, BLACK))

	eval -= popCount(blackAttacks) * mobility

	// during opening, harsh penalty for minor pieces not yet developed
	if b.plyCnt <= 25 {
		whiteKnightsAndBishops := b.getColorPieces(knight, WHITE) | b.getColorPieces(bishop, WHITE)
		blackKnightsAndBishops := b.getColorPieces(knight, BLACK) | b.getColorPieces(bishop, BLACK)

		eval -= popCount(whiteKnightsAndBishops&ranks[R1]) * minorPieceDevelopment
		eval += popCount(blackKnightsAndBishops&ranks[R8]) * minorPieceDevelopment
	}

	if total >= 15 && b.pieces[wQ]|b.pieces[bQ] != 0 {
		// penalty/reward for air around king
		whiteKingAttacks := kingAttacksSquareLookup[bitScanForward(b.getColorPieces(king, WHITE))]
		air := popCount(whiteKingAttacks & b.empty)
		if air > 2 {
			eval -= air * kingAir
		}

		blackKingAttacks := kingAttacksSquareLookup[bitScanForward(b.getColorPieces(king, BLACK))]
		air = popCount(blackKingAttacks & b.empty)
		if air > 2 {
			eval += air * kingAir
		}
	}

	if b.plyCnt <= 25 {
		// during opening, harsh penalty/reward if castling rights are gone and queens are still on
		if !(b.OO || b.OOO) && !((b.squares[g1] == wK && b.squares[h1] != wR) || b.squares[c1] == wK && b.squares[a1] != wR && b.squares[b1] != wR) {
			eval -= noCastlingRights
		}
		if !(b.oo || b.ooo) && !((b.squares[g8] == bK && b.squares[h8] != bR) || b.squares[c8] == bK && b.squares[a8] != bR && b.squares[b8] != bR) {
			eval += noCastlingRights
		}
	}

	// check if pawns are blocked by pieces of same color
	whitePawns := b.getColorPieces(pawn, WHITE)
	blockedCount := popCount(shiftBitboard(whitePawns, NORTH) & b.colors[WHITE])

	eval -= blockedCount * pawnsBlocked

	blackPawns := b.getColorPieces(pawn, BLACK)
	blockedCount = popCount(shiftBitboard(blackPawns, SOUTH) & b.colors[BLACK])

	eval += blockedCount * pawnsBlocked

	// center control
	whiteControl := b.getColorPieces(pawn, WHITE) | whiteAttacks
	whiteCenterControl := (whiteControl & sToBB[e4]) | (whiteControl & sToBB[e5]) | (whiteControl & sToBB[d4]) | (whiteControl & sToBB[d4])

	blackControl := b.getColorPieces(pawn, BLACK) | blackAttacks
	blackCenterControl := (blackControl & sToBB[e4]) | (blackControl & sToBB[e5]) | (blackControl & sToBB[d4]) | (blackControl & sToBB[d4])

	eval += (popCount(whiteCenterControl) - popCount(blackCenterControl)) * centerControl

	return eval
}

func totalMaterialAndPieces(b *Board) (int, int) {
	sum := 0
	total := 0

	for _, piece := range b.squares {
		sum += material[piece]
		if piece != EMPTY {
			total++
		}
	}
	return sum, total
}

func piecePosition(b *Board, totalPieces int) int {
	sum := 0
	noQueensLeft := b.pieces[wQ]|b.pieces[bQ] == 0
	switchEndgame := (totalPieces <= 20 && noQueensLeft) || (totalPieces <= 10)

	for i, piece := range b.squares {
		switch piece {
		case wK:
			if switchEndgame {
				sum += kingSquareTableEndgame[i]
			} else {
				sum += kingSquareTableMiddlegame[i]
			}
		case bK:
			if switchEndgame {
				sum += kingSquareTableEndgame[reversePSQ[i]] * -1
			} else {
				sum += kingSquareTableMiddlegame[reversePSQ[i]] * -1
			}
		case wN:
			sum += knightSquareTable[i]
		case bN:
			sum += knightSquareTable[reversePSQ[i]] * -1
		case wB:
			sum += bishopSquareTable[i]
		case bB:
			sum += bishopSquareTable[reversePSQ[i]] * -1
		case wR:
			sum += rookSquareTable[i]
		case bR:
			sum += rookSquareTable[reversePSQ[i]] * -1
		case wQ:
			sum += queenSquareTable[i]
		case bQ:
			sum += queenSquareTable[reversePSQ[i]] * -1
		case wP:
			sum += pawnSquareTable[i]
		case bP:
			sum += pawnSquareTable[reversePSQ[i]] * -1
		}
	}
	return sum
}
