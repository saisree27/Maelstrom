package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func processPosition(command string) Board {
	b := Board{}
	words := strings.Split(command, " ")
	length := len(words)
	mvStart := 2

	if words[1] != "fen" {
		b.InitStartPos()
		if length == 2 {
			return b
		}
	} else {
		fen := words[2] + " " + words[3] + " " + words[4]
		fen += " " + words[5] + " " + words[6] + " " + words[7]
		b.InitFEN(fen)
		if length == 8 {
			return b
		}
		mvStart = 8
	}

	for i := mvStart + 1; i < length; i++ {
		b.makeMoveFromUCI(words[i])
	}

	return b
}

func processGo(command string, b *Board) {
	words := strings.Split(command, " ")

	var wtime, btime, winc, binc int64

	// default move time = 30s
	var movetime = int64(30000)
	var movetimeSet = false

	for i, word := range words {
		switch word {
		case "movetime":
			movetimeSet = true
			movetime, _ = strconv.ParseInt(words[i+1], 10, 64)
		case "wtime":
			wtime, _ = strconv.ParseInt(words[i+1], 10, 64)
		case "winc":
			winc, _ = strconv.ParseInt(words[i+1], 10, 64)
		case "btime":
			btime, _ = strconv.ParseInt(words[i+1], 10, 64)
		case "binc":
			binc, _ = strconv.ParseInt(words[i+1], 10, 64)
		}
	}

	var bestMove Move

	if movetimeSet {
		bestMove = searchWithTime(b, movetime)
	} else {
		if b.turn == WHITE {
			movetime = wtime/35 + winc*2
			bestMove = searchWithTime(b, movetime)
		} else {
			movetime = btime/35 + binc*2
			bestMove = searchWithTime(b, movetime)
		}
	}

	fmt.Println("bestmove " + bestMove.toUCI())
	// clearTTable()
}

func UciLoop() {
	b := Board{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		command := scanner.Text()

		if command == "uci" {
			initializeEverything()
			fmt.Println("id name Maelstrom")
			fmt.Println("id author saisree27")
			fmt.Println("uciok")
		}

		if command == "isready" {
			fmt.Println("readyok")
		}

		if command == "ucinewgame" {
			b = Board{}
			b.InitStartPos()
		}

		if strings.Contains(command, "position") {
			b = processPosition(command)
		}

		if strings.Contains(command, "go") {
			processGo(command, &b)
		}
	}

}
