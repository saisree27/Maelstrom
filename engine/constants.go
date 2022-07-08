package engine

/**
* Contains constants for pieces, squares, ranks, files, masks, etc.
* Also contains initialization methods for lookup tables for piece attacks
**/

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

func reverseColor(c Color) Color {
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

func (p Piece) getColor() Color {
	if p == wP || p == wN || p == wB || p == wR || p == wQ || p == wK {
		return WHITE
	} else {
		return BLACK
	}
}

type PieceType int

const (
	pawn PieceType = iota
	bishop
	knight
	rook
	queen
	king
)

func getCP(c Color, pt PieceType) Piece {
	return Piece(int(pt) + colorIndexOffset*int(c))
}

// access bitboard of correct color (e.g. pieces[wP + colorIndexOffset * color])
const colorIndexOffset = 6

type MoveType int

const (
	QUIET MoveType = iota
	CAPTURE
	KCASTLE
	QCASTLE
	PROMOTION
	ENPASSANT
	CAPTUREANDPROMOTION
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

type Rank int

const (
	R1 Rank = iota
	R2
	R3
	R4
	R5
	R6
	R7
	R8
)

var almostPromotion = map[Color]Rank{WHITE: R7, BLACK: R2}
var startingRank = map[Color]Rank{WHITE: R2, BLACK: R7}
var pawnPushDirection = map[Color]Direction{WHITE: NORTH, BLACK: SOUTH}

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
	EMPTYSQ
)

var stringToSquareMap = map[string]Square{
	"a1": a1, "a2": a2, "a3": a3, "a4": a4, "a5": a5, "a6": a6, "a7": a7, "a8": a8,
	"b1": b1, "b2": b2, "b3": b3, "b4": b4, "b5": b5, "b6": b6, "b7": b7, "b8": b8,
	"c1": c1, "c2": c2, "c3": c3, "c4": c4, "c5": c5, "c6": c6, "c7": c7, "c8": c8,
	"d1": d1, "d2": d2, "d3": d3, "d4": d4, "d5": d5, "d6": d6, "d7": d7, "d8": d8,
	"e1": e1, "e2": e2, "e3": e3, "e4": e4, "e5": e5, "e6": e6, "e7": e7, "e8": e8,
	"f1": f1, "f2": f2, "f3": f3, "f4": f4, "f5": f5, "f6": f6, "f7": f7, "f8": f8,
	"g1": g1, "g2": g2, "g3": g3, "g4": g4, "g5": g5, "g6": g6, "g7": g7, "g8": g8,
	"h1": h1, "h2": h2, "h3": h3, "h4": h4, "h5": h5, "h6": h6, "h7": h7, "h8": h8}

var squareToStringMap = reversedMap(stringToSquareMap)

func reversedMap(m map[string]Square) map[Square]string {
	n := make(map[Square]string, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

func (sq Square) goDirection(d Direction) Square {
	return Square(int(sq) + int(d))
}

var kingAttacksSquareLookup [64]u64

func initializeKingAttacks() {
	var bb u64 = 1
	for i := 0; i < 64; i++ {
		copy := bb
		attacks := shiftBitboard(copy, EAST) | shiftBitboard(copy, WEST)
		copy |= attacks
		attacks |= shiftBitboard(copy, NORTH) | shiftBitboard(copy, SOUTH)

		kingAttacksSquareLookup[i] = attacks
		bb = bb << 1
	}
}

var knightAttacksSquareLookup [64]u64

func initializeKnightAttacks() {
	var bb u64 = 1
	for i := 0; i < 64; i++ {
		var attacks u64 = 0
		attacks |= (bb << 17) & ^files[A]
		attacks |= (bb << 10) & ^files[A] & ^files[B]
		attacks |= (bb >> 6) & ^files[A] & ^files[B]
		attacks |= (bb >> 15) & ^files[A]
		attacks |= (bb << 15) & ^files[H]
		attacks |= (bb << 6) & ^files[G] & ^files[H]
		attacks |= (bb >> 10) & ^files[G] & ^files[H]
		attacks |= (bb >> 17) & ^files[H]

		knightAttacksSquareLookup[i] = attacks
		bb = bb << 1
	}
}

var whitePawnAttacksSquareLookup [64]u64
var blackPawnAttacksSquareLookup [64]u64

func initializePawnAttacks() {
	var bb u64 = 1
	for i := 0; i < 64; i++ {
		var attacks u64 = shiftBitboard(bb, NE) | shiftBitboard(bb, NW)
		whitePawnAttacksSquareLookup[i] = attacks
		attacks = shiftBitboard(bb, SE) | shiftBitboard(bb, SW)
		blackPawnAttacksSquareLookup[i] = attacks
		bb = bb << 1
	}
}

// get rid of variable unused error while developing, don't really like that lol
func UNUSED(x interface{}) {
	_ = x
}
