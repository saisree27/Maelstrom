package engine

import (
	"fmt"
	"strings"
	"time"
)

func initializeEverything() {
	initializeKingAttacks()
	initializeKnightAttacks()
	initializePawnAttacks()
	initBishopAttacks()
	initRookAttacks()
	initSquaresBetween()
	initLine()
	initializeSQLookup()
	initZobrist()
	initializeTTable()
	initNeighborMasks()
}

func Run(command string, position string, depth int) {
	initializeEverything()
	if command == "perft" {
		RunPerfTests(position, depth)
	}
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

		score := pvs(&b, i, i, -winVal-1, winVal+1, b.turn, true, &line, 100000000, time.Now()) * factor[b.turn]

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
}

func RunSelfPlay(position string, depth int) {
	b := Board{}
	if position == "startpos" {
		b.InitStartPos()
	} else {
		fen := position
		b.InitFEN(fen)
	}

	movesPlayed := []string{}

	legals := b.generateLegalMoves()
	newDepth := depth
	for {
		if len(legals) == 0 {
			break
		}

		b.printFromBitBoards()

		line := []Move{}

		score := 0

		if len(b.generateLegalMoves()) > 30 {
			newDepth--
		} else {
			fmt.Println("Increasing depth")
			newDepth++
		}
		if newDepth < 5 {
			newDepth = 5
		}
		if b.plyCnt <= 20 {
			if newDepth >= 6 {
				newDepth = 6
			}
		}

		for i := 1; i <= newDepth; i++ {
			line = []Move{}

			fmt.Printf("Depth %d: ", i)

			score = pvs(&b, i, i, -winVal-1, winVal+1, b.turn, true, &line, 100000000, time.Now())

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
		}
		bestMove := line[0]

		fmt.Printf("SCORE: %d\n", score*factor[b.turn])

		movesPlayed = append(movesPlayed, bestMove.toUCI())
		b.makeMove(bestMove)

		fmt.Println(movesPlayed)
	}
}

func RunPlay(position string, depth int, player Color) {
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

				score = pvs(&b, i, i, -winVal-1, winVal+1, b.turn, true, &line, 100000, time.Now())

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
