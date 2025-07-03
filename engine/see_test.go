package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestSee(t *testing.T) {
	file, err := os.Open("test_data/see_test.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")

		fen := strings.TrimSpace(parts[0])
		uci := strings.TrimSpace(parts[1])
		scoreStr := strings.TrimSpace(parts[2])
		score, _ := strconv.Atoi(scoreStr)

		fmt.Println("FEN:", fen)
		fmt.Println("Move:", uci)
		fmt.Println("Score:", score)
		fmt.Println()

		b := Board{}
		b.InitFEN(fen + " 0 1")
		move := FromUCI(uci, &b)
		if move.movetype == QUIET || move.movetype == PROMOTION || move.movetype == K_CASTLE || move.movetype == Q_CASTLE || move.movetype == CAPTURE_AND_PROMOTION {
			fmt.Println("skipping since move isn't a capture")
			continue
		}

		b.PrintFromBitBoards()
		res := see(&b, move)

		fmt.Printf("\n\n")
		if res != score {
			t.Fatalf("Expected %d, got %d", score, res)
		}
	}
}
