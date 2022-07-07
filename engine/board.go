package engine

import "fmt"

type u64 uint64

type board struct {
	pieces    [15]u64   // Stores bitboards of all white and black pieces along with EMPTY
	squares   [64]Piece // Stores all 64 squares and which pieces are on them
	colors    [2]u64    // Stores bitboards of both colors
	occupied  u64       // Bits are set when pieces are there
	empty     u64       // Bits are clear when pieces are there
	turn      Color     // Side to move
	enpassant bool      // If en passant is possible at current turn
	kW        bool      // If kingside castling available for White
	qW        bool      // If queenside castling is available for White
	kB        bool      // If kingside castling is available for Black
	qB        bool      // If queenside castling is available for Black
	history   []Move    // Stores move history for board
	zobrist   u64       // Zobrist hash (TODO)
}

func newBoard() *board {
	b := board{}
	b.turn = WHITE
	b.enpassant = false
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
		}
	}
}

func (b *board) initFEN(fen string) {

}

func (b *board) getPieceSet(p Piece) u64 {
	return b.pieces[p]
}

func (b *board) getColorPieces(p Piece, c Color) u64 {
	return b.pieces[int(p)+int(c)*colorIndexOffset]
}

func (b *board) getBothPieces(p1 Piece, p2 Piece) u64 {
	return b.pieces[p1] | b.pieces[p2]
}

func (b *board) putPiece(p Piece, s Square, c Color) {
	b.squares[s] = p

	var square u64 = (1 << s)
	b.colors[c] |= square
	b.pieces[p] |= square
	b.occupied |= square
}

func (b *board) makeMove(mv Move) {
	if mv.movetype == QUIET {
		var from u64 = 1 << mv.from
		var to u64 = 1 << mv.to
		var fromTo u64 = from ^ to
		b.pieces[mv.piece] ^= fromTo
		b.colors[mv.colorMoved] ^= fromTo
		b.occupied ^= fromTo
		b.empty ^= fromTo

		// update mailbox representation
		b.squares[mv.from] = EMPTY
		b.squares[mv.to] = mv.piece
	} else if mv.movetype == CAPTURE {
		var from u64 = 1 << mv.from
		var to u64 = 1 << mv.to
		var fromTo u64 = from ^ to
		b.pieces[mv.piece] ^= fromTo
		b.colors[mv.colorMoved] ^= fromTo
		b.pieces[mv.captured] ^= to
		b.pieces[ReverseColor(mv.colorMoved)] ^= to
		b.occupied ^= from
		b.empty ^= from

		// update mailbox representation
		b.squares[mv.from] = EMPTY
		b.squares[mv.to] = mv.piece
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
