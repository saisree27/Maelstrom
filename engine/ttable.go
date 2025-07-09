package engine

import (
	"math/bits"
	"unsafe"
)

type bound uint8

const (
	UPPER bound = iota
	LOWER
	EXACT
)

type TTEntry struct {
	hash       u64
	score      int32
	bestMove   Move
	depth      uint8
	bd         bound
	age        uint8
	staticEval int32
}

type TranspositionTable struct {
	entries []TTEntry
	count   u64
	age     uint8 // Current age counter
}

var TT TranspositionTable

func InitializeTT(megabytes int) {
	// Total bytes available
	totalBytes := uint64(megabytes) * 1024 * 1024

	// Max number of entries that fit
	entrySize := uint64(unsafe.Sizeof(TTEntry{}))
	numEntries := totalBytes / entrySize

	// Round down to nearest power of two for better cache utilization
	if numEntries == 0 {
		numEntries = 1 // avoid zero-sized allocation
	}
	TT.count = 1 << (bits.Len64(numEntries) - 1)

	// Allocate entries
	TT.entries = make([]TTEntry, TT.count)
	TT.age = 0
}

func ClearTT() {
	for i := range TT.entries {
		TT.entries[i] = TTEntry{}
	}
	TT.age = 0
}

func StoreEntry(b *Board, score int, bd bound, mv Move, depth uint8, staticEval int) {
	entryIndex := b.zobrist % TT.count
	entry := &TT.entries[entryIndex]

	// Replace if:
	// 1. We have a deeper search
	// 2. Newer entry
	if entry.depth <= depth || entry.depth != TT.age {
		*entry = TTEntry{
			bestMove:   mv,
			hash:       b.zobrist,
			bd:         bd,
			score:      int32(score),
			depth:      depth,
			age:        TT.age,
			staticEval: int32(staticEval),
		}
	}
}

func ProbeTT(b *Board, alpha int, beta int, depth uint8, m *Move, staticEval *int) (bool, int) {
	entryIndex := b.zobrist % TT.count
	entry := &TT.entries[entryIndex]

	if entry.hash == b.zobrist {
		// Update age on access
		entry.age = TT.age

		// Get the PV-move
		*m = entry.bestMove
		*staticEval = int(entry.staticEval)
		if entry.depth >= depth {
			score := int(entry.score)
			if entry.bd == LOWER && score >= beta {
				return true, beta
			}

			if entry.bd == UPPER && score <= alpha {
				return true, alpha
			}

			if entry.bd == EXACT {
				return true, score
			}
		}
	}
	return false, 0
}

// Increment age counter periodically
func IncrementTTAge() {
	TT.age++
}
