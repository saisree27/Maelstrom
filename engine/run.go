package engine

import (
	"fmt"
	"strings"
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
	InitializePawnMasks()
	InitializeLMRTable()
	InitializeNNUE()
}

func Run(command string, position string, depth int) {
	InitializeEverythingExceptTTable()
	InitializeTT(256)
	if command == "search" {
		RunSearch(position, depth)
	}
	if command == "selfplay" {
		RunSelfPlay(position, depth)
	}
	if strings.Contains(command, "play") {
		color := strings.Split(command, " ")[1]
		if color == "white" {
			RunPlay(position, depth, WHITE)
		} else {
			RunPlay(position, depth, BLACK)
		}
	}
}

func RunSearch(position string, depth int) {
	b := Board{}

	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}

	b.PrintFromBitBoards()

	for i := 1; i <= depth; i++ {
		line := []Move{}

		fmt.Printf("Depth %d: ", i)

		score, timeout := Pvs(&b, i, i, -WIN_VAL-1, WIN_VAL+1, b.turn, true, &line)
		score *= COLOR_SIGN[b.turn]

		if timeout {
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

func RunSelfPlay(position string, depth int) {
	InitializeEverythingExceptTTable()
	InitializeTT(256)
	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}

	movesPlayed := []string{}

	legals := b.GenerateLegalMoves()

	for {
		if len(legals) == 0 {
			break
		}

		b.PrintFromBitBoards()
		score := 0
		bestMove := SearchWithTime(&b, 10000)
		b.PrintFromBitBoards()

		fmt.Printf("SCORE: %d\n", score*COLOR_SIGN[b.turn])

		movesPlayed = append(movesPlayed, bestMove.ToUCI())
		b.MakeMove(bestMove)

		fmt.Println(movesPlayed)
	}
}

func RunPlay(position string, depth int, player Color) {
	InitializeEverythingExceptTTable()
	InitializeTT(256)
	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}
	movesPlayed := []string{}

	legals := b.GenerateLegalMoves()
	for {
		if len(legals) == 0 {
			break
		}

		b.PrintFromBitBoards()

		if b.turn == player {
			fmt.Println("Your turn: ")
			var move string

			fmt.Scanln(&move)

			b.MakeMoveFromUCI(move)
			movesPlayed = append(movesPlayed, move)
		} else {
			line := []Move{}

			score := 0

			for i := 1; i <= depth; i++ {
				line = []Move{}

				fmt.Printf("Depth %d: ", i)

				score, _ = Pvs(&b, i, i, -WIN_VAL-1, WIN_VAL+1, b.turn, true, &line)

				if score == WIN_VAL || score == -WIN_VAL {
					fmt.Println("Found mate.")
					break
				}

				fmt.Print(score * COLOR_SIGN[b.turn])
				fmt.Print(" ")

				strLine := []string{}
				for i, _ := range line {
					strLine = append(strLine, line[i].ToUCI())
				}

				fmt.Print(strLine)
				fmt.Println()

			}
			bestMove := line[0]

			fmt.Printf("SCORE: %d\n", score*COLOR_SIGN[b.turn])

			movesPlayed = append(movesPlayed, bestMove.ToUCI())
			b.MakeMove(bestMove)

			fmt.Print("Moves played: ")
			fmt.Println(movesPlayed)
		}
	}
}
