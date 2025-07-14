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

//go:embed nn_weights/512.bin
var embeddedWeights []byte

// NETWORK ARCHITECTURE:
// (768 -> 512)x2 -> 1 (Perspective Network)
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
const HIDDEN_LAYER_SIZE = 512
const OUTPUT_LAYER_SIZE = 1

// Quantization constants
const QA int16 = 255
const QB int16 = 64
const SCALE int32 = 400

type Accumulator struct {
	values [HIDDEN_LAYER_SIZE]int16
}

type AccumulatorPair struct {
	white        Accumulator
	black        Accumulator
	dirty        bool
	updateBuffer AccumulatorUpdate
}

type AccumulatorUpdate struct {
	from       Square
	to         Square
	capt       Square
	from2      Square
	to2        Square
	fromType   PieceType
	toType     PieceType
	fromType2  PieceType
	toType2    PieceType
	captType   PieceType
	color      Color
	captColor  Color
	updateType uint8 // 0 for AddSub, 1 for AddSubSub, 2 for AddAddSubSub
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

		for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
			pair.white.values[i] += nnue.accumulator_weights[whiteIndex][i]
			pair.black.values[i] += nnue.accumulator_weights[blackIndex][i]
		}
	}

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		pair.white.values[i] += nnue.accumulator_biases[i]
		pair.black.values[i] += nnue.accumulator_biases[i]
	}

	return pair
}

func (nnue *NNUE) AddSubFeature(to_accs *AccumulatorPair, prev_accs *AccumulatorPair, from Square, to Square, fromType PieceType, toType PieceType, color Color) {
	whiteFromIndex := CalculateIndex(WHITE, from, fromType, color)
	whiteToIndex := CalculateIndex(WHITE, to, toType, color)

	blackFromIndex := CalculateIndex(BLACK, from, fromType, color)
	blackToIndex := CalculateIndex(BLACK, to, toType, color)

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		to_accs.white.values[i] = prev_accs.white.values[i] + nnue.accumulator_weights[whiteToIndex][i] - nnue.accumulator_weights[whiteFromIndex][i]
		to_accs.black.values[i] = prev_accs.black.values[i] + nnue.accumulator_weights[blackToIndex][i] - nnue.accumulator_weights[blackFromIndex][i]
	}
}

func (nnue *NNUE) AddSubSubFeature(to_accs *AccumulatorPair, prev_accs *AccumulatorPair, from Square, to Square, capt Square, fromType PieceType, toType PieceType, captType PieceType, color Color, captColor Color) {
	whiteFromIndex := CalculateIndex(WHITE, from, fromType, color)
	whiteToIndex := CalculateIndex(WHITE, to, toType, color)

	blackFromIndex := CalculateIndex(BLACK, from, fromType, color)
	blackToIndex := CalculateIndex(BLACK, to, toType, color)

	whiteCaptIndex := CalculateIndex(WHITE, capt, captType, captColor)
	blackCaptIndex := CalculateIndex(BLACK, capt, captType, captColor)

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		to_accs.white.values[i] = prev_accs.white.values[i] + nnue.accumulator_weights[whiteToIndex][i] - nnue.accumulator_weights[whiteFromIndex][i] - nnue.accumulator_weights[whiteCaptIndex][i]
		to_accs.black.values[i] = prev_accs.black.values[i] + nnue.accumulator_weights[blackToIndex][i] - nnue.accumulator_weights[blackFromIndex][i] - nnue.accumulator_weights[blackCaptIndex][i]
	}
}

func (nnue *NNUE) AddAddSubSubFeature(to_accs *AccumulatorPair, prev_accs *AccumulatorPair, from1 Square, to1 Square, fromType1 PieceType,
	toType1 PieceType, from2 Square, to2 Square, fromType2 PieceType, toType2 PieceType, color Color) {
	wf1 := CalculateIndex(WHITE, from1, fromType1, color)
	wt1 := CalculateIndex(WHITE, to1, toType1, color)

	bf1 := CalculateIndex(BLACK, from1, fromType1, color)
	bt1 := CalculateIndex(BLACK, to1, toType1, color)

	wf2 := CalculateIndex(WHITE, from2, fromType2, color)
	wt2 := CalculateIndex(WHITE, to2, toType2, color)

	bf2 := CalculateIndex(BLACK, from2, fromType2, color)
	bt2 := CalculateIndex(BLACK, to2, toType2, color)

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		to_accs.white.values[i] = prev_accs.white.values[i] + nnue.accumulator_weights[wt1][i] + nnue.accumulator_weights[wt2][i] - nnue.accumulator_weights[wf1][i] - nnue.accumulator_weights[wf2][i]
		to_accs.black.values[i] = prev_accs.black.values[i] + nnue.accumulator_weights[bt1][i] + nnue.accumulator_weights[bt2][i] - nnue.accumulator_weights[bf1][i] - nnue.accumulator_weights[bf2][i]
	}
}

func StoreAccUpdatesOnMove(b *Board, move Move, stm Color) {
	b.accumulatorStack[b.accumulatorIdx] = b.accumulatorStack[b.accumulatorIdx-1]
	b.accumulatorStack[b.accumulatorIdx].dirty = true

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

		b.accumulatorStack[b.accumulatorIdx].updateBuffer = AccumulatorUpdate{
			from: from, to: to, capt: captureSquare, fromType: pieceType,
			toType: newPieceType, captType: capturedType, color: color, captColor: capturedColor,
			updateType: 1,
		}
	} else {
		b.accumulatorStack[b.accumulatorIdx].updateBuffer = AccumulatorUpdate{
			from: from, to: to, fromType: pieceType,
			toType: newPieceType, color: color,
			updateType: 0,
		}
	}

	// Handle castling
	if move.movetype == K_CASTLE {
		if color == WHITE {
			// Subtract White rook from H1
			// Add White rook to F1
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.from2 = H1
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.to2 = F1
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.fromType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.toType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.updateType = 2
		} else {
			// Subtract Black rook from H8
			// Add Black rook to F8
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.from2 = H8
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.to2 = F8
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.fromType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.toType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.updateType = 2
		}
	} else if move.movetype == Q_CASTLE {
		if color == WHITE {
			// Subtract White rook from A1
			// Add White rook to D1
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.from2 = A1
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.to2 = D1
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.fromType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.toType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.updateType = 2
		} else {
			// Subtract Black rook from A8
			// Add Black rook to D8
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.from2 = A8
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.to2 = D8
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.fromType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.toType2 = ROOK
			b.accumulatorStack[b.accumulatorIdx].updateBuffer.updateType = 2
		}
	}
}

func (n *NNUE) ApplyLazyUpdates(b *Board) {
	currIndex := b.accumulatorIdx
	for currIndex >= 0 && b.accumulatorStack[currIndex].dirty {
		currIndex--
	}

	if currIndex < 0 {
		return
	}

	for currIndex != b.accumulatorIdx {
		update := b.accumulatorStack[currIndex+1].updateBuffer

		if update.updateType == 0 {
			n.AddSubFeature(&b.accumulatorStack[currIndex+1], &b.accumulatorStack[currIndex], update.from, update.to, update.fromType, update.toType, update.color)
		} else if update.updateType == 1 {
			n.AddSubSubFeature(&b.accumulatorStack[currIndex+1], &b.accumulatorStack[currIndex], update.from, update.to, update.capt, update.fromType, update.toType,
				update.captType, update.color, update.captColor)
		} else {
			n.AddAddSubSubFeature(&b.accumulatorStack[currIndex+1], &b.accumulatorStack[currIndex], update.from, update.to, update.fromType, update.toType, update.from2,
				update.to2, update.fromType2, update.toType2, update.color)
		}

		currIndex++
		b.accumulatorStack[currIndex].dirty = false
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
