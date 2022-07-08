package engine

import (
	"fmt"
	"math/rand"
)

func Test() {
	initializeKingAttacks()
	initializeKnightAttacks()
	initializePawnAttacks()

	// printBitBoard(knightAttacksSquareLookup[int(b1)])

	b := newBoard()
	b.print()
	b.printFromBitBoards()
	// for {
	// 	var move string
	// 	fmt.Println("Enter move: ")
	// 	fmt.Scanln(&move)

	// 	b.makeMove(fromUCI(move, b))
	// 	b.print()
	// 	b.printFromBitBoards()
	// }
	for i := 0; i < 10; i++ {
		m := b.generateLegalMoves()
		for _, move := range m {
			fmt.Println(move.toUCI())
		}
		randomMove := m[rand.Intn(len(m))]
		fmt.Printf("Choosing random move: %s", randomMove.toUCI())
		b.makeMove(randomMove)

		b.print()
		b.printFromBitBoards()
	}
}
