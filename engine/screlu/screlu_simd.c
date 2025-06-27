#include <immintrin.h>
#include <stdint.h>

#define QA 255

// Implementation of the Lizard SCReLU code in 
// https://www.chessprogramming.org/NNUE
int32_t screlu_fused_simd_sum(
    const int16_t* stm_values,
    const int16_t* ntm_values,
    const int16_t* weights,
    int len
) {
    const __m256i vec_zero = _mm256_setzero_si256();
    const __m256i vec_qa   = _mm256_set1_epi16(QA);
    __m256i sum = vec_zero;

    for (int i = 0; i < len; i += 16) {
        __m256i stm = _mm256_loadu_si256((const __m256i*)&stm_values[i]);
        __m256i ntm = _mm256_loadu_si256((const __m256i*)&ntm_values[i]);
        __m256i w1  = _mm256_loadu_si256((const __m256i*)&weights[i]);
        __m256i w2  = _mm256_loadu_si256((const __m256i*)&weights[i + len]);

        __m256i stm_clamp = _mm256_min_epi16(_mm256_max_epi16(stm, vec_zero), vec_qa);
        __m256i ntm_clamp = _mm256_min_epi16(_mm256_max_epi16(ntm, vec_zero), vec_qa);

        __m256i res1 = _mm256_madd_epi16(_mm256_mullo_epi16(stm_clamp, w1), stm_clamp);
        __m256i res2 = _mm256_madd_epi16(_mm256_mullo_epi16(ntm_clamp, w2), ntm_clamp);

        sum = _mm256_add_epi32(sum, res1);
        sum = _mm256_add_epi32(sum, res2);
    }

    __m128i lo = _mm256_castsi256_si128(sum);
    __m128i hi = _mm256_extracti128_si256(sum, 1);
    lo = _mm_add_epi32(lo, hi);
    lo = _mm_hadd_epi32(lo, lo);
    lo = _mm_hadd_epi32(lo, lo);

    return _mm_cvtsi128_si32(lo);
}