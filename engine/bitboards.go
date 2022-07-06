package engine

type u64 uint64

type board struct {
	whitePawns   u64
	whiteKnights u64
	whiteBishops u64
	whiteRooks   u64
	whiteQueens  u64
	whiteKing    u64
	blackPawns   u64
	blackKnights u64
	blackBishops u64
	blackRooks   u64
	blackQueens  u64
	blackKing    u64
}
