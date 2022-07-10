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
	initializeSQLookup()

	// // printBitBoard(knightAttacksSquareLookup[int(b1)])

	b := board{}
	b.initFEN("r3k2r/pppbqppp/3p1n2/2b1p1B1/2BnP3/2NP1N2/PPPQ1PPP/1K1R3R b kq - 7 9")

	b.print()
	b.printFromBitBoards()
	fmt.Println(b.kW)
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
