package engine

const HISTORY_MAX_BONUS = 16384

var MVV_LVA_TABLE = [7][7]int{
	{15, 13, 14, 12, 11, 10, 0}, // victim P, attacker P, B, N, R, Q, K, Empty
	{25, 23, 24, 22, 21, 20, 0}, // victim N, attacker P, B, N, R, Q, K, Empty
	{35, 33, 34, 32, 31, 30, 0}, // victim B, attacker P, B, N, R, Q, K, Empty
	{45, 43, 44, 42, 41, 40, 0}, // victim R, attacker P, B, N, R, Q, K, Empty
	{55, 53, 54, 52, 51, 50, 0}, // victim Q, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim K, attacker P, B, N, R, Q, K, Empty
	{0, 0, 0, 0, 0, 0, 0},       // victim Empty, attacker P, B, N, R, Q, K, Empty
}

type Stage uint8

const (
	TT_MOVE Stage = iota
	GEN_CAPTURES
	GOOD_CAPTURES
	PROMOTIONS
	KILLER1
	KILLER2
	COUNTER
	GEN_QUIETS
	HISTORY_QUIETS
	BAD_CAPTURES
)

type ScoredMove struct {
	move  Move
	score int
}

type MovePicker struct {
	stage       Stage
	promotions  []ScoredMove
	captures    []ScoredMove
	quiets      []ScoredMove
	badCaptures []ScoredMove
	ttMove      Move
	killer1     Move
	killer2     Move
	counter     Move
	board       *Board
	history     *[2][64][64]int
	depth       int
	currIdx     int
	lastStage   Stage
	QS          bool
	skipQuiets  bool
}

func NewMovePicker(b *Board, ttMove Move, killer1 Move, killer2 Move, counter Move, depth int, history *[2][64][64]int, fromQS bool) *MovePicker {
	mp := &MovePicker{
		board:      b,
		ttMove:     ttMove,
		stage:      ternary(ttMove.IsEmpty(), TT_MOVE+1, TT_MOVE),
		killer1:    killer1,
		killer2:    killer2,
		counter:    counter,
		depth:      depth,
		history:    history,
		currIdx:    0,
		lastStage:  ternary(fromQS, GOOD_CAPTURES, BAD_CAPTURES),
		QS:         fromQS,
		skipQuiets: fromQS,
	}
	return mp
}

func (mp *MovePicker) processMoves(moves []Move) {
	for _, mv := range moves {
		if mv.movetype == CAPTURE || mv.movetype == EN_PASSANT || mv.movetype == CAPTURE_AND_PROMOTION {
			score := MVV_LVA_TABLE[PieceToPieceType(mv.captured)][PieceToPieceType(mv.piece)]
			if mv.movetype == CAPTURE_AND_PROMOTION {
				score += 1000
			}
			mp.captures = append(mp.captures, ScoredMove{
				move:  mv,
				score: score,
			})
		} else if mv.movetype == PROMOTION {
			mp.promotions = append(mp.promotions, ScoredMove{
				move:  mv,
				score: int(mv.promote),
			})
		} else {
			score := mp.history[mp.board.turn][mv.from][mv.to]
			mp.quiets = append(mp.quiets, ScoredMove{
				move:  mv,
				score: score,
			})
		}
	}
}

func (mp *MovePicker) getNextAndSwap(moves []ScoredMove, idx int) bool {
	if idx >= len(moves) {
		return false
	}

	bestScore := moves[idx].score
	bestIdx := idx

	for i := idx + 1; i < len(moves); i++ {
		if moves[i].score > bestScore {
			bestScore = moves[i].score
			bestIdx = i
		}
	}

	moves[bestIdx], moves[idx] = moves[idx], moves[bestIdx]
	return true
}

func (mp *MovePicker) SkipQuiets() {
	mp.skipQuiets = true
}

func (mp *MovePicker) NextMove() Move {
	for mp.stage <= mp.lastStage {
		switch mp.stage {
		case TT_MOVE:
			mp.stage++
			if mp.board.IsLegal(mp.ttMove) {
				return mp.ttMove
			}
		case GEN_CAPTURES:
			mp.stage++
			if mp.QS {
				moves := mp.board.GenerateCaptures()
				mp.processMoves(moves)
			} else {
				moves := mp.board.GenerateNoisies()
				mp.processMoves(moves)
			}

		case GOOD_CAPTURES:
			if mp.getNextAndSwap(mp.captures, mp.currIdx) {
				move := mp.captures[mp.currIdx].move
				if SEE(move, mp.board) < 0 {
					mp.badCaptures = append(mp.badCaptures, mp.captures[mp.currIdx])
					mp.currIdx++
					continue
				}
				if move == mp.ttMove {
					mp.currIdx++
					continue
				}
				mp.currIdx++
				return move
			}
			mp.currIdx = 0
			mp.stage++

		case PROMOTIONS:
			if mp.getNextAndSwap(mp.promotions, mp.currIdx) {
				move := mp.promotions[mp.currIdx].move
				mp.currIdx++
				return move
			}
			mp.currIdx = 0
			mp.stage++

		case KILLER1:
			mp.stage++
			if mp.skipQuiets {
				continue
			}

			if mp.killer1 != mp.ttMove {
				if mp.board.IsLegal(mp.killer1) {
					return mp.killer1
				}
			}

		case KILLER2:
			mp.stage++
			if mp.skipQuiets {
				continue
			}

			if mp.killer2 != mp.killer1 && mp.killer2 != mp.ttMove {
				if mp.board.IsLegal(mp.killer2) {
					return mp.killer2
				}
			}

		case COUNTER:
			mp.stage++
			if mp.skipQuiets {
				continue
			}

			if mp.counter != mp.killer1 && mp.counter != mp.killer2 && mp.counter != mp.ttMove {
				if mp.board.IsLegal(mp.counter) {
					return mp.counter
				}
			}

		case GEN_QUIETS:
			mp.stage++
			if mp.skipQuiets {
				continue
			}
			mp.processMoves(mp.board.GenerateQuiets())

		case HISTORY_QUIETS:
			if mp.skipQuiets {
				mp.currIdx = 0
				mp.stage++
				continue
			}

			if mp.getNextAndSwap(mp.quiets, mp.currIdx) {
				move := mp.quiets[mp.currIdx].move
				if move == mp.killer1 || move == mp.killer2 || move == mp.ttMove {
					mp.currIdx++
					continue
				}

				mp.currIdx++
				return move
			}

			mp.currIdx = 0
			mp.stage++

		case BAD_CAPTURES:
			if mp.getNextAndSwap(mp.badCaptures, mp.currIdx) {
				move := mp.badCaptures[mp.currIdx].move
				if move == mp.ttMove {
					mp.currIdx++
					continue
				}
				mp.currIdx++
				return move
			}
			mp.stage++
		}
	}
	return Move{}
}
