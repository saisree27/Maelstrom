package engine

import (
	"fmt"
	"testing"
	"time"
)

func TestAccumulatorUpdateMatchesRecompute(t *testing.T) {
	GlobalNNUE = NewRandomNNUE()
	b := Board{}
	b.InitStartPos()

	// Make moves
	b.MakeMoveFromUCI("b1a3")
	b.MakeMoveFromUCI("b8a6")
	b.Undo()
	b.MakeMoveFromUCI("e7e5")
	GlobalNNUE.ApplyLazyUpdates(&b)

	// Recompute from scratch
	expected := GlobalNNUE.RecomputeAccumulators(&b)

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		if b.accumulatorStack[b.accumulatorIdx].white.values[i] != expected.white.values[i] {
			t.Errorf("White accumulator mismatch at %d: got %d, expected %d", i, b.accumulatorStack[b.accumulatorIdx].white.values[i], expected.white.values[i])
		}
		if b.accumulatorStack[b.accumulatorIdx].black.values[i] != expected.black.values[i] {
			t.Errorf("Black accumulator mismatch at %d: got %d, expected %d", i, b.accumulatorStack[b.accumulatorIdx].black.values[i], expected.black.values[i])
		}
	}
}

func TestCalculateIndexSanity(t *testing.T) {
	for sq := A1; sq <= H8; sq++ {
		for pt := PAWN; pt <= KING; pt++ {
			for side := WHITE; side <= BLACK; side++ {
				// Check bounds
				whiteIdx := CalculateIndex(WHITE, sq, pt, side)
				blackIdx := CalculateIndex(BLACK, sq, pt, side)

				if whiteIdx < 0 || whiteIdx >= INPUT_LAYER_SIZE {
					t.Errorf("Invalid white index: sq=%v pt=%v side=%v => %d", sq, pt, side, whiteIdx)
				}
				if blackIdx < 0 || blackIdx >= INPUT_LAYER_SIZE {
					t.Errorf("Invalid black index: sq=%v pt=%v side=%v => %d", sq, pt, side, blackIdx)
				}
			}
		}
	}
}

func PerftWithNNUECheck(b *Board, depth int, t *testing.T) int {
	moves := b.GenerateLegalMoves()

	if depth == 1 {
		return len(moves)
	}

	var numNodes int = 0
	for _, move := range moves {
		b.MakeMove(move)
		GlobalNNUE.ApplyLazyUpdates(b)

		expected := GlobalNNUE.RecomputeAccumulators(b)
		for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
			if b.accumulatorStack[b.accumulatorIdx].white.values[i] != expected.white.values[i] {
				fmt.Println(move.ToUCI())
				b.PrintFromBitBoards()
				t.Fatalf("MakeMove - White accumulator mismatch at %d: got %d, expected %d", i, b.accumulatorStack[b.accumulatorIdx].white.values[i], expected.white.values[i])
			}
			if b.accumulatorStack[b.accumulatorIdx].black.values[i] != expected.black.values[i] {
				t.Fatalf("MakeMove - Black accumulator mismatch at %d: got %d, expected %d", i, b.accumulatorStack[b.accumulatorIdx].black.values[i], expected.black.values[i])
			}
		}

		n := PerftWithNNUECheck(b, depth-1, t)
		numNodes += n

		b.Undo()

		expected = GlobalNNUE.RecomputeAccumulators(b)
		for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
			if b.accumulatorStack[b.accumulatorIdx].white.values[i] != expected.white.values[i] {
				t.Fatalf("Undo - White accumulator mismatch at %d: got %d, expected %d", i, b.accumulatorStack[b.accumulatorIdx].white.values[i], expected.white.values[i])
			}
			if b.accumulatorStack[b.accumulatorIdx].black.values[i] != expected.black.values[i] {
				t.Fatalf("Undo - Black accumulator mismatch at %d: got %d, expected %d", i, b.accumulatorStack[b.accumulatorIdx].black.values[i], expected.black.values[i])
			}
		}
	}
	return numNodes
}

func RunPerfTestsNNUECheck(t *testing.T, position string, maxDepth int, expected int) {
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
		nodes = PerftWithNNUECheck(&b, depth, t)
		duration := time.Since(start)
		fmt.Printf("Depth %d, Nodes: %d, Time: %d µs, NPS: %d\n", depth, nodes, duration.Microseconds(), int(nodes*1000000000/(int(duration.Nanoseconds()+1))))

	}

	if nodes != expected {
		t.Fatalf("TestPerft: got %d nodes, wanted %d", nodes, expected)
	}
}

func TestPerftNNUE(t *testing.T) {
	GlobalNNUE = NewRandomNNUE()
	// Perft tests: https://www.chessprogramming.org/Perft_Results
	RunPerfTestsNNUECheck(t, "startpos", 5, 4865609)

	// // position 2 (works to depth 6) 8031647685 (193690690 = depth 5)
	RunPerfTestsNNUECheck(t, "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 4, 4085603)

	// // position 3 (works to depth 8) 3009794393 (178633661 = depth 7)
	RunPerfTestsNNUECheck(t, "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 6, 11030083)

	// // position 4 (works to depth 6) 706045033 (15833292 = depth 5)
	RunPerfTestsNNUECheck(t, "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 5, 15833292)

	// // position 5 (works to depth 5)  89,941,194
	RunPerfTestsNNUECheck(t, "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 4, 2103487)

	// // position 6 (works to depth 6) 6,923,051,137 (164,075,551 = depth 5)
	RunPerfTestsNNUECheck(t, "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 4, 3894594)

	// More Perft tests: https://www.chessprogramming.net/perfect-perft/
	fmt.Println("\n6: 1134888")
	RunPerfTestsNNUECheck(t, "3k4/3p4/8/K1P4r/8/8/8/8 b - - 0 1", 6, 1134888)

	fmt.Println("\n6: 1015133")
	RunPerfTestsNNUECheck(t, "8/8/4k3/8/2p5/8/B2P2K1/8 w - - 0 1", 6, 1015133)

	fmt.Println("\n6: 1440467")
	RunPerfTestsNNUECheck(t, "8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1", 6, 1440467)

	fmt.Println("\n6: 661072")
	RunPerfTestsNNUECheck(t, "5k2/8/8/8/8/8/8/4K2R w K - 0 1", 6, 661072)

	fmt.Println("\n6: 803711")
	RunPerfTestsNNUECheck(t, "3k4/8/8/8/8/8/8/R3K3 w Q - 0 1", 6, 803711)

	fmt.Println("\n4: 1274206")
	RunPerfTestsNNUECheck(t, "r3k2r/1b4bq/8/8/8/8/7B/R3K2R w KQkq - 0 1", 4, 1274206)

	fmt.Println("\n4: 1720476")
	RunPerfTestsNNUECheck(t, "r3k2r/8/3Q4/8/8/5q2/8/R3K2R b KQkq - 0 1", 4, 1720476)

	fmt.Println("\n6: 3821001")
	RunPerfTestsNNUECheck(t, "2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1", 6, 3821001)

	fmt.Println("\n5: 1004658")
	RunPerfTestsNNUECheck(t, "8/8/1P2K3/8/2n5/1q6/8/5k2 b - - 0 1", 5, 1004658)

	fmt.Println("\n6: 217342")
	RunPerfTestsNNUECheck(t, "4k3/1P6/8/8/8/8/K7/8 w - - 0 1", 6, 217342)

	fmt.Println("\n6: 92683")
	RunPerfTestsNNUECheck(t, "8/P1k5/K7/8/8/8/8/8 w - - 0 1", 6, 92683)

	fmt.Println("\n6: 2217")
	RunPerfTestsNNUECheck(t, "K1k5/8/P7/8/8/8/8/8 w - - 0 1", 6, 2217)

	fmt.Println("\n7: 567584")
	RunPerfTestsNNUECheck(t, "8/k1P5/8/1K6/8/8/8/8 w - - 0 1", 7, 567584)

	fmt.Println("\n4: 23527")
	RunPerfTestsNNUECheck(t, "8/8/2k5/5q2/5n2/8/5K2/8 b - - 0 1", 4, 23527)

	fmt.Println("\n6: 3821001")
	RunPerfTestsNNUECheck(t, "2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1", 6, 3821001)
}

func TestEvalConsistencyAfterUpdate(t *testing.T) {
	GlobalNNUE = NewRandomNNUE()
	b := Board{}
	b.InitStartPos()

	// Save pre-move eval
	beforeEval := Forward(&GlobalNNUE, &b.accumulatorStack[b.accumulatorIdx].white, &b.accumulatorStack[b.accumulatorIdx].black)

	// Apply accumulator update
	b.MakeMoveFromUCI("e2e4")
	GlobalNNUE.ApplyLazyUpdates(&b)

	// Forward after incremental update
	afterEval := Forward(&GlobalNNUE, &b.accumulatorStack[b.accumulatorIdx].white, &b.accumulatorStack[b.accumulatorIdx].black)

	// Fully recompute and eval
	recomputed := GlobalNNUE.RecomputeAccumulators(&b)
	recomputedEval := Forward(&GlobalNNUE, &recomputed.white, &recomputed.black)

	if afterEval != recomputedEval {
		t.Errorf("Eval mismatch after update: incremental=%d, recomputed=%d", afterEval, recomputedEval)
	}

	if beforeEval == afterEval {
		t.Errorf("Eval mismatch after move: before=%d, after=%d", beforeEval, afterEval)
	}
}

func TestEvalOppositeSide(t *testing.T) {
	GlobalNNUE = NewRandomNNUE()
	b := Board{}
	b.InitStartPos()
	b.MakeMoveFromUCI("e2e4")
	GlobalNNUE.ApplyLazyUpdates(&b)

	blackPerspectiveEval := Forward(&GlobalNNUE, &b.accumulatorStack[b.accumulatorIdx].white, &b.accumulatorStack[b.accumulatorIdx].black)

	// Manually change STM
	b.turn = WHITE

	whitePerspectiveEval := -Forward(&GlobalNNUE, &b.accumulatorStack[b.accumulatorIdx].white, &b.accumulatorStack[b.accumulatorIdx].black)
	expected := -blackPerspectiveEval

	if whitePerspectiveEval != expected {
		t.Errorf("Eval mismatch after perspective change: before=%d, after=%d", blackPerspectiveEval, whitePerspectiveEval)
	}
}
