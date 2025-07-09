package engine

const PV_MOVE_BONUS = 2000000
const MVV_LVA_BONUS = 1000000
const PROMOTION_BONUS = 800000
const FIRST_KILLER_MOVE_BONUS = 600000
const SECOND_KILLER_MOVE_BONUS = 590000
const HISTORY_MAX_BONUS = 16384
const BAD_CAPTURE_BONUS = -32768

var MVV_LVA_TABLE = [7][7]int{
	{15, 13, 14, 12, 11, 10, 0}, // victim P, attacker P, B, N, R, Q, K, Empty
	{35, 33, 34, 32, 31, 30, 0}, // victim B, attacker P, B, N, R, Q, K, Empty
	{25, 23, 24, 22, 21, 20, 0}, // victim N, attacker P, B, N, R, Q, K, Empty
	{45, 43, 44, 42, 41, 40, 0}, // victim R, attacker P, B, N, R, Q, K, Empty
	{55, 53, 54, 52, 51, 50, 0}, // victim Q, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim K, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim Empty, attacker P, B, N, R, Q, K, Empty
}

func (s *Searcher) ScoreMove(mv Move, pv Move, depth int) int {
	if mv == pv {
		return PV_MOVE_BONUS
	}

	if mv.movetype == CAPTURE || mv.movetype == CAPTURE_AND_PROMOTION || mv.movetype == EN_PASSANT {
		mvvLvaScore := MVV_LVA_TABLE[PieceToPieceType(mv.captured)][PieceToPieceType(mv.piece)]
		// SEE + MVV/LVA move ordering
		// Rank bad SEE captures below quiets
		if mv.movetype != CAPTURE_AND_PROMOTION {
			if s.SEE(mv) < 0 {
				// Bad SEE capture
				return BAD_CAPTURE_BONUS + mvvLvaScore
			} else {
				return MVV_LVA_BONUS + mvvLvaScore
			}
		} else {
			return MVV_LVA_BONUS + mvvLvaScore
		}
	} else if mv.movetype == PROMOTION {
		return PROMOTION_BONUS // Promotions are valuable
	} else {
		// Quiet moves
		if mv == s.KillerMoves[depth][0] {
			return FIRST_KILLER_MOVE_BONUS // First killer move
		} else if mv == s.KillerMoves[depth][1] {
			return SECOND_KILLER_MOVE_BONUS // Second killer move
		} else {
			return s.History[s.Position.turn][mv.from][mv.to]
		}
	}
}

// Selection-based move ordering. This is more efficient than sorting
// all moves by score since we will likely have cutoffs early in the mvoe list.
func (s *Searcher) SelectMove(idx int, moves []Move, pv Move, depth int) Move {
	bestScore := s.ScoreMove(moves[idx], pv, depth)
	bestIndex := idx
	for j := idx + 1; j < len(moves); j++ {
		score := s.ScoreMove(moves[j], pv, depth)
		if score > bestScore {
			bestScore = score
			bestIndex = j
		}
	}

	moves[idx], moves[bestIndex] = moves[bestIndex], moves[idx]
	return moves[idx]
}
