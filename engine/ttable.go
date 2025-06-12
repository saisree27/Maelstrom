package engine

import (
	"math/bits"
	"unsafe"
)

type bound uint8

const (
	upper bound = iota
	lower
	exact
)

type TTEntry struct {
	hash     u64
	score    int32
	bestMove Move
	depth    uint8
	bd       bound
	age      uint8
	_        [1]byte
}

type TTable struct {
	entries []TTEntry
	count   u64
	age     uint8 // Current age counter
}

var table TTable

func initializeTTable(megabytes int) {
	// Total bytes available
	totalBytes := uint64(megabytes) * 1024 * 1024

	// Max number of entries that fit
	entrySize := uint64(unsafe.Sizeof(TTEntry{}))
	numEntries := totalBytes / entrySize

	// Round down to nearest power of two for better cache utilization
	if numEntries == 0 {
		numEntries = 1 // avoid zero-sized allocation
	}
	table.count = 1 << (bits.Len64(numEntries) - 1)

	// Allocate entries
	table.entries = make([]TTEntry, table.count)
	table.age = 0
}

func clearTTable() {
	for i := range table.entries {
		table.entries[i] = TTEntry{}
	}
	table.age = 0
}

func storeEntry(b *Board, score int, bd bound, mv Move, depth uint8) {
	entryIndex := b.zobrist % table.count
	entry := &table.entries[entryIndex]

	// Replace if:
	// 1. We have a deeper search
	// 2. Newer entry
	if entry.depth <= depth || entry.depth != table.age {
		*entry = TTEntry{
			bestMove: mv,
			hash:     b.zobrist,
			bd:       bd,
			score:    int32(score),
			depth:    depth,
			age:      table.age,
		}
	}
}

func probeTT(b *Board, alpha int, beta int, depth uint8, m *Move) (bool, int) {
	entryIndex := b.zobrist % table.count
	entry := &table.entries[entryIndex]

	if entry.hash == b.zobrist {
		// Update age on access
		entry.age = table.age

		// Get the PV-move
		*m = entry.bestMove
		if entry.depth >= depth {
			score := int(entry.score)
			if entry.bd == lower && score >= beta {
				return true, beta
			}

			if entry.bd == upper && score <= alpha {
				return true, alpha
			}

			if entry.bd == exact {
				return true, score
			}
		}
	}
	return false, 0
}

// Increment age counter periodically
func incrementAge() {
	table.age++
}
