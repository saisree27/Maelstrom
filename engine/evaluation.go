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
	A: -35, B: 0, C: -10, D: -20, E: -20, F: -10, G: 0, H: -35,
}
const tripledPawn int = -40
const isolatedPawn int = -5 
const passedPawn int = 20
const pawnBlockedByPlayer int = -10

// Values for other pieces
const queenEarly int = -30

// Values for king safety
// TODO

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

var center = map[Square]bool {
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
	5, 10, 10, -20, -20, 10, 10, 5,
	5, -5, -10, 0, 0, -10, -5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, 5, 10, 25, 25, 10, 5, 5,
	10, 10, 20, 30, 30, 20, 10, 10,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightSquareTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, -5, -5, 0, -20, -40,
	-50, 0, 10, 15, 15, 10, 0, -50,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopSquareTable = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 5, 0, 0, 0, 0, 5, -10,
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
	20, 30, 10, 0, 0, 10, 30, 20,
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

	evaluatePawns(b, &eval)
	evaluateKnights(b, &eval)
	evaluateBishops(b, &eval)
	evaluateQueens(b, &eval)
	evaluateKings(b, &eval, total)

	return eval
}

func totalMaterialAndPieces(b *Board) (int, int) {
	sum := 0
	total := 0

	for _, piece := range b.squares {
		sum += material[piece]
		if piece != EMPTY {
			total++
		}
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

	for wPawns != 0 {
		square := Square(popLSB(&wPawns))
		*eval += pawnSquareTable[square]
		
		if b.squares[square + Square(NORTH)].getColor() == WHITE {
			if square == c2 {
				*eval += pawnBlockedByPlayer
			}
		}

		// Doubled pawns check
		file := sqToFile(square)
		filesFoundWhite[file]++

		// Isolated pawns check
		if fileNeighbors[file] & wPawnsOrig == 0 {
			*eval += isolatedPawn
		}

		// Passed pawns check
		if (files[file] | fileNeighbors[file]) & bPawnsOrig == 0 {
			*eval += passedPawn
		}
	}

	for bPawns != 0 {
		square := Square(popLSB(&bPawns))
		*eval -= pawnSquareTable[reversePSQ[square]]

		if b.squares[square + Square(SOUTH)].getColor() == BLACK {
			if square == c7 {
				*eval -= pawnBlockedByPlayer
			}
		}

		// Doubled pawns check
		file := sqToFile(square)
		filesFoundBlack[file]++

		// Isolated pawns check
		if fileNeighbors[file] & bPawnsOrig == 0 {
			*eval -= isolatedPawn
		}

		// Passed pawns check
		if (files[file] | fileNeighbors[file]) & wPawnsOrig == 0 {
			*eval -= passedPawn
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
	// TODO: Add specific evaluation for knights
	wKnights := b.getColorPieces(knight, WHITE)
	bKnights := b.getColorPieces(knight, BLACK)
	for wKnights != 0 {
		square := popLSB(&wKnights)
		*eval += knightSquareTable[square]
	}
	for bKnights != 0 {
		square := popLSB(&bKnights)
		*eval -= knightSquareTable[reversePSQ[square]]
	}
}

func evaluateBishops(b *Board, eval *int) {
	// TODO: Add specific evaluation for bishops
	wBishops := b.getColorPieces(bishop, WHITE)
	bBishops := b.getColorPieces(bishop, BLACK)
	for wBishops != 0 {
		square := popLSB(&wBishops)
		*eval += bishopSquareTable[square]
	}
	for bBishops != 0 {
		square := popLSB(&bBishops)
		*eval -= bishopSquareTable[reversePSQ[square]]
	}
}

func evaluateRooks(b *Board, eval *int) {
	// TODO: Add specific evaluation for rooks
	wRooks := b.getColorPieces(rook, WHITE)
	bRooks := b.getColorPieces(rook, BLACK)
	for wRooks != 0 {
		square := popLSB(&wRooks)
		*eval += rookSquareTable[square]
	}
	for bRooks != 0 {
		square := popLSB(&bRooks)
		*eval -= rookSquareTable[reversePSQ[square]]
	}
}

func evaluateQueens(b *Board, eval *int) {
	// TODO: Add specific evaluation for queens
	wQueens := b.getColorPieces(queen, WHITE)
	bQueens := b.getColorPieces(queen, BLACK)
	for wQueens != 0 {
		square := popLSB(&wQueens)
		*eval += queenSquareTable[square]
	}
	for bQueens != 0 {
		square := popLSB(&bQueens)
		*eval -= queenSquareTable[reversePSQ[square]]
	}
}

func evaluateKings(b *Board, eval *int, totalPieces int) {
	// TODO: Add specific evaluation for kings
	noQueensLeft := b.pieces[wQ]|b.pieces[bQ] == 0
	switchEndgame := (totalPieces <= 15 && noQueensLeft) || (totalPieces <= 10)

	wKing := Square(bitScanForward(b.getColorPieces(king, WHITE)))
	bKing := Square(bitScanForward(b.getColorPieces(king, BLACK)))

	if switchEndgame {
		*eval += kingSquareTableEndgame[wKing]
		*eval -= kingSquareTableEndgame[reversePSQ[bKing]]
	} else {
		*eval += kingSquareTableMiddlegame[wKing]
		*eval -= kingSquareTableMiddlegame[reversePSQ[bKing]]
	}
}