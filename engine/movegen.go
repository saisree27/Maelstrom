package engine

func (b *Board) getNumPins(o Color, occupied u64, kingLoc Square) int {
	diagonalThem := b.getColorPieces(rook, o) | b.getColorPieces(queen, o)
	orthogonalThem := b.getColorPieces(bishop, o) | b.getColorPieces(queen, o)

	var pinPoss u64 = (getBishopAttacks(kingLoc, b.colors[o]) & diagonalThem)
	pinPoss |= (getRookAttacks(kingLoc, b.colors[o]) & orthogonalThem)

	var pinned int = 0
	var lsb int
	var piecesBetween u64
	for {
		if pinPoss == 0 {
			break
		}
		lsb = popLSB(&pinPoss)
		piecesBetween = squaresBetween[kingLoc][lsb] & b.pieces[reverseColor(o)]

		if piecesBetween != 0 && (piecesBetween&(piecesBetween-1)) == 0 {
			// only one piece between player and king, since otherwise there is no pin
			pinned++
		}
	}
	return pinned
}

func (b *Board) getAllAttacks(o Color, occupied u64, ortho u64, diag u64) u64 {
	// Pre-calculate occupancy without king to avoid multiple XOR operations
	var playerKing = b.getColorPieces(king, reverseColor(o))
	var occupiedNoKing = occupied ^ playerKing

	// Start with king attacks (fastest to compute)
	var opponentKing = Square(bitScanForward(b.getColorPieces(king, o)))
	var attackedSquares = kingAttacks(opponentKing)

	// Add pawn attacks (also very fast)
	attackedSquares |= allPawnAttacks(b.getColorPieces(pawn, o), o)

	// Add knight attacks (fast due to lookup table)
	var knights = b.getColorPieces(knight, o)
	for knights != 0 {
		attackedSquares |= knightAttacks(Square(popLSB(&knights)))
	}

	// Sliding piece attacks (most expensive)
	// Process diagonal pieces (bishops + queens)
	var diagonalPieces = diag
	for diagonalPieces != 0 {
		attackedSquares |= getBishopAttacks(Square(popLSB(&diagonalPieces)), occupiedNoKing)
	}

	// Process orthogonal pieces (rooks + queens)
	var orthogonalPieces = ortho
	for orthogonalPieces != 0 {
		attackedSquares |= getRookAttacks(Square(popLSB(&orthogonalPieces)), occupiedNoKing)
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

func (b *Board) generateMovesFromLocs(m *[]Move, sq Square, locs u64, c Color) {
	var lsb Square
	var bbLSB u64
	var piece Piece
	var movet MoveType
	for locs != 0 {
		lsb = Square(popLSB(&locs))
		bbLSB = sToBB[lsb]
		piece = b.squares[sq]

		if piece == wP && (bbLSB&ranks[R8] != 0) {
			movet = PROMOTION
			if b.squares[lsb] != EMPTY {
				movet = CAPTUREANDPROMOTION
			}
			// Add all promotion pieces
			*m = append(*m,
				Move{from: sq, to: lsb, piece: wP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: wN},
				Move{from: sq, to: lsb, piece: wP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: wR},
				Move{from: sq, to: lsb, piece: wP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: wQ},
				Move{from: sq, to: lsb, piece: wP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: wB},
			)
		} else if piece == bP && (bbLSB&ranks[R1] != 0) {
			movet = PROMOTION
			if b.squares[lsb] != EMPTY {
				movet = CAPTUREANDPROMOTION
			}
			// Add all promotion pieces
			*m = append(*m,
				Move{from: sq, to: lsb, piece: bP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: bN},
				Move{from: sq, to: lsb, piece: bP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: bR},
				Move{from: sq, to: lsb, piece: bP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: bQ},
				Move{from: sq, to: lsb, piece: bP, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: bB},
			)
		} else {
			*m = append(*m, Move{
				from: sq, to: lsb, piece: piece,
				colorMoved: c, captured: b.squares[lsb],
				movetype: ternary(b.squares[lsb] == EMPTY, QUIET, CAPTURE),
			})
		}
	}
}

// Helper function for ternary operations
func ternary[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

func (b *Board) getKingMoves(m *[]Move, sq Square, attacks u64, player u64, c Color) {
	var kingMoves u64 = kingAttacks(sq) & ^attacks & ^player
	b.generateMovesFromLocs(m, sq, kingMoves, c)

}

func (b *Board) getKnightMoves(m *[]Move, knights u64, pinned u64, player u64, c Color, allowed u64) {
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

func (b *Board) getPawnMoves(m *[]Move, pawns u64, pinned u64, occupied u64, opponents u64, c Color, allowed u64) {
	// have to handle pinned pawns (TODO)
	// promotion & enpassant will be handled in generateMovesFromLocs
	var pinnedPawns u64 = pawns & pinned
	var bbSQ u64
	var kingLoc Square

	if pinnedPawns != 0 {
		var lsb Square
		for {
			if pinnedPawns == 0 {
				break
			}
			lsb = Square(popLSB(&pinnedPawns))
			bbSQ = sToBB[lsb]

			kingLoc = Square(bitScanForward(b.getColorPieces(king, c)))

			var pawnMoves u64 = (shiftBitboard(bbSQ, pawnPushDirection[c]) & ^occupied) & allowed & line[kingLoc][lsb]
			if pawnMoves != 0 {
				if bbSQ&ranks[startingRank[c]] != 0 {
					// two move pushes allowed
					pawnMoves |= (shiftBitboard(bbSQ, pawnPushDirection[c]*2) & ^occupied) & allowed & line[kingLoc][lsb]
				}
			}
			pawnMoves |= colorToPawnLookup[c][lsb] & opponents & allowed & line[kingLoc][lsb]
			b.generateMovesFromLocs(m, lsb, pawnMoves, c)
		}
	}

	var unpinnedPawns u64 = pawns & ^pinned
	var sq Square
	for {
		if unpinnedPawns == 0 {
			break
		}
		sq = Square(popLSB(&unpinnedPawns))
		bbSQ = sToBB[sq]
		var pawnMoves u64 = (shiftBitboard(bbSQ, pawnPushDirection[c]) & ^occupied) & allowed
		if b.squares[sq+Square(pawnPushDirection[c])] == EMPTY {
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

func (b *Board) getBishopMoves(m *[]Move, bishops u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedBishops u64 = bishops & pinned
	var kingLoc Square

	if pinnedBishops != 0 {
		var lsb Square
		for {
			if pinnedBishops == 0 {
				break
			}
			lsb = Square(popLSB(&pinnedBishops))

			kingLoc = Square(bitScanForward(b.getColorPieces(king, c)))

			var possible u64 = getBishopAttacks(lsb, b.occupied) & line[kingLoc][lsb] & allowed & ^player
			b.generateMovesFromLocs(m, lsb, possible, c)
		}
	}

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

func (b *Board) getRookMoves(m *[]Move, rooks u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedRooks u64 = rooks & pinned
	var kingLoc Square
	if pinnedRooks != 0 {
		var lsb Square
		for {
			if pinnedRooks == 0 {
				break
			}
			lsb = Square(popLSB(&pinnedRooks))

			kingLoc = Square(bitScanForward(b.getColorPieces(king, c)))

			var possible u64 = getRookAttacks(lsb, b.occupied) & line[kingLoc][lsb] & allowed & ^player
			b.generateMovesFromLocs(m, lsb, possible, c)
		}
	}

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

func (b *Board) getQueenMoves(m *[]Move, queens u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedQueens u64 = queens & pinned
	var kingLoc Square
	if pinnedQueens != 0 {
		var lsb Square
		for {
			if pinnedQueens == 0 {
				break
			}
			lsb = Square(popLSB(&pinnedQueens))
			kingLoc = Square(bitScanForward(b.getColorPieces(king, c)))

			var possible u64 = (getBishopAttacks(lsb, b.occupied) | getRookAttacks(lsb, b.occupied)) & line[kingLoc][lsb] & allowed & ^player

			b.generateMovesFromLocs(m, lsb, possible, c)
		}
	}

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

func (b *Board) getCastlingMoves(m *[]Move, pKing Square, attacks u64, c Color) {
	var castlingKingsidePossible bool = false
	var castlingQueensidePossible bool = false

	var allowed u64 = ^attacks & b.empty

	if c == WHITE {
		castlingKingsidePossible = (sToBB[f1]&allowed != 0) && (sToBB[g1]&allowed != 0)
		castlingQueensidePossible = (sToBB[b1]&b.empty != 0) && (sToBB[c1]&allowed != 0) && (sToBB[d1]&allowed != 0)
		if b.OO && castlingKingsidePossible && b.squares[h1] == wR {
			*m = append(*m, Move{from: e1, to: g1, piece: wK, movetype: KCASTLE, colorMoved: WHITE})
		}
		if b.OOO && castlingQueensidePossible && b.squares[a1] == wR {
			*m = append(*m, Move{from: e1, to: c1, piece: wK, movetype: QCASTLE, colorMoved: WHITE})
		}
	} else {
		castlingKingsidePossible = (sToBB[f8]&allowed != 0) && (sToBB[g8]&allowed != 0)
		castlingQueensidePossible = (sToBB[b8]&b.empty != 0) && (sToBB[c8]&allowed != 0) && (sToBB[d8]&allowed != 0)
		if b.oo && castlingKingsidePossible && b.squares[h8] == bR {
			*m = append(*m, Move{from: e8, to: g8, piece: bK, movetype: KCASTLE, colorMoved: BLACK})
		}

		if b.ooo && castlingQueensidePossible && b.squares[a8] == bR {
			*m = append(*m, Move{from: e8, to: c8, piece: bK, movetype: QCASTLE, colorMoved: BLACK})
		}
	}
}

func (b *Board) getPinnedPieces(kingSquare Square, c Color) (u64, u64) {
	opponent := reverseColor(c)
	var pinned, checkers u64 = 0, 0

	// Get opponent's sliding pieces
	orthogonalThem := b.getColorPieces(rook, opponent) | b.getColorPieces(queen, opponent)
	diagonalThem := b.getColorPieces(bishop, opponent) | b.getColorPieces(queen, opponent)

	// Check for direct attacks first (faster than pin detection)
	if c == WHITE {
		checkers = whitePawnAttacksSquareLookup[kingSquare] & b.pieces[bP]
	} else {
		checkers = blackPawnAttacksSquareLookup[kingSquare] & b.pieces[wP]
	}
	checkers |= knightAttacks(kingSquare) & b.getColorPieces(knight, opponent)

	// Get potential pinners on diagonals
	potentialPinners := getBishopAttacks(kingSquare, b.colors[opponent]) & diagonalThem
	var pinner u64
	for potentialPinners != 0 {
		pinner = sToBB[Square(popLSB(&potentialPinners))]
		between := squaresBetween[kingSquare][bitScanForward(pinner)] & b.occupied
		if popCount(between) == 1 && (between&b.colors[c]) != 0 {
			pinned |= between
		} else if popCount(between) == 0 {
			checkers |= pinner
		}
	}

	// Get potential pinners on ranks/files
	potentialPinners = getRookAttacks(kingSquare, b.colors[opponent]) & orthogonalThem
	for potentialPinners != 0 {
		pinner = sToBB[Square(popLSB(&potentialPinners))]
		between := squaresBetween[kingSquare][bitScanForward(pinner)] & b.occupied
		if popCount(between) == 1 && (between&b.colors[c]) != 0 {
			pinned |= between
		} else if popCount(between) == 0 {
			checkers |= pinner
		}
	}

	return pinned, checkers
}

func (b *Board) generateLegalMoves() []Move {
	// Setup move list with pre-allocated capacity
	m := make([]Move, 0, 48)

	player := b.turn
	opponent := reverseColor(player)

	// Get player pieces and opponent pieces
	playerPieces := b.colors[player]
	opponentPieces := b.colors[opponent]

	// Get player king and calculate pins/checkers
	playerKing := Square(bitScanForward(b.getColorPieces(king, player)))
	pinned, checkers := b.getPinnedPieces(playerKing, player)

	// Get attack map from opponent
	orthogonalThem := b.getColorPieces(rook, opponent) | b.getColorPieces(queen, opponent)
	diagonalThem := b.getColorPieces(bishop, opponent) | b.getColorPieces(queen, opponent)
	attacks := b.getAllAttacks(opponent, b.occupied, orthogonalThem, diagonalThem)

	// Generate king moves first
	b.getKingMoves(&m, playerKing, attacks, playerPieces, player)

	numCheckers := popCount(checkers)
	if numCheckers >= 2 {
		return m // Only king moves in double check
	}

	// Get allowed squares (all squares in single check, only blocking squares in check)
	allowed := u64(0xFFFFFFFFFFFFFFFF)
	if numCheckers == 1 {
		checkerSquare := Square(bitScanForward(checkers))
		checkerPiece := b.squares[checkerSquare]
		allowed = squaresBetween[playerKing][checkerSquare] | checkers

		// Handle en passant capture of checking pawn
		if b.enpassant != EMPTYSQ {
			enPassantShift := b.enpassant + Square(pawnPushDirection[opponent])
			if checkerSquare == enPassantShift && (checkerPiece == wP || checkerPiece == bP) {
				pawns := colorToPawnLookup[opponent][b.enpassant] & b.getColorPieces(pawn, player)

				// Handle unpinned pawns
				unpinnedPawns := pawns & ^pinned
				for unpinnedPawns != 0 {
					lsb := Square(popLSB(&unpinnedPawns))
					move := Move{from: lsb, to: b.enpassant, movetype: ENPASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[lsb]}

					b.makeMove(move)
					if !b.isCheck(player) {
						b.undo()
						m = append(m, move)
					} else {
						b.undo()
					}
				}

				// Handle pinned pawns
				pinnedPawns := pawns & pinned & line[checkerSquare][playerKing]
				if pinnedPawns != 0 {
					sq := Square(bitScanForward(pinnedPawns))
					m = append(m, Move{from: sq, to: b.enpassant, movetype: ENPASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[sq]})
				}
			}
		}
	}

	// Generate other piece moves
	b.getKnightMoves(&m, b.getColorPieces(knight, player), pinned, playerPieces, player, allowed)
	b.getPawnMoves(&m, b.getColorPieces(pawn, player), pinned, b.occupied, opponentPieces, player, allowed)
	b.getBishopMoves(&m, b.getColorPieces(bishop, player)|b.getColorPieces(queen, player), pinned, playerPieces, opponentPieces, player, allowed)
	b.getRookMoves(&m, b.getColorPieces(rook, player)|b.getColorPieces(queen, player), pinned, playerPieces, opponentPieces, player, allowed)

	// Only try castling if not in check
	if numCheckers == 0 {
		b.getCastlingMoves(&m, playerKing, attacks, player)

		// Handle en passant when not in check
		if b.enpassant != EMPTYSQ {
			pawns := colorToPawnLookup[opponent][b.enpassant] & b.getColorPieces(pawn, player)

			// Handle unpinned pawns
			unpinnedPawns := pawns & ^pinned
			for unpinnedPawns != 0 {
				lsb := Square(popLSB(&unpinnedPawns))
				move := Move{from: lsb, to: b.enpassant, movetype: ENPASSANT, captured: b.squares[b.enpassant], colorMoved: player, piece: b.squares[lsb]}

				b.makeMoveNoUpdate(move)
				if !b.isCheck(player) {
					b.undoNoUpdate(move)
					m = append(m, move)
				} else {
					b.undoNoUpdate(move)
				}
			}

			// Handle pinned pawns
			pinnedPawns := pawns & pinned & line[b.enpassant][playerKing]
			if pinnedPawns != 0 {
				sq := Square(bitScanForward(pinnedPawns))
				m = append(m, Move{from: sq, to: b.enpassant, movetype: ENPASSANT, captured: b.squares[b.enpassant], colorMoved: player, piece: b.squares[sq]})
			}
		}
	}

	return m
}

func (b *Board) generateCaptures() []Move {
	var m []Move = []Move{}

	// Get player and opponent colors
	var player Color = b.turn
	var opponent Color = reverseColor(b.turn)

	// Lookup all player pieces and opponent pieces
	var playerPieces u64 = b.colors[player]
	var opponentPieces u64 = b.colors[opponent]

	// get player pawns
	var playerPawns u64 = b.getColorPieces(pawn, player)

	// Mask that specifies which squares are pieces allowed to move
	// Set to opponent pieces since we want only captures
	var allowed u64 = opponentPieces

	var orthogonalUs u64 = b.getColorPieces(rook, player) | b.getColorPieces(queen, player)
	var diagonalUs u64 = b.getColorPieces(bishop, player) | b.getColorPieces(queen, player)
	var orthogonalThem u64 = b.getColorPieces(rook, opponent) | b.getColorPieces(queen, opponent)
	var diagonalThem u64 = b.getColorPieces(bishop, opponent) | b.getColorPieces(queen, opponent)

	// Find all squares controlled by opponent
	var attacks u64 = b.getAllAttacks(opponent, b.occupied, orthogonalThem, diagonalThem)
	// printBitBoard(attacks)

	// STEP 1: Generate king moves, since we can look for checks after
	// Get player king square
	var playerKing Square = Square(bitScanForward(b.getColorPieces(king, player)))
	b.getKingMoves(&m, playerKing, attacks|^allowed, playerPieces, player)

	var checkers u64 = 0

	// 2a: get checks of pawns and knight since we don't need to check for pieces in between
	if player == WHITE {
		checkers |= (whitePawnAttacksSquareLookup[playerKing] & b.pieces[bP])
	} else {
		checkers |= (blackPawnAttacksSquareLookup[playerKing] & b.pieces[wP])
	}
	checkers |= (knightAttacks(playerKing) & b.getColorPieces(knight, opponent))

	// 2b: get bishop/rook/queen attacks and check if there are pieces between them and the king
	var pinPoss u64 = (getBishopAttacks(playerKing, b.colors[opponent]) & diagonalThem)
	pinPoss |= (getRookAttacks(playerKing, b.colors[opponent]) & orthogonalThem)

	var pinned u64 = 0
	var lsb int
	var piecesBetween u64
	for {
		if pinPoss == 0 {
			break
		}
		lsb = popLSB(&pinPoss)
		piecesBetween = squaresBetween[playerKing][lsb] & playerPieces

		if piecesBetween == 0 {
			checkers ^= sToBB[lsb]
		} else if piecesBetween != 0 && (piecesBetween&(piecesBetween-1)) == 0 {
			// only one piece between player and king, since otherwise there is no pin
			pinned ^= piecesBetween
		}
	}

	var numCheckers int = popCount(checkers)
	if numCheckers == 2 {
		// double check so only king moves allowed
		return m
	} else if numCheckers == 1 {
		var checker Square = Square(bitScanForward(checkers))
		var checkerPiece = b.squares[checker]

		var enPassantShift Square = b.enpassant + Square(pawnPushDirection[opponent])

		if checker == enPassantShift && (checkerPiece == wP || checkerPiece == bP) {
			// en passant check capture
			pawns := colorToPawnLookup[opponent][b.enpassant] & playerPawns

			unpinnedPawns := pawns & ^pinned
			for {
				if unpinnedPawns == 0 {
					break
				}
				lsb := Square(popLSB(&unpinnedPawns))
				move := Move{from: lsb, to: b.enpassant, movetype: ENPASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[lsb]}

				b.makeMove(move)
				if b.isCheck(player) {
					b.undo()
				} else {
					b.undo()
					m = append(m, move)
				}
			}

			pinnedPawns := pawns & pinned & line[checker][playerKing]
			if pinnedPawns != 0 {
				sq := Square(bitScanForward(pinnedPawns))
				m = append(m, Move{from: sq, to: checker, movetype: ENPASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[lsb]})
			}

		}

		// get possible captures of the checker
		var possibleCaptures u64 = 0
		if player == WHITE {
			possibleCaptures |= blackPawnAttacksSquareLookup[checker] & b.pieces[wP]
			possibleCaptures |= knightAttacks(checker) & b.pieces[wN]
			possibleCaptures |= getBishopAttacks(checker, b.occupied) & diagonalUs
			possibleCaptures |= getRookAttacks(checker, b.occupied) & orthogonalUs
			possibleCaptures &= ^pinned
		} else {
			possibleCaptures |= whitePawnAttacksSquareLookup[checker] & b.pieces[bP]
			possibleCaptures |= knightAttacks(checker) & b.pieces[bN]
			possibleCaptures |= getBishopAttacks(checker, b.occupied) & diagonalUs
			possibleCaptures |= getRookAttacks(checker, b.occupied) & orthogonalUs
			possibleCaptures &= ^pinned
		}

		var lsb Square
		for {
			if possibleCaptures == 0 {
				break
			}
			lsb = Square(popLSB(&possibleCaptures))

			b.generateMovesFromLocs(&m, lsb, sToBB[checker], player)
		}

		if checkerPiece == bN || checkerPiece == wN {
			// can only capture checker and move king, which has been handled above
			return m
		} else if checkerPiece == bP || checkerPiece == wP {
			// TODO: add enpassant capture of checks
			return m
		}
		return m
	}

	var playerKnights u64 = b.getColorPieces(knight, player)
	b.getKnightMoves(&m, playerKnights, pinned, playerPieces, player, allowed)

	// STEP 3: Calculate pawn moves
	b.getPawnMoves(&m, playerPawns, pinned, b.occupied, opponentPieces, player, allowed)

	// STEP 4: Calculate bishop moves
	b.getBishopMoves(&m, diagonalUs, pinned, playerPieces, opponentPieces, player, allowed)

	// STEP 5: Calculate rook moves
	b.getRookMoves(&m, orthogonalUs, pinned, playerPieces, opponentPieces, player, allowed)

	// STEP 6: Castling
	b.getCastlingMoves(&m, playerKing, attacks, player)

	// STEP 7: En passant
	if b.enpassant != EMPTYSQ {
		pawns := colorToPawnLookup[opponent][b.enpassant] & playerPawns

		unpinnedPawns := pawns & ^pinned
		for {
			if unpinnedPawns == 0 {
				break
			}
			lsb := Square(popLSB(&unpinnedPawns))
			move := Move{from: lsb, to: b.enpassant, movetype: ENPASSANT, captured: b.squares[b.enpassant], colorMoved: player, piece: b.squares[lsb]}

			b.makeMoveNoUpdate(move)
			if b.isCheck(player) {
				b.undoNoUpdate(move)
			} else {
				b.undoNoUpdate(move)
				m = append(m, move)
			}
		}

		pinnedPawns := pawns & pinned & line[b.enpassant][playerKing]
		if pinnedPawns != 0 {
			sq := Square(bitScanForward(pinnedPawns))
			m = append(m, Move{from: sq, to: b.enpassant, movetype: ENPASSANT, captured: b.squares[b.enpassant], colorMoved: player, piece: b.squares[lsb]})
		}
	}

	return m
}
