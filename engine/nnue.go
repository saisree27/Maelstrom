package engine

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"log"
	"maelstrom/engine/screlu"
	"math/rand"
	"time"
)

//go:embed nn_weights/network.bin
var embeddedWeights []byte

// NETWORK ARCHITECTURE:
// (768 -> 256)x2 -> 1 (Perspective Network)
// This means there are two sets of inputs, x and x^ where x is from STM perspective, x^ is from NTM perspective
// Input size: 768 flattened = 12 * 64. Indices 0 through 11 will represent friendly_pawn, friendly_bishop, ..., enemy_king
// We will have two accumulators:
// 	a  = Hx  + b
//  a^ = Hx^ + b
// where H is the hidden layer weights and b is the hidden layer biases
// The inference will then be:
//  y = O(concat(a + a^)) + c
// where O is the output layer weights and c is the output layer biases

const INPUT_LAYER_SIZE = 768
const HIDDEN_LAYER_SIZE = 256
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

func (nnue *NNUE) RecomputeAccumulators(b *Board) AccumulatorPair {
	pair := AccumulatorPair{
		white: Accumulator{},
		black: Accumulator{},
	}

	for sq := A1; sq <= H8; sq++ {
		piece := b.squares[sq]
		if piece == EMPTY {
			continue
		}

		pieceType := PieceToPieceType(piece)
		pieceColor := piece.GetColor()

		whiteIndex := CalculateIndex(WHITE, sq, pieceType, pieceColor)
		blackIndex := CalculateIndex(BLACK, sq, pieceType, pieceColor)

		AccumulatorAdd(nnue, &pair, whiteIndex, blackIndex)
	}

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		pair.white.values[i] += nnue.accumulator_biases[i]
		pair.black.values[i] += nnue.accumulator_biases[i]
	}

	return pair
}

func AccumulatorAdd(network *NNUE, accs *AccumulatorPair, indexWhite int, indexBlack int) {
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		accs.white.values[i] += network.accumulator_weights[indexWhite][i]
		accs.black.values[i] += network.accumulator_weights[indexBlack][i]
	}
}

func AccumulatorSub(network *NNUE, accs *AccumulatorPair, indexWhite int, indexBlack int) {
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		accs.white.values[i] -= network.accumulator_weights[indexWhite][i]
		accs.black.values[i] -= network.accumulator_weights[indexBlack][i]
	}
}

func AccumulatorAddSub(network *NNUE, accs *AccumulatorPair, indexWhiteFrom int, indexBlackFrom int, indexWhiteTo int, indexBlackTo int) {
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		accs.white.values[i] += network.accumulator_weights[indexWhiteTo][i] - network.accumulator_weights[indexWhiteFrom][i]
		accs.black.values[i] += network.accumulator_weights[indexBlackTo][i] - network.accumulator_weights[indexBlackFrom][i]
	}
}

func AccumulatorAddSubSub(network *NNUE, accs *AccumulatorPair, indexWhiteFrom int, indexBlackFrom int, indexWhiteTo int, indexBlackTo int, indexWhiteCapt int, indexBlackCapt int) {
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		accs.white.values[i] += network.accumulator_weights[indexWhiteTo][i] - network.accumulator_weights[indexWhiteFrom][i] - network.accumulator_weights[indexWhiteCapt][i]
		accs.black.values[i] += network.accumulator_weights[indexBlackTo][i] - network.accumulator_weights[indexBlackFrom][i] - network.accumulator_weights[indexBlackCapt][i]
	}
}

func (nnue *NNUE) AddFeature(accs *AccumulatorPair, sq Square, pieceType PieceType, color Color) {
	whiteIndex := CalculateIndex(WHITE, sq, pieceType, color)
	blackIndex := CalculateIndex(BLACK, sq, pieceType, color)

	AccumulatorAdd(nnue, accs, whiteIndex, blackIndex)
}

func (nnue *NNUE) SubFeature(accs *AccumulatorPair, sq Square, pieceType PieceType, color Color) {
	whiteIndex := CalculateIndex(WHITE, sq, pieceType, color)
	blackIndex := CalculateIndex(BLACK, sq, pieceType, color)

	AccumulatorSub(nnue, accs, whiteIndex, blackIndex)
}

func (nnue *NNUE) AddSubFeature(accs *AccumulatorPair, from Square, to Square, fromType PieceType, toType PieceType, color Color) {
	whiteFromIndex := CalculateIndex(WHITE, from, fromType, color)
	whiteToIndex := CalculateIndex(WHITE, to, toType, color)

	blackFromIndex := CalculateIndex(BLACK, from, fromType, color)
	blackToIndex := CalculateIndex(BLACK, to, toType, color)

	AccumulatorAddSub(nnue, accs, whiteFromIndex, blackFromIndex, whiteToIndex, blackToIndex)
}

func (nnue *NNUE) AddSubSubFeature(accs *AccumulatorPair, from Square, to Square, capt Square, fromType PieceType, toType PieceType, captType PieceType, color Color, captColor Color) {
	whiteFromIndex := CalculateIndex(WHITE, from, fromType, color)
	whiteToIndex := CalculateIndex(WHITE, to, toType, color)

	blackFromIndex := CalculateIndex(BLACK, from, fromType, color)
	blackToIndex := CalculateIndex(BLACK, to, toType, color)

	whiteCaptIndex := CalculateIndex(WHITE, capt, captType, captColor)
	blackCaptIndex := CalculateIndex(BLACK, capt, captType, captColor)

	AccumulatorAddSubSub(nnue, accs, whiteFromIndex, blackFromIndex, whiteToIndex, blackToIndex, whiteCaptIndex, blackCaptIndex)
}

func (network *NNUE) UpdateAccumulatorOnMove(b *Board, move Move, stm Color) {
	from := move.from
	to := move.to

	movingPiece := move.piece
	pieceType := PieceToPieceType(movingPiece)
	color := movingPiece.GetColor()

	// If promotion, change the new piece type to the promoting piece
	newPieceType := pieceType
	if move.movetype == PROMOTION || move.movetype == CAPTURE_AND_PROMOTION {
		newPieceType = PieceToPieceType(move.promote)
	}

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

		network.AddSubSubFeature(&b.accumulators, from, to, captureSquare, pieceType, newPieceType, capturedType, color, capturedColor)
	} else {
		network.AddSubFeature(&b.accumulators, from, to, pieceType, newPieceType, color)
	}

	// Handle castling
	if move.movetype == K_CASTLE {
		if color == WHITE {
			// Subtract White rook from H1
			// Add White rook to F1
			network.AddSubFeature(&b.accumulators, H1, F1, ROOK, ROOK, WHITE)
		} else {
			// Subtract Black rook from H8
			// Add Black rook to F8
			network.AddSubFeature(&b.accumulators, H8, F8, ROOK, ROOK, BLACK)
		}
	} else if move.movetype == Q_CASTLE {
		if color == WHITE {
			// Subtract White rook from A1
			// Add White rook to D1
			network.AddSubFeature(&b.accumulators, A1, D1, ROOK, ROOK, WHITE)
		} else {
			// Subtract Black rook from A8
			// Add Black rook to D8
			network.AddSubFeature(&b.accumulators, A8, D8, ROOK, ROOK, BLACK)
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

func LoadNNUEFromBytes(data []byte) (*NNUE, error) {
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

func InitializeNNUE() {
	fmt.Println("loading NNUE weights from embedded data")
	nnue, err := LoadNNUEFromBytes(embeddedWeights)
	if err != nil {
		log.Fatalf("Failed to load NNUE: %v", err)
	}
	GlobalNNUE = *nnue
}
