package engine

type Move struct {
	to         Square   // square which piece moves to
	from       Square   // square which piece moves from
	movetype   MoveType // type of the move (quiet, capture, etc)
	piece      Piece    // piece which is moving
	captured   Piece    // piece which may have been captured
	colorMoved Color    // side which is moving
	promote    Piece    // promotion piece
}
