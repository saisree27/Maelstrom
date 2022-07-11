package engine

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	initializeKingAttacks()
	initializeKnightAttacks()
	initializePawnAttacks()
	initBishopAttacks()
	initRookAttacks()
	initSquaresBetween()
	initLine()
	initializeSQLookup()
}

// Move generation testing suite. A lot of the test-case FENs are from this awesome repo:
// https://github.com/schnitzi/chessmovegen/tree/master/src/main/resources/testcases

// Test InitStartPos() method by making sure of correct default values
func TestStartPos(t *testing.T) {
	b := Board{}
	b.InitStartPos()

	name := "castlingRights"
	want := true
	got := b.kB && b.qB && b.kW && b.qW

	if got != want {
		t.Errorf("TestStartPos (%q): got %t, wanted %t", name, got, want)
	}

	name = "enpassant"
	want2 := EMPTYSQ
	got2 := b.enpassant

	if got2 != want2 {
		t.Errorf("TestStartPos (%q): got %q, wanted %q", name, squareToStringMap[got2], squareToStringMap[want2])
	}
}

// Test enpassant when in pseudo-pin
func TestEnPassantPseudoPin(t *testing.T) {
	b := Board{}
	b.InitFEN("rnbq1bnr/ppp1pppp/8/8/k2p3R/8/PPPPPPPP/RNBQKBN1 w - - 0 1")
	b.makeMoveFromUCI("e2e4")

	if b.enpassant != e3 {
		t.Errorf("TestEnPassantPseudoPin (square): got %q, wanted %q", squareToStringMap[b.enpassant], squareToStringMap[e3])
	}

	moves := b.generateLegalMoves()

	uciMoves := []interface{}{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	if contains(uciMoves, "d4e3") != false {
		t.Errorf("TestEnPassantPseudoPin (enpassant move): got %t, wanted %t", contains(uciMoves, "d4e3"), false)
	}
}

// Two en passant options allowed
func TestTwoEnPassant(t *testing.T) {
	fen := "7k/8/8/8/pPp5/8/8/7K b - b3 0 1"
	b := Board{}
	b.InitFEN(fen)

	moves := b.generateLegalMoves()
	uciMoves := []string{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	actualMoves := []string{
		"a4a3", "c4c3", "a4b3", "c4b3", "h8h7", "h8g7", "h8g8",
	}

	result := checkSameElements(actualMoves, uciMoves)
	if result == false {
		t.Errorf("TestTwoEnPassant: got %v, wanted %v", uciMoves, actualMoves)
	}
}

// Two en passant options but one of the pawns is pinned
func TestTwoEnpassantOneLegal(t *testing.T) {
	fen := "8/8/4k3/8/2pPp3/8/B7/7K b - d3 0 1"
	b := Board{}
	b.InitFEN(fen)

	moves := b.generateLegalMoves()
	uciMoves := []string{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	actualMoves := []string{
		"e4e3", "e4d3", "e6d5", "e6f5", "e6d6", "e6f6", "e6d7", "e6e7", "e6f7",
	}

	result := checkSameElements(actualMoves, uciMoves)
	if result == false {
		t.Errorf("TestTwoEnPassant: got %v, wanted %v", uciMoves, actualMoves)
	}
}

// All 8 pawns present for both sides but no pawn moves
func TestNoPawnMoves(t *testing.T) {
	fen := "8/4k3/1p1p1p1p/pPpPpPpP/P1P1P1P1/8/5K2/8 w - - 0 1"
	b := Board{}
	b.InitFEN(fen)

	moves := b.generateLegalMoves()
	uciMoves := []string{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	actualMoves := []string{
		"f2e3", "f2f3", "f2g3", "f2e2", "f2g2", "f2e1", "f2f1", "f2g1",
	}

	result := checkSameElements(actualMoves, uciMoves)
	if result == false {
		t.Errorf("TestTwoEnPassant: got %v, wanted %v", uciMoves, actualMoves)
	}

}

// Test that castling is not possible when squares between are attacked
func TestCastlingIfSquaresAttacked(t *testing.T) {
	b := Board{}
	b.InitFEN("rnbq1rk1/pppp1ppp/5n2/2b1p3/2B1P3/5P2/PPPPN1PP/RNBQK2R w KQ - 5 5")

	moves := b.generateLegalMoves()

	uciMoves := []interface{}{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	if contains(uciMoves, "e1g1") != false {
		t.Errorf("TestCastlingIfSquaresAttacked: got %t, wanted %t", contains(uciMoves, "e1g1"), false)
	}
}

// Test that all moves are correct on a complex position
func TestAllMoves(t *testing.T) {
	fen := "r3r1k1/pp3pbp/1qp1b1p1/2B5/2BP4/Q1n2N2/P4PPP/3R1K1R w - - 4 18"
	b := Board{}
	b.InitFEN(fen)

	moves := b.generateLegalMoves()
	uciMoves := []string{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	actualMoves := []string{
		"d1c1", "d1b1", "d1a1", "d1e1", "d1d2", "d1d3", "h1g1",
		"f1e1", "f1g1", "g2g3", "g2g4", "h2h3", "h2h4", "a3b2",
		"a3c1", "a3b3", "a3c3", "a3a4", "a3a5", "a3a6", "a3a7",
		"a3b4", "f3e1", "f3g1", "f3d2", "f3h4", "f3e5", "f3g5",
		"c4b3", "c4d3", "c4e2", "c4b5", "c4d5", "c4a6", "c4e6",
		"d4d5", "c5b4", "c5b6", "c5d6", "c5e7", "c5f8",
	}

	result := checkSameElements(actualMoves, uciMoves)
	if result == false {
		t.Errorf("TestAllMoves: got %v, wanted %v", uciMoves, actualMoves)
	}

}

// Test number of generated moves for a board with a large amount of moves
func TestAllMovesHuge(t *testing.T) {
	fen := "R6R/3Q4/1Q4Q1/4Q3/2Q4Q/Q4Q2/pp1Q4/kBNN1KB1 w - - 0 1"
	b := Board{}
	b.InitFEN(fen)

	moves := b.generateLegalMoves()
	uciMoves := []string{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	lenActualMoves := 218

	if len(uciMoves) != lenActualMoves {
		t.Errorf("TestAllMovesHuge: got %v, wanted %v", len(uciMoves), lenActualMoves)
	}
}

// Test case where all eight black pawns promote with captures
func TestPromotion(t *testing.T) {
	b := Board{}
	b.InitFEN("3k4/8/1K6/8/8/8/pppppppp/RRRRRRRR b - - 0 1")

	moves := b.generateLegalMoves()
	uciMoves := []string{}
	for _, move := range moves {
		uciMoves = append(uciMoves, move.toUCI())
	}

	actualMoves := []string{
		"a2b1q", "a2b1r", "a2b1n", "a2b1b",
		"b2a1q", "b2a1r", "b2a1n", "b2a1b",
		"b2c1q", "b2c1r", "b2c1n", "b2c1b",
		"c2b1q", "c2b1r", "c2b1n", "c2b1b",
		"c2d1q", "c2d1r", "c2d1n", "c2d1b",
		"e2d1q", "e2d1r", "e2d1n", "e2d1b",
		"e2f1q", "e2f1r", "e2f1n", "e2f1b",
		"f2e1q", "f2e1r", "f2e1n", "f2e1b",
		"f2g1q", "f2g1r", "f2g1n", "f2g1b",
		"g2f1q", "g2f1r", "g2f1n", "g2f1b",
		"g2h1q", "g2h1r", "g2h1n", "g2h1b",
		"h2g1q", "h2g1r", "h2g1n", "h2g1b",
		"d8c8", "d8d7", "d8e7", "d8e8",
	}

	result := checkSameElements(actualMoves, uciMoves)
	if result == false {
		t.Errorf("TestPromotion: got %v, wanted %v", uciMoves, actualMoves)
	}
}

// Test that make and undo move with promotion (capture and promotion) work
func TestMakeAndUndoMovePromotion(t *testing.T) {
	b := Board{}
	b.InitFEN("3k4/8/1K6/8/8/8/pppppppp/RRRRRRRR b - - 0 1")
	orig := b.getStringFromBitBoards()

	b.makeMoveFromUCI("b2a1q")
	length := len(b.history)

	if b.squares[a1] != bQ {
		t.Errorf("TestMakeAndUndoMovePromotion (promotion->queen): got %q, wanted %q", b.squares[a1].toString(), bQ.toString())
	}

	if b.squares[b2] != EMPTY {
		t.Errorf("TestMakeAndUndoMovePromotion (promotion->pawn): got %q, wanted %q", b.squares[b2].toString(), EMPTY.toString())
	}

	b.undo()
	newLength := len(b.history)

	if b.squares[a1] != wR {
		t.Errorf("TestMakeAndUndoMovePromotion (undo->rook): got %q, wanted %q", b.squares[a1].toString(), wR.toString())
	}

	if b.squares[b2] != bP {
		t.Errorf("TestMakeAndUndoMovePromotion (undo->pawn): got %q, wanted %q", b.squares[b2].toString(), bP.toString())
	}

	if length-newLength != 1 {
		t.Errorf("TestMakeAndUndoMovePromotion (history): got %d, wanted %d", length-newLength, 1)
	}

	new := b.getStringFromBitBoards()
	if orig != new {
		t.Errorf("TestMakeAndUndoMovePromotion (bitboard): got %q, wanted %q", new, orig)
	}
}

// Test that make and undo move with en passant work
func TestMakeAndUndoEnPassant(t *testing.T) {
	fen := "7k/8/8/8/pPp5/8/8/7K b - b3 0 1"
	b := Board{}
	b.InitFEN(fen)

	orig := b.getStringFromBitBoards()

	b.makeMoveFromUCI("a4b3")

	if b.squares[b3] != bP {
		t.Errorf("TestMakeAndUndoEnPassant (e.p.->ourpawn): got %q, wanted %q", b.squares[b3].toString(), bP.toString())
	}

	if b.squares[b4] != EMPTY {
		t.Errorf("TestMakeAndUndoEnPassant (e.p.->theirpawn): got %q, wanted %q", b.squares[b4].toString(), EMPTY.toString())
	}

	b.undo()

	if b.squares[b3] != EMPTY {
		t.Errorf("TestMakeAndUndoEnPassant (after e.p.->ourpawn): got %q, wanted %q", b.squares[b3].toString(), EMPTY.toString())
	}

	if b.squares[b4] != wP {
		t.Errorf("TestMakeAndUndoEnPassant (after e.p.->theirpawn): got %q, wanted %q", b.squares[b4].toString(), wP.toString())
	}

	if b.squares[a4] != bP {
		t.Errorf("TestMakeAndUndoEnPassant (after e.p.->ourpawn back): got %q, wanted %q", b.squares[a4].toString(), bP.toString())
	}

	new := b.getStringFromBitBoards()

	if orig != new {
		t.Errorf("TestMakeAndUndoEnpassant (bitboard): got %q, wanted %q", new, orig)
	}
}

// Test undoing a capture
func TestUndoMoveCapture(t *testing.T) {
	fen := "r3r1k1/pp3pbp/1qp1b1p1/2B5/2BP4/Q1n2N2/P4PPP/3R1K1R w - - 4 18"
	b := Board{}
	b.InitFEN(fen)

	orig := b.getStringFromBitBoards()

	b.makeMoveFromUCI("c4e6")
	b.undo()

	if b.squares[e6] != bB {
		t.Errorf("TestUndoMoveCapture (theirbishop): got %q, wanted %q", b.squares[e6].toString(), bP.toString())
	}

	if b.squares[c4] != wB {
		t.Errorf("TestUndoMoveCapture (ourbishop): got %q, wanted %q", b.squares[c4].toString(), wB.toString())
	}

	new := b.getStringFromBitBoards()

	if orig != new {
		t.Errorf("TestUndoMoveCapture (bitboard): got %q, wanted %q", new, orig)
	}
}
