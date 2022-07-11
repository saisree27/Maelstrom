package engine

import (
	"math/bits"
	"reflect"
	"sort"
)

/**
* Contains constants for pieces, squares, ranks, files, masks, etc.
* Also contains initialization methods for lookup tables for piece attacks
**/

// Bitboard constants
// Using Little-Endian Rank-File Mapping

// Bitboard mask of a-h files
var files = [8]u64{
	0x101010101010101, 0x202020202020202, 0x404040404040404, 0x808080808080808,
	0x1010101010101010, 0x2020202020202020, 0x4040404040404040, 0x8080808080808080,
}

// Bitboard mask of ranks 1-8
var ranks = [8]u64{
	0xff, 0xff00, 0xff0000, 0xff000000,
	0xff00000000, 0xff0000000000, 0xff000000000000, 0xff00000000000000,
}

var diagonals = [15]u64{
	0x80, 0x8040, 0x804020,
	0x80402010, 0x8040201008, 0x804020100804,
	0x80402010080402, 0x8040201008040201, 0x4020100804020100,
	0x2010080402010000, 0x1008040201000000, 0x804020100000000,
	0x402010000000000, 0x201000000000000, 0x100000000000000,
}

var antiDiagonals = [15]u64{
	0x1, 0x102, 0x10204,
	0x1020408, 0x102040810, 0x10204081020,
	0x1020408102040, 0x102040810204080, 0x204081020408000,
	0x408102040800000, 0x810204080000000, 0x1020408000000000,
	0x2040800000000000, 0x4080000000000000, 0x8000000000000000,
}

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

var stringToPieceMap = map[string]Piece{
	"P": wP, "B": wB, "N": wN, "R": wR, "Q": wQ, "K": wK,
	"p": bP, "b": bB, "n": bN, "r": bR, "q": bQ, "k": bK,
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
	"h1": h1, "h2": h2, "h3": h3, "h4": h4, "h5": h5, "h6": h6, "h7": h7, "h8": h8,
}

func sqToRank(s Square) Rank {
	return Rank(s >> 3)
}

func sqToFile(s Square) File {
	return File(s & 0b111)
}

func sqToDiag(s Square) u64 {
	return diagonals[int(sqToRank(s))-int(sqToFile(s))+7]
}

func sqToAntiDiag(s Square) u64 {
	return antiDiagonals[int(sqToRank(s))+int(sqToFile(s))]
}

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

var colorToPawnLookup = map[Color]*[64]u64{
	WHITE: &whitePawnAttacksSquareLookup, BLACK: &blackPawnAttacksSquareLookup,
}

var colorToPawnLookupReverse = map[Color]*[64]u64{
	WHITE: &blackPawnAttacksSquareLookup, BLACK: &whitePawnAttacksSquareLookup,
}

var colorToKingLookup = map[Color]Piece{
	WHITE: wK, BLACK: bK,
}

var sToBB [64]u64

func initializeSQLookup() {
	for i := 0; i < 64; i++ {
		sToBB[i] = 1 << i
	}
}

// Magic bitboard initialization for bishop and rook moves:
//
// Don't really quite understand how magics work yet so a lot of the
// implementations are derived from chessprogramming wiki and open-source
// move generators
//

// Function to get sliding attacks on a given line/diagonal
// Implementation from https://github.com/nkarve/surge
func slidingAttacks(s Square, b u64, locs u64) u64 {
	a := (locs & b) - (1<<s)*2
	c := (bits.Reverse64(uint64(b&locs)) - bits.Reverse64(1<<s)*2)
	d := bits.Reverse64(c)
	return (a ^ u64(d)) & locs
}

// masks
var bishopMasks [64]u64
var rookMasks [64]u64

// attacks
var bishopAttacks [64][512]u64
var rookAttacks [64][4096]u64

// magics from Chess Programming YT video
var rookMagics = [64]u64{
	0xa8002c000108020, 0x6c00049b0002001, 0x100200010090040, 0x2480041000800801,
	0x280028004000800, 0x900410008040022, 0x280020001001080, 0x2880002041000080,
	0xa000800080400034, 0x4808020004000, 0x2290802004801000, 0x411000d00100020,
	0x402800800040080, 0xb000401004208, 0x2409000100040200, 0x1002100004082,
	0x22878001e24000, 0x1090810021004010, 0x801030040200012, 0x500808008001000,
	0xa08018014000880, 0x8000808004000200, 0x201008080010200, 0x801020000441091,
	0x800080204005, 0x1040200040100048, 0x120200402082, 0xd14880480100080,
	0x12040280080080, 0x100040080020080, 0x9020010080800200, 0x813241200148449,
	0x491604001800080, 0x100401000402001, 0x4820010021001040, 0x400402202000812,
	0x209009005000802, 0x810800601800400, 0x4301083214000150, 0x204026458e001401,
	0x40204000808000, 0x8001008040010020, 0x8410820820420010, 0x1003001000090020,
	0x804040008008080, 0x12000810020004, 0x1000100200040208, 0x430000a044020001,
	0x280009023410300, 0xe0100040002240, 0x200100401700, 0x2244100408008080,
	0x8000400801980, 0x2000810040200, 0x8010100228810400, 0x2000009044210200,
	0x4080008040102101, 0x40002080411d01, 0x2005524060000901, 0x502001008400422,
	0x489a000810200402, 0x1004400080a13, 0x4000011008020084, 0x26002114058042,
}

var bishopMagics = [64]u64{
	0x89a1121896040240, 0x2004844802002010, 0x2068080051921000, 0x62880a0220200808,
	0x4042004000000, 0x100822020200011, 0xc00444222012000a, 0x28808801216001,
	0x400492088408100, 0x201c401040c0084, 0x840800910a0010, 0x82080240060,
	0x2000840504006000, 0x30010c4108405004, 0x1008005410080802, 0x8144042209100900,
	0x208081020014400, 0x4800201208ca00, 0xf18140408012008, 0x1004002802102001,
	0x841000820080811, 0x40200200a42008, 0x800054042000, 0x88010400410c9000,
	0x520040470104290, 0x1004040051500081, 0x2002081833080021, 0x400c00c010142,
	0x941408200c002000, 0x658810000806011, 0x188071040440a00, 0x4800404002011c00,
	0x104442040404200, 0x511080202091021, 0x4022401120400, 0x80c0040400080120,
	0x8040010040820802, 0x480810700020090, 0x102008e00040242, 0x809005202050100,
	0x8002024220104080, 0x431008804142000, 0x19001802081400, 0x200014208040080,
	0x3308082008200100, 0x41010500040c020, 0x4012020c04210308, 0x208220a202004080,
	0x111040120082000, 0x6803040141280a00, 0x2101004202410000, 0x8200000041108022,
	0x21082088000, 0x2410204010040, 0x40100400809000, 0x822088220820214,
	0x40808090012004, 0x910224040218c9, 0x402814422015008, 0x90014004842410,
	0x1000042304105, 0x10008830412a00, 0x2520081090008908, 0x40102000a0a60140,
}

// rook occupancy bits
var rookShifts = [64]int{}

// bishop occupancy bits
var bishopShifts = [64]int{}

func slidingBishopAttacksForInitialization(s Square, b u64) u64 {
	var attacks u64 = 0
	attacks |= slidingAttacks(s, b, sqToAntiDiag(s))
	attacks |= slidingAttacks(s, b, sqToDiag(s))
	return attacks
}

func initBishopAttacks() {
	for s := a1; s <= h8; s++ {
		var edgeMask u64 = (files[A] | files[H]) & ^files[sqToFile(s)]
		edgeMask |= (ranks[R1] | ranks[R8]) & ^ranks[sqToRank(s)]
		// printBitBoard(edgeMask)
		// remove board edges from attack squares
		bishopMasks[s] = (sqToAntiDiag(s) ^ sqToDiag(s)) & ^edgeMask

		// printBitBoard(bishopMasks[s])
		// number of movement possibilities from square s
		bishopShifts[s] = 64 - popCount(bishopMasks[s])

		var i u64 = 0
		for {
			j := i
			j *= bishopMagics[s]
			j >>= bishopShifts[s]
			bishopAttacks[s][j] = slidingBishopAttacksForInitialization(s, i)
			i = (i - bishopMasks[s]) & bishopMasks[s]
			if i == 0 {
				break
			}
		}
	}
}

func slidingRookAttacksForInitialization(s Square, b u64) u64 {
	var attacks u64 = 0
	attacks |= slidingAttacks(s, b, files[sqToFile(s)])
	attacks |= slidingAttacks(s, b, ranks[sqToRank(s)])
	return attacks
}

func initRookAttacks() {
	for s := a1; s <= h8; s++ {
		var edgeMask u64 = (files[A] | files[H]) & ^files[sqToFile(s)]
		edgeMask |= (ranks[R1] | ranks[R8]) & ^ranks[sqToRank(s)]

		// remove board edges from attack squares
		rookMasks[s] = (files[sqToFile(s)] ^ ranks[sqToRank(s)]) & ^edgeMask

		// number of movement possibilities from square s
		rookShifts[s] = 64 - popCount(rookMasks[s])

		var i u64 = 0
		for {
			j := i
			j *= rookMagics[s]
			j >>= rookShifts[s]
			rookAttacks[s][j] = slidingRookAttacksForInitialization(s, i)
			i = (i - rookMasks[s]) & rookMasks[s]
			if i == 0 {
				break
			}
		}
	}
}

// helpful tables found in nkarve/surge
var squaresBetween [64][64]u64

func initSquaresBetween() {
	var squares u64
	for i := a1; i <= h8; i++ {
		for j := a1; j <= h8; j++ {
			squares = 1<<i | 1<<j
			if (sqToFile(i) == sqToFile(j)) || (sqToRank(i) == sqToRank(j)) {
				squaresBetween[i][j] =
					slidingRookAttacksForInitialization(i, squares) & slidingRookAttacksForInitialization(j, squares)
			} else if (sqToDiag(i) == sqToDiag(j)) || (sqToAntiDiag(i) == sqToAntiDiag(j)) {
				squaresBetween[i][j] =
					slidingBishopAttacksForInitialization(i, squares) & slidingBishopAttacksForInitialization(j, squares)
			}
		}
	}
}

var line [64][64]u64

func initLine() {
	for i := a1; i <= h8; i++ {
		for j := a1; j <= h8; j++ {
			if (sqToFile(i) == sqToFile(j)) || (sqToRank(i) == sqToRank(j)) {
				line[i][j] =
					(slidingRookAttacksForInitialization(i, 0) & slidingRookAttacksForInitialization(j, 0)) | 1<<i | 1<<j
			} else if (sqToDiag(i) == sqToDiag(j)) || (sqToAntiDiag(i) == sqToAntiDiag(j)) {
				line[i][j] =
					(slidingBishopAttacksForInitialization(i, 0) & slidingBishopAttacksForInitialization(j, 0)) | 1<<i | 1<<j
			}
		}
	}
}

// Some slice functions for unit testing and possibly other things in the engine
func contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkSameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	a_copy := make([]string, len(a))
	b_copy := make([]string, len(b))

	copy(a_copy, a)
	copy(b_copy, b)

	sort.Strings(a_copy)
	sort.Strings(b_copy)

	return reflect.DeepEqual(a_copy, b_copy)
}

// get rid of variable unused error while developing, don't really like that lol
func UNUSED(x interface{}) {
	_ = x
}
