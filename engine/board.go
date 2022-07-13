package engine

import (
	"fmt"
	"strconv"
	"strings"
)

type u64 uint64

type Board struct {
	pieces    [14]u64   // Stores bitboards of all white and black pieces
	squares   [64]Piece // Stores all 64 squares (not used for move generation)
	colors    [2]u64    // Stores bitboards of both colors
	occupied  u64       // Bits are set when pieces are there
	empty     u64       // Bits are clear when pieces are there
	turn      Color     // Side to move
	enpassant Square    // En passant square. If not possible, stores EMPTY
	OO        bool      // If kingside castling available for White
	OOO       bool      // If queenside castling is available for White
	oo        bool      // If kingside castling is available for Black
	ooo       bool      // If queenside castling is available for Black
	history   []prev    // Stores history for board
	zobrist   u64       // Zobrist hash (TODO)
	plyCnt    int       // Stores number of half moves played
	moveCount int       // Stores which move currently we are at
}

type prev struct {
	move      Move   // Stores previous move made
	OO        bool   // White Kingside castling history
	OOO       bool   // White Queenside castling history
	oo        bool   // Black Kingside castling history
	ooo       bool   // Black Queenside castling history
	enpassant Square // En passant square history
	hash      u64    // Zobrist hash of prev position
}

func newBoard() *Board {
	b := Board{}
	b.turn = WHITE
	b.enpassant = EMPTYSQ
	b.OO, b.OOO, b.oo, b.ooo = true, true, true, true
	b.InitStartPos()
	return &b
}

func (b *Board) InitStartPos() {
	b.zobrist = 0
	b.squares = [64]Piece{
		wR, wN, wB, wQ, wK, wB, wN, wR,
		wP, wP, wP, wP, wP, wP, wP, wP,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		bP, bP, bP, bP, bP, bP, bP, bP,
		bR, bN, bB, bQ, bK, bB, bN, bR,
	}

	for i := 0; i < 64; i++ {
		if i < 16 {
			b.putPiece(b.squares[i], Square(i), WHITE)
		} else if i >= 48 {
			b.putPiece(b.squares[i], Square(i), BLACK)
		} else {
			b.empty ^= sToBB[i]
		}
	}
	b.turn = WHITE
	b.enpassant = EMPTYSQ
	b.OO = true
	b.OOO = true
	b.oo = true
	b.ooo = true
}

func (b *Board) InitFEN(fen string) {
	b.zobrist = 0
	for s := a1; s <= h8; s++ {
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
				b.putPiece(piece, sq, piece.getColor())
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
		b.enpassant = EMPTYSQ
	} else {
		b.enpassant = stringToSquareMap[enpassant]
	}

	halfMoveClock := attrs[4]
	b.plyCnt, _ = strconv.Atoi(halfMoveClock)

	moveCount := attrs[5]
	b.moveCount, _ = strconv.Atoi(moveCount)
}

func (b *Board) getColorPieces(p PieceType, c Color) u64 {
	return b.pieces[int(p)+int(c)*colorIndexOffset]
}

func (b *Board) putPiece(p Piece, s Square, c Color) {
	b.squares[s] = p

	if p != EMPTY {
		var square u64 = sToBB[s]
		b.colors[c] |= square
		b.pieces[p] |= square
		b.occupied |= square
		b.empty &= ^square
		b.zobrist ^= zobristTable[p][s]
	} else {
		var square u64 = sToBB[s]
		b.empty |= square
		b.occupied &= ^square
		b.zobrist ^= zobristTable[p][s]
	}
}

func (b *Board) movePiece(p Piece, mvfrom Square, mvto Square, c Color) {
	var from u64 = sToBB[mvfrom]
	var to u64 = sToBB[mvto]
	var fromTo u64 = from ^ to

	b.pieces[p] ^= fromTo
	b.colors[c] ^= fromTo
	b.occupied ^= fromTo
	b.empty ^= fromTo

	// update mailbox representation
	b.squares[mvfrom] = EMPTY
	b.squares[mvto] = p

	b.zobrist ^= zobristTable[p][mvfrom] ^ zobristTable[p][mvto] ^ zobristTable[EMPTY][mvto] ^ zobristTable[EMPTY][mvfrom]
}

func (b *Board) capturePiece(p Piece, q Piece, mvfrom Square, mvto Square, c Color) {
	var from u64 = sToBB[mvfrom]
	var to u64 = sToBB[mvto]
	var fromTo u64 = from ^ to
	b.pieces[p] ^= fromTo
	b.colors[c] ^= fromTo
	b.pieces[q] ^= to
	b.colors[reverseColor(c)] ^= to
	b.occupied ^= from
	b.empty ^= from

	// update mailbox representation
	b.squares[mvfrom] = EMPTY
	b.squares[mvto] = p

	b.zobrist ^= zobristTable[p][mvfrom] ^ zobristTable[p][mvto] ^ zobristTable[q][mvto] ^ zobristTable[EMPTY][mvfrom]
}

func (b *Board) replacePiece(p Piece, q Piece, sq Square) {
	var square u64 = sToBB[sq]
	b.pieces[p] ^= square
	b.pieces[q] ^= square

	// update mailbox representation
	b.squares[sq] = q

	b.zobrist ^= zobristTable[p][sq] ^ zobristTable[q][sq]
}

func (b *Board) removePiece(p Piece, sq Square, c Color) {
	var square u64 = sToBB[sq]
	b.pieces[p] ^= square
	b.occupied ^= square
	b.empty ^= square
	b.colors[c] ^= square
	b.squares[sq] = EMPTY
	b.zobrist ^= zobristTable[p][sq]
}

func (b *Board) makeMoveFromUCI(uci string) {
	b.makeMove(fromUCI(uci, *b))
}

func (b *Board) makeMoveNoUpdate(mv Move) {
	switch mv.movetype {
	case QUIET:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
	case CAPTURE:
		b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
	case PROMOTION:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
		b.replacePiece(mv.piece, mv.promote, mv.to)
	case CAPTUREANDPROMOTION:
		b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
		b.replacePiece(mv.piece, mv.promote, mv.to)
	case KCASTLE:
		if mv.colorMoved == WHITE {
			b.movePiece(wK, e1, g1, WHITE)
			b.movePiece(wR, h1, f1, WHITE)
		} else {
			b.movePiece(bK, e8, g8, BLACK)
			b.movePiece(bR, h8, f8, BLACK)
		}
	case QCASTLE:
		if mv.colorMoved == WHITE {
			b.movePiece(wK, e1, c1, WHITE)
			b.movePiece(wR, a1, d1, WHITE)
		} else {
			b.movePiece(bK, e8, c8, BLACK)
			b.movePiece(bR, a8, d8, BLACK)
		}
	case ENPASSANT:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
		if mv.colorMoved == WHITE {
			b.removePiece(bP, mv.to.goDirection(SOUTH), BLACK)
		} else {
			b.removePiece(wP, mv.to.goDirection(NORTH), WHITE)
		}
	}
}

func (b *Board) makeMove(mv Move) {
	var entry prev = prev{move: mv, OO: b.OO, OOO: b.OOO, oo: b.oo, ooo: b.ooo, enpassant: b.enpassant, hash: b.zobrist}
	b.history = append(b.history, entry)

	switch mv.movetype {
	case QUIET:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
	case CAPTURE:
		b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
	case PROMOTION:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
		b.replacePiece(mv.piece, mv.promote, mv.to)
	case CAPTUREANDPROMOTION:
		b.capturePiece(mv.piece, mv.captured, mv.from, mv.to, mv.colorMoved)
		b.replacePiece(mv.piece, mv.promote, mv.to)
	case KCASTLE:
		if mv.colorMoved == WHITE {
			b.movePiece(wK, e1, g1, WHITE)
			b.movePiece(wR, h1, f1, WHITE)
		} else {
			b.movePiece(bK, e8, g8, BLACK)
			b.movePiece(bR, h8, f8, BLACK)
		}
	case QCASTLE:
		if mv.colorMoved == WHITE {
			b.movePiece(wK, e1, c1, WHITE)
			b.movePiece(wR, a1, d1, WHITE)
		} else {
			b.movePiece(bK, e8, c8, BLACK)
			b.movePiece(bR, a8, d8, BLACK)
		}
	case ENPASSANT:
		b.movePiece(mv.piece, mv.from, mv.to, mv.colorMoved)
		if mv.colorMoved == WHITE {
			b.removePiece(bP, mv.to.goDirection(SOUTH), BLACK)
		} else {
			b.removePiece(wP, mv.to.goDirection(NORTH), WHITE)
		}
	}

	b.turn = reverseColor(b.turn)

	// update castling rights
	if mv.piece == wK {
		b.OO = false
		b.OOO = false
	}
	if mv.piece == bK {
		b.oo = false
		b.ooo = false
	}
	if mv.piece == wR {
		if mv.from == a1 {
			b.OOO = false
		}
		if mv.from == h1 {
			b.OO = false
		}
	}
	if mv.piece == bR {
		if mv.from == a8 {
			b.ooo = false
		}
		if mv.from == h8 {
			b.oo = false
		}
	}

	// update en passant square
	dist := Direction(int(mv.to - mv.from))
	if dist == 2*NORTH && mv.piece == wP {
		b.enpassant = mv.from + Square(NORTH)
	} else if dist == 2*SOUTH && mv.piece == bP {
		b.enpassant = mv.from + Square(SOUTH)
	} else {
		b.enpassant = EMPTYSQ
	}

	b.plyCnt++
}

func (b *Board) undoNoUpdate(prevMove Move) {
	switch prevMove.movetype {
	case QUIET:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
	case CAPTURE:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
		b.putPiece(prevMove.captured, prevMove.to, reverseColor(prevMove.colorMoved))
	case PROMOTION:
		b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
		b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
	case CAPTUREANDPROMOTION:
		b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
		b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
		b.putPiece(prevMove.captured, prevMove.to, reverseColor(prevMove.colorMoved))
	case KCASTLE:
		if prevMove.colorMoved == WHITE {
			b.movePiece(wK, g1, e1, WHITE)
			b.movePiece(wR, f1, h1, WHITE)
		} else {
			b.movePiece(bK, g8, e8, BLACK)
			b.movePiece(bR, f8, h8, BLACK)
		}
	case QCASTLE:
		if prevMove.colorMoved == WHITE {
			b.movePiece(wK, c1, e1, WHITE)
			b.movePiece(wR, d1, a1, WHITE)
		} else {
			b.movePiece(bK, c8, e8, BLACK)
			b.movePiece(bR, d8, a8, BLACK)
		}
	case ENPASSANT:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
		if prevMove.colorMoved == WHITE {
			b.putPiece(bP, prevMove.to.goDirection(SOUTH), BLACK)
		} else {
			b.putPiece(wP, prevMove.to.goDirection(NORTH), WHITE)
		}
	}
}

func (b *Board) undo() {
	prevEntry := b.history[len(b.history)-1]
	prevMove := prevEntry.move

	switch prevMove.movetype {
	case QUIET:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
	case CAPTURE:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
		b.putPiece(prevMove.captured, prevMove.to, reverseColor(prevMove.colorMoved))
	case PROMOTION:
		b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
		b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
	case CAPTUREANDPROMOTION:
		b.removePiece(prevMove.promote, prevMove.to, prevMove.colorMoved)
		b.putPiece(prevMove.piece, prevMove.from, prevMove.colorMoved)
		b.putPiece(prevMove.captured, prevMove.to, reverseColor(prevMove.colorMoved))
	case KCASTLE:
		if prevMove.colorMoved == WHITE {
			b.movePiece(wK, g1, e1, WHITE)
			b.movePiece(wR, f1, h1, WHITE)
		} else {
			b.movePiece(bK, g8, e8, BLACK)
			b.movePiece(bR, f8, h8, BLACK)
		}
	case QCASTLE:
		if prevMove.colorMoved == WHITE {
			b.movePiece(wK, c1, e1, WHITE)
			b.movePiece(wR, d1, a1, WHITE)
		} else {
			b.movePiece(bK, c8, e8, BLACK)
			b.movePiece(bR, d8, a8, BLACK)
		}
	case ENPASSANT:
		b.movePiece(prevMove.piece, prevMove.to, prevMove.from, prevMove.colorMoved)
		if prevMove.colorMoved == WHITE {
			b.putPiece(bP, prevMove.to.goDirection(SOUTH), BLACK)
		} else {
			b.putPiece(wP, prevMove.to.goDirection(NORTH), WHITE)
		}
	}
	b.OO = prevEntry.OO
	b.OOO = prevEntry.OOO
	b.oo = prevEntry.oo
	b.ooo = prevEntry.ooo
	b.enpassant = prevEntry.enpassant
	b.zobrist = prevEntry.hash
	b.turn = reverseColor(b.turn)

	b.history = b.history[:len(b.history)-1]
	b.plyCnt--
}

func (b *Board) isCheck(c Color) bool {
	player := c
	opponent := reverseColor(c)

	playerKing := b.getColorPieces(king, player)
	return b.attacksOn(opponent, b.occupied, playerKing, true)
}

func (b *Board) isThreeFoldRep() bool {
	hash := b.zobrist
	matches := 0
	for i := len(b.history) - 1; i >= 0; i-- {
		entry := b.history[i]
		if entry.hash == hash {
			matches++
		}
		if matches >= 2 {
			break
		}
	}
	return matches >= 2
}

func (b *Board) print() {
	s := "\n"
	for i := 56; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			s += b.squares[i+j].toString() + " "
		}
		s += "\n"
	}
	fmt.Print(s)
}

func (b *Board) printFromBitBoards() {
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
							s += "| " + Piece(k).toString() + " "
						}
					}
				}
				if !found {
					fmt.Println("Piece is in occupied bitboard not not present in any of the pieces bitboard...")
				}
			} else if b.empty&u64(1<<(i+j)) != 0 {
				s += "| " + EMPTY.toString() + " "
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
func (b *Board) getStringFromBitBoards() string {
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
							s += Piece(k).toString() + " "
						}
					}
				}
				if !found {
					fmt.Println("Piece is in occupied bitboard not not present in any of the pieces bitboard...")
				}
			} else if b.empty&u64(1<<(i+j)) != 0 {
				s += EMPTY.toString() + " "
			} else {
				fmt.Println("Square is not represented in either occupied or empty...")
			}
		}
		s += "\n"
	}
	return s
}
