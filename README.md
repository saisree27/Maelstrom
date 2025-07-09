<div align="center">
  <img src="maelstrom-logo.png" width="250" height="250" style="border-radius:5%">
</div>

<div align="center">

  ![](https://github.com/saisree27/Maelstrom/actions/workflows/go.yml/badge.svg)
  ![](https://img.shields.io/github/v/release/saisree27/Maelstrom)
  ![](https://img.shields.io/github/commits-since/saisree27/Maelstrom/v3.1.0)

</div>

# Maelstrom

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
 - Futility pruning
 - Quiescence search
 - Static Exchange Evaluation (SEE) pruning and move ordering
 - NNUE Evaluation using a (768->256)x2 -> 1 architecture using a SIMD SCReLU activation function, trained on Lc0/SF data
 - UCI protocol implementation, so you can run the engine using a UCI-supported GUI such as [CuteChess](https://github.com/cutechess/cutechess/releases)
 - Time management with soft/hard bounds
 - Pondering

## Releases
Checkout and download binaries and source code from the Releases page.

## Elo Progression
Table summarizing Elo progression tests documented in release notes as well as rating by CCRL. Time controls I used for testing: STC 8s+0.08s, LTC 40s+0.4s.
|        Version      |   STC Elo |  LTC Elo | Estimated Elo (vs. Stash) | CCRL Blitz |
|:-------------------:|:-----------:|:---------:|:------------:|:------------:|
| v3.1.0    | +258.7 +/- 51.8 | +165.7 +/- 37.6 | ~3040 |              |
| v3.0.0    | +343 +/- 65.5 | +515.5 +/- 119.7 | ~2820 |              |
| v2.1.0    | +249 +/- 46.1 | +190.8 +/- 70.1  | |              |
| v2.0.0    |              |                 | |     [2109](https://computerchess.org.uk/ccrl/404/cgi/engine_details.cgi?print=Details&each_game=1&eng=Maelstrom%202.0.0%2064-bit#Maelstrom_2_0_0_64-bit)     |

## Building from Source
Requirements:
- go version 1.23.0 or later
- any C compiler
- AVX2 enabled processor (if not enabled, update `engine/screlu/screlu.go` with `AVX2_ENABLED=false` and remove the AVX2 CFLAGS)

Clone the repository, then run `go build maelstrom/main.go`. The engine binary will be built into the project root folder as the binary `main`. Run this executable to start the CLI, which uses the [UCI-protocol](https://official-stockfish.github.io/docs/stockfish-wiki/UCI-&-Commands.html).
Enter the following commands to run the engine on starting position from binary:

```
> uci
id name Maelstrom v3.1.0
id author Saigautam Bonam
option name Hash type spin default 256 min 1 max 4096
option name Ponder type check default false
uciok
> isready
readyok
> position startpos
> go infinite
```

## Engine Testing

SPRT command:
```
cutechess-cli -engine proto=uci cmd={BINARY_TO_TEST} name={TEST_NAME} -engine proto=uci cmd={EXISTING_VERISON_BINARY} name={EXISTING_NAME} -each tc=8+0.08 option.Hash=32 -games 2 -rounds 1000 -repeat -concurrency 8 -openings file={PATH_TO_EPD} format=epd order=random -pgnout {PATH_TO_PGN} -sprt elo0=0 elo1=5 alpha=0.05 beta=0.1 -ratinginterval 10
```

## References and Acknowledgements
- Definitely the most helpful reference in developing this engine for me has been the [Chess Programming wiki](https://www.chessprogramming.org/Main_Page)! If you're interested in developing your own chess engine or move library, this website has everything.
- Engine references that helped me improve the engine:
  - [Blunder](https://github.com/deanmchris/blunder)
  - [Carballo](https://github.com/albertoruibal/carballo)
  - [Ethereal](https://github.com/AndyGrant/Ethereal.git)
  - [Stockfish](https://github.com/official-stockfish/Stockfish)
  - [Zahak](https://github.com/amanjpro/zahak)
  - [Stormphrax](https://github.com/Ciekce/Stormphrax)
  - [Viridithas](https://github.com/cosmobobak/viridithas)
  - [Alexandria](https://github.com/PGG106/Alexandria)
  - [Stash](https://gitlab.com/mhouppin/stash-bot)
  - [Starzix](https://github.com/zzzzz151/Starzix)
  - and many more open-source engines, these are just the ones I can name off the top of my head!
- [bullet](https://github.com/jw1912/bullet) for allowing me to easily train the NNUE.
- Engine Programmers and Stockfish discord servers for their huge knowledge base/resources and advice.
- Huge thanks to Gabor Szots and the folks at CCRL for rating the engine!
