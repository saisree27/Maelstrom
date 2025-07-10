package engine

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type UCIManager struct {
	Version                 string
	Author                  string
	SearchThread            Searcher
	HashSize                int64
	PonderingEnabled        bool
	PonderHit               bool
	TunableParams           *TunableParameters
	ExposeTunableParameters bool
}

func (uci *UCIManager) Initialize() {
	fmt.Println("initializing...")

	// Set default UCI options
	uci.HashSize = int64(256)
	uci.PonderingEnabled = false
	uci.TunableParams = &Params
	uci.Version = "v3.1.1"
	uci.Author = "Saigautam Bonam"
	uci.SearchThread = Searcher{}

	// For SPSA tuning, ExposeTunableParameters needs to be true
	uci.ExposeTunableParameters = false

	// Initialize
	InitializeEverythingExceptTTable()
	InitializeTT(int(uci.HashSize))
	uci.SearchThread.Position = NewBoard()

	fmt.Println("done, ready for UCI commands")
}

func (uci *UCIManager) UCI() {
	fmt.Printf("id name Maelstrom %s\n", uci.Version)
	fmt.Printf("id author %s\n", uci.Author)
	fmt.Printf("option name Hash type spin default %d min 1 max 4096\n", uci.HashSize)
	fmt.Printf("option name Ponder type check default %t\n", uci.PonderingEnabled)

	if uci.ExposeTunableParameters {
		val := reflect.ValueOf(*uci.TunableParams)
		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			name := typ.Field(i).Name
			defaultVal := val.Field(i).Int()
			fmt.Printf("option name %s type spin default %d min 0 max 10000\n", name, defaultVal)
		}
	}
	fmt.Println("uciok")
}

func (uci *UCIManager) IsReady() {
	fmt.Println("readyok")
}

func (uci *UCIManager) UCINewGame() {
	uci.SearchThread.Position = NewBoard()
	uci.SearchThread.ClearHistory()
	uci.SearchThread.ClearKillers()
	ClearTT()
}

func (uci *UCIManager) Stop() {
	Timer.Stop = true
}

func (uci *UCIManager) PonderHitUpdate() {
	uci.PonderHit = true
	Timer.Stop = true
}

func (uci *UCIManager) Position(position string) {
	Timer.Stop = true
	*uci.SearchThread.Position = uci.processPosition(position)
}

func (uci *UCIManager) Go(options string) {
	uci.processGo(options)
}

func (uci *UCIManager) processPosition(command string) Board {
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

func (uci *UCIManager) processGo(command string) {
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

	Timer.Calculate(uci.SearchThread.Position.turn, wtime, btime, winc, binc, movestogo, depth, nodes, movetime, infinite)

	// Start search in a goroutine
	go func() {
		// Ponder workflow:
		// Server sends go ponder ...
		// We start ponder search until either the server sends:
		// - ponderhit
		// - stop ...
		// When ponderhit is sent, switch to normal search with reduced movetime
		// When stop is sent, stop search, print bestmove and return
		if uci.PonderingEnabled && shouldPonder && uci.SearchThread.Info.PonderMove.to != uci.SearchThread.Info.PonderMove.from {
			// Start ponder
			uci.SearchThread.Info.IsPondering = true
			uci.SearchThread.SearchPosition()
			uci.SearchThread.Info.IsPondering = false

			if uci.PonderHit {
				Timer.Calculate(uci.SearchThread.Position.turn, wtime, btime, winc, binc, movestogo, depth, nodes, movetime, false)
				// Reduce time spent since we got a ponderhit
				Timer.softLimit /= 2
				Timer.hardLimit /= 2

				uci.PonderHit = false
			} else {
				fmt.Println("bestmove")
				return
			}
		}

		bestMove := uci.SearchThread.SearchPosition()

		if uci.PonderingEnabled && uci.SearchThread.Info.PonderMove.to != uci.SearchThread.Info.PonderMove.from {
			fmt.Println("bestmove " + bestMove.ToUCI() + " ponder " + uci.SearchThread.Info.PonderMove.ToUCI())
			return
		}

		fmt.Println("bestmove " + bestMove.ToUCI())
	}()
}

func (uci *UCIManager) SetOption(option string) {
	words := strings.Split(option, " ")
	if len(words) >= 5 && words[1] == "name" && words[3] == "value" {
		paramName := words[2]
		if paramName == "Hash" {
			uci.HashSize, _ = strconv.ParseInt(words[4], 10, 64)
			InitializeTT(int(uci.HashSize))
			return
		} else if paramName == "Ponder" {
			ponder, _ := strconv.ParseBool(words[4])
			uci.PonderingEnabled = ponder
			uci.SearchThread.Info.PonderingEnabled = ponder
			return
		}
		paramValue, err := strconv.Atoi(words[4])
		if err != nil {
			fmt.Println("info string invalid value")
			return
		}

		pVal := reflect.ValueOf(uci.TunableParams).Elem()
		pField := pVal.FieldByName(paramName)
		if pField.IsValid() && pField.CanSet() && pField.Kind() == reflect.Int {
			pField.SetInt(int64(paramValue))
		}
	}

}

func (uci *UCIManager) DebugPosition() {
	uci.SearchThread.Position.PrintFromBitBoards()
}

func (uci *UCIManager) StaticEvaluate() {
	eval := EvaluateNNUE(uci.SearchThread.Position)
	fmt.Println(eval)
}

func (uci *UCIManager) UciLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		command := scanner.Text()

		if command == "uci" {
			uci.UCI()
		} else if command == "isready" {
			uci.IsReady()
		} else if command == "ucinewgame" {
			uci.UCINewGame()
		} else if command == "quit" {
			os.Exit(0)
		} else if command == "stop" {
			uci.Stop()
		} else if command == "ponderhit" {
			uci.PonderHitUpdate()
		} else if strings.Contains(command, "position") {
			uci.Position(command)
		} else if strings.Contains(command, "go") {
			uci.Go(command)
		} else if strings.Contains(command, "setoption") {
			uci.SetOption(command)
		} else if command == "d" {
			uci.DebugPosition()
		} else if command == "eval" {
			uci.StaticEvaluate()
		}
	}
}
