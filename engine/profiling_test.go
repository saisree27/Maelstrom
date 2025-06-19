package engine

import "testing"

func TestSearchWithProfiling(t *testing.T) {
	// Initialize the engine
	InitializeEverythingExceptTTable()
	InitializeTT(256)

	// Create a board with a complex position for testing
	b := Board{}
	b.InitFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")

	// Run search with profiling for 15 seconds
	move := SearchWithProfiling(&b, 30000) // 15 seconds

	// Verify that we got a valid move
	if move.from == 0 && move.to == 0 {
		t.Error("Expected a valid move, got empty move")
	}
}
