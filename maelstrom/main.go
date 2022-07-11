package main

import (
	"maelstrom/engine"
)

func main() {
	// engine.RunPerfTests("startpos")
	// engine.RunDivide("rnb1kbnr/pp1ppppp/8/q1p5/1P6/3P4/P1P1PPPP/RNBQKBNR w KQkq - 1 3", 1)
	// engine.RunDivide("rnbqkbnr/pppppppp/8/8/8/2P5/PP1PPPPP/RNBQKBNR b KQkq - 0 1", 3)
	// engine.RunDivide("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 6)
	// engine.RunDivide("r6r/p1pkqpb1/bn2pQp1/3P4/1p2P3/2N4p/PPPBBPPP/R3K2R b KQ - 0 2", 3)
	engine.RunDivide("r6r/p1p1qpb1/bn1kpQp1/3P4/Pp2P3/2N4p/1PPBBPPP/R3K2R b KQ a3 0 3", 1)
}
