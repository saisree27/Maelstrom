package engine

// Values for material and checkmate
const WIN_VAL int = 1000000

var MG_VALUES = []int{82, 365, 337, 477, 1025, 0} // pawn, bishop, knight, rook, queen, king
var EG_VALUES = []int{94, 297, 281, 512, 936, 0}
var PHASE_VALUES = []int{0, 1, 1, 2, 4, 0}

var REVERSE_SQUARE = [64]int{
	56, 57, 58, 59, 60, 61, 62, 63,
	48, 49, 50, 51, 52, 53, 54, 55,
	40, 41, 42, 43, 44, 45, 46, 47,
	32, 33, 34, 35, 36, 37, 38, 39,
	24, 25, 26, 27, 28, 29, 30, 31,
	16, 17, 18, 19, 20, 21, 22, 23,
	8, 9, 10, 11, 12, 13, 14, 15,
	0, 1, 2, 3, 4, 5, 6, 7,
}

// PeSTO evaluation tables
var PAWN_ST_MG = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	-35, -1, -20, -23, -15, 24, 38, -22,
	-26, -4, -4, -10, 3, 3, 33, -12,
	-27, -2, -5, 12, 17, 6, 10, -25,
	-14, 13, 6, 21, 23, 12, 17, -23,
	-6, 7, 26, 31, 65, 56, 25, -20,
	98, 134, 61, 95, 68, 126, 34, -11,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var PAWN_ST_EG = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	13, 8, 8, 10, 13, 0, 2, -7,
	4, 7, -6, 1, 0, -5, -1, -8,
	13, 9, -3, -7, -7, -8, 3, -1,
	32, 24, 13, 5, -2, 4, 17, 17,
	94, 100, 85, 67, 56, 53, 82, 84,
	178, 173, 158, 134, 147, 132, 165, 187,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var KNIGHT_ST_MG = [64]int{
	-105, -21, -58, -33, -17, -28, -19, -23,
	-29, -53, -12, -3, -1, 18, -14, -19,
	-23, -9, 12, 10, 19, 17, 25, -16,
	-13, 4, 16, 13, 28, 19, 21, -8,
	-9, 17, 19, 53, 37, 69, 18, 22,
	-47, 60, 37, 65, 84, 129, 73, 44,
	-73, -41, 72, 36, 23, 62, 7, -17,
	-167, -89, -34, -49, 61, -97, -15, -107,
}

var KNIGHT_ST_EG = [64]int{
	-29, -51, -23, -15, -22, -18, -50, -64,
	-42, -20, -10, -5, -2, -20, -23, -44,
	-23, -3, -1, 15, 10, -3, -20, -22,
	-18, -6, 16, 25, 16, 17, 4, -18,
	-17, 3, 22, 22, 22, 11, 8, -18,
	-24, -20, 10, 9, -1, -9, -19, -41,
	-25, -8, -25, -2, -9, -25, -24, -52,
	-58, -38, -13, -28, -31, -27, -63, -99,
}

var BISHOP_ST_MG = [64]int{
	-33, -3, -14, -21, -13, -12, -39, -21,
	4, 15, 16, 0, 7, 21, 33, 1,
	0, 15, 15, 15, 14, 27, 18, 10,
	-6, 13, 13, 26, 34, 12, 10, 4,
	-4, 5, 19, 50, 37, 37, 7, -2,
	-16, 37, 43, 40, 35, 50, 37, -2,
	-26, 16, -18, -13, 30, 59, 18, -47,
	-29, 4, -82, -37, -25, -42, 7, -8,
}

var BISHOP_ST_EG = [64]int{
	-23, -9, -23, -5, -9, -16, -5, -17,
	-14, -18, -7, -1, 4, -9, -15, -27,
	-12, -3, 8, 10, 13, 3, -7, -15,
	-6, 3, 13, 19, 7, 10, -3, -9,
	-3, 9, 12, 9, 14, 10, 3, 2,
	2, -8, 0, -1, -2, 6, 0, 4,
	-8, -4, 7, -12, -3, -13, -4, -14,
	-14, -21, -11, -8, -7, -9, -17, -24,
}

var ROOK_ST_MG = [64]int{
	-19, -13, 1, 17, 16, 7, -37, -26,
	-44, -16, -20, -9, -1, 11, -6, -71,
	-45, -25, -16, -17, 3, 0, -5, -33,
	-36, -26, -12, -1, 9, -7, 6, -23,
	-24, -11, 7, 26, 24, 35, -8, -20,
	-5, 19, 26, 36, 17, 45, 61, 16,
	27, 32, 58, 62, 80, 67, 26, 44,
	32, 42, 32, 51, 63, 9, 31, 43,
}

var ROOK_ST_EG = [64]int{
	-9, 2, 3, -1, -5, -13, 4, -20,
	-6, -6, 0, 2, -9, -9, -11, -3,
	-4, 0, -5, -1, -7, -12, -8, -16,
	3, 5, 8, 4, -5, -6, -8, -11,
	4, 3, 13, 1, 2, 1, -1, 2,
	7, 7, 7, 5, 4, -3, -5, -3,
	11, 13, 13, 11, -3, 3, 8, 3,
	13, 10, 18, 15, 12, 12, 8, 5,
}

var QUEEN_ST_MG = [64]int{
	-1, -18, -9, 10, -15, -25, -31, -50,
	-35, -8, 11, 2, 8, 15, -3, 1,
	-14, 2, -11, -2, -5, 2, 14, 5,
	-9, -26, -9, -10, -2, -4, 3, -3,
	-27, -27, -16, -16, -1, 17, -2, 1,
	-13, -17, 7, 8, 29, 56, 47, 57,
	-24, -39, -5, 1, -16, 57, 28, 54,
	-28, 0, 29, 12, 59, 44, 43, 45,
}

var QUEEN_ST_EG = [64]int{
	-33, -28, -22, -43, -5, -32, -20, -41,
	-22, -23, -30, -16, -16, -23, -36, -32,
	-16, -27, 15, 6, 9, 17, 10, 5,
	-18, 28, 19, 47, 31, 34, 39, 23,
	3, 22, 24, 45, 57, 40, 57, 36,
	-20, 6, 9, 49, 47, 35, 19, 9,
	-17, 20, 32, 41, 58, 25, 30, 0,
	-9, 22, 22, 27, 27, 19, 10, 20,
}

var KING_ST_MG = [64]int{
	-15, 36, 12, -54, 8, -28, 24, 14,
	1, 7, -8, -64, -43, -16, 9, 8,
	-14, -14, -22, -46, -44, -30, -15, -27,
	-49, -1, -27, -39, -46, -44, -33, -51,
	-17, -20, -12, -27, -30, -25, -14, -36,
	-9, 24, 2, -16, -20, 6, 22, -22,
	29, -1, -20, -7, -8, -4, -38, -29,
	-65, 23, 16, -15, -56, -34, 2, 13,
}

var KING_ST_EG = [64]int{
	-53, -34, -21, -11, -28, -14, -24, -43,
	-27, -11, 4, 13, 14, 4, -5, -17,
	-19, -3, 11, 21, 23, 16, 7, -9,
	-18, -4, 21, 24, 27, 23, 9, -11,
	-8, 22, 24, 27, 26, 33, 26, 3,
	10, 17, 23, 15, 20, 45, 44, 13,
	-12, 17, 14, 17, 17, 38, 23, 11,
	-74, -35, -18, -18, -11, 15, 4, -17,
}

type PieceTables struct {
	MG [64]int
	EG [64]int
}

var PIECE_TABLES = []PieceTables{
	{MG: PAWN_ST_MG, EG: PAWN_ST_EG},     // index 0: pawn
	{MG: BISHOP_ST_MG, EG: BISHOP_ST_EG}, // index 1: bishop
	{MG: KNIGHT_ST_MG, EG: KNIGHT_ST_EG}, // index 2: knight
	{MG: ROOK_ST_MG, EG: ROOK_ST_EG},     // index 3: rook
	{MG: QUEEN_ST_MG, EG: QUEEN_ST_EG},   // index 4: queen
	{MG: KING_ST_MG, EG: KING_ST_EG},     // index 5: king
}

// Mobility bonus values for each piece (values from Blunder)
var MOBILITY_BONUS_MG = [6]int{0, 3, 5, 3, 0, 0}
var MOBILITY_BONUS_EG = [6]int{0, 2, 3, 2, 6, 0}

// Pawn structure penalties and masks (isolated/doubled/passed)
var ISOLATED_PAWN_MASKS [8]u64
var DOUBLED_PAWN_MASKS [2][64]u64
var PASSED_PAWN_MASKS [2][64]u64

const ISOLATED_PAWN_PENALTY_MG = 14
const ISOLATED_PAWN_PENALTY_EG = 3
const DOUBLED_PAWN_PENALTY_MG = 1
const DOUBLED_PAWN_PENALTY_EG = 20

// Passed pawn bonus tables (values from Blunder)
var PASSED_PAWN_ST_MG = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	8, 9, 2, -8, -3, 8, 16, 9,
	5, 3, -3, -14, -3, 10, 13, 19,
	14, 0, -9, -7, -13, -7, 9, 16,
	28, 17, 13, 10, 10, 19, 6, 1,
	48, 43, 43, 30, 24, 31, 12, 2,
	45, 52, 42, 43, 28, 34, 19, 9,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var PASSED_PAWN_ST_EG = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	2, 3, -4, 0, -2, -1, 7, 6,
	8, 6, 5, 1, 1, -1, 14, 7,
	29, 26, 21, 18, 17, 19, 34, 30,
	55, 52, 42, 35, 30, 34, 56, 52,
	91, 83, 66, 40, 30, 61, 67, 84,
	77, 74, 63, 53, 59, 60, 72, 77,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// King safety evaluation penalties
const PAWN_SHIELD_LEFT_PENALTY = 5
const PAWN_SHIELD_RIGHT_PENALTY = 10
const PAWN_SHIELD_MIDDLE_PENALTY = 12
const NOT_CASTLED_PENALTY = 15

// Bishop pair bonus
const BISHOP_PAIR_BONUS_MG = 20
const BISHOP_PAIR_BONUS_EG = 30

// Returns an evaluation of the position in cp
// 1000000 or -1000000 is designated as checkmate
// Evaluations are returned in White's perspective

func Evaluate(b *Board) int {
	if b.IsThreeFoldRep() || b.IsInsufficientMaterial() {
		return 0
	}

	mgEval := 0
	egEval := 0
	totalPhase := 0

	for piece := PAWN; piece <= KING; piece++ {
		idx := piece
		tbls := PIECE_TABLES[idx]
		mgVal := MG_VALUES[idx]
		egVal := EG_VALUES[idx]
		phase := PHASE_VALUES[idx]

		wpieces := b.GetColorPieces(piece, WHITE)
		bpieces := b.GetColorPieces(piece, BLACK)

		mobilityMG := MOBILITY_BONUS_MG[piece]
		mobilityEG := MOBILITY_BONUS_EG[piece]

		bishopCounter := 0
		for wpieces != 0 {
			totalPhase += phase
			sq := Square(PopLSB(&wpieces))
			mgEval += mgVal + tbls.MG[sq]
			egEval += egVal + tbls.EG[sq]

			moves := 0
			if piece == BISHOP {
				bishopCounter += 1
				moves = PopCount(BishopAttacks(sq, b.occupied))
			} else if piece == KNIGHT {
				moves = PopCount(KnightAttacks(sq))
			} else if piece == ROOK {
				moves = PopCount(RookAttacks(sq, b.occupied))
			} else if piece == QUEEN {
				moves = PopCount(BishopAttacks(sq, b.occupied) | RookAttacks(sq, b.occupied))
			} else if piece == PAWN {
				m, e := evaluatePawn(b, sq, WHITE)
				mgEval += m
				egEval += e
			} else if piece == KING {
				m := evaluateKingSafetyWhite(b, sq)
				mgEval += m
			}

			mgEval += moves * mobilityMG
			egEval += moves * mobilityEG
		}

		if bishopCounter >= 2 {
			mgEval += BISHOP_PAIR_BONUS_MG
			egEval += BISHOP_PAIR_BONUS_EG
		}

		bishopCounter = 0
		for bpieces != 0 {
			totalPhase += phase
			sq := Square(PopLSB(&bpieces))
			rsq := REVERSE_SQUARE[sq]
			mgEval -= mgVal + tbls.MG[rsq]
			egEval -= egVal + tbls.EG[rsq]

			moves := 0
			if piece == BISHOP {
				bishopCounter += 1
				moves = PopCount(BishopAttacks(sq, b.occupied))
			} else if piece == KNIGHT {
				moves = PopCount(KnightAttacks(sq))
			} else if piece == ROOK {
				moves = PopCount(RookAttacks(sq, b.occupied))
			} else if piece == QUEEN {
				moves = PopCount(BishopAttacks(sq, b.occupied) | RookAttacks(sq, b.occupied))
			} else if piece == PAWN {
				m, e := evaluatePawn(b, sq, BLACK)
				mgEval -= m
				egEval -= e
			} else if piece == KING {
				m := evaluateKingSafetyBlack(b, sq)
				mgEval -= m
			}

			mgEval -= moves * mobilityMG
			egEval -= moves * mobilityEG
		}

		if bishopCounter >= 2 {
			mgEval -= BISHOP_PAIR_BONUS_MG
			egEval -= BISHOP_PAIR_BONUS_EG
		}
	}

	mgPhase := totalPhase
	if mgPhase > 24 {
		mgPhase = 24
	}
	egPhase := 24 - mgPhase

	return (mgEval*mgPhase + egEval*egPhase) / 24
}

func evaluatePawn(b *Board, sq Square, color Color) (int, int) {
	ourPawns := b.GetColorPieces(PAWN, color)
	enemyPawns := b.GetColorPieces(PAWN, ReverseColor(color))
	mg := 0
	eg := 0

	if ISOLATED_PAWN_MASKS[SquareToFile(sq)]&ourPawns == 0 {
		mg -= ISOLATED_PAWN_PENALTY_MG
		eg -= ISOLATED_PAWN_PENALTY_EG
	}

	if DOUBLED_PAWN_MASKS[color][sq]&ourPawns != 0 {
		mg -= DOUBLED_PAWN_PENALTY_MG
		eg -= DOUBLED_PAWN_PENALTY_EG
	}

	if PASSED_PAWN_MASKS[color][sq]&enemyPawns == 0 && ourPawns&DOUBLED_PAWN_MASKS[color][sq] == 0 {
		mg += PASSED_PAWN_ST_MG[sq]
		eg += PASSED_PAWN_ST_EG[sq]
	}

	return mg, eg
}

func evaluateKingSafetyWhite(b *Board, sq Square) int {
	if sq == E1 {
		return -NOT_CASTLED_PENALTY
	}

	score := 0
	pawns := b.pieces[W_P] | b.pieces[W_B] // include bishops in the case of fianchetto
	bb := SQUARE_TO_BITBOARD[sq]

	nw := ShiftBitboard(bb, NW) & pawns
	n := ShiftBitboard(bb, NORTH) & pawns
	ne := ShiftBitboard(bb, NE) & pawns

	onKingside := (b.pieces[W_K] & KINGSIDE) != 0

	if nw == 0 {
		if onKingside {
			score -= PAWN_SHIELD_LEFT_PENALTY
		} else {
			score -= PAWN_SHIELD_RIGHT_PENALTY
		}
	}
	if ne == 0 {
		if onKingside {
			score -= PAWN_SHIELD_RIGHT_PENALTY
		} else {
			score -= PAWN_SHIELD_LEFT_PENALTY
		}
	}
	if n == 0 {
		score -= PAWN_SHIELD_MIDDLE_PENALTY
	}

	return score
}

func evaluateKingSafetyBlack(b *Board, sq Square) int {
	if sq == E8 {
		return -NOT_CASTLED_PENALTY
	}

	score := 0
	pawns := b.pieces[B_P] | b.pieces[B_B] // include bishops for the case of fianchetto
	bb := SQUARE_TO_BITBOARD[sq]
	sw := ShiftBitboard(bb, SW) & pawns
	s := ShiftBitboard(bb, SOUTH) & pawns
	se := ShiftBitboard(bb, SE) & pawns

	onKingside := (b.pieces[B_K] & KINGSIDE) != 0

	if se == 0 {
		if onKingside {
			score -= PAWN_SHIELD_LEFT_PENALTY
		} else {
			score -= PAWN_SHIELD_RIGHT_PENALTY
		}
	}
	if sw == 0 {
		if onKingside {
			score -= PAWN_SHIELD_LEFT_PENALTY
		} else {
			score -= PAWN_SHIELD_RIGHT_PENALTY
		}
	}
	if s == 0 {
		score -= PAWN_SHIELD_MIDDLE_PENALTY
	}

	return score
}

func InitializePawnMasks() {
	for file := 0; file < 8; file++ {
		mask := u64(0)
		if file > 0 {
			mask |= FILES[file-1]
		}
		if file < 7 {
			mask |= FILES[file+1]
		}
		ISOLATED_PAWN_MASKS[file] = mask
	}

	for sq := 0; sq < 64; sq++ {
		file := sq & 7
		rank := sq >> 3
		maskW := u64(0)
		maskB := u64(0)
		for r := rank + 1; r < 8; r++ {
			maskW |= 1 << (r*8 + file)
		}
		for r := 0; r < rank; r++ {
			maskB |= 1 << (r*8 + file)
		}
		DOUBLED_PAWN_MASKS[WHITE][sq] = maskW
		DOUBLED_PAWN_MASKS[BLACK][sq] = maskB
	}

	for sq := 0; sq < 64; sq++ {
		file := sq & 7
		rank := sq >> 3

		maskW := u64(0)
		maskB := u64(0)

		for f := file - 1; f <= file+1; f++ {
			if f < 0 || f > 7 {
				continue
			}
			for r := rank + 1; r < 8; r++ {
				maskW |= 1 << (r*8 + f)
			}
			for r := 0; r < rank; r++ {
				maskB |= 1 << (r*8 + f)
			}
		}
		PASSED_PAWN_MASKS[WHITE][sq] = maskW
		PASSED_PAWN_MASKS[BLACK][sq] = maskB
	}
}
