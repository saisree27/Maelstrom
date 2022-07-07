package engine

import "fmt"

func Test() {
	b := newBoard()
	b.print()
	b.printFromBitBoards()
	for {
		var move string
		fmt.Println("Enter move: ")
		fmt.Scanln(&move)

		b.makeMove(fromUCI(move, b))
		b.print()
		b.printFromBitBoards()
	}
}
