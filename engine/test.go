package engine

func Test() {
	b := newBoard()
	b.print()
	printBitBoard(b.occupied)

	m := Move{from: e2, to: e4, piece: wP, movetype: QUIET, captured: EMPTY, colorMoved: WHITE, promote: EMPTY}
	b.makeMove(m)

	b.print()

	printBitBoard(b.occupied)
}
