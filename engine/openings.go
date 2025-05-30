package engine

import (
	"math/rand"
	"time"
)

// OpeningMove represents a single move in an opening line
type OpeningMove struct {
	move     string // UCI format move string
	children []*OpeningMove
	name     string // Name of the variation (empty if not a notable position)
	zobrist  u64    // Zobrist hash of the position after this move
}

// OpeningBook represents the entire opening book
type OpeningBook struct {
	whiteRepertoire *OpeningMove
	blackRepertoire *OpeningMove
}

// NewOpeningBook creates and initializes a new opening book with predefined openings
func NewOpeningBook() *OpeningBook {
	rand.Seed(time.Now().UnixNano())

	book := &OpeningBook{
		whiteRepertoire: &OpeningMove{},
		blackRepertoire: &OpeningMove{},
	}

	// Initialize both repertoires with starting position
	startBoard := newBoard()
	book.whiteRepertoire.zobrist = startBoard.zobrist
	book.blackRepertoire.zobrist = startBoard.zobrist

	// White repertoire
	book.addWhiteLine([]string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5", "a7a6", "b5a4", "g8f6", "e1g1", "f8e7"}, "Ruy Lopez")
	book.addWhiteLine([]string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "f8c5", "c2c3", "d7d6", "d2d3"}, "Giuoco Piano")
	book.addWhiteLine([]string{"e2e4", "c7c5", "g1f3", "b8c6", "f1b5", "g8f6", "b5c6", "d7c6", "d2d3"}, "Sicilian Rossolimo")
	book.addWhiteLine([]string{"e2e4", "c7c5", "g1f3", "d7d6", "f1b5"}, "Sicilian Canal")
	book.addWhiteLine([]string{"d2d4", "d7d5", "c2c4", "e7e6", "g1f3", "g8f6", "b1c3"}, "QGD")
	book.addWhiteLine([]string{"d2d4", "d7d5", "c2c4", "d5c4", "g1f3", "g8f6", "e2e3", "b7b5", "a2a4", "c7c6", "a4b5", "c6b5", "b2b3"}, "QGA")
	book.addWhiteLine([]string{"d2d4", "d7d5", "c2c4", "d5c4", "g1f3", "c7c5", "e2e4", "c5d4", "d1d4", "d8d4", "f3d4", "g8f6", "d4b5", "b8a6", "f2f3"}, "QGA")
	book.addWhiteLine([]string{"d2d4", "g8f6", "c2c4", "e7e6", "b1c3", "f8b4"}, "Nimzo-Indian")
	book.addWhiteLine([]string{"d2d4", "g8f6", "c2c4", "e7e6", "g1f3", "d7d5"}, "QGD")

	// Black repertoire
	book.addWhiteLine([]string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5", "a7a6", "b5a4", "g8f6", "e1g1", "f8e7"}, "Ruy Lopez")
	book.addWhiteLine([]string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "f8c5", "c2c3", "d7d6"}, "Giuoco Piano")
	book.addWhiteLine([]string{"e2e4", "c7c5", "g1f3", "b8c6", "f1b5", "g8f6", "b5c6", "d7c6", "d2d3", "c8g4"}, "Sicilian Rossolimo")
	book.addWhiteLine([]string{"e2e4", "c7c5", "g1f3", "b8c6", "d2d4", "c5d4", "f3d4", "g8f6", "d4c6", "b7c6"}, "Sicilian")
	book.addWhiteLine([]string{"e2e4", "c7c5", "g1f3", "b8c6", "d2d4", "c5d4", "f3d4", "g8f6", "b1c3", "e7e6"}, "Sicilian")
	book.addWhiteLine([]string{"d2d4", "d7d5", "c2c4", "e7e6", "g1f3", "g8f6", "b1c3", "f8e7"}, "QGD")
	book.addBlackLine([]string{"d2d4", "d7d5", "c2c4", "e7e6", "b1c3", "g8f6", "c1g5", "f8e7"}, "QGD")
	book.addBlackLine([]string{"d2d4", "g8f6", "c2c4", "e7e6", "b1c3", "f8b4"}, "Nimzo-Indian")
	book.addBlackLine([]string{"c2c4", "e7e6", "d2d4", "g8f6", "g1f3", "d7d5"}, "English")

	return book
}

// addWhiteLine adds a sequence of moves to the white repertoire
func (book *OpeningBook) addWhiteLine(moves []string, name string) {
	current := book.whiteRepertoire
	board := newBoard()

	for i, move := range moves {
		board.makeMoveFromUCI(move)

		found := false
		for _, child := range current.children {
			if child.move == move {
				current = child
				found = true
				break
			}
		}

		if !found {
			newMove := &OpeningMove{
				move:    move,
				zobrist: board.zobrist,
			}
			if i == len(moves)-1 {
				newMove.name = name
			}
			current.children = append(current.children, newMove)
			current = newMove
		}
	}
}

// addBlackLine adds a sequence of moves to the black repertoire
func (book *OpeningBook) addBlackLine(moves []string, name string) {
	current := book.blackRepertoire
	board := newBoard()

	for i, move := range moves {
		board.makeMoveFromUCI(move)

		found := false
		for _, child := range current.children {
			if child.move == move {
				current = child
				found = true
				break
			}
		}

		if !found {
			newMove := &OpeningMove{
				move:    move,
				zobrist: board.zobrist,
			}
			if i == len(moves)-1 {
				newMove.name = name
			}
			current.children = append(current.children, newMove)
			current = newMove
		}
	}
}

// LookupPosition returns the next book move (if any) for the given position
func (book *OpeningBook) LookupPosition(b *Board, color Color) (string, string) {
	// Choose the appropriate repertoire based on color
	root := book.whiteRepertoire
	if color == BLACK {
		root = book.blackRepertoire
	}

	// Find the node matching our current position
	var current *OpeningMove
	var findPosition func(*OpeningMove) *OpeningMove
	findPosition = func(node *OpeningMove) *OpeningMove {
		if node.zobrist == b.zobrist {
			return node
		}

		for _, child := range node.children {
			if found := findPosition(child); found != nil {
				return found
			}
		}
		return nil
	}

	current = findPosition(root)
	if current == nil {
		return "", ""
	}

	// If we have book moves available, randomly select one
	if len(current.children) > 0 {
		randIndex := rand.Intn(len(current.children))
		nextMove := current.children[randIndex]
		return nextMove.move, nextMove.name
	}

	return "", ""
}

// IsInBook checks if the current position is still within the opening book
func (book *OpeningBook) IsInBook(b *Board, color Color) bool {
	move, _ := book.LookupPosition(b, color)
	return move != ""
}
