package engine

import (
	"fmt"
	"strconv"
	"strings"
)

type u64 uint64

type board struct {
	pieces    [14]u64   // Stores bitboards of all white and black pieces
	squares   [64]Piece // Stores all 64 squares (not used for move generation)
	colors    [2]u64    // Stores bitboards of both colors
	occupied  u64       // Bits are set when pieces are there
	empty     u64       // Bits are clear when pieces are there
	turn      Color     // Side to move
	enpassant Square    // En passant square. If not possible, stores EMPTY
	kW        bool      // If kingside castling available for White
	qW        bool      // If queenside castling is available for White
	kB        bool      // If kingside castling is available for Black
	qB        bool      // If queenside castling is available for Black
	history   []Move    // Stores move history for board
	zobrist   u64       // Zobrist hash (TODO)
	plyCnt    int       // Stores number of half moves played
	moveCount int       // Stores which move currently we are at
}

func newBoard() *board {
	b := board{}
	b.turn = WHITE
	b.enpassant = EMPTYSQ
	b.kW, b.qW, b.kB, b.qB = true, true, true, true
	b.initStartPos()
	return &b
}

func (b *board) initStartPos() {
	b.squares = [64]Piece{
		wR, wN, wB, wQ, wK, wB, wN, wR,
		wP, wP, wP, wP, wP, wP, wP, wP,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		bP, bP, bP, bP, bP, bP, bP, bP,
		bR, bN, bB, bQ, bK, bB, bN, bR}

	for i := 0; i < 64; i++ {
		if i < 16 {
			b.putPiece(b.squares[i], Square(i), WHITE)
		} else if i >= 48 {
			b.putPiece(b.squares[i], Square(i), BLACK)
		} else {
			b.empty ^= u64(1 << i)
		}
	}
}

func (b *board) initFEN(fen string) {
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
		b.kW = true
	}
	if strings.Contains(castling, "Q") {
		b.qW = true
	}
	if strings.Contains(castling, "k") {
		b.kB = true
	}
	if strings.Contains(castling, "q") {
		b.qB = true
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

func (b *board) getColorPieces(p PieceType, c Color) u64 {
	return b.pieces[int(p)+int(c)*colorIndexOffset]
}

func (b *board) putPiece(p Piece, s Square, c Color) {
	b.squares[s] = p

	if p != EMPTY {
		var square u64 = (1 << s)
		b.colors[c] |= square
		b.pieces[p] |= square
		b.occupied |= square
		b.empty &= ^square
	} else {
		var square u64 = (1 << s)
		b.empty |= square
		b.occupied &= ^square
	}
}

func (b *board) movePiece(p Piece, mvfrom Square, mvto Square, c Color) {
	var from u64 = 1 << mvfrom
	var to u64 = 1 << mvto
	var fromTo u64 = from ^ to

	b.pieces[p] ^= fromTo
	b.colors[c] ^= fromTo
	b.occupied ^= fromTo
	b.empty ^= fromTo

	// update mailbox representation
	b.squares[mvfrom] = EMPTY
	b.squares[mvto] = p
}

func (b *board) capturePiece(p Piece, q Piece, mvfrom Square, mvto Square, c Color) {
	var from u64 = 1 << mvfrom
	var to u64 = 1 << mvto
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
}

func (b *board) replacePiece(p Piece, q Piece, sq Square) {
	var square u64 = 1 << sq
	b.pieces[p] ^= square
	b.pieces[q] ^= square

	// update mailbox representation
	b.squares[sq] = q
}

func (b *board) removePiece(p Piece, sq Square) {
	var square u64 = 1 << sq
	b.pieces[p] ^= square
	b.occupied ^= square
	b.empty ^= square
	b.squares[sq] = EMPTY
}

func (b *board) makeMove(mv Move) {
	b.history = append(b.history, mv)

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
			b.removePiece(bP, mv.to.goDirection(SOUTH))
		} else {
			b.removePiece(wP, mv.to.goDirection(NORTH))
		}
	}

	b.turn = reverseColor(b.turn)

	// update castling rights
	if mv.piece == wK {
		b.kW = false
		b.qW = false
	}
	if mv.piece == bK {
		b.kB = false
		b.qB = false
	}
	if mv.piece == wR {
		if mv.from == a1 {
			b.qW = false
		}
		if mv.from == h1 {
			b.kW = false
		}
	}
	if mv.piece == bR {
		if mv.from == a8 {
			b.qB = false
		}
		if mv.from == h8 {
			b.kB = false
		}
	}
}

func (b *board) print() {
	s := "\n"
	for i := 56; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			s += b.squares[i+j].toString() + " "
		}
		s += "\n"
	}
	fmt.Print(s)
}

func (b *board) printFromBitBoards() {
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
	fmt.Print(s)
}
