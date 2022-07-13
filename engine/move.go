package engine

import (
	"strings"
)

type Move struct {
	to         Square   // square which piece moves to
	from       Square   // square which piece moves from
	movetype   MoveType // type of the move (quiet, capture, etc)
	piece      Piece    // piece which is moving
	captured   Piece    // piece which may have been captured
	colorMoved Color    // side which is moving
	promote    Piece    // promotion piece
	null       bool     // null move (default false)
}

func (m Move) toUCI() string {
	var s = ""
	s += squareToStringMap[m.from]
	s += squareToStringMap[m.to]
	if m.movetype == PROMOTION || m.movetype == CAPTUREANDPROMOTION {
		s += strings.ToLower(m.promote.toString())
	}
	return s
}

func fromUCI(uci string, b Board) Move {
	// parse UCI string into Move
	// should only be necessary for testing, as UCI gives fen position
	var m = Move{}
	var from, to Square = stringToSquareMap[uci[:2]], stringToSquareMap[uci[2:4]]
	var promotion = EMPTY

	m.movetype = QUIET
	m.from = from
	m.to = to
	m.piece = b.squares[from]
	m.colorMoved = m.piece.getColor()

	if len(uci) == 5 {
		m.movetype = PROMOTION
		var s = string(uci[4])
		switch s {
		case "n":
			if m.colorMoved == WHITE {
				promotion = wN
			} else {
				promotion = bN
			}
		case "b":
			if m.colorMoved == WHITE {
				promotion = wB
			} else {
				promotion = bB
			}
		case "r":
			if m.colorMoved == WHITE {
				promotion = wR
			} else {
				promotion = bR
			}
		case "q":
			if m.colorMoved == WHITE {
				promotion = wQ
			} else {
				promotion = bQ
			}
		}
	}

	m.promote = promotion

	if b.squares[to] != EMPTY {
		if m.promote != EMPTY {
			m.movetype = CAPTUREANDPROMOTION
		} else {
			m.movetype = CAPTURE
		}
		m.captured = b.squares[to]
	}

	if (from == e1 && to == g1 && m.piece == wK) || (from == e8 && to == g8 && m.piece == bK) {
		m.movetype = KCASTLE
	}

	if (from == e1 && to == c1 && m.piece == wK) || (from == e8 && to == c8 && m.piece == bK) {
		m.movetype = QCASTLE
	}

	var oneSquare = int(to-from) == int(NORTH)
	var twoSquares = int(to-from) == 2*int(NORTH)

	if m.piece == wP && !(oneSquare || twoSquares) && m.movetype == QUIET {
		m.movetype = ENPASSANT
	}

	oneSquare = int(to-from) == int(SOUTH)
	twoSquares = int(to-from) == 2*int(SOUTH)

	if m.piece == bP && !(oneSquare || twoSquares) && m.movetype == QUIET {
		m.movetype = ENPASSANT
	}

	return m
}
