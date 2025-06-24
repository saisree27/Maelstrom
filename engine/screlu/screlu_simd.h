#ifndef SCRELU_SIMD_H
#define SCRELU_SIMD_H

#include <stdint.h>

int32_t screlu_fused_simd_sum(
    const int16_t* stm_values,
    const int16_t* ntm_values,
    const int16_t* weights,
    int len,
    int16_t QA
);

#endif