package main

import "maelstrom/engine"

func main() {
	b := engine.Board{}
	b.InitStartPos()
	engine.Search(&b)
}
