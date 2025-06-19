package engine

func (b *Board) GetNumPins(o Color, occupied u64, kingLoc Square) int {
	diagonalThem := b.GetColorPieces(ROOK, o) | b.GetColorPieces(QUEEN, o)
	orthogonalThem := b.GetColorPieces(BISHOP, o) | b.GetColorPieces(QUEEN, o)

	var pinPoss u64 = (BishopAttacks(kingLoc, b.colors[o]) & diagonalThem)
	pinPoss |= (RookAttacks(kingLoc, b.colors[o]) & orthogonalThem)

	var pinned int = 0
	var lsb int
	var piecesBetween u64
	for {
		if pinPoss == 0 {
			break
		}
		lsb = PopLSB(&pinPoss)
		piecesBetween = SQUARES_BETWEEN[kingLoc][lsb] & b.pieces[ReverseColor(o)]

		if piecesBetween != 0 && (piecesBetween&(piecesBetween-1)) == 0 {
			// only one piece between player and king, since otherwise there is no pin
			pinned++
		}
	}
	return pinned
}

func (b *Board) GetAllAttacks(o Color, occupied u64, ortho u64, diag u64) u64 {
	// Pre-calculate occupancy without king to avoid multiple XOR operations
	var playerKing = b.GetColorPieces(KING, ReverseColor(o))
	var occupiedNoKing = occupied ^ playerKing

	// Start with king attacks (fastest to compute)
	var opponentKing = Square(BitScanForward(b.GetColorPieces(KING, o)))
	var attackedSquares = KingAttacks(opponentKing)

	// Add pawn attacks (also very fast)
	attackedSquares |= AllPawnAttacks(b.GetColorPieces(PAWN, o), o)

	// Add knight attacks (fast due to lookup table)
	var knights = b.GetColorPieces(KNIGHT, o)
	for knights != 0 {
		attackedSquares |= KnightAttacks(Square(PopLSB(&knights)))
	}

	// Sliding piece attacks (most expensive)
	// Process diagonal pieces (bishops + queens)
	var diagonalPieces = diag
	for diagonalPieces != 0 {
		attackedSquares |= BishopAttacks(Square(PopLSB(&diagonalPieces)), occupiedNoKing)
	}

	// Process orthogonal pieces (rooks + queens)
	var orthogonalPieces = ortho
	for orthogonalPieces != 0 {
		attackedSquares |= RookAttacks(Square(PopLSB(&orthogonalPieces)), occupiedNoKing)
	}

	return attackedSquares
}

func AllPawnAttacks(b u64, c Color) u64 {
	if c == WHITE {
		return ShiftBitboard(b, NE) | ShiftBitboard(b, NW)
	} else {
		return ShiftBitboard(b, SE) | ShiftBitboard(b, SW)
	}
}

func KingAttacks(sq Square) u64 {
	return KING_ATTACKS_LOOKUP[sq]
}

func KnightAttacks(sq Square) u64 {
	return KNIGHT_ATTACKS_LOOKUP[sq]
}

func BishopAttacks(sq Square, bb u64) u64 {
	magicIndex := (bb & BISHOP_MASKS[sq]) * BISHOP_MAGICS[sq]
	return BISHOP_ATTACKS[sq][magicIndex>>BISHOP_SHIFTS[sq]]
}

func RookAttacks(sq Square, bb u64) u64 {
	magicIndex := (bb & ROOK_MASKS[sq]) * ROOK_MAGICS[sq]
	return ROOK_ATTACKS[sq][magicIndex>>ROOK_SHIFTS[sq]]
}

func (b *Board) GenerateMovesFromLocs(m *[]Move, sq Square, locs u64, c Color) {
	var lsb Square
	var bbLSB u64
	var piece Piece
	var movet MoveType
	for locs != 0 {
		lsb = Square(PopLSB(&locs))
		bbLSB = SQUARE_TO_BITBOARD[lsb]
		piece = b.squares[sq]

		if piece == W_P && (bbLSB&RANKS[R8] != 0) {
			movet = PROMOTION
			if b.squares[lsb] != EMPTY {
				movet = CAPTURE_AND_PROMOTION
			}
			// Add all promotion pieces
			*m = append(*m,
				Move{from: sq, to: lsb, piece: W_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: W_N},
				Move{from: sq, to: lsb, piece: W_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: W_R},
				Move{from: sq, to: lsb, piece: W_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: W_Q},
				Move{from: sq, to: lsb, piece: W_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: W_B},
			)
		} else if piece == B_P && (bbLSB&RANKS[R1] != 0) {
			movet = PROMOTION
			if b.squares[lsb] != EMPTY {
				movet = CAPTURE_AND_PROMOTION
			}
			// Add all promotion pieces
			*m = append(*m,
				Move{from: sq, to: lsb, piece: B_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: B_N},
				Move{from: sq, to: lsb, piece: B_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: B_R},
				Move{from: sq, to: lsb, piece: B_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: B_Q},
				Move{from: sq, to: lsb, piece: B_P, colorMoved: c, captured: b.squares[lsb], movetype: movet, promote: B_B},
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

func (b *Board) KingMoves(m *[]Move, sq Square, attacks u64, player u64, c Color) {
	var kingMoves u64 = KingAttacks(sq) & ^attacks & ^player
	b.GenerateMovesFromLocs(m, sq, kingMoves, c)

}

func (b *Board) KnightMoves(m *[]Move, knights u64, pinned u64, player u64, c Color, allowed u64) {
	// since a pinned knight means no possible moves, can just mask by non-pinned
	knights &= ^pinned

	var sq Square
	for {
		if knights == 0 {
			break
		}

		sq = Square(PopLSB(&knights))

		var knightMoves u64 = KnightAttacks(Square(sq)) & ^player & allowed
		b.GenerateMovesFromLocs(m, sq, knightMoves, c)
	}
}

func (b *Board) PawnMoves(m *[]Move, pawns u64, pinned u64, occupied u64, opponents u64, c Color, allowed u64) {
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
			lsb = Square(PopLSB(&pinnedPawns))
			bbSQ = SQUARE_TO_BITBOARD[lsb]

			kingLoc = Square(BitScanForward(b.GetColorPieces(KING, c)))

			var pawnMoves u64 = (ShiftBitboard(bbSQ, PAWN_PUSH_DIRECTION[c]) & ^occupied) & allowed & LINE[kingLoc][lsb]
			if pawnMoves != 0 {
				if bbSQ&RANKS[STARTING_RANK[c]] != 0 {
					// two move pushes allowed
					pawnMoves |= (ShiftBitboard(bbSQ, PAWN_PUSH_DIRECTION[c]*2) & ^occupied) & allowed & LINE[kingLoc][lsb]
				}
			}
			pawnMoves |= COLOR_TO_PAWN_LOOKUP[c][lsb] & opponents & allowed & LINE[kingLoc][lsb]
			b.GenerateMovesFromLocs(m, lsb, pawnMoves, c)
		}
	}

	var unpinnedPawns u64 = pawns & ^pinned
	var sq Square
	for {
		if unpinnedPawns == 0 {
			break
		}
		sq = Square(PopLSB(&unpinnedPawns))
		bbSQ = SQUARE_TO_BITBOARD[sq]
		var pawnMoves u64 = (ShiftBitboard(bbSQ, PAWN_PUSH_DIRECTION[c]) & ^occupied) & allowed
		if b.squares[sq+Square(PAWN_PUSH_DIRECTION[c])] == EMPTY {
			if bbSQ&RANKS[STARTING_RANK[c]] != 0 {
				// two move pushes allowed
				pawnMoves |= (ShiftBitboard(bbSQ, PAWN_PUSH_DIRECTION[c]*2) & ^occupied) & allowed
			}
		}
		// add captures
		if c == WHITE {
			pawnMoves |= (WHITE_PAWN_ATTACKS_LOOKUP[sq] & opponents) & allowed
		} else {
			pawnMoves |= (BLACK_PAWN_ATTACKS_LOOKUP[sq] & opponents) & allowed
		}

		b.GenerateMovesFromLocs(m, sq, pawnMoves, c)
	}
}

func (b *Board) BishopMoves(m *[]Move, bishops u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedBishops u64 = bishops & pinned
	var kingLoc Square

	if pinnedBishops != 0 {
		var lsb Square
		for {
			if pinnedBishops == 0 {
				break
			}
			lsb = Square(PopLSB(&pinnedBishops))

			kingLoc = Square(BitScanForward(b.GetColorPieces(KING, c)))

			var possible u64 = BishopAttacks(lsb, b.occupied) & LINE[kingLoc][lsb] & allowed & ^player
			b.GenerateMovesFromLocs(m, lsb, possible, c)
		}
	}

	var unpinnedBishops u64 = bishops & ^pinned
	var sq Square
	for {
		if unpinnedBishops == 0 {
			break
		}
		sq = Square(PopLSB(&unpinnedBishops))

		var bishopMoves u64 = BishopAttacks(sq, player|opponents) & ^player & allowed

		b.GenerateMovesFromLocs(m, sq, bishopMoves, c)
	}
}

func (b *Board) RookMoves(m *[]Move, rooks u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedRooks u64 = rooks & pinned
	var kingLoc Square
	if pinnedRooks != 0 {
		var lsb Square
		for {
			if pinnedRooks == 0 {
				break
			}
			lsb = Square(PopLSB(&pinnedRooks))

			kingLoc = Square(BitScanForward(b.GetColorPieces(KING, c)))

			var possible u64 = RookAttacks(lsb, b.occupied) & LINE[kingLoc][lsb] & allowed & ^player
			b.GenerateMovesFromLocs(m, lsb, possible, c)
		}
	}

	var unpinnedRooks u64 = rooks & ^pinned
	var sq Square
	for {
		if unpinnedRooks == 0 {
			break
		}
		sq = Square(PopLSB(&unpinnedRooks))

		var rookMoves u64 = RookAttacks(sq, player|opponents) & ^player & allowed

		b.GenerateMovesFromLocs(m, sq, rookMoves, c)
	}
}

func (b *Board) QueenMoves(m *[]Move, queens u64, pinned u64, player u64, opponents u64, c Color, allowed u64) {
	var pinnedQueens u64 = queens & pinned
	var kingLoc Square
	if pinnedQueens != 0 {
		var lsb Square
		for {
			if pinnedQueens == 0 {
				break
			}
			lsb = Square(PopLSB(&pinnedQueens))
			kingLoc = Square(BitScanForward(b.GetColorPieces(KING, c)))

			var possible u64 = (BishopAttacks(lsb, b.occupied) | RookAttacks(lsb, b.occupied)) & LINE[kingLoc][lsb] & allowed & ^player

			b.GenerateMovesFromLocs(m, lsb, possible, c)
		}
	}

	var unpinnedQueens u64 = queens & ^pinned
	var sq Square
	for {
		if unpinnedQueens == 0 {
			break
		}
		sq = Square(PopLSB(&unpinnedQueens))

		var rookMoves u64 = RookAttacks(sq, player|opponents) & ^player & allowed
		var bishopMoves u64 = BishopAttacks(sq, player|opponents) & ^player & allowed

		b.GenerateMovesFromLocs(m, sq, rookMoves|bishopMoves, c)
	}
}

func (b *Board) CastlingMoves(m *[]Move, pKing Square, attacks u64, c Color) {
	var castlingKingsidePossible bool = false
	var castlingQueensidePossible bool = false

	var allowed u64 = ^attacks & b.empty

	if c == WHITE {
		castlingKingsidePossible = (SQUARE_TO_BITBOARD[F1]&allowed != 0) && (SQUARE_TO_BITBOARD[G1]&allowed != 0)
		castlingQueensidePossible = (SQUARE_TO_BITBOARD[B1]&b.empty != 0) && (SQUARE_TO_BITBOARD[C1]&allowed != 0) && (SQUARE_TO_BITBOARD[D1]&allowed != 0)
		if b.OO && castlingKingsidePossible && b.squares[H1] == W_R {
			*m = append(*m, Move{from: E1, to: G1, piece: W_K, movetype: K_CASTLE, colorMoved: WHITE})
		}
		if b.OOO && castlingQueensidePossible && b.squares[A1] == W_R {
			*m = append(*m, Move{from: E1, to: C1, piece: W_K, movetype: Q_CASTLE, colorMoved: WHITE})
		}
	} else {
		castlingKingsidePossible = (SQUARE_TO_BITBOARD[F8]&allowed != 0) && (SQUARE_TO_BITBOARD[G8]&allowed != 0)
		castlingQueensidePossible = (SQUARE_TO_BITBOARD[B8]&b.empty != 0) && (SQUARE_TO_BITBOARD[C8]&allowed != 0) && (SQUARE_TO_BITBOARD[D8]&allowed != 0)
		if b.oo && castlingKingsidePossible && b.squares[H8] == B_R {
			*m = append(*m, Move{from: E8, to: G8, piece: B_K, movetype: K_CASTLE, colorMoved: BLACK})
		}

		if b.ooo && castlingQueensidePossible && b.squares[A8] == B_R {
			*m = append(*m, Move{from: E8, to: C8, piece: B_K, movetype: Q_CASTLE, colorMoved: BLACK})
		}
	}
}

func (b *Board) PinnedPieces(kingSquare Square, c Color) (u64, u64) {
	opponent := ReverseColor(c)
	var pinned, checkers u64 = 0, 0

	// Get opponent's sliding pieces
	orthogonalThem := b.GetColorPieces(ROOK, opponent) | b.GetColorPieces(QUEEN, opponent)
	diagonalThem := b.GetColorPieces(BISHOP, opponent) | b.GetColorPieces(QUEEN, opponent)

	// Check for direct attacks first (faster than pin detection)
	if c == WHITE {
		checkers = WHITE_PAWN_ATTACKS_LOOKUP[kingSquare] & b.pieces[B_P]
	} else {
		checkers = BLACK_PAWN_ATTACKS_LOOKUP[kingSquare] & b.pieces[W_P]
	}
	checkers |= KnightAttacks(kingSquare) & b.GetColorPieces(KNIGHT, opponent)

	// Get potential pinners on diagonals
	potentialPinners := BishopAttacks(kingSquare, b.colors[opponent]) & diagonalThem
	var pinner u64
	for potentialPinners != 0 {
		pinner = SQUARE_TO_BITBOARD[Square(PopLSB(&potentialPinners))]
		between := SQUARES_BETWEEN[kingSquare][BitScanForward(pinner)] & b.occupied
		if PopCount(between) == 1 && (between&b.colors[c]) != 0 {
			pinned |= between
		} else if PopCount(between) == 0 {
			checkers |= pinner
		}
	}

	// Get potential pinners on ranks/files
	potentialPinners = RookAttacks(kingSquare, b.colors[opponent]) & orthogonalThem
	for potentialPinners != 0 {
		pinner = SQUARE_TO_BITBOARD[Square(PopLSB(&potentialPinners))]
		between := SQUARES_BETWEEN[kingSquare][BitScanForward(pinner)] & b.occupied
		if PopCount(between) == 1 && (between&b.colors[c]) != 0 {
			pinned |= between
		} else if PopCount(between) == 0 {
			checkers |= pinner
		}
	}

	return pinned, checkers
}

func (b *Board) GenerateLegalMoves() []Move {
	// Setup move list with pre-allocated capacity
	m := make([]Move, 0, 48)

	player := b.turn
	opponent := ReverseColor(player)

	// Get player pieces and opponent pieces
	playerPieces := b.colors[player]
	opponentPieces := b.colors[opponent]

	// Get player king and calculate pins/checkers
	playerKing := Square(BitScanForward(b.GetColorPieces(KING, player)))
	pinned, checkers := b.PinnedPieces(playerKing, player)

	// Get attack map from opponent
	orthogonalThem := b.GetColorPieces(ROOK, opponent) | b.GetColorPieces(QUEEN, opponent)
	diagonalThem := b.GetColorPieces(BISHOP, opponent) | b.GetColorPieces(QUEEN, opponent)
	attacks := b.GetAllAttacks(opponent, b.occupied, orthogonalThem, diagonalThem)

	// Generate king moves first
	b.KingMoves(&m, playerKing, attacks, playerPieces, player)

	numCheckers := PopCount(checkers)
	if numCheckers >= 2 {
		return m // Only king moves in double check
	}

	// Get allowed squares (all squares in single check, only blocking squares in check)
	allowed := u64(0xFFFFFFFFFFFFFFFF)
	if numCheckers == 1 {
		checkerSquare := Square(BitScanForward(checkers))
		checkerPiece := b.squares[checkerSquare]
		allowed = SQUARES_BETWEEN[playerKing][checkerSquare] | checkers

		// Handle en passant capture of checking pawn
		if b.enPassant != EMPTY_SQ {
			enPassantShift := b.enPassant + Square(PAWN_PUSH_DIRECTION[opponent])
			if checkerSquare == enPassantShift && (checkerPiece == W_P || checkerPiece == B_P) {
				pawns := COLOR_TO_PAWN_LOOKUP[opponent][b.enPassant] & b.GetColorPieces(PAWN, player)

				// Handle unpinned pawns
				unpinnedPawns := pawns & ^pinned
				for unpinnedPawns != 0 {
					lsb := Square(PopLSB(&unpinnedPawns))
					move := Move{from: lsb, to: b.enPassant, movetype: EN_PASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[lsb]}

					b.MakeMove(move)
					if !b.IsCheck(player) {
						b.Undo()
						m = append(m, move)
					} else {
						b.Undo()
					}
				}

				// Handle pinned pawns
				pinnedPawns := pawns & pinned & LINE[checkerSquare][playerKing]
				if pinnedPawns != 0 {
					sq := Square(BitScanForward(pinnedPawns))
					m = append(m, Move{from: sq, to: b.enPassant, movetype: EN_PASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[sq]})
				}
			}
		}
	}

	// Generate other piece moves
	b.KnightMoves(&m, b.GetColorPieces(KNIGHT, player), pinned, playerPieces, player, allowed)
	b.PawnMoves(&m, b.GetColorPieces(PAWN, player), pinned, b.occupied, opponentPieces, player, allowed)
	b.BishopMoves(&m, b.GetColorPieces(BISHOP, player)|b.GetColorPieces(QUEEN, player), pinned, playerPieces, opponentPieces, player, allowed)
	b.RookMoves(&m, b.GetColorPieces(ROOK, player)|b.GetColorPieces(QUEEN, player), pinned, playerPieces, opponentPieces, player, allowed)

	// Only try castling if not in check
	if numCheckers == 0 {
		b.CastlingMoves(&m, playerKing, attacks, player)

		// Handle en passant when not in check
		if b.enPassant != EMPTY_SQ {
			pawns := COLOR_TO_PAWN_LOOKUP[opponent][b.enPassant] & b.GetColorPieces(PAWN, player)

			// Handle unpinned pawns
			unpinnedPawns := pawns & ^pinned
			for unpinnedPawns != 0 {
				lsb := Square(PopLSB(&unpinnedPawns))
				move := Move{from: lsb, to: b.enPassant, movetype: EN_PASSANT, captured: b.squares[b.enPassant], colorMoved: player, piece: b.squares[lsb]}

				b.MakeMoveNoUpdate(move)
				if !b.IsCheck(player) {
					b.UndoNoUpdate(move)
					m = append(m, move)
				} else {
					b.UndoNoUpdate(move)
				}
			}

			// Handle pinned pawns
			pinnedPawns := pawns & pinned & LINE[b.enPassant][playerKing]
			if pinnedPawns != 0 {
				sq := Square(BitScanForward(pinnedPawns))
				m = append(m, Move{from: sq, to: b.enPassant, movetype: EN_PASSANT, captured: b.squares[b.enPassant], colorMoved: player, piece: b.squares[sq]})
			}
		}
	}

	return m
}

func (b *Board) GenerateCaptures() []Move {
	var m []Move = []Move{}

	// Get player and opponent colors
	var player Color = b.turn
	var opponent Color = ReverseColor(b.turn)

	// Lookup all player pieces and opponent pieces
	var playerPieces u64 = b.colors[player]
	var opponentPieces u64 = b.colors[opponent]

	// get player pawns
	var playerPawns u64 = b.GetColorPieces(PAWN, player)

	// Mask that specifies which squares are pieces allowed to move
	// Set to opponent pieces since we want only captures
	var allowed u64 = opponentPieces

	var orthogonalUs u64 = b.GetColorPieces(ROOK, player) | b.GetColorPieces(QUEEN, player)
	var diagonalUs u64 = b.GetColorPieces(BISHOP, player) | b.GetColorPieces(QUEEN, player)
	var orthogonalThem u64 = b.GetColorPieces(ROOK, opponent) | b.GetColorPieces(QUEEN, opponent)
	var diagonalThem u64 = b.GetColorPieces(BISHOP, opponent) | b.GetColorPieces(QUEEN, opponent)

	// Find all squares controlled by opponent
	var attacks u64 = b.GetAllAttacks(opponent, b.occupied, orthogonalThem, diagonalThem)
	// printBitBoard(attacks)

	// STEP 1: Generate king moves, since we can look for checks after
	// Get player king square
	var playerKing Square = Square(BitScanForward(b.GetColorPieces(KING, player)))
	b.KingMoves(&m, playerKing, attacks|^allowed, playerPieces, player)

	var checkers u64 = 0

	// 2a: get checks of pawns and knight since we don't need to check for pieces in between
	if player == WHITE {
		checkers |= (WHITE_PAWN_ATTACKS_LOOKUP[playerKing] & b.pieces[B_P])
	} else {
		checkers |= (BLACK_PAWN_ATTACKS_LOOKUP[playerKing] & b.pieces[W_P])
	}
	checkers |= (KnightAttacks(playerKing) & b.GetColorPieces(KNIGHT, opponent))

	// 2b: get bishop/rook/queen attacks and check if there are pieces between them and the king
	var pinPoss u64 = (BishopAttacks(playerKing, b.colors[opponent]) & diagonalThem)
	pinPoss |= (RookAttacks(playerKing, b.colors[opponent]) & orthogonalThem)

	var pinned u64 = 0
	var lsb int
	var piecesBetween u64
	for {
		if pinPoss == 0 {
			break
		}
		lsb = PopLSB(&pinPoss)
		piecesBetween = SQUARES_BETWEEN[playerKing][lsb] & playerPieces

		if piecesBetween == 0 {
			checkers ^= SQUARE_TO_BITBOARD[lsb]
		} else if piecesBetween != 0 && (piecesBetween&(piecesBetween-1)) == 0 {
			// only one piece between player and king, since otherwise there is no pin
			pinned ^= piecesBetween
		}
	}

	var numCheckers int = PopCount(checkers)
	if numCheckers == 2 {
		// double check so only king moves allowed
		return m
	} else if numCheckers == 1 {
		var checker Square = Square(BitScanForward(checkers))
		var checkerPiece = b.squares[checker]

		var enPassantShift Square = b.enPassant + Square(PAWN_PUSH_DIRECTION[opponent])

		if checker == enPassantShift && (checkerPiece == W_P || checkerPiece == B_P) {
			// en passant check capture
			pawns := COLOR_TO_PAWN_LOOKUP[opponent][b.enPassant] & playerPawns

			unpinnedPawns := pawns & ^pinned
			for {
				if unpinnedPawns == 0 {
					break
				}
				lsb := Square(PopLSB(&unpinnedPawns))
				move := Move{from: lsb, to: b.enPassant, movetype: EN_PASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[lsb]}

				b.MakeMove(move)
				if b.IsCheck(player) {
					b.Undo()
				} else {
					b.Undo()
					m = append(m, move)
				}
			}

			pinnedPawns := pawns & pinned & LINE[checker][playerKing]
			if pinnedPawns != 0 {
				sq := Square(BitScanForward(pinnedPawns))
				m = append(m, Move{from: sq, to: checker, movetype: EN_PASSANT, captured: checkerPiece, colorMoved: player, piece: b.squares[lsb]})
			}

		}

		// get possible captures of the checker
		var possibleCaptures u64 = 0
		if player == WHITE {
			possibleCaptures |= BLACK_PAWN_ATTACKS_LOOKUP[checker] & b.pieces[W_P]
			possibleCaptures |= KnightAttacks(checker) & b.pieces[W_N]
			possibleCaptures |= BishopAttacks(checker, b.occupied) & diagonalUs
			possibleCaptures |= RookAttacks(checker, b.occupied) & orthogonalUs
			possibleCaptures &= ^pinned
		} else {
			possibleCaptures |= WHITE_PAWN_ATTACKS_LOOKUP[checker] & b.pieces[B_P]
			possibleCaptures |= KnightAttacks(checker) & b.pieces[B_N]
			possibleCaptures |= BishopAttacks(checker, b.occupied) & diagonalUs
			possibleCaptures |= RookAttacks(checker, b.occupied) & orthogonalUs
			possibleCaptures &= ^pinned
		}

		var lsb Square
		for {
			if possibleCaptures == 0 {
				break
			}
			lsb = Square(PopLSB(&possibleCaptures))

			b.GenerateMovesFromLocs(&m, lsb, SQUARE_TO_BITBOARD[checker], player)
		}

		if checkerPiece == B_N || checkerPiece == W_N {
			// can only capture checker and move king, which has been handled above
			return m
		} else if checkerPiece == B_P || checkerPiece == W_P {
			// TODO: add enpassant capture of checks
			return m
		}
		return m
	}

	var playerKnights u64 = b.GetColorPieces(KNIGHT, player)
	b.KnightMoves(&m, playerKnights, pinned, playerPieces, player, allowed)

	// STEP 3: Calculate pawn moves
	b.PawnMoves(&m, playerPawns, pinned, b.occupied, opponentPieces, player, allowed)

	// STEP 4: Calculate bishop moves
	b.BishopMoves(&m, diagonalUs, pinned, playerPieces, opponentPieces, player, allowed)

	// STEP 5: Calculate rook moves
	b.RookMoves(&m, orthogonalUs, pinned, playerPieces, opponentPieces, player, allowed)

	// STEP 6: Castling
	b.CastlingMoves(&m, playerKing, attacks, player)

	// STEP 7: En passant
	if b.enPassant != EMPTY_SQ {
		pawns := COLOR_TO_PAWN_LOOKUP[opponent][b.enPassant] & playerPawns

		unpinnedPawns := pawns & ^pinned
		for {
			if unpinnedPawns == 0 {
				break
			}
			lsb := Square(PopLSB(&unpinnedPawns))
			move := Move{from: lsb, to: b.enPassant, movetype: EN_PASSANT, captured: b.squares[b.enPassant], colorMoved: player, piece: b.squares[lsb]}

			b.MakeMoveNoUpdate(move)
			if b.IsCheck(player) {
				b.UndoNoUpdate(move)
			} else {
				b.UndoNoUpdate(move)
				m = append(m, move)
			}
		}

		pinnedPawns := pawns & pinned & LINE[b.enPassant][playerKing]
		if pinnedPawns != 0 {
			sq := Square(BitScanForward(pinnedPawns))
			m = append(m, Move{from: sq, to: b.enPassant, movetype: EN_PASSANT, captured: b.squares[b.enPassant], colorMoved: player, piece: b.squares[lsb]})
		}
	}

	return m
}
