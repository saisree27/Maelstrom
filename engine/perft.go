package engine

import (
	"fmt"
	"time"
)

func init() {
	initializeKingAttacks()
	initializeKnightAttacks()
	initializePawnAttacks()
	initBishopAttacks()
	initRookAttacks()
	initSquaresBetween()
	initLine()
	initializeSQLookup()
	fmt.Println("Finished initialization.")
}

func Perft(b *Board, depth int) int {
	if depth <= 0 {
		return 1
	}

	moves := b.generateLegalMoves()
	if depth == 1 {
		return len(moves)
	}

	var numNodes int = 0
	for _, move := range moves {
		b.makeMove(move)
		numNodes += Perft(b, depth-1)
		b.undo()
	}
	return numNodes
}

func Divide(b *Board, depth int) int {
	moves := b.generateLegalMoves()

	var numNodes int = 0

	for _, move := range moves {
		var branch int = 0
		b.makeMove(move)

		branch = Perft(b, depth-1)
		fmt.Printf("%q: %d nodes\n", move.toUCI(), branch)
		numNodes += branch
		b.undo()
	}
	return numNodes
}

func RunPerfTests(position string, maxDepth int) {
	fmt.Println("------RUNNING PERFT------")
	fmt.Println("Input position: ")

	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		b.InitFEN(position)
	}

	b.printFromBitBoards()
	fmt.Println()

	for depth := 0; depth <= maxDepth; depth++ {
		start := time.Now()
		nodes := Perft(&b, depth)
		duration := time.Since(start)
		fmt.Printf("Depth %d, Nodes: %d, Time: %d µs, NPS: %d\n", depth, nodes, duration.Microseconds(), int(nodes*1000000000/(int(duration.Nanoseconds()))))

	}
}

func RunDivide(position string, depth int) {
	fmt.Println("------RUNNING DIVIDE------")
	fmt.Println("Input position: ")

	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		b.InitFEN(position)
	}

	b.printFromBitBoards()
	fmt.Println()

	fmt.Printf("Depth %d finished, Nodes: %d\n", depth, Divide(&b, depth))
}

func RunTests() {
	// Perft tests: https://www.chessprogramming.org/Perft_Results

	fmt.Println("Position 1: ")
	// // startpos (works to depth 6)
	RunPerfTests("startpos", 6)

	fmt.Println("Position 2: ")
	// // position 2 (works to depth 6)
	RunPerfTests("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 6)

	fmt.Println("Position 3: ")
	// // position 3 (works to depth 8)
	RunPerfTests("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 8)

	fmt.Println("Position 4: ")
	// // position 4 (works to depth 6)
	RunPerfTests("r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 6)

	fmt.Println("Position 5: ")
	// // position 5 (works to depth 5)
	RunPerfTests("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 5)

	fmt.Println("Position 6: ")
	// // position 6 (works to depth 6)
	RunPerfTests("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 6)

	// More Perft tests: https://www.chessprogramming.net/perfect-perft/
	// Perft 6: 1134888
	RunPerfTests("3k4/3p4/8/K1P4r/8/8/8/8 b - - 0 1", 6)

	// Perft 6: 1015133
	RunPerfTests("8/8/4k3/8/2p5/8/B2P2K1/8 w - - 0 1", 6)

	// Perft 6: 1440467
	RunPerfTests("8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1", 6)

	// Perft 6: 661072
	RunPerfTests("5k2/8/8/8/8/8/8/4K2R w K - 0 1", 6)

	// Perft 6: 803711
	RunPerfTests("3k4/8/8/8/8/8/8/R3K3 w Q - 0 1", 6)

	// Perft 4: 1274206
	RunPerfTests("r3k2r/1b4bq/8/8/8/8/7B/R3K2R w KQkq - 0 1", 4)

	// Perft 4: 1720476
	RunPerfTests("r3k2r/8/3Q4/8/8/5q2/8/R3K2R b KQkq - 0 1", 4)

	// Perft 6: 3821001
	RunPerfTests("2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1", 6)

	// Perft 5: 1004658
	RunPerfTests("8/8/1P2K3/8/2n5/1q6/8/5k2 b - - 0 1", 5)

	// Perft 6: 217342
	RunPerfTests("4k3/1P6/8/8/8/8/K7/8 w - - 0 1", 6)

	// Perft 6: 92683
	RunPerfTests("8/P1k5/K7/8/8/8/8/8 w - - 0 1", 6)

	// Perft 6: 2217
	RunPerfTests("K1k5/8/P7/8/8/8/8/8 w - - 0 1", 6)

	// Perft 7: 567584
	RunPerfTests("8/k1P5/8/1K6/8/8/8/8 w - - 0 1", 7)

	// Perft 4: 23527
	RunPerfTests("8/8/2k5/5q2/5n2/8/5K2/8 b - - 0 1", 4)
}
