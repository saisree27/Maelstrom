package engine

type SearchParams struct {
	RFP_MULT            int
	RFP_MAX_DEPTH       int
	RAZORING_MULT       int
	RAZORING_MAX_DEPTH  int
	FUTILITY_BASE       int
	FUTILITY_MULT       int
	FUTILITY_MAX_DEPTH  int
	IIR_MIN_DEPTH       int
	IIR_DEPTH_REDUCTION int
	LMR_MIN_DEPTH       int
	NMP_MIN_DEPTH       int
}

var Params = SearchParams{
	RFP_MULT:            120,
	RFP_MAX_DEPTH:       9,
	RAZORING_MULT:       220,
	RAZORING_MAX_DEPTH:  2,
	FUTILITY_BASE:       50,
	FUTILITY_MULT:       150,
	FUTILITY_MAX_DEPTH:  8,
	IIR_MIN_DEPTH:       4,
	IIR_DEPTH_REDUCTION: 1,
	LMR_MIN_DEPTH:       3,
	NMP_MIN_DEPTH:       3,
}
