package engine

import "testing"

func TestPerft(t *testing.T) {
	// Current fastest: 25.023s
	RunTests(t)
}

func TestSearchPosition(t *testing.T) {
	// Current best: 5.415s
	InitializeEverythingExceptTTable()
	InitializeTT(256)
	RunSearch("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 8)
}
