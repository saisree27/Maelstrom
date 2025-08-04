package engine

import (
	"fmt"
	"time"
)

const INF_TIME = 10000000 // Time for `go infinite`
var MOVE_STABILITY_FACTOR = [5]float32{
	2.5, 1.2, 0.9, 0.8, 0.75,
}

var SCORE_STABILITY_FACTOR = [5]float32{
	1.25, 1.15, 1.0, 0.94, 0.88,
}

type TimeManager struct {
	softLimit       int64   // Soft limit, checked at the end of each ID iteration
	hardLimit       int64   // Hard limit, checked during search
	softScale       float32 // Scale of soft limit
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
	moveStability   int  // Keeps track of how often the best move stays the same
	scoreStability  int  // Keeps track of how often the best move stays the same
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
	t.softScale = 1

	if infinite {
		t.softLimit = INF_TIME
		t.hardLimit = INF_TIME
		t.moveTime = INF_TIME
		return
	}

	if movetime > 0 {
		t.softLimit = INF_TIME
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
	t.softScale = 1
	t.moveStability = 0
	t.scoreStability = 0
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
	if t.Delta() > int64(float32(t.softLimit)*t.softScale) {
		t.Stop = true
	}

	if t.maxNodes > 0 && info.NodesSearched > int(t.maxNodes) {
		t.Stop = true
	}

	if t.maxDepth > 0 && depth >= int(t.maxDepth) {
		t.Stop = true
	}
}

func (t *TimeManager) UpdateSoftLimit(info *SearchInfo, bestmove Move, prevBest Move, score int, prevScore int) {
	///////////////////////////////////////////////////////////////////////////////
	// SCORE STABILITY
	// Keeping track of the best score over each iteration of IID, we can judge
	// how complex a position is. The more the score changes after each iteration,
	// it is likely to be more tactical, so we should give the search more time.
	// We keep track of a running count of how stable the score is in a window and
	// then use it to index into a decreasing array of time multipliers.
	///////////////////////////////////////////////////////////////////////////////
	if score <= prevScore+Params.TM_STABILITY_WINDOW && score >= prevScore-Params.TM_STABILITY_WINDOW {
		t.scoreStability++
		t.scoreStability = Min(t.scoreStability, 4)
	} else {
		t.scoreStability = 0
	}

	///////////////////////////////////////////////////////////////////////////////
	// BEST MOVE STABILITY
	// Keeping track of the best move over each iteration of IID, we can judge
	// how complex a position is. The more the move changes after each iteration,
	// it is likely to be more tactical, so we should give the search more time.
	// We keep track of a running count of how often we have the same best move
	// then use it to index into a decreasing array of time multipliers.
	///////////////////////////////////////////////////////////////////////////////
	if prevBest == bestmove {
		t.moveStability++
		t.moveStability = Min(t.moveStability, 4)
	} else {
		t.moveStability = 0
	}

	///////////////////////////////////////////////////////////////////////////////
	// NODE COUNT RATIO
	// We can keep track of the ratio of how many nodes are spent searching the
	// best move compared to the overall nodes in the iterations. The higher the
	// ratio, the more stable the best move should be, so we can spend less time on
	// the move. We calculate the ratio and subtract it from a constant to
	// determine the time multiplier.
	///////////////////////////////////////////////////////////////////////////////
	totalNodesInIteration := 0
	for _, value := range info.NodesPerMove {
		totalNodesInIteration += value
	}

	scoreStabilityFactor := SCORE_STABILITY_FACTOR[t.scoreStability]
	moveStabilityFactor := MOVE_STABILITY_FACTOR[t.moveStability]
	nodeRatioFactor := float32(Params.TM_NODE_COUNT_CONSTANT)/10 - float32(info.NodesPerMove[bestmove])/float32(totalNodesInIteration)
	t.softScale = moveStabilityFactor * scoreStabilityFactor * nodeRatioFactor
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
