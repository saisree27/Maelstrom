package engine

import "fmt"

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

func (b *board) getKingMoves(m *[]Move, sq Square, attacks u64, player u64, c Color) {
	var kingMoves u64 = kingAttacks(sq) & ^attacks & ^player
	b.generateMovesFromLocs(m, sq, kingMoves, c)

}

func (b *board) getKnightMoves(m *[]Move, knights u64, pinned u64, player u64, c Color, allowed u64) {
	// since a pinned knight means no possible moves, can just mask by non-pinned
	knights &= ^pinned

	var sq Square
	for {
		if knights == 0 {
			break
		}

		sq = Square(popLSB(&knights))

		var knightMoves u64 = knightAttacks(Square(sq)) & ^player & allowed
		b.generateMovesFromLocs(m, sq, knightMoves, c)
	}
}

func (b *board) getPawnMoves(m *[]Move, pawns u64, pinned u64, occupied u64, opponents u64, c Color, allowed u64) {
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
		var pawnMoves u64 = (shiftBitboard(bbSQ, pawnPushDirection[c]) & ^occupied) & allowed
		if pawnMoves != 0 {
			if bbSQ&ranks[startingRank[c]] != 0 {
				// two move pushes allowed
				pawnMoves |= (shiftBitboard(bbSQ, pawnPushDirection[c]*2) & ^occupied) & allowed
			}
		}
		// add captures
		if c == WHITE {
			pawnMoves |= (whitePawnAttacksSquareLookup[sq] & opponents) & allowed
		} else {
			pawnMoves |= (blackPawnAttacksSquareLookup[sq] & opponents) & allowed
		}

		b.generateMovesFromLocs(m, sq, pawnMoves, c)
	}
}

func (b *board) getBishopMoves(m *[]Move, bishops u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedBishops u64 = bishops & pinned
	UNUSED(pinnedBishops)

	var unpinnedBishops u64 = bishops & ^pinned
	var sq Square
	for {
		if unpinnedBishops == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedBishops))

		var bishopMoves u64 = getBishopAttacks(sq, player|opponents) & ^player & allowed

		b.generateMovesFromLocs(m, sq, bishopMoves, c)
	}
}

func (b *board) getRookMoves(m *[]Move, rooks u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedRooks u64 = rooks & pinned
	UNUSED(pinnedRooks)

	var unpinnedRooks u64 = rooks & ^pinned
	var sq Square
	for {
		if unpinnedRooks == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedRooks))

		var rookMoves u64 = getRookAttacks(sq, player|opponents) & ^player & allowed

		b.generateMovesFromLocs(m, sq, rookMoves, c)
	}
}

func (b *board) getQueenMoves(m *[]Move, queens u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedQueens u64 = queens & pinned
	UNUSED(pinnedQueens)

	var unpinnedQueens u64 = queens & ^pinned
	var sq Square
	for {
		if unpinnedQueens == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedQueens))

		var rookMoves u64 = getRookAttacks(sq, player|opponents) & ^player & allowed
		var bishopMoves u64 = getBishopAttacks(sq, player|opponents) & ^player & allowed

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

	// Mask that specifies which squares are pieces allowed to move
	// Set to all bits by default
	var allowed u64 = 0xFFFFFFFFFFFFFFFF

	// Find all squares controlled by opponent
	var attacks u64 = b.getAllAttacks(opponent, b.occupied)
	// printBitBoard(attacks)

	// STEP 1: Generate king moves, since we can look for checks after
	// Get player king square
	var playerKing Square = Square(bitScanForward(b.getColorPieces(king, player)))
	b.getKingMoves(&m, playerKing, attacks, playerPieces, player)

	// STEP 2: Calculate checks and pins (TODO)
	var checkers u64 = 0

	// 2a: get checks of pawns and knight since we don't need to check for pieces in between
	if player == WHITE {
		checkers |= (whitePawnAttacksSquareLookup[playerKing] & b.pieces[bP])
	} else {
		checkers |= (blackPawnAttacksSquareLookup[playerKing] & b.pieces[wP])
	}
	checkers |= (knightAttacks(playerKing) & b.getColorPieces(knight, opponent))

	// 2b: get bishop/rook/queen attacks and check if there are pieces between them and the king
	var pinPoss u64 = (getBishopAttacks(playerKing, b.colors[opponent]) & (b.getColorPieces(bishop, opponent) | b.getColorPieces(queen, opponent)))
	pinPoss |= (getRookAttacks(playerKing, b.colors[opponent]) & (b.getColorPieces(rook, opponent) | b.getColorPieces(queen, opponent)))

	var pinned u64 = 0
	var lsb int
	var piecesBetween u64
	for {
		if pinPoss == 0 {
			break
		}
		lsb = popLSB(&pinPoss)
		piecesBetween = squaresBetween[playerKing][lsb] & playerPieces

		fmt.Println("SQUARES BETWEEN: ")
		printBitBoard(piecesBetween)

		if piecesBetween == 0 {
			checkers ^= 1 << lsb
		} else if piecesBetween != 0 && (piecesBetween&(piecesBetween-1)) == 0 {
			// only one piece between player and king, since otherwise there is no pin
			pinned ^= piecesBetween
		}
	}

	fmt.Println("\nCheckers: ")
	printBitBoard(checkers)
	fmt.Println("\nPinned: ")
	printBitBoard(pinned)

	var numCheckers int = popCount(checkers)
	if numCheckers == 2 {
		// double check so only king moves allowed
		return m
	} else if numCheckers == 1 {
		var checker Square = Square(bitScanForward(checkers))
		var checkerPiece = b.squares[checker]

		// get possible captures of the checker
		var possibleCaptures u64 = 0
		if player == WHITE {
			possibleCaptures |= blackPawnAttacksSquareLookup[checker] & b.pieces[wP]
			possibleCaptures |= knightAttacks(checker) & b.pieces[wN]
			possibleCaptures |= getBishopAttacks(checker, b.occupied) & (b.pieces[wB] | b.pieces[wQ])
			possibleCaptures |= getRookAttacks(checker, b.occupied) & (b.pieces[wR] | b.pieces[wQ])
			possibleCaptures &= ^pinned
		} else {
			possibleCaptures |= whitePawnAttacksSquareLookup[checker] & b.pieces[bP]
			possibleCaptures |= knightAttacks(checker) & b.pieces[bN]
			possibleCaptures |= getBishopAttacks(checker, b.occupied) & (b.pieces[bB] | b.pieces[bQ])
			possibleCaptures |= getRookAttacks(checker, b.occupied) & (b.pieces[bR] | b.pieces[bQ])
			possibleCaptures &= ^pinned
		}

		var lsb Square
		for {
			if possibleCaptures == 0 {
				break
			}
			lsb = Square(popLSB(&possibleCaptures))

			b.generateMovesFromLocs(&m, lsb, 1<<checker, player)
		}

		if checkerPiece == bN || checkerPiece == wN {
			// can only capture checker and move king, which has been handled above
			return m
		} else if checkerPiece == bP || checkerPiece == wP {
			// TODO: add enpassant capture of checks
			return m
		}

		var blockSquares u64 = squaresBetween[playerKing][checker]
		allowed = blockSquares

		b.getKnightMoves(&m, b.getColorPieces(knight, player), pinned, playerPieces, player, allowed)
		b.getPawnMoves(&m, b.getColorPieces(pawn, player), pinned, b.occupied, opponentPieces, player, allowed)
		b.getBishopMoves(&m, b.getColorPieces(bishop, player), pinned, playerPieces, opponentPieces, player, allowed)
		b.getRookMoves(&m, b.getColorPieces(rook, player), pinned, playerPieces, opponentPieces, player, allowed)
		b.getQueenMoves(&m, b.getColorPieces(queen, player), pinned, playerPieces, opponentPieces, player, allowed)

		return m
	}

	// STEP 3: Calculate knight moves
	var playerKnights u64 = b.getColorPieces(knight, player)
	b.getKnightMoves(&m, playerKnights, pinned, playerPieces, player, allowed)

	// STEP 3: Calculate pawn moves
	var playerPawns u64 = b.getColorPieces(pawn, player)
	b.getPawnMoves(&m, playerPawns, pinned, b.occupied, opponentPieces, player, allowed)

	// STEP 4: Calculate bishop moves
	var playerBishops u64 = b.getColorPieces(bishop, player)
	b.getBishopMoves(&m, playerBishops, pinned, playerPieces, opponentPieces, player, allowed)

	// STEP 5: Calculate rook moves
	var playerRooks u64 = b.getColorPieces(rook, player)
	b.getRookMoves(&m, playerRooks, pinned, playerPieces, opponentPieces, player, allowed)

	// STEP 5: Calculate rook moves
	var playerQueens u64 = b.getColorPieces(queen, player)
	b.getQueenMoves(&m, playerQueens, pinned, playerPieces, opponentPieces, player, allowed)

	return m
}
