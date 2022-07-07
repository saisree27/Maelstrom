package engine

// Bitboard constants
// Using Little-Endian Rank-File Mapping

// Bitboard mask of a-h files
var files = [8]u64{
	0x101010101010101, 0x202020202020202, 0x404040404040404, 0x808080808080808,
	0x1010101010101010, 0x2020202020202020, 0x4040404040404040, 0x8080808080808080}

// Bitboard mask of ranks 1-8
var ranks = [8]u64{
	0xff, 0xff00, 0xff0000, 0xff000000,
	0xff00000000, 0xff0000000000, 0xff000000000000, 0xff00000000000000}

const a1h8Diagonal = 0x8040201008040201
const h1a8Diagonal = 0x0102040810204080
const lightSquares = 0x55AA55AA55AA55AA
const darkSquares = 0xAA55AA55AA55AA55

// Defining color (white = 0, black = 1)
type Color int

const (
	WHITE Color = iota
	BLACK
)

func ReverseColor(c Color) Color {
	return Color(c ^ 1)
}

type Piece int

const (
	wP Piece = iota
	wB
	wN
	wR
	wQ
	wK
	bP
	bB
	bN
	bR
	bQ
	bK
	EMPTY
)

func (p Piece) toString() string {
	switch p {
	case wP:
		return "P"
	case wB:
		return "B"
	case wN:
		return "N"
	case wR:
		return "R"
	case wQ:
		return "Q"
	case wK:
		return "K"
	case bP:
		return "p"
	case bB:
		return "b"
	case bN:
		return "n"
	case bR:
		return "r"
	case bQ:
		return "q"
	case bK:
		return "k"
	case EMPTY:
		return "."
	}
	return "."
}

// access bitboard of correct color (e.g. pieces[wP + colorIndexOffset * color])
const colorIndexOffset = 6

type MoveType int

const (
	QUIET MoveType = iota
	CAPTURE
	CHECK
	CASTLE
	PROMOTION
	ENPASSANT
)

type Direction int

const (
	NORTH Direction = 8
	NW    Direction = 7
	NE    Direction = 9
	WEST  Direction = -1
	EAST  Direction = 1
	SW    Direction = -9
	SOUTH Direction = -8
	SE    Direction = -7
)

type File int

const (
	A File = iota
	B
	C
	D
	E
	F
	G
	H
)

type Square int

// LERM ordering
const (
	a1 Square = iota
	b1
	c1
	d1
	e1
	f1
	g1
	h1
	a2
	b2
	c2
	d2
	e2
	f2
	g2
	h2
	a3
	b3
	c3
	d3
	e3
	f3
	g3
	h3
	a4
	b4
	c4
	d4
	e4
	f4
	g4
	h4
	a5
	b5
	c5
	d5
	e5
	f5
	g5
	h5
	a6
	b6
	c6
	d6
	e6
	f6
	g6
	h6
	a7
	b7
	c7
	d7
	e7
	f7
	g7
	h7
	a8
	b8
	c8
	d8
	e8
	f8
	g8
	h8
)
