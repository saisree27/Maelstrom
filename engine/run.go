package engine

import (
	"fmt"
)

func InitializeEverythingExceptTTable() {
	InitializeKingAttacks()
	InitializeKnightAttacks()
	InitializePawnAttacks()
	InitBishopAttacks()
	InitRookAttacks()
	InitSquaresBetween()
	InitLine()
	InitializeSQLookup()
	InitZobrist()
	InitNeighborMasks()
	InitializeLMRTable()
	InitializeNNUE()
}

func Run(command string, position string, depth int) {
	InitializeEverythingExceptTTable()
	InitializeTT(256)
	if command == "search" {
		RunSearch(position, depth)
	}
}

func RunSearch(position string, depth int) {
	s := Searcher{}
	s.Position = NewBoard()

	if position != "startpos" {
		fen := position
		s.Position.InitFEN(fen)
	} else {
		s.Position.InitStartPos()
	}

	s.Position.PrintFromBitBoards()

	for i := 1; i <= depth; i++ {
		line := []Move{}

		fmt.Printf("Depth %d: ", i)

		score := s.Pvs(i, -WIN_VAL-1, WIN_VAL+1, true, &line)
		score *= COLOR_SIGN[s.Position.turn]

		if Timer.Stop {
			break
		}

		fmt.Print(score)
		fmt.Print(" ")

		strLine := []string{}
		for i, _ := range line {
			strLine = append(strLine, line[i].ToUCI())
		}

		fmt.Print(strLine)
		fmt.Println()

		if score == WIN_VAL || score == -WIN_VAL {
			fmt.Println("Found mate.")
			break
		}

	}

	fmt.Println("Done")
}
