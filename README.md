# Maelstrom
UCI-compliant Golang chess engine in development from scratch. This is my first experience coding in Go, and so far it's been awesome!

## Features
 - Fast bitboard move generation (magic bitboards for sliding pieces)
 - Iterative deepening principal variation search with aspiration window
 - Null move pruning and transposition table
 - Move ordering and late move reductions
 - Quiescence search
 - Evaluation utilizing material, piece square tables, pawn structure, and more 
 - UCI protocol implementation, so you can run the engine using a UCI-supported GUI such as [CuteChess](https://github.com/cutechess/cutechess/releases).

## Releases
I plan to release binaries for Windows and MacOS available soon, once I am happy with Maelstrom's strength as a starting point. I also plan to host Maelstrom on a Lichess bot account so you can play against it online.

## Building from Source
Clone the repository, then run `go build maelstrom/main.go`. The engine binary will be built into the project root folder.

## References
Definitely the most helpful reference in developing this engine for me has been the Chess Programming [wiki](https://www.chessprogramming.org/Main_Page)! If you're interested in developing your own chess engine or move library, this website has everything.
