package main

import (
	"maelstrom/engine"
)

func main() {
	uci := engine.UCIManager{}
	uci.Initialize()
	uci.UciLoop()
}
