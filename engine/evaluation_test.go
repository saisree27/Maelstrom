package engine

import "testing"

func TestMaterial(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	b := Board{}
	b.InitFEN(fen)

	material, _ := totalMaterialAndPieces(&b)
	if material != 0 {
		t.Errorf("TestMaterial (equal): got %d, wanted %d", material, 0)
	}

	fen = "rnbqkbnr/ppp1pppp/8/3P4/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2"
	b = Board{}
	b.InitFEN(fen)

	material, _ = totalMaterialAndPieces(&b)
	if material != pawnVal {
		t.Errorf("TestMaterial (white pawn up): got %d, wanted %d", material, pawnVal)
	}

	fen = "rnb1kb1r/ppp1pppp/5n2/3N4/8/8/PPPP1PPP/R1BQKBNR b KQkq - 0 4"
	b = Board{}
	b.InitFEN(fen)

	material, _ = totalMaterialAndPieces(&b)
	if material != queenVal {
		t.Errorf("TestMaterial (white queen up): got %d, wanted %d", material, queenVal)
	}

	fen = "rnb1kbnr/ppp1pppp/8/7q/8/8/PPPP1PPP/RNB1KBNR w KQkq - 0 4"
	b = Board{}
	b.InitFEN(fen)

	material, _ = totalMaterialAndPieces(&b)
	if material != -queenVal {
		t.Errorf("TestMaterial (black queen up): got %d, wanted %d", material, -queenVal)
	}

	fen = "2r5/1p1k1pp1/p2p4/P7/2R4p/1P1b1B1P/2r2PP1/3KR3 w - - 0 31"
	b = Board{}
	b.InitFEN(fen)

	material, _ = totalMaterialAndPieces(&b)
	if material != -pawnVal {
		t.Errorf("TestMaterial (black pawn up): got %d, wanted %d", material, -pawnVal)
	}
}

func TestInsufficientMaterial(t *testing.T) {
	fen := "8/4k3/8/2n5/8/8/8/3K4 w - - 0 1"
	b := Board{}
	b.InitFEN(fen)

	if !b.isInsufficientMaterial() {
		t.Errorf("TestInsufficientMaterial (b+k vs K): got %t, wanted %t", b.isInsufficientMaterial(), true)
	}

	fen = "8/4k3/2b5/3n4/8/8/8/3K4 w - - 0 1"
	b = Board{}
	b.InitFEN(fen)

	if b.isInsufficientMaterial() {
		t.Errorf("TestInsufficientMaterial (b+n+k vs K): got %t, wanted %t", b.isInsufficientMaterial(), false)
	}

	fen = "8/4k3/2n5/3n4/8/8/8/3K4 w - - 0 1"
	b = Board{}
	b.InitFEN(fen)

	if !b.isInsufficientMaterial() {
		t.Errorf("TestInsufficientMaterial (n+n+k vs K): got %t, wanted %t", b.isInsufficientMaterial(), true)
	}

	fen = "8/4k3/2n5/3n4/8/8/4P3/3K4 w - - 0 1"
	b = Board{}
	b.InitFEN(fen)

	if b.isInsufficientMaterial() {
		t.Errorf("TestInsufficientMaterial (n+n+k vs K+P): got %t, wanted %t", b.isInsufficientMaterial(), false)
	}

	fen = "8/4k3/8/8/8/8/4P3/3K4 w - - 0 1"
	b = Board{}
	b.InitFEN(fen)

	if b.isInsufficientMaterial() {
		t.Errorf("TestInsufficientMaterial (k vs K+P): got %t, wanted %t", b.isInsufficientMaterial(), false)
	}
}
