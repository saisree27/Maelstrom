package engine

import (
	"unsafe"
)

type bound int

const (
	upper bound = iota
	lower
	exact
)

type TTEntry struct {
	hash     u64
	bestMove Move
	score    int
	depth    int
	bd       bound
}

type TTable struct {
	entries []TTEntry
	count   u64
}

var table TTable

func initializeTTable(megabytes int) {
	table.count = u64(megabytes) * 1024 * 1024 / u64(unsafe.Sizeof(TTEntry{}))
	table.entries = make([]TTEntry, table.count)
}

func clearTTable() {
	table.entries = make([]TTEntry, table.count)
}

func storeEntry(b *Board, score int, bd bound, mv Move, depth int) {
	entryIndex := b.zobrist % table.count
	if !(mv.from == a1 && mv.to == a1) {
		table.entries[entryIndex] = TTEntry{
			bestMove: mv,
			hash:     b.zobrist,
			bd:       bd,
			score:    score,
			depth:    depth,
		}
	}
}

func probeTT(b *Board, score *int, alpha *int, beta *int, depth int, rd int, m *Move) (bool, int) {
	entryIndex := b.zobrist % table.count
	entry := table.entries[entryIndex]
	if entry.hash == b.zobrist {
		// Get the PV-move
		*m = entry.bestMove
		if entry.depth > depth {
			*score = entry.score
			switch entry.bd {
			case upper:
				if *score < *beta && depth != rd {
					*beta = *score
				}
			case lower:
				if *score > *alpha && depth != rd {
					*alpha = *score
				}
			case exact:
				return true, *score
			}
			if *alpha >= *beta {
				return true, *score
			}
		}
	}
	return false, 0
}
