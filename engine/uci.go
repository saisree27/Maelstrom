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
var StopChannel chan struct{}
var SearchMutex sync.Mutex
var IsSearching bool
var UseOpeningBook bool = false
var UseTablebase bool = false

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
		b.MakeMoveFromUCI(words[i])
	}

	return b
}

func processGo(command string, b *Board) {
	SearchMutex.Lock()
	if IsSearching {
		SearchMutex.Unlock()
		return
	}
	IsSearching = true
	SearchMutex.Unlock()

	// Create a new stop channel for this search
	StopChannel = make(chan struct{})

	words := strings.Split(command, " ")

	var wtime, btime, winc, binc, depth, movetime int64
	var infinite bool
	movetimeSet := false
	movesToGo := int64(-1)

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
		case "movestogo":
			if i+1 < len(words) {
				movesToGo, _ = strconv.ParseInt(words[i+1], 10, 64)
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
		} else if movesToGo > 0 {
			if b.turn == WHITE {
				movetime = wtime/movesToGo - 400
			} else {
				movetime = btime/movesToGo - 400
			}
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

		bestMove = SearchWithTime(b, movetime)

		// Only output bestmove if we're still searching (i.e., not stopped)
		SearchMutex.Lock()
		if IsSearching {
			fmt.Println("bestmove " + bestMove.ToUCI())
		}
		IsSearching = false
		SearchMutex.Unlock()
	}()
}

func UciLoop() {
	ttSize := int64(256)
	InitializeEverythingExceptTTable()
	InitializeTT(int(ttSize))

	b := Board{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		command := scanner.Text()

		if command == "uci" {
			fmt.Println("id name Maelstrom v2.1.0")
			fmt.Println("id author saisree27")
			fmt.Println("option name Hash type spin default 256 min 1 max 4096")
			fmt.Println("option name UseBook type check default false")
			fmt.Println("option name UseLichessTB type check default false")
			fmt.Println("uciok")
		} else if command == "isready" {
			fmt.Println("readyok")
		} else if command == "ucinewgame" {
			b = Board{}
			b.InitStartPos()
			ClearTT()
		} else if command == "quit" {
			os.Exit(0)
		} else if command == "stop" {
			SearchMutex.Lock()
			if IsSearching && StopChannel != nil {
				close(StopChannel)
			}
			SearchMutex.Unlock()
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
				InitializeTT(int(ttSize))
			}

			if words[1] == "name" && words[2] == "UseBook" && words[3] == "value" {
				UseOpeningBook, _ = strconv.ParseBool(words[4])
			}

			if words[1] == "name" && words[2] == "UseLichessTB" && words[3] == "value" {
				UseTablebase, _ = strconv.ParseBool(words[4])
			}
		} else if command == "d" {
			// Debug command to print current position
			b.PrintFromBitBoards()
		}
	}
}
