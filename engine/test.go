package engine

func Test() {
	b := newBoard()

	b.print()
	b.printFromBitBoards()

	m := Move{from: e2, to: e4, piece: wP, movetype: QUIET, captured: EMPTY, colorMoved: WHITE, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: e7, to: e5, piece: bP, movetype: QUIET, captured: EMPTY, colorMoved: BLACK, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: g1, to: f3, piece: wN, movetype: QUIET, captured: EMPTY, colorMoved: WHITE, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: g8, to: f6, piece: bN, movetype: QUIET, captured: EMPTY, colorMoved: BLACK, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: f3, to: e5, piece: wN, movetype: CAPTURE, captured: bP, colorMoved: WHITE, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: d7, to: d5, piece: bP, movetype: QUIET, captured: EMPTY, colorMoved: BLACK, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: d2, to: d3, piece: wP, movetype: QUIET, captured: EMPTY, colorMoved: WHITE, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: d5, to: d4, piece: bP, movetype: QUIET, captured: EMPTY, colorMoved: BLACK, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: c2, to: c4, piece: wP, movetype: QUIET, captured: EMPTY, colorMoved: WHITE, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: d4, to: c3, piece: bP, movetype: ENPASSANT, captured: wP, colorMoved: BLACK, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: f1, to: e2, piece: wB, movetype: QUIET, captured: EMPTY, colorMoved: WHITE, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

	m = Move{from: f8, to: e7, piece: bB, movetype: QUIET, captured: EMPTY, colorMoved: BLACK, promote: EMPTY}
	b.makeMove(m)

	b.print()
	b.printFromBitBoards()

}
