package engine

// Values for material and checkmate
const winVal int = 1000000
const pawnVal int = 100
const knightVal int = 350
const bishopVal int = 350
const rookVal int = 525
const queenVal int = 1000
const kingVal int = 100000

// Values for rook open/semi-open file
const rookOpenFile int = 15
const rookSemiOpenFile int = 7
const twoRooksOnSeventh int = 15

// Values for pawn structure
var doubledPawnByFile = []int{
	A: -25, B: -5, C: -30, D: -20, E: -20, F: -20, G: -5, H: -20,
}

const tripledPawn int = -50
const isolatedPawn int = -15
const doubledAndIsolated int = -35
const isolatedPawnBlocked int = -15
const passedPawn int = 15
const phalanx int = 30
const passedPawnBlocked int = -20
const cdPawnBlockedByPlayer int = -50

var passedPawnRankWhite = []int{
	-5, -5, 5, 5, 25, 45, 150, 0,
}

var passedPawnRankBlack = []int{
	0, 150, 45, 25, 5, 5, -5, -5,
}

// Values for other pieces
const queenEarly int = -20
const queensNotTradedWhenNotCastled = 15
const bishopPair int = 45
const bishopMobility int = 2
const rookMobility int = 4
const queenMobility int = 1

// Values for king safety
const pawnShieldLeft int = -15
const pawnShieldUpDown int = -50
const pawnShieldRight int = -15
const kingAir int = -10
const notCastled int = -30
const cannotCastle int = -80

const samePieceTwice int = -15
const piecesOnBackRank int = -15

var factor = []int{
	WHITE: 1, BLACK: -1,
}

var material = map[Piece]int{
	wP: pawnVal, bP: -pawnVal,
	wN: knightVal, bN: -knightVal,
	wB: bishopVal, bB: -bishopVal,
	wR: rookVal, bR: -rookVal,
	wQ: queenVal, bQ: -queenVal,
	wK: kingVal, bK: -kingVal,
	EMPTY: 0,
}

var center = map[Square]bool{
	e4: true, d4: true, e5: true, d5: true,
}

var reversePSQ = [64]int{
	56, 57, 58, 59, 60, 61, 62, 63,
	48, 49, 50, 51, 52, 53, 54, 55,
	40, 41, 42, 43, 44, 45, 46, 47,
	32, 33, 34, 35, 36, 37, 38, 39,
	24, 25, 26, 27, 28, 29, 30, 31,
	16, 17, 18, 19, 20, 21, 22, 23,
	8, 9, 10, 11, 12, 13, 14, 15,
	0, 1, 2, 3, 4, 5, 6, 7,
}

var pawnSquareTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, -10, -20, -20, 10, 10, 5,
	5, 5, 5, 0, 0, -10, 5, 5,
	0, 0, 10, 20, 20, 0, 0, 0,
	5, 5, 10, 25, 25, 10, 5, 5,
	10, 10, 20, 30, 30, 20, 10, 10,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightSquareTable = [64]int{
	-50, -30, -30, -30, -30, -30, -30, -50,
	-40, -20, 0, -5, -5, 0, -20, -40,
	-40, 0, 10, 15, 15, 10, 0, -40,
	-50, 5, 15, 20, 20, 15, 5, -50,
	-45, 0, 15, 20, 20, 15, 0, -45,
	-50, 5, 10, 15, 15, 10, 5, -50,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopSquareTable = [64]int{
	-20, -10, -5, -10, -10, -10, -10, -20,
	-10, 10, 0, 0, 0, 0, 10, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookSquareTable = [64]int{
	-5, 0, 0, 5, 5, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var queenSquareTable = [64]int{
	-20, -10, -10, 5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, -5, -5, -5, -5, -5, 0, -10,
	0, 0, 5, 5, 5, 5, 0, -5,
	-5, 0, 5, 5, 5, 5, 0, -5,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingSquareTableMiddlegame = [64]int{
	0, 30, 10, 0, 0, 10, 30, 0,
	-30, -30, -30, -30, -30, -30, -30, -30,
	-50, -50, -50, -50, -50, -50, -50, -50,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
	-70, -70, -70, -70, -70, -70, -70, -70,
}

var kingSquareTableEndgame = [64]int{
	-50, -10, 0, 0, 0, 0, -10, -50,
	-10, 0, 10, 10, 10, 10, 0, -10,
	0, 10, 15, 15, 15, 15, 10, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 10, 15, 20, 20, 15, 10, 0,
	0, 10, 15, 15, 15, 15, 10, 0,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-50, -10, 0, 0, 0, 0, -10, -50,
}

// Returns an evaluation of the position in cp
// 1000000 or -1000000 is designated as checkmate
// Evaluations are returned in White's perspective

func evaluate(b *Board) int {
	moves := b.generateLegalMoves()
	if len(moves) == 0 {
		if b.isCheck(b.turn) {
			return winVal * factor[reverseColor(b.turn)]
		} else {
			return 0
		}
	}

	if b.isThreeFoldRep() {
		return 0
	}

	if b.isInsufficientMaterial() {
		return 0
	}

	eval := 0

	material, total := totalMaterialAndPieces(b)
	eval += material

	evaluateKnights(b, &eval)
	evaluateBishops(b, &eval)
	evaluateRooks(b, &eval)
	evaluateQueens(b, &eval)
	evaluateKings(b, &eval, total)
	evaluatePawns(b, &eval)

	if b.plyCnt <= 20 {
		length := len(b.history)
		for i := length - 3; i > 0; i-- {
			if b.history[i+2].move.from == b.history[i].move.to {
				if reverseColor(b.turn) == WHITE && b.history[i+2].move.piece != wP {
					eval += samePieceTwice
				} else if b.history[i+2].move.piece != bP {
					eval -= samePieceTwice
				}
			}
		}
	}

	if b.plyCnt <= 25 {
		wPiecesOnBackRank := (b.colors[WHITE] ^ b.pieces[wR]) & ranks[R1]
		bPiecesOnBackRank := (b.colors[BLACK] ^ b.pieces[bR]) & ranks[R8]

		eval += (popCount(wPiecesOnBackRank) - popCount(bPiecesOnBackRank)) * piecesOnBackRank
	}

	if b.plyCnt >= 25 {
		if !b.whiteCastled {
			eval += notCastled
		}
		if !b.blackCastled {
			eval -= notCastled
		}
	}

	if !b.whiteCastled && b.pieces[bQ] != 0 {
		eval -= queensNotTradedWhenNotCastled
	}
	if !b.blackCastled && b.pieces[wQ] != 0 {
		eval += queensNotTradedWhenNotCastled
	}

	return eval
}

func totalMaterialAndPieces(b *Board) (int, int) {
	sum := 0
	total := 0

	// Use bitboard operations for faster material counting
	for piece := wP; piece <= bK; piece++ {
		count := popCount(b.pieces[piece])
		sum += material[piece] * count
		total += count
	}
	return sum, total
}

func evaluatePawns(b *Board, eval *int) {
	// TODO: Add pawn hash table to reduce cost of doing this entire method
	wPawnsOrig := b.getColorPieces(pawn, WHITE)
	bPawnsOrig := b.getColorPieces(pawn, BLACK)

	wPawns := wPawnsOrig
	bPawns := bPawnsOrig

	filesFoundWhite := [8]int{}
	filesFoundBlack := [8]int{}

	gotPhalanx := false
	for wPawns != 0 {
		square := Square(popLSB(&wPawns))
		*eval += pawnSquareTable[square]

		// Doubled pawns check
		file := sqToFile(square)
		filesFoundWhite[file]++

		neighborPlayerPawns := fileNeighbors[file] & wPawnsOrig
		// Isolated pawns check
		if neighborPlayerPawns == 0 {
			if filesFoundWhite[file] >= 2 {
				*eval += doubledAndIsolated
			} else {
				*eval += isolatedPawn
			}
			if b.squares[square+Square(NORTH)].getColor() == BLACK {
				*eval += isolatedPawnBlocked
			}
		} else if !gotPhalanx && sqToRank(square) >= R4 && sqToFile(square) >= C && sqToFile(square) <= F {
			for neighborPlayerPawns != 0 {
				sq := Square(popLSB(&neighborPlayerPawns))
				if sqToRank(sq) == sqToRank(square) {
					gotPhalanx = true
					*eval += phalanx
				}
			}
		}

		// Passed pawns check
		neighborPawns := (files[file] | fileNeighbors[file]) & bPawnsOrig
		if neighborPawns == 0 {
			*eval += passedPawn
			*eval += passedPawnRankWhite[sqToRank(square)]
		} else {
			passedAhead := true
			for neighborPawns != 0 {
				p := Square(popLSB(&neighborPawns))
				if sqToRank(p) > sqToRank(square) {
					passedAhead = false
				}
			}

			if passedAhead {
				*eval += passedPawn
				*eval += passedPawnRankWhite[sqToRank(square)]
			}
		}
	}

	gotPhalanx = false
	for bPawns != 0 {
		square := Square(popLSB(&bPawns))
		*eval -= pawnSquareTable[reversePSQ[square]]

		// Doubled pawns check
		file := sqToFile(square)
		filesFoundBlack[file]++

		neighborPlayerPawns := fileNeighbors[file] & bPawnsOrig
		// Isolated pawns check
		if neighborPlayerPawns == 0 {
			if filesFoundBlack[file] >= 2 {
				*eval -= doubledAndIsolated
			} else {
				*eval -= isolatedPawn
			}
			if b.squares[square+Square(SOUTH)].getColor() == WHITE {
				*eval -= isolatedPawnBlocked
			}
		} else if !gotPhalanx && sqToRank(square) <= R5 && sqToFile(square) >= C && sqToFile(square) <= F {
			for neighborPlayerPawns != 0 {
				sq := Square(popLSB(&neighborPlayerPawns))
				if sqToRank(sq) == sqToRank(square) {
					gotPhalanx = true
					*eval -= phalanx
				}
			}
		}

		neighborPawns := (files[file] | fileNeighbors[file]) & wPawnsOrig
		// Passed pawns check
		if neighborPawns == 0 {
			*eval -= passedPawn
			*eval -= passedPawnRankBlack[sqToRank(square)]
		} else {
			passedAhead := true
			for neighborPawns != 0 {
				p := Square(popLSB(&neighborPawns))
				if sqToRank(p) < sqToRank(square) {
					passedAhead = false
				}
			}

			if passedAhead {
				*eval -= passedPawn
				*eval -= passedPawnRankBlack[sqToRank(square)]
			}
		}
	}

	// Assign penalties for doubled and tripled pawns
	for i := A; i <= H; i++ {
		if filesFoundWhite[i] == 2 {
			*eval += doubledPawnByFile[i]
		}
		if filesFoundWhite[i] == 3 {
			*eval += tripledPawn
		}
		if filesFoundBlack[i] == 2 {
			*eval -= doubledPawnByFile[i]
		}
		if filesFoundBlack[i] == 3 {
			*eval -= tripledPawn
		}
	}
}

func evaluateKnights(b *Board, eval *int) {
	wKnights := b.getColorPieces(knight, WHITE)
	bKnights := b.getColorPieces(knight, BLACK)
	for wKnights != 0 {
		square := Square(popLSB(&wKnights))

		if square == c3 && b.squares[c2] == wP {
			*eval += cdPawnBlockedByPlayer
		}

		*eval += knightSquareTable[square]
	}
	for bKnights != 0 {
		square := Square(popLSB(&bKnights))

		if square == c6 && b.squares[c7] == bP {
			*eval -= cdPawnBlockedByPlayer
		}

		*eval -= knightSquareTable[reversePSQ[square]]
	}
}

func evaluateBishops(b *Board, eval *int) {
	// TODO: Add specific evaluation for bishops
	wBishops := b.getColorPieces(bishop, WHITE)
	wNum := 0
	bBishops := b.getColorPieces(bishop, BLACK)
	bNum := 0
	for wBishops != 0 {
		square := Square(popLSB(&wBishops))

		if square == d3 && b.squares[d2] == wP {
			*eval += cdPawnBlockedByPlayer
		}

		attacks := getBishopAttacks(square, b.occupied)

		*eval += popCount(attacks) * bishopMobility
		*eval += bishopSquareTable[square]
		wNum++
	}

	if wNum >= 2 {
		*eval += bishopPair
	}

	for bBishops != 0 {
		square := Square(popLSB(&bBishops))

		if square == d6 && b.squares[d7] == bP {
			*eval -= cdPawnBlockedByPlayer
		}

		attacks := getBishopAttacks(square, b.occupied)

		*eval -= popCount(attacks) * bishopMobility
		*eval -= bishopSquareTable[reversePSQ[square]]
		bNum++
	}

	if bNum >= 2 {
		*eval -= bishopPair
	}
}

func evaluateRooks(b *Board, eval *int) {
	wRooks := b.getColorPieces(rook, WHITE)
	bRooks := b.getColorPieces(rook, BLACK)
	pawns := b.pieces[wP] | b.pieces[bP]
	for wRooks != 0 {
		square := Square(popLSB(&wRooks))
		attacks := getRookAttacks(square, b.occupied)
		piecesOnFile := files[sqToFile(square)] & pawns
		if piecesOnFile == 0 {
			*eval += rookOpenFile
		} else if piecesOnFile == 1 {
			*eval += rookSemiOpenFile
		}
		*eval += popCount(attacks) * rookMobility
		*eval += rookSquareTable[square]
	}
	for bRooks != 0 {
		square := Square(popLSB(&bRooks))
		attacks := getRookAttacks(square, b.occupied)
		piecesOnFile := files[sqToFile(square)] & pawns
		if piecesOnFile == 0 {
			*eval -= rookOpenFile
		} else if piecesOnFile == 1 {
			*eval -= rookSemiOpenFile
		}
		*eval -= popCount(attacks) * rookMobility
		*eval -= rookSquareTable[reversePSQ[square]]
	}
}

func evaluateQueens(b *Board, eval *int) {
	wQueens := b.getColorPieces(queen, WHITE)
	bQueens := b.getColorPieces(queen, BLACK)
	for wQueens != 0 {
		square := Square(popLSB(&wQueens))
		if square != d1 && b.plyCnt <= 15 {
			*eval += queenEarly
		}
		attacks := getBishopAttacks(square, b.occupied) | getRookAttacks(square, b.occupied)

		*eval += popCount(attacks) * queenMobility
		*eval += queenSquareTable[square]
	}
	for bQueens != 0 {
		square := Square(popLSB(&bQueens))
		if square != d8 && b.plyCnt <= 15 {
			*eval -= queenEarly
		}
		attacks := getBishopAttacks(square, b.occupied) | getRookAttacks(square, b.occupied)

		*eval -= popCount(attacks) * queenMobility
		*eval -= queenSquareTable[reversePSQ[square]]
	}
}

func evaluateKings(b *Board, eval *int, totalPieces int) {
	noQueensLeft := b.pieces[wQ]|b.pieces[bQ] == 0
	switchEndgame := totalPieces <= 25 && noQueensLeft

	wKing := Square(bitScanForward(b.getColorPieces(king, WHITE)))
	bKing := Square(bitScanForward(b.getColorPieces(king, BLACK)))

	if switchEndgame {
		*eval += kingSquareTableEndgame[wKing]
		*eval -= kingSquareTableEndgame[reversePSQ[bKing]]
	} else {
		*eval += kingSquareTableMiddlegame[wKing]
		*eval -= kingSquareTableMiddlegame[reversePSQ[bKing]]

		// White king:
		if wKing != e1 && !b.whiteCastled {
			*eval += cannotCastle
		}

		if sToBB[wKing]&ranks[R1] != 0 { // If first rank
			// Mask out squares in the same rank
			empty := b.empty // Cache empty squares bitboard
			attacks := kingAttacks(wKing) & empty
			attacks &= ^ranks[R1]
			airW := popCount(attacks)
			*eval += airW * kingAir
		}

		// Pawn shield
		pawns := b.pieces[wP]
		nw := shiftBitboard(sToBB[wKing], NW) & pawns
		n := shiftBitboard(sToBB[wKing], NORTH) & pawns
		ne := shiftBitboard(sToBB[wKing], NE) & pawns

		if nw == 0 {
			*eval += pawnShieldLeft
		}
		if n == 0 {
			*eval += pawnShieldUpDown
		}
		if ne == 0 {
			*eval += pawnShieldRight
		}

		// Black king:
		if bKing != e8 && !b.blackCastled {
			*eval -= cannotCastle
		}

		if sToBB[bKing]&ranks[R8] != 0 { // If first rank
			// Mask out squares in the same rank
			empty := b.empty // Cache empty squares bitboard
			attacks := kingAttacks(bKing) & empty
			attacks &= ^ranks[R8]
			airB := popCount(attacks)
			*eval -= airB * kingAir
		}

		// Pawn shield
		pawns = b.pieces[bP]
		sw := shiftBitboard(sToBB[bKing], SW) & pawns
		s := shiftBitboard(sToBB[bKing], SOUTH) & pawns
		se := shiftBitboard(sToBB[bKing], SE) & pawns

		if sw == 0 {
			*eval -= pawnShieldLeft
		}
		if s == 0 {
			*eval -= pawnShieldUpDown
		}
		if se == 0 {
			*eval -= pawnShieldRight
		}
	}
}
