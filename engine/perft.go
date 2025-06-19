package engine

import (
	"fmt"
	"testing"
	"time"
)

func init() {
	InitializeKingAttacks()
	InitializeKnightAttacks()
	InitializePawnAttacks()
	InitBishopAttacks()
	InitRookAttacks()
	InitSquaresBetween()
	InitLine()
	InitializeSQLookup()
	InitZobrist()
}

func Perft(b *Board, depth int) int {
	if depth <= 0 {
		return 1
	}

	moves := b.GenerateLegalMoves()
	if depth == 1 {
		return len(moves)
	}

	var numNodes int = 0
	for _, move := range moves {
		b.MakeMove(move)
		numNodes += Perft(b, depth-1)
		b.Undo()
	}
	return numNodes
}

func Divide(b *Board, depth int) int {
	moves := b.GenerateLegalMoves()

	var numNodes int = 0

	for _, move := range moves {
		var branch int = 0
		b.MakeMove(move)

		branch = Perft(b, depth-1)
		fmt.Printf("%q: %d nodes\n", move.ToUCI(), branch)
		numNodes += branch
		b.Undo()
	}
	return numNodes
}

func RunPerfTests(t *testing.T, position string, maxDepth int, expected int) {
	fmt.Println("------RUNNING PERFT------")
	fmt.Println("Input position: ")

	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		b.InitFEN(position)
	}

	b.PrintFromBitBoards()
	fmt.Println()
	nodes := 0

	for depth := 1; depth <= maxDepth; depth++ {
		start := time.Now()
		nodes = Perft(&b, depth)
		duration := time.Since(start)
		fmt.Printf("Depth %d, Nodes: %d, Time: %d µs, NPS: %d\n", depth, nodes, duration.Microseconds(), int(nodes*1000000000/(int(duration.Nanoseconds()+1))))

	}

	if nodes != expected {
		t.Fatalf("TestPerft: got %d, wanted %d", nodes, expected)
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

	b.PrintFromBitBoards()
	fmt.Println()

	fmt.Printf("Depth %d finished, Nodes: %d\n", depth, Divide(&b, depth))
}

func RunTests(t *testing.T) {
	// Perft tests: https://www.chessprogramming.org/Perft_Results

	fmt.Println("\nPosition 1: (works to depth 6) 119,060,324 (4,865,609 = depth 5)")
	// // startpos (works to depth 6) 119,060,324 (4,865,609 = depth 5)
	RunPerfTests(t, "startpos", 6, 119060324)

	fmt.Println("\nPosition 2: (works to depth 6) 8031647685 (193690690 = depth 5)")
	// // position 2 (works to depth 6) 8031647685 (193690690 = depth 5)
	RunPerfTests(t, "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 5, 193690690)

	fmt.Println("\nPosition 3:(works to depth 8) 3009794393 (178633661 = depth 7) ")
	// // position 3 (works to depth 8) 3009794393 (178633661 = depth 7)
	RunPerfTests(t, "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 7, 178633661)

	fmt.Println("\nPosition 4: (works to depth 6) 706045033 (15833292 = depth 5)")
	// // position 4 (works to depth 6) 706045033 (15833292 = depth 5)
	RunPerfTests(t, "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 5, 15833292)

	fmt.Println("\nPosition 5: (works to depth 5)  89,941,194")
	// // position 5 (works to depth 5)  89,941,194
	RunPerfTests(t, "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 5, 89941194)

	fmt.Println("\nPosition 6: (works to depth 6) 6,923,051,137 (164,075,551 = depth 5) ")
	// // position 6 (works to depth 6) 6,923,051,137 (164,075,551 = depth 5)
	RunPerfTests(t, "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 5, 164075551)

	// More Perft tests: https://www.chessprogramming.net/perfect-perft/
	fmt.Println("\n6: 1134888")
	RunPerfTests(t, "3k4/3p4/8/K1P4r/8/8/8/8 b - - 0 1", 6, 1134888)

	fmt.Println("\n6: 1015133")
	RunPerfTests(t, "8/8/4k3/8/2p5/8/B2P2K1/8 w - - 0 1", 6, 1015133)

	fmt.Println("\n6: 1440467")
	RunPerfTests(t, "8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1", 6, 1440467)

	fmt.Println("\n6: 661072")
	RunPerfTests(t, "5k2/8/8/8/8/8/8/4K2R w K - 0 1", 6, 661072)

	fmt.Println("\n6: 803711")
	RunPerfTests(t, "3k4/8/8/8/8/8/8/R3K3 w Q - 0 1", 6, 803711)

	fmt.Println("\n4: 1274206")
	RunPerfTests(t, "r3k2r/1b4bq/8/8/8/8/7B/R3K2R w KQkq - 0 1", 4, 1274206)

	fmt.Println("\n4: 1720476")
	RunPerfTests(t, "r3k2r/8/3Q4/8/8/5q2/8/R3K2R b KQkq - 0 1", 4, 1720476)

	fmt.Println("\n6: 3821001")
	RunPerfTests(t, "2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1", 6, 3821001)

	fmt.Println("\n5: 1004658")
	RunPerfTests(t, "8/8/1P2K3/8/2n5/1q6/8/5k2 b - - 0 1", 5, 1004658)

	fmt.Println("\n6: 217342")
	RunPerfTests(t, "4k3/1P6/8/8/8/8/K7/8 w - - 0 1", 6, 217342)

	fmt.Println("\n6: 92683")
	RunPerfTests(t, "8/P1k5/K7/8/8/8/8/8 w - - 0 1", 6, 92683)

	fmt.Println("\n6: 2217")
	RunPerfTests(t, "K1k5/8/P7/8/8/8/8/8 w - - 0 1", 6, 2217)

	fmt.Println("\n7: 567584")
	RunPerfTests(t, "8/k1P5/8/1K6/8/8/8/8 w - - 0 1", 7, 567584)

	fmt.Println("\n4: 23527")
	RunPerfTests(t, "8/8/2k5/5q2/5n2/8/5K2/8 b - - 0 1", 4, 23527)
}
