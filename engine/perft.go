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

func RunPerfTests(position string) {
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

	for depth := 0; depth <= 6; depth++ {
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
