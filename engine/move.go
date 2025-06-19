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

func (m Move) ToUCI() string {
	var s = ""
	s += SQUARE_TO_STRING_MAP[m.from]
	s += SQUARE_TO_STRING_MAP[m.to]
	if m.movetype == PROMOTION || m.movetype == CAPTURE_AND_PROMOTION {
		s += strings.ToLower(m.promote.ToString())
	}
	return s
}

func FromUCI(uci string, b *Board) Move {
	// parse UCI string into Move
	// should only be necessary for testing, as UCI gives fen position
	var m = Move{}
	var from, to Square = STRING_TO_SQUARE_MAP[uci[:2]], STRING_TO_SQUARE_MAP[uci[2:4]]
	var promotion = EMPTY

	m.movetype = QUIET
	m.from = from
	m.to = to
	m.piece = b.squares[from]
	m.colorMoved = m.piece.GetColor()

	if len(uci) == 5 {
		m.movetype = PROMOTION
		var s = string(uci[4])
		switch s {
		case "n":
			if m.colorMoved == WHITE {
				promotion = W_N
			} else {
				promotion = B_N
			}
		case "b":
			if m.colorMoved == WHITE {
				promotion = W_B
			} else {
				promotion = B_B
			}
		case "r":
			if m.colorMoved == WHITE {
				promotion = W_R
			} else {
				promotion = B_R
			}
		case "q":
			if m.colorMoved == WHITE {
				promotion = W_Q
			} else {
				promotion = B_Q
			}
		}
	}

	m.promote = promotion

	if b.squares[to] != EMPTY {
		if m.promote != EMPTY {
			m.movetype = CAPTURE_AND_PROMOTION
		} else {
			m.movetype = CAPTURE
		}
		m.captured = b.squares[to]
	}

	if (from == E1 && to == G1 && m.piece == W_K) || (from == E8 && to == G8 && m.piece == B_K) {
		m.movetype = K_CASTLE
	}

	if (from == E1 && to == C1 && m.piece == W_K) || (from == E8 && to == C8 && m.piece == B_K) {
		m.movetype = Q_CASTLE
	}

	var oneSquare = int(to-from) == int(NORTH)
	var twoSquares = int(to-from) == 2*int(NORTH)

	if m.piece == W_P && !(oneSquare || twoSquares) && m.movetype == QUIET {
		m.movetype = EN_PASSANT
	}

	oneSquare = int(to-from) == int(SOUTH)
	twoSquares = int(to-from) == 2*int(SOUTH)

	if m.piece == B_P && !(oneSquare || twoSquares) && m.movetype == QUIET {
		m.movetype = EN_PASSANT
	}

	return m
}
