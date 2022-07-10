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
	b.initFEN("4r2k/6p1/7p/1p1N2nP/1P2P3/b3BP2/2R3K1/8 b - - 18 58")

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
