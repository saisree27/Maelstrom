package engine

func (b *board) getAllAttacks(o Color, occupied u64) u64 {
	// Generate all squares attacked/defended by opponent
	var attackedSquares u64 = 0

	// Begin first with king
	var opponentKing Square = Square(bitScanForward(b.getColorPieces(king, o)))
	attackedSquares |= kingAttacks(opponentKing)

	// Begin first with opponent pawns
	var opponentPawns u64 = b.getColorPieces(pawn, o)
	attackedSquares |= allPawnAttacks(opponentPawns, o)

	// Next, add opponent knights
	// Loop over knight locations and clear them
	var opponentKnights u64 = b.getColorPieces(knight, o)
	var lsb int
	for {
		if opponentKnights == 0 {
			break
		}
		lsb = bitScanForward(opponentKnights)
		opponentKnights &= ^(1 << lsb)
		attackedSquares |= knightAttacks(Square(lsb))
	}

	// TODO: Slider pieces (bishop, rook, queen)
	var opponentBishops u64 = b.getColorPieces(bishop, o)
	for {
		if opponentBishops == 0 {
			break
		}
		lsb = bitScanForward(opponentBishops)
		opponentBishops &= ^(1 << lsb)
		attackedSquares |= getBishopAttacks(Square(lsb), occupied)
	}

	var opponentRooks u64 = b.getColorPieces(rook, o)
	for {
		if opponentRooks == 0 {
			break
		}
		lsb = bitScanForward(opponentRooks)
		opponentRooks &= ^(1 << lsb)
		attackedSquares |= getRookAttacks(Square(lsb), occupied)
	}

	var opponentQueens u64 = b.getColorPieces(queen, o)
	for {
		if opponentQueens == 0 {
			break
		}
		lsb = bitScanForward(opponentQueens)
		opponentQueens &= ^(1 << lsb)
		attackedSquares |= (getRookAttacks(Square(lsb), occupied) | getBishopAttacks(Square(lsb), occupied))
	}

	return attackedSquares
}

func allPawnAttacks(b u64, c Color) u64 {
	if c == WHITE {
		return shiftBitboard(b, NE) | shiftBitboard(b, NW)
	} else {
		return shiftBitboard(b, SE) | shiftBitboard(b, SW)
	}
}

func kingAttacks(sq Square) u64 {
	return kingAttacksSquareLookup[sq]
}

func knightAttacks(sq Square) u64 {
	return knightAttacksSquareLookup[sq]
}

func getBishopAttacks(sq Square, bb u64) u64 {
	magicIndex := (bb & bishopMasks[sq]) * bishopMagics[sq]
	return bishopAttacks[sq][magicIndex>>bishopShifts[sq]]
}

func getRookAttacks(sq Square, bb u64) u64 {
	magicIndex := (bb & rookMasks[sq]) * rookMagics[sq]
	return rookAttacks[sq][magicIndex>>rookShifts[sq]]
}

func (b *board) generateMovesFromLocs(m *[]Move, sq Square, locs u64, c Color) {
	var lsb Square
	for {
		if locs == 0 {
			break
		}

		lsb = Square(popLSB(&locs))

		var newMove Move = Move{
			from: sq, to: Square(lsb), piece: b.squares[sq],
			colorMoved: c, captured: b.squares[lsb]}
		if b.squares[lsb] == EMPTY {
			newMove.movetype = QUIET
		} else {
			newMove.movetype = CAPTURE
		}

		*m = append(*m, newMove)
	}
}

func (b *board) getKingMoves(m *[]Move, sq Square, attacks u64, occupied u64, c Color) {
	var kingMoves u64 = kingAttacks(sq) & ^attacks & ^occupied
	b.generateMovesFromLocs(m, sq, kingMoves, c)

}

func (b *board) getKnightMoves(m *[]Move, knights u64, pinned u64, occupied u64, c Color) {
	// since a pinned knight means no possible moves, can just mask by non-pinned
	knights &= ^pinned

	var sq Square
	for {
		if knights == 0 {
			break
		}

		sq = Square(popLSB(&knights))

		var knightMoves u64 = knightAttacks(Square(sq)) & ^occupied
		b.generateMovesFromLocs(m, sq, knightMoves, c)
	}
}

func (b *board) getPawnMoves(m *[]Move, pawns u64, pinned u64, occupied u64, opponents u64, c Color) {
	// have to handle pinned pawns (TODO)
	// promotion & enpassant will be handled in generateMovesFromLocs
	var pinnedPawns u64 = pawns & pinned
	UNUSED(pinnedPawns)

	var unpinnedPawns u64 = pawns & ^pinned
	var sq Square
	var bbSQ u64
	for {
		if unpinnedPawns == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedPawns))
		bbSQ = 1 << sq
		var pawnMoves u64 = shiftBitboard(bbSQ, pawnPushDirection[c]) & ^occupied
		if bbSQ&ranks[startingRank[c]] != 0 {
			// two move pushes allowed
			pawnMoves |= shiftBitboard(bbSQ, pawnPushDirection[c]*2) & ^occupied
		}
		// add captures
		if c == WHITE {
			pawnMoves |= (whitePawnAttacksSquareLookup[sq] & opponents)
		} else {
			pawnMoves |= (blackPawnAttacksSquareLookup[sq] & opponents)
		}

		b.generateMovesFromLocs(m, sq, pawnMoves, c)
	}
}

func (b *board) getBishopMoves(m *[]Move, bishops u64, pinned u64, occupied u64, opponents u64, c Color) {
	var pinnedBishops u64 = bishops & pinned
	UNUSED(pinnedBishops)

	var unpinnedBishops u64 = bishops & ^pinned
	var sq Square
	for {
		if unpinnedBishops == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedBishops))

		var bishopMoves u64 = getBishopAttacks(sq, occupied|opponents) & ^occupied

		b.generateMovesFromLocs(m, sq, bishopMoves, c)
	}
}

func (b *board) getRookMoves(m *[]Move, rooks u64, pinned u64, occupied u64, opponents u64, c Color) {
	var pinnedRooks u64 = rooks & pinned
	UNUSED(pinnedRooks)

	var unpinnedRooks u64 = rooks & ^pinned
	var sq Square
	for {
		if unpinnedRooks == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedRooks))

		var rookMoves u64 = getRookAttacks(sq, occupied|opponents) & ^occupied

		b.generateMovesFromLocs(m, sq, rookMoves, c)
	}
}

func (b *board) getQueenMoves(m *[]Move, queens u64, pinned u64, occupied u64, opponents u64, c Color) {
	var pinnedQueens u64 = queens & pinned
	UNUSED(pinnedQueens)

	var unpinnedQueens u64 = queens & ^pinned
	var sq Square
	for {
		if unpinnedQueens == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedQueens))

		var rookMoves u64 = getRookAttacks(sq, occupied|opponents) & ^occupied
		var bishopMoves u64 = getBishopAttacks(sq, occupied|opponents) & ^occupied

		b.generateMovesFromLocs(m, sq, rookMoves|bishopMoves, c)
	}
}

func (b *board) generateLegalMoves() []Move {
	// Setup move list. Will be appending moves to this
	var m []Move = []Move{}

	// Get player and opponent colors
	var player Color = b.turn
	var opponent Color = reverseColor(b.turn)

	// Lookup all player pieces and opponent pieces
	var playerPieces u64 = b.colors[player]
	var opponentPieces u64 = b.colors[opponent]

	// Find all squares controlled by opponent
	var attacks u64 = b.getAllAttacks(opponent, b.occupied)
	printBitBoard(attacks)

	// STEP 1: Generate king moves, since we can look for checks after
	// Get player king square
	var playerKing Square = Square(bitScanForward(b.getColorPieces(king, player)))
	b.getKingMoves(&m, playerKing, attacks, playerPieces, player)

	// STEP 2: Calculate checks and pins (TODO)
	var checkers u64 = 0
	UNUSED(checkers)
	var pinned u64 = 0

	// STEP 3: Calculate knight moves
	var playerKnights u64 = b.getColorPieces(knight, player)
	b.getKnightMoves(&m, playerKnights, pinned, playerPieces, player)

	// STEP 3: Calculate pawn moves
	var playerPawns u64 = b.getColorPieces(pawn, player)
	b.getPawnMoves(&m, playerPawns, pinned, playerPieces, opponentPieces, player)

	// STEP 4: Calculate bishop moves
	var playerBishops u64 = b.getColorPieces(bishop, player)
	b.getBishopMoves(&m, playerBishops, pinned, playerPieces, opponentPieces, player)

	// STEP 5: Calculate rook moves
	var playerRooks u64 = b.getColorPieces(rook, player)
	b.getRookMoves(&m, playerRooks, pinned, playerPieces, opponentPieces, player)

	// STEP 5: Calculate rook moves
	var playerQueens u64 = b.getColorPieces(queen, player)
	b.getQueenMoves(&m, playerQueens, pinned, playerPieces, opponentPieces, player)

	return m
}
