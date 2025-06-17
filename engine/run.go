package engine

import (
	"fmt"
	"strings"
)

func initializeEverythingExceptTTable() {
	initializeKingAttacks()
	initializeKnightAttacks()
	initializePawnAttacks()
	initBishopAttacks()
	initRookAttacks()
	initSquaresBetween()
	initLine()
	initializeSQLookup()
	initZobrist()
	initNeighborMasks()
	initializePawnMasks()
}

func Run(command string, position string, depth int) {
	initializeEverythingExceptTTable()
	initializeTTable(256)
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

	b.printFromBitBoards()

	for i := 1; i <= depth; i++ {
		line := []Move{}

		fmt.Printf("Depth %d: ", i)

		score, timeout := pvs(&b, i, i, -winVal-1, winVal+1, b.turn, true, &line)
		score *= factor[b.turn]

		if timeout {
			break
		}

		fmt.Print(score)
		fmt.Print(" ")

		strLine := []string{}
		for i, _ := range line {
			strLine = append(strLine, line[i].toUCI())
		}

		fmt.Print(strLine)
		fmt.Println()

		if score == winVal || score == -winVal {
			fmt.Println("Found mate.")
			break
		}

	}

	fmt.Println("Done")
}

func RunSelfPlay(position string, depth int) {
	initializeEverythingExceptTTable()
	initializeTTable(256)
	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}

	movesPlayed := []string{}

	legals := b.generateLegalMoves()

	for {
		if len(legals) == 0 {
			break
		}

		b.printFromBitBoards()
		score := 0
		bestMove := searchWithTime(&b, 10000)
		b.printFromBitBoards()

		fmt.Printf("SCORE: %d\n", score*factor[b.turn])

		movesPlayed = append(movesPlayed, bestMove.toUCI())
		b.makeMove(bestMove)

		fmt.Println(movesPlayed)
	}
}

func RunPlay(position string, depth int, player Color) {
	initializeEverythingExceptTTable()
	initializeTTable(256)
	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}
	movesPlayed := []string{}

	legals := b.generateLegalMoves()
	for {
		if len(legals) == 0 {
			break
		}

		b.printFromBitBoards()

		if b.turn == player {
			fmt.Println("Your turn: ")
			var move string

			fmt.Scanln(&move)

			b.makeMoveFromUCI(move)
			movesPlayed = append(movesPlayed, move)
		} else {
			line := []Move{}

			score := 0

			for i := 1; i <= depth; i++ {
				line = []Move{}

				fmt.Printf("Depth %d: ", i)

				score, _ = pvs(&b, i, i, -winVal-1, winVal+1, b.turn, true, &line)

				if score == winVal || score == -winVal {
					fmt.Println("Found mate.")
					break
				}

				fmt.Print(score * factor[b.turn])
				fmt.Print(" ")

				strLine := []string{}
				for i, _ := range line {
					strLine = append(strLine, line[i].toUCI())
				}

				fmt.Print(strLine)
				fmt.Println()

			}
			bestMove := line[0]

			fmt.Printf("SCORE: %d\n", score*factor[b.turn])

			movesPlayed = append(movesPlayed, bestMove.toUCI())
			b.makeMove(bestMove)

			fmt.Print("Moves played: ")
			fmt.Println(movesPlayed)
		}
	}
}
