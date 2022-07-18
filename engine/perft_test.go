package engine

import "testing"

func TestPerft(t *testing.T) {
	// Current fastest: 22.031s
	RunTests()
}

func TestSearch(t *testing.T) {
	// Current fastest: 36.926s
	initializeTTable(4096)
	RunSearch("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 5)
}
