package engine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"maelstrom/engine/screlu"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// NETWORK ARCHITECTURE:
// 768 -> (1024)x2 -> 1 (Perspective Network)
// This means there are two sets of inputs, x and x^ where x is from STM perspective, x^ is from NTM perspective
// Input size: 768 flattened = 12 * 64. Indices 0 through 11 will represent friendly_pawn, friendly_bishop, ..., enemy_king
// We will have two accumulators:
// 	a  = Hx  + b
//  a^ = Hx^ + b
// where H is the hidden layer weights and b is the hidden layer biases
// The inference will then be:
//  y = O(concat(a + a^)) + c
// where O is the output layer weights and c is the output layer biases

const WEIGHTS_FILENAME = "network.bin"

const INPUT_LAYER_SIZE = 768
const HIDDEN_LAYER_SIZE = 128
const OUTPUT_LAYER_SIZE = 1

// Quantization constants
const QA int16 = 255
const QB int16 = 64
const SCALE int32 = 400

type Accumulator struct {
	values [HIDDEN_LAYER_SIZE]int16
}

type AccumulatorPair struct {
	white Accumulator
	black Accumulator
}

type NNUE struct {
	accumulator_weights [INPUT_LAYER_SIZE][HIDDEN_LAYER_SIZE]int16
	accumulator_biases  [HIDDEN_LAYER_SIZE]int16
	output_weights      [2 * HIDDEN_LAYER_SIZE]int16
	output_bias         int16
}

func AccumulatorAdd(network *NNUE, accumulator *Accumulator, index int) {
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		accumulator.values[i] += network.accumulator_weights[index][i]
	}
}

func AccumulatorSub(network *NNUE, accumulator *Accumulator, index int) {
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		accumulator.values[i] -= network.accumulator_weights[index][i]
	}
}

func (nnue *NNUE) RecomputeAccumulators(b *Board) AccumulatorPair {
	var whiteAccumulator Accumulator
	var blackAccumulator Accumulator

	for sq := A1; sq <= H8; sq++ {
		piece := b.squares[sq]
		if piece == EMPTY {
			continue
		}

		pieceType := PieceToPieceType(piece)
		pieceColor := piece.GetColor()

		whiteIndex := CalculateIndex(WHITE, sq, pieceType, pieceColor)
		blackIndex := CalculateIndex(BLACK, sq, pieceType, pieceColor)

		AccumulatorAdd(nnue, &whiteAccumulator, whiteIndex)
		AccumulatorAdd(nnue, &blackAccumulator, blackIndex)
	}

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		whiteAccumulator.values[i] += nnue.accumulator_biases[i]
		blackAccumulator.values[i] += nnue.accumulator_biases[i]
	}

	return AccumulatorPair{white: whiteAccumulator, black: blackAccumulator}
}

func (nnue *NNUE) AddFeature(perspective Color, acc *Accumulator, sq Square, pieceType PieceType, color Color) {
	index := CalculateIndex(perspective, sq, pieceType, color)
	AccumulatorAdd(nnue, acc, index)
}

func (nnue *NNUE) SubFeature(perspective Color, acc *Accumulator, sq Square, pieceType PieceType, color Color) {
	index := CalculateIndex(perspective, sq, pieceType, color)
	AccumulatorSub(nnue, acc, index)
}

func (network *NNUE) UpdateAccumulatorOnMove(b *Board, move Move, stm Color) {
	from := move.from
	to := move.to

	movingPiece := move.piece
	pieceType := PieceToPieceType(movingPiece)
	color := movingPiece.GetColor()

	// Subtract moving piece from old square
	network.SubFeature(WHITE, &b.accumulators.white, from, pieceType, color)
	network.SubFeature(BLACK, &b.accumulators.black, from, pieceType, color)

	// If capture, subtract the captured piece at destination
	if move.movetype == CAPTURE || move.movetype == EN_PASSANT || move.movetype == CAPTURE_AND_PROMOTION {
		captureSquare := to
		capturedPiece := move.captured
		if move.movetype == EN_PASSANT {
			if color == WHITE {
				captureSquare = move.to.GoDirection(SOUTH)
				capturedPiece = B_P
			} else {
				captureSquare = move.to.GoDirection(NORTH)
				capturedPiece = W_P
			}
		}

		capturedType := PieceToPieceType(capturedPiece)
		capturedColor := ReverseColor(color)

		network.SubFeature(WHITE, &b.accumulators.white, captureSquare, capturedType, capturedColor)
		network.SubFeature(BLACK, &b.accumulators.black, captureSquare, capturedType, capturedColor)
	}

	// Add moving piece to new square (or promotion)
	newPieceType := pieceType
	if move.movetype == PROMOTION || move.movetype == CAPTURE_AND_PROMOTION {
		newPieceType = PieceToPieceType(move.promote)
	}
	network.AddFeature(WHITE, &b.accumulators.white, to, newPieceType, color)
	network.AddFeature(BLACK, &b.accumulators.black, to, newPieceType, color)

	// Handle castling
	if move.movetype == K_CASTLE {
		if color == WHITE {
			// Subtract White rook from H1
			network.SubFeature(WHITE, &b.accumulators.white, H1, ROOK, WHITE)
			network.SubFeature(BLACK, &b.accumulators.black, H1, ROOK, WHITE)

			// Add White rook to F1
			network.AddFeature(WHITE, &b.accumulators.white, F1, ROOK, WHITE)
			network.AddFeature(BLACK, &b.accumulators.black, F1, ROOK, WHITE)
		} else {
			// Subtract Black rook from H8
			network.SubFeature(WHITE, &b.accumulators.white, H8, ROOK, BLACK)
			network.SubFeature(BLACK, &b.accumulators.black, H8, ROOK, BLACK)

			// Add Black rook to F8
			network.AddFeature(WHITE, &b.accumulators.white, F8, ROOK, BLACK)
			network.AddFeature(BLACK, &b.accumulators.black, F8, ROOK, BLACK)
		}
	} else if move.movetype == Q_CASTLE {
		if color == WHITE {
			// Subtract White rook from A1
			network.SubFeature(WHITE, &b.accumulators.white, A1, ROOK, WHITE)
			network.SubFeature(BLACK, &b.accumulators.black, A1, ROOK, WHITE)

			// Add White rook to D1
			network.AddFeature(WHITE, &b.accumulators.white, D1, ROOK, WHITE)
			network.AddFeature(BLACK, &b.accumulators.black, D1, ROOK, WHITE)
		} else {
			// Subtract Black rook from A8
			network.SubFeature(WHITE, &b.accumulators.white, A8, ROOK, BLACK)
			network.SubFeature(BLACK, &b.accumulators.black, A8, ROOK, BLACK)

			// Add Black rook to D8
			network.AddFeature(WHITE, &b.accumulators.white, D8, ROOK, BLACK)
			network.AddFeature(BLACK, &b.accumulators.black, D8, ROOK, BLACK)
		}
	}
}

func Forward(nnue *NNUE, stmAccumulator *Accumulator, ntmAccumulator *Accumulator) int32 {
	eval := screlu.SCReLUFusedSIMDSum(
		stmAccumulator.values[:],
		ntmAccumulator.values[:],
		nnue.output_weights[:],
		QA,
	)

	// Need this scaling when using SCReLU
	eval /= int32(QA)

	eval += int32(nnue.output_bias)
	eval *= SCALE
	eval /= int32(QA * QB)

	return eval
}

func CalculateIndex(perspective Color, sq Square, pieceType PieceType, side Color) int {
	if perspective == BLACK {
		side = ReverseColor(side)
		sq = Square(REVERSE_SQUARE[sq])
	}

	return int(side)*64*6 + int(pieceType)*64 + int(sq)
}

func NewRandomNNUE() NNUE {
	rand.Seed(time.Now().UnixNano())

	network := NNUE{}

	// Initialize accumulator weights: [INPUT_LAYER_SIZE][HIDDEN_LAYER_SIZE]
	for i := 0; i < INPUT_LAYER_SIZE; i++ {
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			network.accumulator_weights[i][j] = int16(rand.Intn(256) - 128) // [-128, 127]
		}
	}

	// Initialize accumulator biases: [HIDDEN_LAYER_SIZE]
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		network.accumulator_biases[i] = int16(rand.Intn(256) - 128) // [-128, 127]
	}

	// Initialize output weights: [2 * HIDDEN_LAYER_SIZE]
	for i := 0; i < 2*HIDDEN_LAYER_SIZE; i++ {
		network.output_weights[i] = int16(rand.Intn(128) - 64) // [-64, 63]
	}

	// Initialize output bias
	network.output_bias = int16(rand.Intn(128) - 64)

	return network
}

func LoadNNUEFromFile(path string) (*NNUE, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	reader := bytes.NewReader(data)
	nnue := &NNUE{}

	readInt16 := func(dst *int16) error {
		return binary.Read(reader, binary.LittleEndian, dst)
	}

	for i := 0; i < INPUT_LAYER_SIZE; i++ {
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			if err := readInt16(&nnue.accumulator_weights[i][j]); err != nil {
				return nil, fmt.Errorf("accumulator_weights[%d][%d]: %w", i, j, err)
			}
		}
	}

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		if err := readInt16(&nnue.accumulator_biases[i]); err != nil {
			return nil, fmt.Errorf("accumulator_biases[%d]: %w", i, err)
		}
	}

	for i := 0; i < 2*HIDDEN_LAYER_SIZE; i++ {
		if err := readInt16(&nnue.output_weights[i]); err != nil {
			return nil, fmt.Errorf("output_weights[%d]: %w", i, err)
		}
	}

	if err := readInt16(&nnue.output_bias); err != nil {
		return nil, fmt.Errorf("output_bias: %w", err)
	}

	return nnue, nil
}

var GlobalNNUE NNUE

func GetProjectRootPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not get caller info")
	}
	dir := filepath.Dir(filename)
	for !folderExists(filepath.Join(dir, "nn_weights")) && dir != filepath.Dir(dir) {
		dir = filepath.Dir(dir)
	}
	if !folderExists(filepath.Join(dir, "nn_weights")) {
		return "", fmt.Errorf("nn_weights folder not found")
	}
	return dir, nil
}

func folderExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
func InitializeNNUE() {
	root, err := GetProjectRootPath()
	if err != nil {
		log.Fatalf("Failed to find nn_weights folder: %v", err)
	}
	path := filepath.Join(root, "nn_weights", WEIGHTS_FILENAME)
	fmt.Printf("loading NNUE weights from file %v\n", path)
	nnue, err := LoadNNUEFromFile(path)
	if err != nil {
		log.Fatalf("Failed to load NNUE: %v", err)
	}
	GlobalNNUE = *nnue
}
