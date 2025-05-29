package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// TablebaseMove represents a single move in the tablebase response
type TablebaseMove struct {
	Uci     string `json:"uci"`     // Move in UCI format
	San     string `json:"san"`     // Move in SAN format
	Dtz     int    `json:"dtz"`     // Distance to zero
	Dtm     int    `json:"dtm"`     // Distance to mate
	Wdl     int    `json:"wdl"`     // Win, draw, or loss
	Zeroing bool   `json:"zeroing"` // Whether this is a zeroing move
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

// probeTablebase queries the lichess tablebase API for a position
// Returns (score, bestMove, found)
// score: tablebase score (mate score if DTM available, otherwise uses DTZ)
// bestMove: best move in UCI format
// found: whether the position was found in the tablebase
func probeTablebase(fen string) (int, string, bool) {
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

	// No moves available
	if len(result.Moves) == 0 {
		return 0, "", false
	}

	// Convert category to score
	var score int
	switch result.Category {
	case "win":
		score = winVal - 100
	case "cursed-win":
		score = 300
	case "blessed-loss":
		score = -300
	case "loss":
		score = -winVal + 100
	case "draw":
		score = 0
	default:
		score = 0
	}

	// Select best move based on category
	var bestMove TablebaseMove
	if result.Category == "win" || result.Category == "cursed-win" {
		// In winning positions, choose fastest win
		bestMove = result.Moves[0] // Moves are already sorted by DTZ
	} else if result.Category == "loss" || result.Category == "blessed-loss" {
		// In losing positions, choose slowest loss
		bestMove = result.Moves[len(result.Moves)-1]
	} else {
		// In drawn positions, prefer zeroing moves
		bestMove = result.Moves[0]
		for _, move := range result.Moves {
			if move.Zeroing {
				bestMove = move
				break
			}
		}
	}

	return score, bestMove.Uci, true
}

// isTablebasePosition checks if a position should be probed in the tablebase
func isTablebasePosition(b *Board) bool {
	// Count total pieces
	whitePieces := popCount(b.colors[WHITE])
	blackPieces := popCount(b.colors[BLACK])
	totalPieces := whitePieces + blackPieces

	// Only probe if 7 or fewer pieces total
	return totalPieces <= 7
}
