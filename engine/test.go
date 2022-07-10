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

	// // printBitBoard(knightAttacksSquareLookup[int(b1)])

	b := board{}
	b.initFEN("r3k2r/2qn1p1p/3bpn2/1Ppp4/P3P1p1/8/1B1N1PPP/RN1QK2R w KQkq - 2 15")

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
	completedMoves := []Move{}

	for i := 0; i < 10; i++ {
		m := b.generateLegalMoves()
		for _, move := range m {
			fmt.Println(move.toUCI())
		}
		randomMove := m[rand.Intn(len(m))]
		fmt.Printf("Choosing random move: %s", randomMove.toUCI())
		b.makeMove(randomMove)
		completedMoves = append(completedMoves, randomMove)

		b.print()
		b.printFromBitBoards()
	}

	for _, move := range completedMoves {
		fmt.Print(move.toUCI() + " ")
	}
	fmt.Print("\n")
}
