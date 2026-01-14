package lib

import (
	"math"
	"sort"
)

// Quantile calculates the q-th quantile of a series.
// q should be between 0 and 1 (e.g., 0.5 for median).
func Quantile(series []float64, q float64) float64 {
	if len(series) == 0 || q < 0 || q > 1 {
		return math.NaN()
	}

	// Filter out NaN values and copy
	values := make([]float64, 0, len(series))
	for _, v := range series {
		if !math.IsNaN(v) {
			values = append(values, v)
		}
	}

	if len(values) == 0 {
		return math.NaN()
	}

	// Sort values
	sort.Float64s(values)

	// Calculate index
	n := float64(len(values))
	index := q * (n - 1)

	// Linear interpolation
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper || upper >= len(values) {
		return values[lower]
	}

	// Interpolate
	frac := index - float64(lower)
	return values[lower]*(1-frac) + values[upper]*frac
}

// RollingQuantile calculates the rolling q-th quantile over a window.
func RollingQuantile(series []float64, window int, q float64) []float64 {
	if len(series) == 0 || window <= 0 || q < 0 || q > 1 {
		return nil
	}

	result := make([]float64, len(series))
	for i := range result {
		result[i] = math.NaN()
	}

	for i := window - 1; i < len(series); i++ {
		windowSlice := series[i-window+1 : i+1]
		result[i] = Quantile(windowSlice, q)
	}

	return result
}

// Median calculates the median of a series.
func Median(series []float64) float64 {
	return Quantile(series, 0.5)
}

// RollingMedian calculates the rolling median over a window.
func RollingMedian(series []float64, window int) []float64 {
	return RollingQuantile(series, window, 0.5)
}

// Percentile calculates the p-th percentile of a series.
// p should be between 0 and 100.
func Percentile(series []float64, p float64) float64 {
	return Quantile(series, p/100)
}

// RollingPercentile calculates the rolling p-th percentile over a window.
func RollingPercentile(series []float64, window int, p float64) []float64 {
	return RollingQuantile(series, window, p/100)
}

// IQR calculates the Interquartile Range (Q3 - Q1).
func IQR(series []float64) float64 {
	q1 := Quantile(series, 0.25)
	q3 := Quantile(series, 0.75)
	if math.IsNaN(q1) || math.IsNaN(q3) {
		return math.NaN()
	}
	return q3 - q1
}
