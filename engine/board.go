package engine

import (
	"fmt"
	"strconv"
	"strings"
)

type u64 uint64

type Board struct {
	pieces       [14]u64   // Stores bitboards of all white and black pieces
	squares      [64]Piece // Stores all 64 squares (not used for move generation)
	colors       [2]u64    // Stores bitboards of both colors
	occupied     u64       // Bits are set when pieces are there
	empty        u64       // Bits are clear when pieces are there
	turn         Color     // Side to move
	enPassant    Square    // En passant square. If not possible, stores EMPTY
	OO           bool      // If kingside castling available for White
	OOO          bool      // If queenside castling is available for White
	oo           bool      // If kingside castling is available for Black
	ooo          bool      // If queenside castling is available for Black
	history      []prev    // Stores history for board
	zobrist      u64       // Zobrist hash (TODO)
	plyCnt       int       // Stores number of half moves played
	moveCount    int       // Stores which move currently we are at
	whiteCastled bool      // Stores whether white has previously castled
	blackCastled bool      // Stores whether black has previously castled
}

type prev struct {
	move         Move   // Stores previous move made
	OO           bool   // White Kingside castling history
	OOO          bool   // White Queenside castling history
	oo           bool   // Black Kingside castling history
	ooo          bool   // Black Queenside castling history
	enPassant    Square // En passant square history
	hash         u64    // Zobrist hash of prev position
	whiteCastled bool   // Stores whether white has previously castled
	blackCastled bool   // Stores whether black has previously castled
}

func NewBoard() *Board {
	b := Board{}
	b.turn = WHITE
	b.enPassant = EMPTY_SQ
	b.OO, b.OOO, b.oo, b.ooo = true, true, true, true
	b.InitStartPos()
	return &b
}

func (b *Board) InitStartPos() {
	b.zobrist = 0
	b.squares = [64]Piece{
		W_R, W_N, W_B, W_Q, W_K, W_B, W_N, W_R,
		W_P, W_P, W_P, W_P, W_P, W_P, W_P, W_P,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		B_P, B_P, B_P, B_P, B_P, B_P, B_P, B_P,
		B_R, B_N, B_B, B_Q, B_K, B_B, B_N, B_R,
	}

	for i := 0; i < 64; i++ {
		if i < 16 {
			b.putPiece(b.squares[i], Square(i), WHITE)
		} else if i >= 48 {
			b.putPiece(b.squares[i], Square(i), BLACK)
		} else {
			b.empty ^= SQUARE_TO_BITBOARD[i]
		}
	}
	b.turn = WHITE
	b.enPassant = EMPTY_SQ
	b.OO = true
	b.OOO = true
	b.oo = true
	b.ooo = true
	b.zobrist ^= TURN_HASH
}

func (b *Board) InitFEN(fen string) {
	b.zobrist = 0
	for s := A1; s <= H8; s++ {
		b.putPiece(EMPTY, s, WHITE)
	}

	attrs := strings.Split(fen, " ")

	pieces := attrs[0]
	ranksPieces := strings.Split(pieces, "/")

	for i, stRank := range ranksPieces {
		rank := 8 * (7 - i)
		sq := Square(rank)
		for _, piece := range stRank {
			if strings.Contains("PBNRQKpbnrqk", string(piece)) {
				piece := stringToPieceMap[string(piece)]
				b.putPiece(piece, sq, piece.GetColor())
				sq++
			} else {
				num, _ := strconv.Atoi(string(piece))
				for k := 0; k < num; k++ {
					newSq := Square(int(sq) + int(k))
					b.putPiece(EMPTY, newSq, WHITE)
				}
				sq += Square(num)
			}
		}
	}

	turn := attrs[1]
	if turn == "w" {
		b.turn = WHITE
	} else {
		b.turn = BLACK
	}

	castling := attrs[2]
	if strings.Contains(castling, "K") {
		b.OO = true
	}
	if strings.Contains(castling, "Q") {
		b.OOO = true
	}
	if strings.Contains(castling, "k") {
		b.oo = true
	}
	if strings.Contains(castling, "q") {
		b.ooo = true
	}

	enpassant := attrs[3]
	if enpassant == "-" {
		b.enPassant = EMPTY_SQ
	} else {
		b.enPassant = STRING_TO_SQUARE_MAP[enpassant]
	}

	halfMoveClock := attrs[4]
	b.plyCnt, _ = strconv.Atoi(halfMoveClock)

	moveCount := attrs[5]
	b.moveCount, _ = strconv.Atoi(moveCount)
	b.zobrist ^= TURN_HASH

	b.plyCnt = b.moveCount * 2
}

func (b *Board) GetColorPieces(p PieceType, c Color) u64 {
	return b.pieces[int(p)+int(c)*COLOR_INDEX_OFFSET]
}

func (b *Board) putPiece(p Piece, s Square, c Color) {
	b.squares[s] = p

	if p != EMPTY {
		var square u64 = SQUARE_TO_BITBOARD[s]
		b.colors[c] |= square
		b.pieces[p] |= square
		b.occupied |= square
		b.empty &= ^square
		b.zobrist ^= ZOBRIST_TABLE[p][s]
	} else {
		var square u64 = SQUARE_TO_BITBOARD[s]
		b.empty |= square
		b.occupied &= ^square
		b.zobrist ^= ZOBRIST_TABLE[p][s]
	}
}

func (b *Board) movePiece(p Piece, mvfrom Square, mvto Square, c Color) {
	var fromTo u64 = SQUARE_TO_BITBOARD[mvfrom] ^ SQUARE_TO_BITBOARD[mvto]

	// Update all bitboards in one pass
	b.pieces[p] ^= fromTo
	b.colors[c] ^= fromTo
	b.occupied ^= fromTo
	b.empty ^= fromTo

	// Update mailbox and zobrist in one pass
	b.squares[mvto] = p
	b.squares[mvfrom] = EMPTY
	b.zobrist ^= ZOBRIST_TABLE[p][mvfrom] ^ ZOBRIST_TABLE[p][mvto] ^ ZOBRIST_TABLE[EMPTY][mvto] ^ ZOBRIST_TABLE[EMPTY][mvfrom]
}

func (b *Board) capturePiece(p Piece, captured Piece, mvfrom Square, mvto Square, c Color) {
	var from u64 = SQUARE_TO_BITBOARD[mvfrom]
	var to u64 = SQUARE_TO_BITBOARD[mvto]

	// Update moving piece
	b.pieces[p] ^= from ^ to
	b.colors[c] ^= from ^ to

	// Remove captured piece
	b.pieces[captured] ^= to
	b.colors[ReverseColor(c)] ^= to

	// Update occupancy
	b.occupied ^= from
	b.empty ^= from

	// Update mailbox and zobrist in one pass
	b.squares[mvto] = p
	b.squares[mvfrom] = EMPTY
	b.zobrist ^= ZOBRIST_TABLE[p][mvfrom] ^ ZOBRIST_TABLE[p][mvto] ^ ZOBRIST_TABLE[captured][mvto] ^ ZOBRIST_TABLE[EMPTY][mvfrom]
}

func (b *Board) replacePiece(p Piece, q Piece, sq Square) {
	var square u64 = SQUARE_TO_BITBOARD[sq]
	b.pieces[p] ^= square
	b.pieces[q] ^= square

	// update mailbox representation
	b.squares[sq] = q

	b.zobrist ^= ZOBRIST_TABLE[p][sq] ^ ZOBRIST_TABLE[q][sq]
}

func (b *Board) removePiece(p Piece, sq Square, c Color) {
	var square u64 = SQUARE_TO_BITBOARD[sq]
	b.pieces[p] ^= square
	b.occupied ^= square
	b.empty ^= square
	b.colors[c] ^= square
	b.squares[sq] = EMPTY
	b.zobrist ^= ZOBRIST_TABLE[p][sq]
}

func (b *Board) MakeMoveFromUCI(uci string) {
	b.MakeMove(FromUCI(uci, b))
}

func (b *Board) MakeMoveNoUpdate(mv Move) {
	switch mv.movetype {
	case QUIET:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
	case CAPTURE:
		b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
	case PROMOTION:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
		b.replacePiece(mv.piece, mv.promote, mv.to)
	case CAPTURE_AND_PROMOTION:
		b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
		b.replacePiece(mv.piece, mv.promote, mv.to)
	case K_CASTLE:
		if mv.colorMoved == WHITE {
			b.movePiece(W_K, E1, G1, WHITE)
			b.movePiece(W_R, H1, F1, WHITE)
		} else {
			b.movePiece(B_K, E8, G8, BLACK)
			b.movePiece(B_R, H8, F8, BLACK)
		}
	case Q_CASTLE:
		if mv.colorMoved == WHITE {
			b.movePiece(W_K, E1, C1, WHITE)
			b.movePiece(W_R, A1, D1, WHITE)
		} else {
			b.movePiece(B_K, E8, C8, BLACK)
			b.movePiece(B_R, A8, D8, BLACK)
		}
	case EN_PASSANT:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
		if mv.colorMoved == WHITE {
			b.removePiece(B_P, mv.to.GoDirection(SOUTH), BLACK)
		} else {
			b.removePiece(W_P, mv.to.GoDirection(NORTH), WHITE)
		}
	}
}

func (b *Board) MakeMove(mv Move) {
	var entry prev = prev{move: mv, OO: b.OO, OOO: b.OOO, oo: b.oo, ooo: b.ooo, enPassant: b.enPassant, hash: b.zobrist, whiteCastled: b.whiteCastled, blackCastled: b.blackCastled}
	b.history = append(b.history, entry)

	if !mv.null {
		switch mv.movetype {
		case QUIET:
			b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
		case CAPTURE:
			b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
		case PROMOTION:
			b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
			b.replacePiece(mv.piece, mv.promote, mv.to)
		case CAPTURE_AND_PROMOTION:
			b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
			b.replacePiece(mv.piece, mv.promote, mv.to)
		case K_CASTLE:
			if mv.colorMoved == WHITE {
				b.movePiece(W_K, E1, G1, WHITE)
				b.movePiece(W_R, H1, F1, WHITE)
				b.whiteCastled = true
			} else {
				b.movePiece(B_K, E8, G8, BLACK)
				b.movePiece(B_R, H8, F8, BLACK)
				b.blackCastled = true
			}
		case Q_CASTLE:
			if mv.colorMoved == WHITE {
				b.movePiece(W_K, E1, C1, WHITE)
				b.movePiece(W_R, A1, D1, WHITE)
				b.whiteCastled = true
			} else {
				b.movePiece(B_K, E8, C8, BLACK)
				b.movePiece(B_R, A8, D8, BLACK)
				b.blackCastled = true
			}
		case EN_PASSANT:
			b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
			if mv.colorMoved == WHITE {
				b.removePiece(B_P, mv.to.GoDirection(SOUTH), BLACK)
			} else {
				b.removePiece(W_P, mv.to.GoDirection(NORTH), WHITE)
			}
		}
	}

	b.turn = ReverseColor(b.turn)

	// update castling rights
	if mv.piece == W_K {
		b.OO = false
		b.OOO = false
	}
	if mv.piece == B_K {
		b.oo = false
		b.ooo = false
	}
	if mv.piece == W_R {
		if mv.from == A1 {
			b.OOO = false
		}
		if mv.from == H1 {
			b.OO = false
		}
	}
	if mv.piece == B_R {
		if mv.from == A8 {
			b.ooo = false
		}
		if mv.from == H8 {
			b.oo = false
		}
	}

	// update en passant square
	dist := Direction(int(mv.to - mv.from))
	if dist == 2*NORTH && mv.piece == W_P {
		b.enPassant = mv.from + Square(NORTH)
	} else if dist == 2*SOUTH && mv.piece == B_P {
		b.enPassant = mv.from + Square(SOUTH)
	} else {
		b.enPassant = EMPTY_SQ
	}

	b.plyCnt++
	b.zobrist ^= TURN_HASH
}

func (b *Board) UndoNoUpdate(prevMove Move) {
	switch prevMove.movetype {
	case QUIET:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
	case CAPTURE:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
		b.putPiece(prevMove.captured, prevMove.to, ReverseColor(prevMove.colorMoved))
	case PROMOTION:
		b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
		b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
	case CAPTURE_AND_PROMOTION:
		b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
		b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
		b.putPiece(prevMove.captured, prevMove.to, ReverseColor(prevMove.colorMoved))
	case K_CASTLE:
		if prevMove.colorMoved == WHITE {
			b.movePiece(W_K, G1, E1, WHITE)
			b.movePiece(W_R, F1, H1, WHITE)
		} else {
			b.movePiece(B_K, G8, E8, BLACK)
			b.movePiece(B_R, F8, H8, BLACK)
		}
	case Q_CASTLE:
		if prevMove.colorMoved == WHITE {
			b.movePiece(W_K, C1, E1, WHITE)
			b.movePiece(W_R, D1, A1, WHITE)
		} else {
			b.movePiece(B_K, C8, E8, BLACK)
			b.movePiece(B_R, D8, A8, BLACK)
		}
	case EN_PASSANT:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
		if prevMove.colorMoved == WHITE {
			b.putPiece(B_P, prevMove.to.GoDirection(SOUTH), BLACK)
		} else {
			b.putPiece(W_P, prevMove.to.GoDirection(NORTH), WHITE)
		}
	}
}

func (b *Board) Undo() {
	prevEntry := b.history[len(b.history)-1]
	prevMove := prevEntry.move

	if !prevMove.null {
		switch prevMove.movetype {
		case QUIET:
			b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
		case CAPTURE:
			b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
			b.putPiece(prevMove.captured, prevMove.to, ReverseColor(prevMove.colorMoved))
		case PROMOTION:
			b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
			b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
		case CAPTURE_AND_PROMOTION:
			b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
			b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
			b.putPiece(prevMove.captured, prevMove.to, ReverseColor(prevMove.colorMoved))
		case K_CASTLE:
			if prevMove.colorMoved == WHITE {
				b.movePiece(W_K, G1, E1, WHITE)
				b.movePiece(W_R, F1, H1, WHITE)
			} else {
				b.movePiece(B_K, G8, E8, BLACK)
				b.movePiece(B_R, F8, H8, BLACK)
			}
		case Q_CASTLE:
			if prevMove.colorMoved == WHITE {
				b.movePiece(W_K, C1, E1, WHITE)
				b.movePiece(W_R, D1, A1, WHITE)
			} else {
				b.movePiece(B_K, C8, E8, BLACK)
				b.movePiece(B_R, D8, A8, BLACK)
			}
		case EN_PASSANT:
			b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
			if prevMove.colorMoved == WHITE {
				b.putPiece(B_P, prevMove.to.GoDirection(SOUTH), BLACK)
			} else {
				b.putPiece(W_P, prevMove.to.GoDirection(NORTH), WHITE)
			}
		}
	}

	b.OO = prevEntry.OO
	b.OOO = prevEntry.OOO
	b.oo = prevEntry.oo
	b.ooo = prevEntry.ooo
	b.enPassant = prevEntry.enPassant
	b.zobrist = prevEntry.hash
	b.turn = ReverseColor(b.turn)
	b.whiteCastled = prevEntry.whiteCastled
	b.blackCastled = prevEntry.blackCastled

	b.history = b.history[:len(b.history)-1]
	b.plyCnt--
}

func (b *Board) MakeNullMove() {
	b.MakeMove(Move{null: true})
}

func (b *Board) UndoNullMove() {
	b.Undo()
}

func (b *Board) IsCheck(c Color) bool {
	checkers := u64(0)
	playerKing := Square(BitScanForward(b.GetColorPieces(KING, c)))
	orthogonalThem := b.GetColorPieces(ROOK, ReverseColor(c)) | b.GetColorPieces(QUEEN, ReverseColor(c))
	diagonalThem := b.GetColorPieces(BISHOP, ReverseColor(c)) | b.GetColorPieces(QUEEN, ReverseColor(c))

	if c == WHITE {
		checkers |= (WHITE_PAWN_ATTACKS_LOOKUP[playerKing] & b.pieces[B_P])
	} else {
		checkers |= (BLACK_PAWN_ATTACKS_LOOKUP[playerKing] & b.pieces[W_P])
	}

	if checkers != 0 {
		return true
	}

	checkers |= (KnightAttacks(playerKing) & b.GetColorPieces(KNIGHT, ReverseColor(c)))
	if checkers != 0 {
		return true
	}

	var pinPoss u64 = (BishopAttacks(playerKing, b.colors[ReverseColor(c)]) & diagonalThem)
	pinPoss |= (RookAttacks(playerKing, b.colors[ReverseColor(c)]) & orthogonalThem)

	var lsb int
	var piecesBetween u64
	for {
		if pinPoss == 0 {
			break
		}
		lsb = PopLSB(&pinPoss)
		piecesBetween = SQUARES_BETWEEN[playerKing][lsb] & b.colors[c]

		if piecesBetween == 0 {
			checkers ^= SQUARE_TO_BITBOARD[lsb]
			return true
		}
	}

	return checkers != 0
}

func (b *Board) IsThreeFoldRep() bool {
	hash := b.zobrist
	matches := 0
	for i := len(b.history) - 1; i >= 0; i-- {
		entry := b.history[i]
		if entry.hash == hash && !entry.move.null {
			matches++
		}
		if matches >= 2 {
			break
		}
	}
	return matches >= 2
}

func (b *Board) IsTwoFold() bool {
	hash := b.zobrist
	matches := 1
	for i := len(b.history) - 1; i >= 0; i-- {
		entry := b.history[i]
		if entry.hash == hash && !entry.move.null {
			matches++
		}
		if matches >= 2 {
			break
		}
	}
	return matches >= 2
}

func (b *Board) IsInsufficientMaterial() bool {
	numPieces := PopCount(b.occupied)
	if numPieces == 2 {
		return true
	}
	if numPieces == 3 && (b.pieces[W_N] != 0 || b.pieces[B_N] != 0 || b.pieces[W_B] != 0 || b.pieces[B_B] != 0) {
		return true
	}
	if numPieces == 4 {
		if b.pieces[W_N] != 0 {
			count := PopCount(b.pieces[W_N])
			if count == 2 {
				return true
			}
		}
		if b.pieces[B_N] != 0 {
			count := PopCount(b.pieces[B_N])
			if count == 2 {
				return true
			}
		}
	}

	return false
}

func (b *Board) Print() {
	s := "\n"
	for i := 56; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			s += b.squares[i+j].ToString() + " "
		}
		s += "\n"
	}
	fmt.Print(s)
}

func (b *Board) PrintFromBitBoards() {
	s := "   +---+---+---+---+---+---+---+---+\n"
	for i := 56; i >= 0; i -= 8 {
		s += " " + fmt.Sprint(i/8+1) + " "
		for j := 0; j < 8; j++ {
			if b.occupied&u64(1<<(i+j)) != 0 {
				var found = false
				for k := 0; k < 14; k++ {
					if b.pieces[k]&u64(1<<(i+j)) != 0 {
						if found {
							fmt.Println("Duplicate pieces...")
						} else {
							found = true
							s += "| " + Piece(k).ToString() + " "
						}
					}
				}
				if !found {
					fmt.Println("Piece is in occupied bitboard not not present in any of the pieces bitboard...")
				}
			} else if b.empty&u64(1<<(i+j)) != 0 {
				s += "| " + EMPTY.ToString() + " "
			} else {
				fmt.Println("Square is not represented in either occupied or empty...")
			}
		}
		s += "| " + "\n"
		s += "   +---+---+---+---+---+---+---+---+\n"
	}

	s += "     A   B   C   D   E   F   G   H\n"
	fmt.Print(s)
}

// For testing purposes only
func (b *Board) GetStringFromBitBoards() string {
	s := "\n"
	for i := 56; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			if b.occupied&u64(1<<(i+j)) != 0 {
				var found = false
				for k := 0; k < 14; k++ {
					if b.pieces[k]&u64(1<<(i+j)) != 0 {
						if found {
							fmt.Println("Duplicate pieces...")
						} else {
							found = true
							s += Piece(k).ToString() + " "
						}
					}
				}
				if !found {
					fmt.Println("Piece is in occupied bitboard not not present in any of the pieces bitboard...")
				}
			} else if b.empty&u64(1<<(i+j)) != 0 {
				s += EMPTY.ToString() + " "
			} else {
				fmt.Println("Square is not represented in either occupied or empty...")
			}
		}
		s += "\n"
	}
	return s
}

// ToFEN generates a FEN string from the current board position
func (b *Board) ToFEN() string {
	var fen strings.Builder

	// Piece placement
	emptyCount := 0
	for i := 56; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			piece := b.squares[i+j]
			if piece == EMPTY {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				fen.WriteString(piece.ToString())
			}
		}
		if emptyCount > 0 {
			fen.WriteString(strconv.Itoa(emptyCount))
			emptyCount = 0
		}
		if i > 0 {
			fen.WriteString("/")
		}
	}

	// Active color
	if b.turn == WHITE {
		fen.WriteString(" w ")
	} else {
		fen.WriteString(" b ")
	}

	// Castling availability
	hasCastling := false
	if b.OO {
		fen.WriteString("K")
		hasCastling = true
	}
	if b.OOO {
		fen.WriteString("Q")
		hasCastling = true
	}
	if b.oo {
		fen.WriteString("k")
		hasCastling = true
	}
	if b.ooo {
		fen.WriteString("q")
		hasCastling = true
	}
	if !hasCastling {
		fen.WriteString("-")
	}

	// En passant target square
	fen.WriteString(" ")
	if b.enPassant == EMPTY_SQ {
		fen.WriteString("-")
	} else {
		fen.WriteString(SQUARE_TO_STRING_MAP[b.enPassant])
	}

	// Halfmove clock and fullmove number
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.plyCnt))
	fen.WriteString(" ")
	fen.WriteString(strconv.Itoa(b.moveCount))

	return fen.String()
}
