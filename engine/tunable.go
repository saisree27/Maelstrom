package engine

type TunableParameters struct {
	ASPIRATION_WINDOW_SIZE      int
	RFP_MULT                    int
	RFP_MAX_DEPTH               int
	RAZORING_MULT               int
	RAZORING_MAX_DEPTH          int
	FUTILITY_BASE               int
	FUTILITY_MULT               int
	FUTILITY_MAX_DEPTH          int
	IIR_MIN_DEPTH               int
	IIR_DEPTH_REDUCTION         int
	LMR_MIN_DEPTH               int
	LMR_CHECK                   int
	LMR_TT_CAPTURE              int
	LMR_NOT_PV                  int
	NMP_MIN_DEPTH               int
	LMP_MAX_DEPTH               int
	LMP_BASE                    int
	LMP_MULT                    int
	SEE_PAWN_VALUE              int
	SEE_KNIGHT_VALUE            int
	SEE_BISHOP_VALUE            int
	SEE_ROOK_VALUE              int
	SEE_QUEEN_VALUE             int
	SEE_QUIET_PRUNING_MAX_DEPTH int
	SEE_QUIET_PRUNING_MULT      int
	SEE_CAPTURE_PRUNING_MULT    int
	TIME_DIVISOR                int64
	INC_FRACTION                int64
	HARD_LIMIT_MULT             int64
	TM_STABILITY_WINDOW         int
	TM_NODE_COUNT_CONSTANT      int
}

var Params = TunableParameters{
	ASPIRATION_WINDOW_SIZE:   25,
	RFP_MULT:                 132,
	RFP_MAX_DEPTH:            11,
	RAZORING_MULT:            228,
	RAZORING_MAX_DEPTH:       2,
	FUTILITY_BASE:            51,
	FUTILITY_MULT:            142,
	FUTILITY_MAX_DEPTH:       10,
	IIR_MIN_DEPTH:            4,
	IIR_DEPTH_REDUCTION:      1,
	LMR_MIN_DEPTH:            2,
	LMR_CHECK:                1170,
	LMR_TT_CAPTURE:           1130,
	LMR_NOT_PV:               1050,
	NMP_MIN_DEPTH:            2,
	LMP_MAX_DEPTH:            7,
	LMP_BASE:                 5,
	LMP_MULT:                 2,
	SEE_PAWN_VALUE:           100,
	SEE_KNIGHT_VALUE:         300,
	SEE_BISHOP_VALUE:         300,
	SEE_ROOK_VALUE:           500,
	SEE_QUEEN_VALUE:          900,
	SEE_QUIET_PRUNING_MULT:   20,
	SEE_CAPTURE_PRUNING_MULT: 100,
	TIME_DIVISOR:             20,
	INC_FRACTION:             2,
	HARD_LIMIT_MULT:          2,
	TM_STABILITY_WINDOW:      10,
	TM_NODE_COUNT_CONSTANT:   15,
}
