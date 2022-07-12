package main

import "maelstrom/engine"

func main() {
	// engine.Run("search", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 6)
	// engine.Run("search", "startpos", 6)
	engine.Run("selfplay", "startpos", 6)
	// engine.Run("selfplay", "r2r1nk1/ppp2ppp/3p4/8/1PP1pP2/4P2P/P1P3P1/1K1R2NR b - - 0 17", 7)
	// engine.Run("search", "r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4", 10)
	// engine.Run("search", "r5rk/5p1p/5R2/4B3/8/8/7P/7K w - - 0 0", 15)
	// engine.Run("search", "8/7b/7b/p7/Pp2k3/1P6/KP2p2p/3N4 b - - 0 1", 10)
}
