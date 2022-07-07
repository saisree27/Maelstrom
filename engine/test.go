package engine

import "fmt"

func Test() {
	b := newBoard()
	m := []string{
		"h2h4", "a7a5", "h4h5", "g7g5", "h5g6", "a5a4", "b2b4", "a4b3", "g6g7", "b3b2", "g7h8q", "b2a1b"}
	for i := 0; i < len(m); i++ {
		n := fromUCI(m[i], b)
		b.makeMove(n)
		fmt.Println(n.toUCI())
	}
	b.print()
	b.printFromBitBoards()
}
