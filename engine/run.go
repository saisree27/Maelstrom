package engine

import "strings"

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
