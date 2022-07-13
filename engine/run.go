package engine

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
}
