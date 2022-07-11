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
	b.initStartPos()
	// b.initFEN("r1bqkbnr/pPppp3/2n2ppp/8/8/8/1PPPPPPP/RNBQKBNR w KQkq - 0 5")
	// b.makeMove(fromUCI("b7c8r", b))
	// b.initFEN("rnb1kbnr/pp2pppp/2ppq3/4P3/8/8/PPPP1PPP/RNBQKBNR w kq - 0 6")
	// b.initFEN("r1Rqkb1r/p1ppp3/5npp/1P3pN1/5P2/1nRP4/2P1P1PP/1NBQKB1R b Kk - 0 13")
	// b.makeMove(fromUCI("b7c8r", b))
	// b.initFEN("6K1/8/8/8/1k6/8/8/8 b - - 0 1")
	// b.makeMove(fromUCI("b4a3", b))
	b.print()
	b.printFromBitBoards()
	fmt.Println(b.kB)
	fmt.Println(b.qB)
	fmt.Println(b.kW)
	fmt.Println(b.qW)
	// for {
	// 	var move string
	// 	fmt.Println("Enter move: ")
	// 	fmt.Scanln(&move)

	// 	b.makeMove(fromUCI(move, b))
	// 	b.print()
	// 	b.printFromBitBoards()
	// }

	// p := b.generateLegalMoves()
	// for _, move := range p {
	// 	fmt.Print(move.toUCI() + " ")
	// }
	// fmt.Println("done")
	completedMoves := []Move{}

	var i u64
	for i = 0; i < 1500; i++ {
		m := b.generateLegalMoves()
		// for _, move := range m {
		// fmt.Println(move.toUCI())
		// }
		fmt.Println(i)
		if len(m) == 0 {
			break
		}
		randomMove := m[rand.Intn(len(m))]
		fmt.Printf("Choosing random move: %s\n", randomMove.toUCI())
		b.makeMove(randomMove)
		completedMoves = append(completedMoves, randomMove)

		b.print()
		b.printFromBitBoards()
	}

	// for _, move := range completedMoves {
	// fmt.Print(move.toUCI() + " ")
	// }

	b.printFromBitBoards()
	fmt.Print("\n")
}
