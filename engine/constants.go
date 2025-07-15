package engine

import (
	"math/bits"
	"math/rand"
	"reflect"
	"sort"
)

/**
* Contains constants for pieces, squares, ranks, files, masks, etc.
* Also contains initialization methods for lookup tables for piece attacks
**/

// Bitboard constants
// Using Little-Endian Rank-File Mapping

// Bitboard mask of a-h FILES
var FILES = [8]u64{
	0x101010101010101, 0x202020202020202, 0x404040404040404, 0x808080808080808,
	0x1010101010101010, 0x2020202020202020, 0x4040404040404040, 0x8080808080808080,
}

// Bitboard mask of RANKS 1-8
var RANKS = [8]u64{
	0xff, 0xff00, 0xff0000, 0xff000000,
	0xff00000000, 0xff0000000000, 0xff000000000000, 0xff00000000000000,
}

var DIAGONALS = [15]u64{
	0x80, 0x8040, 0x804020,
	0x80402010, 0x8040201008, 0x804020100804,
	0x80402010080402, 0x8040201008040201, 0x4020100804020100,
	0x2010080402010000, 0x1008040201000000, 0x804020100000000,
	0x402010000000000, 0x201000000000000, 0x100000000000000,
}

var ANTI_DIAGONALS = [15]u64{
	0x1, 0x102, 0x10204,
	0x1020408, 0x102040810, 0x10204081020,
	0x1020408102040, 0x102040810204080, 0x204081020408000,
	0x408102040800000, 0x810204080000000, 0x1020408000000000,
	0x2040800000000000, 0x4080000000000000, 0x8000000000000000,
}

const A1_H8_DIAGONAL = 0x8040201008040201
const H1_A8_DIAGONAL = 0x0102040810204080
const LIGHT_SQUARES = 0x55AA55AA55AA55AA
const DARK_SQUARES = 0xAA55AA55AA55AA55

const QUEENSIDE = 0xf0f0f0f0f0f0f0f
const KINGSIDE = 0xf0f0f0f0f0f0f0f0

// Defining color (white = 0, black = 1)
type Color uint8

const (
	WHITE Color = iota
	BLACK
	NONE
)

func ReverseColor(c Color) Color {
	return Color(c ^ 1)
}

type Piece uint8

const (
	W_P Piece = iota
	W_N
	W_B
	W_R
	W_Q
	W_K
	B_P
	B_N
	B_B
	B_R
	B_Q
	B_K
	EMPTY
)

func (p Piece) ToString() string {
	switch p {
	case W_P:
		return "P"
	case W_B:
		return "B"
	case W_N:
		return "N"
	case W_R:
		return "R"
	case W_Q:
		return "Q"
	case W_K:
		return "K"
	case B_P:
		return "p"
	case B_B:
		return "b"
	case B_N:
		return "n"
	case B_R:
		return "r"
	case B_Q:
		return "q"
	case B_K:
		return "k"
	case EMPTY:
		return "."
	}
	return "."
}

var stringToPieceMap = map[string]Piece{
	"P": W_P, "B": W_B, "N": W_N, "R": W_R, "Q": W_Q, "K": W_K,
	"p": B_P, "b": B_B, "n": B_N, "r": B_R, "q": B_Q, "k": B_K,
}

func (p Piece) GetColor() Color {
	if p == EMPTY {
		return NONE
	}
	if p == W_P || p == W_N || p == W_B || p == W_R || p == W_Q || p == W_K {
		return WHITE
	} else {
		return BLACK
	}
}

type PieceType uint8

const (
	PAWN PieceType = iota
	KNIGHT
	BISHOP
	ROOK
	QUEEN
	KING
)

func PieceTypeToPiece(c Color, pt PieceType) Piece {
	return Piece(int(pt) + COLOR_INDEX_OFFSET*int(c))
}

func PieceToPieceType(p Piece) PieceType {
	if p == EMPTY {
		panic("cannot convert EMPTY to PieceType")
	}
	return PieceType(p % COLOR_INDEX_OFFSET)
}

// access bitboard of correct color (e.g. pieces[wP + COLOR_INDEX_OFFSET * color])
const COLOR_INDEX_OFFSET = 6

type MoveType uint8

const (
	QUIET MoveType = iota
	CAPTURE
	K_CASTLE
	Q_CASTLE
	PROMOTION
	EN_PASSANT
	CAPTURE_AND_PROMOTION
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

type File uint8

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

var FILE_NEIGHBORS [8]u64

func InitNeighborMasks() {
	for f := A; f <= H; f++ {
		if f == A {
			FILE_NEIGHBORS[f] = FILES[B]
		} else if f == H {
			FILE_NEIGHBORS[f] = FILES[G]
		} else {
			FILE_NEIGHBORS[f] = FILES[f-1] | FILES[f+1]
		}
	}
}

type Rank uint8

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

var ALMOST_PROMOTION = []Rank{WHITE: R7, BLACK: R2}
var STARTING_RANK = []Rank{WHITE: R2, BLACK: R7}
var PAWN_PUSH_DIRECTION = []Direction{WHITE: NORTH, BLACK: SOUTH}

type Square int8

// LERM ordering
const (
	A1 Square = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	EMPTY_SQ
)

var STRING_TO_SQUARE_MAP = map[string]Square{
	"a1": A1, "a2": A2, "a3": A3, "a4": A4, "a5": A5, "a6": A6, "a7": A7, "a8": A8,
	"b1": B1, "b2": B2, "b3": B3, "b4": B4, "b5": B5, "b6": B6, "b7": B7, "b8": B8,
	"c1": C1, "c2": C2, "c3": C3, "c4": C4, "c5": C5, "c6": C6, "c7": C7, "c8": C8,
	"d1": D1, "d2": D2, "d3": D3, "d4": D4, "d5": D5, "d6": D6, "d7": D7, "d8": D8,
	"e1": E1, "e2": E2, "e3": E3, "e4": E4, "e5": E5, "e6": E6, "e7": E7, "e8": E8,
	"f1": F1, "f2": F2, "f3": F3, "f4": F4, "f5": F5, "f6": F6, "f7": F7, "f8": F8,
	"g1": G1, "g2": G2, "g3": G3, "g4": G4, "g5": G5, "g6": G6, "g7": G7, "g8": G8,
	"h1": H1, "h2": H2, "h3": H3, "h4": H4, "h5": H5, "h6": H6, "h7": H7, "h8": H8,
}

func SquareToRank(s Square) Rank {
	return Rank(s >> 3)
}

func SquareToFile(s Square) File {
	return File(s & 0b111)
}

func SquareToDiagonal(s Square) u64 {
	return DIAGONALS[int(SquareToRank(s))-int(SquareToFile(s))+7]
}

func SquareToAntiDiagonal(s Square) u64 {
	return ANTI_DIAGONALS[int(SquareToRank(s))+int(SquareToFile(s))]
}

var SQUARE_TO_STRING_MAP = reversedMap(STRING_TO_SQUARE_MAP)

func reversedMap(m map[string]Square) map[Square]string {
	n := make(map[Square]string, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

func (sq Square) GoDirection(d Direction) Square {
	return Square(int(sq) + int(d))
}

var KING_ATTACKS_LOOKUP [64]u64

func InitializeKingAttacks() {
	var bb u64 = 1
	for i := 0; i < 64; i++ {
		copy := bb
		attacks := ShiftBitboard(copy, EAST) | ShiftBitboard(copy, WEST)
		copy |= attacks
		attacks |= ShiftBitboard(copy, NORTH) | ShiftBitboard(copy, SOUTH)

		KING_ATTACKS_LOOKUP[i] = attacks
		bb = bb << 1
	}
}

var KNIGHT_ATTACKS_LOOKUP [64]u64

func InitializeKnightAttacks() {
	var bb u64 = 1
	for i := 0; i < 64; i++ {
		var attacks u64 = 0
		attacks |= (bb << 17) & ^FILES[A]
		attacks |= (bb << 10) & ^FILES[A] & ^FILES[B]
		attacks |= (bb >> 6) & ^FILES[A] & ^FILES[B]
		attacks |= (bb >> 15) & ^FILES[A]
		attacks |= (bb << 15) & ^FILES[H]
		attacks |= (bb << 6) & ^FILES[G] & ^FILES[H]
		attacks |= (bb >> 10) & ^FILES[G] & ^FILES[H]
		attacks |= (bb >> 17) & ^FILES[H]

		KNIGHT_ATTACKS_LOOKUP[i] = attacks
		bb = bb << 1
	}
}

var WHITE_PAWN_ATTACKS_LOOKUP [64]u64
var BLACK_PAWN_ATTACKS_LOOKUP [64]u64

func InitializePawnAttacks() {
	var bb u64 = 1
	for i := 0; i < 64; i++ {
		var attacks u64 = ShiftBitboard(bb, NE) | ShiftBitboard(bb, NW)
		WHITE_PAWN_ATTACKS_LOOKUP[i] = attacks
		attacks = ShiftBitboard(bb, SE) | ShiftBitboard(bb, SW)
		BLACK_PAWN_ATTACKS_LOOKUP[i] = attacks
		bb = bb << 1
	}
}

var COLOR_TO_PAWN_LOOKUP = []*[64]u64{
	WHITE: &WHITE_PAWN_ATTACKS_LOOKUP, BLACK: &BLACK_PAWN_ATTACKS_LOOKUP,
}

var COLOR_TO_PAWN_LOOKUP_REVERSE = []*[64]u64{
	WHITE: &BLACK_PAWN_ATTACKS_LOOKUP, BLACK: &WHITE_PAWN_ATTACKS_LOOKUP,
}

var COLOR_TO_KING_LOOKUP = []Piece{
	WHITE: W_K, BLACK: B_K,
}

var SQUARE_TO_BITBOARD [64]u64

func InitializeSQLookup() {
	for i := 0; i < 64; i++ {
		SQUARE_TO_BITBOARD[i] = 1 << i
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
func SlidingAttacks(s Square, b u64, locs u64) u64 {
	a := (locs & b) - (SQUARE_TO_BITBOARD[s])*2
	c := (bits.Reverse64(uint64(b&locs)) - bits.Reverse64(uint64(SQUARE_TO_BITBOARD[s]))*2)
	d := bits.Reverse64(c)
	return (a ^ u64(d)) & locs
}

// masks
var BISHOP_MASKS [64]u64
var ROOK_MASKS [64]u64

// attacks
var BISHOP_ATTACKS [64][512]u64
var ROOK_ATTACKS [64][4096]u64

// magics from Chess Programming YT video
var ROOK_MAGICS = [64]u64{
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

var BISHOP_MAGICS = [64]u64{
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
var ROOK_SHIFTS = [64]int{}

// bishop occupancy bits
var BISHOP_SHIFTS = [64]int{}

func slidingBishopAttacksForInitialization(s Square, b u64) u64 {
	var attacks u64 = 0
	attacks |= SlidingAttacks(s, b, SquareToAntiDiagonal(s))
	attacks |= SlidingAttacks(s, b, SquareToDiagonal(s))
	return attacks
}

func InitBishopAttacks() {
	for s := A1; s <= H8; s++ {
		var edgeMask u64 = (FILES[A] | FILES[H]) & ^FILES[SquareToFile(s)]
		edgeMask |= (RANKS[R1] | RANKS[R8]) & ^RANKS[SquareToRank(s)]
		// printBitBoard(edgeMask)
		// remove board edges from attack squares
		BISHOP_MASKS[s] = (SquareToAntiDiagonal(s) ^ SquareToDiagonal(s)) & ^edgeMask

		// printBitBoard(bishopMasks[s])
		// number of movement possibilities from square s
		BISHOP_SHIFTS[s] = 64 - PopCount(BISHOP_MASKS[s])

		var i u64 = 0
		for {
			j := i
			j *= BISHOP_MAGICS[s]
			j >>= BISHOP_SHIFTS[s]
			BISHOP_ATTACKS[s][j] = slidingBishopAttacksForInitialization(s, i)
			i = (i - BISHOP_MASKS[s]) & BISHOP_MASKS[s]
			if i == 0 {
				break
			}
		}
	}
}

func slidingRookAttacksForInitialization(s Square, b u64) u64 {
	var attacks u64 = 0
	attacks |= SlidingAttacks(s, b, FILES[SquareToFile(s)])
	attacks |= SlidingAttacks(s, b, RANKS[SquareToRank(s)])
	return attacks
}

func InitRookAttacks() {
	for s := A1; s <= H8; s++ {
		var edgeMask u64 = (FILES[A] | FILES[H]) & ^FILES[SquareToFile(s)]
		edgeMask |= (RANKS[R1] | RANKS[R8]) & ^RANKS[SquareToRank(s)]

		// remove board edges from attack squares
		ROOK_MASKS[s] = (FILES[SquareToFile(s)] ^ RANKS[SquareToRank(s)]) & ^edgeMask

		// number of movement possibilities from square s
		ROOK_SHIFTS[s] = 64 - PopCount(ROOK_MASKS[s])

		var i u64 = 0
		for {
			j := i
			j *= ROOK_MAGICS[s]
			j >>= ROOK_SHIFTS[s]
			ROOK_ATTACKS[s][j] = slidingRookAttacksForInitialization(s, i)
			i = (i - ROOK_MASKS[s]) & ROOK_MASKS[s]
			if i == 0 {
				break
			}
		}
	}
}

// helpful tables found in nkarve/surge
var SQUARES_BETWEEN [64][64]u64

func InitSquaresBetween() {
	var squares u64
	for i := A1; i <= H8; i++ {
		for j := A1; j <= H8; j++ {
			squares = 1<<i | 1<<j
			if (SquareToFile(i) == SquareToFile(j)) || (SquareToRank(i) == SquareToRank(j)) {
				SQUARES_BETWEEN[i][j] =
					slidingRookAttacksForInitialization(i, squares) & slidingRookAttacksForInitialization(j, squares)
			} else if (SquareToDiagonal(i) == SquareToDiagonal(j)) || (SquareToAntiDiagonal(i) == SquareToAntiDiagonal(j)) {
				SQUARES_BETWEEN[i][j] =
					slidingBishopAttacksForInitialization(i, squares) & slidingBishopAttacksForInitialization(j, squares)
			}
		}
	}
}

var LINE [64][64]u64

func InitLine() {
	for i := A1; i <= H8; i++ {
		for j := A1; j <= H8; j++ {
			if (SquareToFile(i) == SquareToFile(j)) || (SquareToRank(i) == SquareToRank(j)) {
				LINE[i][j] =
					(slidingRookAttacksForInitialization(i, 0) & slidingRookAttacksForInitialization(j, 0)) | 1<<i | 1<<j
			} else if (SquareToDiagonal(i) == SquareToDiagonal(j)) || (SquareToAntiDiagonal(i) == SquareToAntiDiagonal(j)) {
				LINE[i][j] =
					(slidingBishopAttacksForInitialization(i, 0) & slidingBishopAttacksForInitialization(j, 0)) | 1<<i | 1<<j
			}
		}
	}
}

const (
	WHITE_K_CASTLE = uint8(1 << 0)
	WHITE_Q_CASTLE = uint8(1 << 1)
	BLACK_K_CASTLE = uint8(1 << 2)
	BLACK_Q_CASTLE = uint8(1 << 3)
)

var ZOBRIST_TABLE [12][64]u64
var CASTLING_HASH [16]u64
var EP_FILE_HASH [8]u64
var TURN_HASH u64

func InitZobrist() {
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 12; i++ {
		for j := 0; j < 64; j++ {
			ZOBRIST_TABLE[i][j] = u64(r.Uint64())
		}
	}
	for i := 0; i < 16; i++ {
		CASTLING_HASH[i] = u64(r.Uint64())
	}

	for i := 0; i < 8; i++ {
		EP_FILE_HASH[i] = u64(r.Uint64())
	}
	TURN_HASH = u64(r.Uint64())
}

// Some slice functions for unit testing and possibly other things in the engine
func Contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CheckSameElements(a, b []string) bool {
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

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
