# Maelstrom
UCI-compliant Golang chess engine in development from scratch. This is my first experience coding in Go, and so far it's been awesome!

## Features
 - Fast bitboard move generation (magic bitboards for sliding pieces)
 - Iterative deepening principal variation search with aspiration window
 - Null move pruning and transposition table
 - Quiscence search
 - UCI protocol implementation (can use with a UCI-supported GUI such as CuteChess)

## Building from Source
Clone the repository, then run `go build maelstrom/main.go`.
