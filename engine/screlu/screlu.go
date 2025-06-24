package screlu

/*
#cgo CFLAGS: -mavx2 -march=haswell -O3
#include "screlu_simd.h"
*/
import "C"
import (
	"unsafe"
)

var AVX2_ENABLED bool = true

// values: accumulator
// weights: output layer weights
func SCReLUFusedSIMDSum(stmValues []int16, ntmValues []int16, weights []int16, QA int16) int32 {
	length := len(stmValues)

	if AVX2_ENABLED && length%16 == 0 {
		return int32(C.screlu_fused_simd_sum(
			(*C.int16_t)(unsafe.Pointer(&stmValues[0])),
			(*C.int16_t)(unsafe.Pointer(&ntmValues[0])),
			(*C.int16_t)(unsafe.Pointer(&weights[0])),
			C.int(length),
			C.int16_t(QA),
		))
	}

	// fallback scalar sum
	var total int32
	for i := 0; i < length; i++ {
		s := stmValues[i]
		t := ntmValues[i]

		if s < 0 {
			s = 0
		}
		if s > QA {
			s = QA
		}
		if t < 0 {
			t = 0
		}
		if t > QA {
			t = QA
		}

		total += int32(s) * int32(s) * int32(weights[i])
		total += int32(t) * int32(t) * int32(weights[i+length])
	}
	return total
}
