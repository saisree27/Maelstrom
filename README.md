# Maelstrom
![](https://github.com/saisree27/Maelstrom/actions/workflows/go.yml/badge.svg)

<p align="center">
  <img src="maelstrom-logo.png" />
</p>

UCI-compliant Golang chess engine in development from scratch.

## Play/watch games
Maelstrom often plays on Lichess [here](https://lichess.org/@/Maelstrom-Chess). Please feel free to challenge the engine on lichess whenever it is online.

## Features
 - Fast bitboard move generation (magic bitboards for sliding pieces)
 - Iterative deepening principal variation search with aspiration windows
 - Move ordering with MVV-LVA, history, and killer heuristic
 - Transposition table
 - Null move pruning
 - Static null move pruning
 - Late move reductions
 - Check extensions
 - Razoring
 - Quiescence search with delta pruning
 - Hand-crafted evaluation (HCE) utilizing PeSTO tapered evaluation, mobility, pawn structure, bishop pair, and pawn shield king safety
 - UCI protocol implementation, so you can run the engine using a UCI-supported GUI such as [CuteChess](https://github.com/cutechess/cutechess/releases).
 - Time management
 - Custom opening book (by default opening book is off, but can be enabled with `setoption name UseBook value true`)
 - Integration with lichess 7-man tablebase (by default tablebase is off, but can be enabled with `setoption name UseLichessTB value true`) 

## Releases
Checkout and download binaries and source code from the Releases page.

## Building from Source
Clone the repository, then run `go build maelstrom/main.go`. The engine binary will be built into the project root folder as the binary `main`. Run this executable to start the CLI, which uses the [UCI-protocol](https://official-stockfish.github.io/docs/stockfish-wiki/UCI-&-Commands.html).
Enter the following commands to run the engine on starting position from binary:

```
uci
...
isready
...
position startpos
go infinite
```

## Engine Testing

SPRT command:
```
cutechess-cli -engine proto=uci cmd={BINARY_TO_TEST} name={TEST_NAME} -engine proto=uci cmd={EXISTING_VERISON_BINARY} name={EXISTING_NAME} -each tc=8+0.08 option.Hash=32 -games 2 -rounds 1000 -repeat -concurrency 8 -openings file={PATH_TO_EPD} format=epd order=random -pgnout {PATH_TO_PGN} -sprt elo0=0 elo1=5 alpha=0.05 beta=0.1 -ratinginterval 10
```

## References
Definitely the most helpful reference in developing this engine for me has been the Chess Programming [wiki](https://www.chessprogramming.org/Main_Page)! If you're interested in developing your own chess engine or move library, this website has everything. Also want to shoutout [Blunder](https://github.com/deanmchris/blunder) as a great and readable reference for helping me improve the engine!
