package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Global channel to signal search stop
var stopChannel chan struct{}
var searchMutex sync.Mutex
var isSearching bool
var useOpeningBook bool = false

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
		if words[2] == "moves" {
			mvStart = 3
		}
	} else {
		fen := words[2] + " " + words[3] + " " + words[4]
		fen += " " + words[5] + " " + words[6] + " " + words[7]
		b.InitFEN(fen)
		if length == 8 {
			return b
		}
		mvStart = 8
		if length > 8 && words[8] == "moves" {
			mvStart = 9
		}
	}

	for i := mvStart; i < length; i++ {
		b.makeMoveFromUCI(words[i])
	}

	return b
}

func processGo(command string, b *Board) {
	searchMutex.Lock()
	if isSearching {
		searchMutex.Unlock()
		return
	}
	isSearching = true
	searchMutex.Unlock()

	// Create a new stop channel for this search
	stopChannel = make(chan struct{})

	words := strings.Split(command, " ")

	var wtime, btime, winc, binc, depth, movetime int64
	var infinite bool
	movetimeSet := false

	for i := 0; i < len(words); i++ {
		switch words[i] {
		case "infinite":
			infinite = true
		case "depth":
			if i+1 < len(words) {
				depth, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "movetime":
			if i+1 < len(words) {
				movetimeSet = true
				movetime, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "wtime":
			if i+1 < len(words) {
				wtime, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "btime":
			if i+1 < len(words) {
				btime, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "winc":
			if i+1 < len(words) {
				winc, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "binc":
			if i+1 < len(words) {
				binc, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		}
	}

	// Start search in a goroutine
	go func() {
		var bestMove Move

		if infinite {
			movetime = 1000000000 // Use a very large time for infinite search
		} else if depth > 0 {
			movetime = 1000000000 // Use a very large time when searching to a specific depth
		} else if movetimeSet {
			// Use specified movetime
		} else {
			// Calculate time based on remaining time and increment
			if b.turn == WHITE {
				if winc >= 0 {
					movetime = wtime/25 + winc - 200
				} else {
					movetime = wtime / 30
				}
			} else {
				if binc >= 0 {
					movetime = btime/25 + binc - 200
				} else {
					movetime = btime / 30
				}
			}
		}

		bestMove = searchWithTime(b, movetime)

		// Only output bestmove if we're still searching (i.e., not stopped)
		searchMutex.Lock()
		if isSearching {
			fmt.Println("bestmove " + bestMove.toUCI())
		}
		isSearching = false
		searchMutex.Unlock()

		if b.plyCnt%10 == 0 {
			clearTTable()
		}
	}()
}

func UciLoop() {
	ttSize := int64(256)

	b := Board{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		command := scanner.Text()

		if command == "uci" {
			initializeEverythingExceptTTable()
			fmt.Println("id name Maelstrom")
			fmt.Println("id author saisree27")
			fmt.Println("option name Hash type spin default 256 min 1 max 1024")
			fmt.Println("uciok")
		} else if command == "isready" {
			initializeTTable(int(ttSize))
			fmt.Println("readyok")
		} else if command == "ucinewgame" {
			b = Board{}
			b.InitStartPos()
			clearTTable()
		} else if command == "quit" {
			os.Exit(0)
		} else if command == "stop" {
			searchMutex.Lock()
			if isSearching && stopChannel != nil {
				close(stopChannel)
			}
			searchMutex.Unlock()
		} else if command == "ponderhit" {
			// Currently we don't support pondering, so treat it like a regular move
			continue
		} else if strings.Contains(command, "position") {
			b = processPosition(command)
		} else if strings.Contains(command, "go") {
			processGo(command, &b)
		} else if strings.Contains(command, "setoption") {
			words := strings.Split(command, " ")
			if len(words) >= 5 && words[1] == "name" && words[2] == "Hash" && words[3] == "value" {
				ttSize, _ = strconv.ParseInt(words[4], 10, 64)
				initializeTTable(int(ttSize))
			}

			if words[1] == "name" && words[2] == "UseBook" && words[3] == "value" {
				useOpeningBook, _ = strconv.ParseBool(words[4])
			}
		} else if command == "d" {
			// Debug command to print current position
			b.printFromBitBoards()
		}
	}
}
