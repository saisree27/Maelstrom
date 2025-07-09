package engine

import (
	"fmt"
	"os"
	"runtime/pprof"
)

// SearchWithProfiling runs searchWithTime with CPU and memory profiling enabled.
// It creates two profile files:
// - cpu.prof: CPU profiling data
// - mem.prof: Memory profiling data
func SearchWithProfiling(b *Board, movetime int64) Move {
	// Start CPU profiling
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Printf("Could not create CPU profile: %v\n", err)
		return Move{}
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		fmt.Printf("Could not start CPU profile: %v\n", err)
		return Move{}
	}
	defer pprof.StopCPUProfile()

	// Run the search
	Timer.SetMoveTime(movetime)

	s := Searcher{}
	s.Position = b
	move := s.SearchPosition()

	// Write memory profile
	memFile, err := os.Create("mem.prof")
	if err != nil {
		fmt.Printf("Could not create memory profile: %v\n", err)
		return move
	}
	defer memFile.Close()

	if err := pprof.WriteHeapProfile(memFile); err != nil {
		fmt.Printf("Could not write memory profile: %v\n", err)
	}

	return move
}
