package engine

import (
	"fmt"
	"strings"
)

func uciLoop() {
	b := Board{}
	for {
		var command string
		fmt.Scanln(&command)

		if command == "uci" {
			fmt.Println("id name Maelstrom")
			fmt.Println("id author saisree27")
			initializeEverything()
			fmt.Println("uciok")
		}

		if command == "isready" {
			fmt.Println("readyok")
		}

		if command == "ucinewgame" {
			b = Board{}
		}

		if strings.Contains(command, "position") {
			words := strings.Split(command, " ")
			if len(words) == 2 {
				b.InitStartPos()
			} else {
				b.InitFEN(words[2])
			}
		}

	}
}
