package engine

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var PONDERING_ENABLED = false
var IS_PONDERING = false
var PONDER_HIT = false

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

func processGo(command string, b Board) {
	words := strings.Split(command, " ")

	var wtime, btime, winc, binc, depth, movetime, nodes, movestogo int64
	var infinite bool
	var shouldPonder bool

	for i := 0; i < len(words); i++ {
		switch words[i] {
		case "infinite":
			infinite = true
		case "depth":
			if i+1 < len(words) {
				depth, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "nodes":
			if i+1 < len(words) {
				nodes, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "movetime":
			if i+1 < len(words) {
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
				movestogo, _ = strconv.ParseInt(words[i+1], 10, 64)
				i++
			}
		case "ponder":
			shouldPonder = true
			infinite = true
		}
	}

	Timer.Calculate(b.turn, wtime, btime, winc, binc, movestogo, depth, nodes, movetime, infinite)

	// Start search in a goroutine
	go func() {
		// Ponder workflow:
		// Server sends go ponder ...
		// We start ponder search until either the server sends:
		// - ponderhit
		// - stop ...
		// When ponderhit is sent, switch to normal search with reduced movetime
		// When stop is sent, stop search, print bestmove and return
		if PONDERING_ENABLED && shouldPonder && PonderMove.to != PonderMove.from {
			// Start ponder
			IS_PONDERING = true
			SearchPosition(&b)
			IS_PONDERING = false

			if PONDER_HIT {
				Timer.Calculate(b.turn, wtime, btime, winc, binc, movestogo, depth, nodes, movetime, false)
				// Reduce time spent since we got a ponderhit
				Timer.softLimit /= 2
				Timer.hardLimit /= 2

				PONDER_HIT = false
			} else {
				fmt.Println("bestmove")
				return
			}
		}

		bestMove := SearchPosition(&b)

		if PONDERING_ENABLED && PonderMove.to != PonderMove.from {
			fmt.Println("debug ponder move - " + PonderMove.ToUCI())
			fmt.Println("bestmove " + bestMove.ToUCI() + " ponder " + PonderMove.ToUCI())
			return
		}

		fmt.Println("bestmove " + bestMove.ToUCI())
	}()
}

func UciLoop() {
	fmt.Println("initializing...")
	ttSize := int64(256)
	InitializeEverythingExceptTTable()
	InitializeTT(int(ttSize))
	fmt.Println("done, ready for UCI commands")

	b := Board{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		command := scanner.Text()

		if command == "uci" {
			fmt.Println("id name Maelstrom v3.1.0")
			fmt.Println("id author Saigautam Bonam")
			fmt.Println("option name Hash type spin default 256 min 1 max 4096")
			fmt.Println("option name Ponder type check default false")

			if TUNING_EXPOSE_UCI {
				val := reflect.ValueOf(Params)
				typ := val.Type()

				for i := 0; i < val.NumField(); i++ {
					name := typ.Field(i).Name
					defaultVal := val.Field(i).Int()
					fmt.Printf("option name %s type spin default %d min 0 max 10000\n", name, defaultVal)
				}
			}
			fmt.Println("uciok")
		} else if command == "isready" {
			fmt.Println("readyok")
		} else if command == "ucinewgame" {
			b = Board{}
			b.InitStartPos()
			ClearTT()
			ClearHistory()
			ClearKillers()
		} else if command == "quit" {
			os.Exit(0)
		} else if command == "stop" {
			Timer.Stop = true
		} else if command == "ponderhit" {
			PONDER_HIT = true
			// Stop any pondering
			Timer.Stop = true
		} else if strings.Contains(command, "position") {
			// Stop any pondering
			Timer.Stop = true
			b = processPosition(command)
		} else if strings.Contains(command, "go") {
			processGo(command, b)
		} else if strings.Contains(command, "setoption") {
			words := strings.Split(command, " ")
			if len(words) >= 5 && words[1] == "name" && words[3] == "value" {
				paramName := words[2]
				if paramName == "Hash" {
					ttSize, _ = strconv.ParseInt(words[4], 10, 64)
					InitializeTT(int(ttSize))
					continue
				} else if paramName == "Ponder" {
					ponder, _ := strconv.ParseBool(words[4])
					PONDERING_ENABLED = ponder
					continue
				}
				paramValue, err := strconv.Atoi(words[4])
				if err != nil {
					fmt.Println("info string invalid value")
					return
				}

				pVal := reflect.ValueOf(&Params).Elem()
				pField := pVal.FieldByName(paramName)
				if pField.IsValid() && pField.CanSet() && pField.Kind() == reflect.Int {
					pField.SetInt(int64(paramValue))
				}
			}
		} else if command == "d" {
			// Debug command to print current position
			b.PrintFromBitBoards()
		} else if command == "eval" {
			eval := EvaluateNNUE(&b)
			fmt.Println(eval)
		}
	}
}
