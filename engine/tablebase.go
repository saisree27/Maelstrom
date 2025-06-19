package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TablebaseMove represents a single move in the tablebase response
type TablebaseMove struct {
	Uci      string `json:"uci"`      // Move in UCI format
	San      string `json:"san"`      // Move in SAN format
	Dtz      int    `json:"dtz"`      // Distance to zero
	Dtm      int    `json:"dtm"`      // Distance to mate
	Wdl      int    `json:"wdl"`      // Win, draw, or loss
	Zeroing  bool   `json:"zeroing"`  // Whether this is a zeroing move
	Category string `json:"category"` // "win", "loss", "draw", "unknown", "cursed-win", "blessed-loss"
}

// TablebaseResult represents the response from the lichess tablebase API
type TablebaseResult struct {
	Category    string          `json:"category"`     // "win", "loss", "draw", "unknown", "cursed-win", "blessed-loss"
	Dtz         int             `json:"dtz"`          // Distance to zero (50-move rule) in plies
	Dtm         int             `json:"dtm"`          // Distance to mate in plies (not always available)
	Wdl         int             `json:"wdl"`          // Win, draw, or loss
	Checkmate   bool            `json:"checkmate"`    // True if position is checkmate
	Stalemate   bool            `json:"stalemate"`    // True if position is stalemate
	VariantWin  bool            `json:"variant_win"`  // True if position is won due to variant rules
	VariantLoss bool            `json:"variant_loss"` // True if position is lost due to variant rules
	Moves       []TablebaseMove `json:"moves"`        // Available moves and their evaluations
}

// ProbeTablebase queries the lichess tablebase API for a position
// Returns (score, bestMove, found)
// score: tablebase score (mate score if DTM available, otherwise uses DTZ)
// bestMove: best move in UCI format (or comma-separated list of drawing moves if position is a draw)
// found: whether the position was found in the tablebase
func ProbeTablebase(fen string) (int, string, bool) {
	// Encode FEN for URL
	encodedFEN := url.QueryEscape(fen)
	url := fmt.Sprintf("https://tablebase.lichess.ovh/standard?fen=%s", encodedFEN)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		return 0, "", false
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Received status code %d\n", resp.StatusCode)
		return 0, "", false
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return 0, "", false
	}

	// Parse JSON response
	var result TablebaseResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return 0, "", false
	}

	fmt.Printf("Tablebase result: %+v\n", result)

	// No moves available
	if len(result.Moves) == 0 {
		return 0, "", false
	}

	// Convert category to score
	var score int
	switch result.Category {
	case "win":
		score = WIN_VAL - 100
	case "cursed-win":
		score = 300
	case "blessed-loss":
		score = -300
	case "loss":
		score = -WIN_VAL + 100
	case "draw":
		score = 0
	default:
		score = 0
	}

	if result.Category == "draw" {
		// Filter moves to only include those matching the position's category
		var filteredMoves []TablebaseMove
		for _, move := range result.Moves {
			if move.Category == result.Category {
				filteredMoves = append(filteredMoves, move)
			}
		}

		// Update result.Moves to filtered list
		result.Moves = filteredMoves
		var drawingMoves []string
		for _, move := range filteredMoves {
			drawingMoves = append(drawingMoves, move.Uci)
		}
		return 0, fmt.Sprintf("draw:%s", strings.Join(drawingMoves, ",")), true
	}

	if len(result.Moves) == 0 {
		return 0, "", false
	}

	var bestMove TablebaseMove = result.Moves[0]

	return score, bestMove.Uci, true
}

// IsTablebasePosition checks if a position should be probed in the tablebase
func IsTablebasePosition(b *Board) bool {
	// Count total pieces
	whitePieces := PopCount(b.colors[WHITE])
	blackPieces := PopCount(b.colors[BLACK])
	totalPieces := whitePieces + blackPieces

	// Only probe if 7 or fewer pieces total
	return totalPieces <= 7
}
