package engine

import (
	"fmt"
	"math/rand"
	"time"
)

func Test() {
	rand.Seed(time.Now().UnixNano())

	initializeKingAttacks()
	initializeKnightAttacks()
	initializePawnAttacks()
	initBishopAttacks()
	initRookAttacks()
	initSquaresBetween()
	initLine()

	// // printBitBoard(knightAttacksSquareLookup[int(b1)])

	b := board{}
	b.initFEN("8/1k1R1p1p/5n2/3q4/P4B2/1P6/4Q1P1/7K b - - 0 38")

	b.print()
	b.printFromBitBoards()

	printBitBoard(b.occupied)
	// for {
	// 	var move string
	// 	fmt.Println("Enter move: ")
	// 	fmt.Scanln(&move)

	// 	b.makeMove(fromUCI(move, b))
	// 	b.print()
	// 	b.printFromBitBoards()
	// }

	p := b.generateLegalMoves()
	for _, move := range p {
		fmt.Print(move.toUCI() + " ")
	}
	fmt.Println("done")
	// completedMoves := []Move{}

	// for i := 0; i < 10; i++ {
	// 	m := b.generateLegalMoves()
	// 	for _, move := range m {
	// 		fmt.Println(move.toUCI())
	// 	}
	// 	randomMove := m[rand.Intn(len(m))]
	// 	fmt.Printf("Choosing random move: %s", randomMove.toUCI())
	// 	b.makeMove(randomMove)
	// 	completedMoves = append(completedMoves, randomMove)

	// 	b.print()
	// 	b.printFromBitBoards()
	// }

	// for _, move := range completedMoves {
	// 	fmt.Print(move.toUCI() + " ")
	// }
	// fmt.Print("\n")
}
