package engine

type TunableParameters struct {
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
	NMP_MIN_DEPTH               int
	SEE_QUIET_PRUNING_MAX_DEPTH int
	SEE_QUIET_PRUNING_MULT      int
	SEE_CAPTURE_PRUNING_MULT    int
	TIME_DIVISOR                int64
	INC_FRACTION                int64
	HARD_LIMIT_MULT             int64
}

var Params = TunableParameters{
	RFP_MULT:                 132,
	RFP_MAX_DEPTH:            9,
	RAZORING_MULT:            228,
	RAZORING_MAX_DEPTH:       2,
	FUTILITY_BASE:            51,
	FUTILITY_MULT:            142,
	FUTILITY_MAX_DEPTH:       8,
	IIR_MIN_DEPTH:            4,
	IIR_DEPTH_REDUCTION:      1,
	LMR_MIN_DEPTH:            2,
	NMP_MIN_DEPTH:            2,
	SEE_QUIET_PRUNING_MULT:   20,
	SEE_CAPTURE_PRUNING_MULT: 100,
	TIME_DIVISOR:             25,
	INC_FRACTION:             2,
	HARD_LIMIT_MULT:          2,
}

var TUNING_EXPOSE_UCI = false
