package engine

import (
	"fmt"
	"time"
)

const INF_TIME = 10000000 // Time for `go infinite`

type TimeManager struct {
	softLimit       int64 // Soft limit, checked at the end of each ID iteration
	hardLimit       int64 // Hard limit, checked during search
	searchStartTime time.Time
	wTime           int64
	bTime           int64
	wInc            int64
	bInc            int64
	movesToGo       int64
	moveTime        int64 // Will be 0 if movetime not specified
	maxDepth        int64 // Will be 0 if max depth not specified
	maxNodes        int64 // Will be 0 if max nodes not specified
	infinite        bool
	Stop            bool // If set to true, will immediately stop search
}

var Timer TimeManager

func (t *TimeManager) Calculate(c Color, wtime int64, btime int64, winc int64, binc int64, movestogo int64, depth int64, nodes int64, movetime int64, infinite bool) {
	t.wTime = wtime
	t.wInc = winc
	t.bTime = btime
	t.bInc = binc
	t.movesToGo = movestogo
	t.maxDepth = depth
	t.maxNodes = nodes
	t.moveTime = movetime
	t.infinite = infinite

	if infinite {
		t.softLimit = INF_TIME
		t.hardLimit = INF_TIME
		t.moveTime = INF_TIME
		return
	}

	if movetime > 0 {
		t.softLimit = movetime
		t.hardLimit = movetime
		return
	}

	if depth > 0 {
		// Searching up to a max depth, so set time bounds to infinite and nodes to -1
		t.softLimit = INF_TIME
		t.hardLimit = INF_TIME
		return
	}

	if nodes > 0 {
		// Searching up to a max number of nodes, so set time bounds to infinite and depth to -1
		t.softLimit = INF_TIME
		t.hardLimit = INF_TIME
		return
	}

	remainingTime := ternary(c == WHITE, t.wTime, t.bTime)
	increment := ternary(c == WHITE, t.wInc, t.bInc)

	// General time management formula
	if t.movesToGo > 0 {
		t.softLimit = remainingTime/t.movesToGo + increment/Params.INC_FRACTION
	} else {
		t.softLimit = remainingTime/Params.TIME_DIVISOR + increment/Params.INC_FRACTION
	}

	t.hardLimit = t.softLimit * Params.HARD_LIMIT_MULT

	// Cap hard and soft limits to the remainingTime - 200 so we don't flag
	if t.hardLimit > remainingTime-200 {
		t.hardLimit = remainingTime - 200
	}

	if t.softLimit > remainingTime-200 {
		t.softLimit = remainingTime - 200
	}
}

// Primarily for tests
func (t *TimeManager) SetMoveTime(movetime int64) {
	t.hardLimit = movetime
	t.softLimit = movetime
	t.maxNodes = 0
	t.maxDepth = 0
}

func (t *TimeManager) StartSearch() {
	t.searchStartTime = time.Now()
	t.Stop = false
}

func (t *TimeManager) Delta() int64 {
	return time.Since(t.searchStartTime).Milliseconds()
}

// Only called during PVS, checks hard limit timeout and nodes
func (t *TimeManager) CheckPVS(info *SearchInfo) {
	if t.Delta() > t.hardLimit {
		t.Stop = true
	}

	if t.maxNodes > 0 && info.NodesSearched > int(t.maxNodes) {
		t.Stop = true
	}
}

// Called during iterative deepening, checks soft limit, nodes, current depth
// TODO: add soft time bound scaling based on ID statistics
func (t *TimeManager) CheckID(info *SearchInfo, depth int) {
	if t.Delta() > t.softLimit {
		t.Stop = true
	}

	if t.maxNodes > 0 && info.NodesSearched > int(t.maxNodes) {
		t.Stop = true
	}

	if t.maxDepth > 0 && depth >= int(t.maxDepth) {
		t.Stop = true
	}
}

func (t *TimeManager) PrintConditions() {
	fmt.Println("searching with conditions: ")
	fmt.Printf("\thard limit: %d\n", t.hardLimit)
	fmt.Printf("\tsoft limit: %d\n", t.softLimit)

	if t.maxNodes > 0 {
		fmt.Printf("\tmax nodes: %d\n", t.maxNodes)
	}

	if t.maxDepth > 0 {
		fmt.Printf("\tmax depth: %d\n", t.maxDepth)
	}
}
