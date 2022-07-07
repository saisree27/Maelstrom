package engine

import (
	"fmt"
	"math/bits"
)

func printBitBoard(u u64) {
	s := "\n"
	for i := 56; i >= 0; i -= 8 {
		for j := 0; j < 8; j++ {
			// lookup bit at square i + j
			if u&u64(1<<(i+j)) != 0 {
				s += "1" + " "
			} else {
				s += "0" + " "
			}
		}
		s += "\n"
	}
	fmt.Print(s)
}

func popCount(u u64) int {
	return bits.OnesCount64(uint64(u))
}

// Below bitboard operations are for ease of use when generating legal moves
// Found on chessprogramming wiki
var index64 = [64]int{
	0, 47, 1, 56, 48, 27, 2, 60,
	57, 49, 41, 37, 28, 16, 3, 61,
	54, 58, 35, 52, 50, 42, 21, 44,
	38, 32, 29, 23, 17, 11, 4, 62,
	46, 55, 26, 59, 40, 36, 15, 53,
	34, 51, 20, 43, 31, 22, 10, 45,
	25, 39, 14, 33, 19, 30, 9, 24,
	13, 18, 8, 12, 7, 6, 5, 63}

func bitScanForward(bb u64) int {
	const debruijn64 = u64(0x03f79d71b4cb0a89)
	return index64[((bb^(bb-1))*debruijn64)>>58]
}

func flipVertical(x u64) u64 {
	const k1 = u64(0x00FF00FF00FF00FF)
	const k2 = u64(0x0000FFFF0000FFFF)

	x = ((x >> 8) & k1) | ((x & k1) << 8)
	x = ((x >> 16) & k2) | ((x & k2) << 16)
	x = (x >> 32) | (x << 32)

	return x
}

func mirrorHorizontal(x u64) u64 {
	const k1 = u64(0x5555555555555555)
	const k2 = u64(0x3333333333333333)
	const k4 = u64(0x0f0f0f0f0f0f0f0f)
	x = ((x >> 1) & k1) + 2*(x&k1)
	x = ((x >> 2) & k2) + 4*(x&k2)
	x = ((x >> 4) & k4) + 16*(x&k4)
	return x
}

func flipDiagA1H8(x u64) u64 {
	var t u64
	const k1 = u64(0x5500550055005500)
	const k2 = u64(0x3333000033330000)
	const k4 = u64(0x0f0f0f0f00000000)
	t = k4 & (x ^ (x << 28))
	x ^= t ^ (t >> 28)
	t = k2 & (x ^ (x << 14))
	x ^= t ^ (t >> 14)
	t = k1 & (x ^ (x << 7))
	x ^= t ^ (t >> 7)
	return x
}

func flipDiagA8H1(x u64) u64 {
	var t u64
	const k1 = u64(0xaa00aa00aa00aa00)
	const k2 = u64(0xcccc0000cccc0000)
	const k4 = u64(0xf0f0f0f00f0f0f0f)
	t = x ^ (x << 36)
	x ^= k4 & (t ^ (x >> 36))
	t = k2 & (x ^ (x << 18))
	x ^= t ^ (t >> 18)
	t = k1 & (x ^ (x << 9))
	x ^= t ^ (t >> 9)
	return x
}

// all rotations below are done clockwise
func rotate90(x u64) u64 {
	return flipVertical(flipDiagA1H8(x))
}
func rotate180(x u64) u64 {
	return mirrorHorizontal(flipVertical(x))
}
func rotate270(x u64) u64 {
	return flipDiagA1H8(flipVertical(x))
}

// bitwise rotate (shift but wrap around)
func rotateLeft(x u64, n int) u64 {
	return u64(bits.RotateLeft64(uint64(x), n))
}

func rotateRight(x u64, n int) u64 {
	return (x >> n) | (x << (64 - n))
}
