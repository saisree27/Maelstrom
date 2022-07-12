package engine

import "testing"

func TestCheckmate(t *testing.T) {
	fen := "r1bqkb1r/pppp1Qpp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4"
	b := Board{}
	b.InitFEN(fen)

	res := evaluate(&b)
	if res != winVal {
		t.Errorf("TestCheckmate (white): got %d, wanted %d", res, winVal)
	}

	fen = "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3"
	b = Board{}
	b.InitFEN(fen)

	res = evaluate(&b)
	if res != -winVal {
		t.Errorf("TestCheckmate (black): got %d, wanted %d", res, -winVal)
	}
}

func TestStalemate(t *testing.T) {
	fen := "7K/5k1P/8/8/8/8/8/8 w - - 0 1"
	b := Board{}
	b.InitFEN(fen)

	res := evaluate(&b)
	if res != 0 {
		t.Errorf("TestStalemate (white): got %d, wanted %d", res, 0)
	}

	fen = "8/8/8/8/8/8/p1K5/k7 b - - 0 1"
	b = Board{}
	b.InitFEN(fen)

	res = evaluate(&b)
	if res != 0 {
		t.Errorf("TestStalemate (black): got %d, wanted %d", res, 0)
	}
}

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
	if material != 100 {
		t.Errorf("TestMaterial (white pawn up): got %d, wanted %d", material, 100)
	}

	fen = "rnb1kb1r/ppp1pppp/5n2/3N4/8/8/PPPP1PPP/R1BQKBNR b KQkq - 0 4"
	b = Board{}
	b.InitFEN(fen)

	material, _ = totalMaterialAndPieces(&b)
	if material != 900 {
		t.Errorf("TestMaterial (white queen up): got %d, wanted %d", material, 900)
	}

	fen = "rnb1kbnr/ppp1pppp/8/7q/8/8/PPPP1PPP/RNB1KBNR w KQkq - 0 4"
	b = Board{}
	b.InitFEN(fen)

	material, _ = totalMaterialAndPieces(&b)
	if material != -900 {
		t.Errorf("TestMaterial (black queen up): got %d, wanted %d", material, -900)
	}

	fen = "2r5/1p1k1pp1/p2p4/P7/2R4p/1P1b1B1P/2r2PP1/3KR3 w - - 0 31"
	b = Board{}
	b.InitFEN(fen)

	material, _ = totalMaterialAndPieces(&b)
	if material != -100 {
		t.Errorf("TestMaterial (black pawn up): got %d, wanted %d", material, -100)
	}
}
