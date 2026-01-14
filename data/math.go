package data

import (
	"math"
)

// SMA calculates Simple Moving Average.
// Returns NaN for the first (period-1) values.
func SMA(values []float64, period int) []float64 {
	n := len(values)
	result := make([]float64, n)

	if period <= 0 || period > n {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Fill initial values with NaN
	for i := 0; i < period-1; i++ {
		result[i] = math.NaN()
	}

	// Calculate first SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate remaining SMAs using sliding window
	for i := period; i < n; i++ {
		sum = sum - values[i-period] + values[i]
		result[i] = sum / float64(period)
	}

	return result
}

// EMA calculates Exponential Moving Average.
// Uses the standard multiplier: 2 / (period + 1)
// Returns NaN for the first (period-1) values.
func EMA(values []float64, period int) []float64 {
	n := len(values)
	result := make([]float64, n)

	if period <= 0 || period > n {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Fill initial values with NaN
	for i := 0; i < period-1; i++ {
		result[i] = math.NaN()
	}

	// First EMA is SMA of first period values
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate remaining EMAs
	multiplier := 2.0 / float64(period+1)
	for i := period; i < n; i++ {
		result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}

// StdDev calculates rolling standard deviation (sample std dev).
// Returns NaN for the first (period-1) values.
func StdDev(values []float64, period int) []float64 {
	n := len(values)
	result := make([]float64, n)

	if period <= 1 || period > n {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Fill initial values with NaN
	for i := 0; i < period-1; i++ {
		result[i] = math.NaN()
	}

	// Calculate for each window
	for i := period - 1; i < n; i++ {
		// Calculate mean
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += values[j]
		}
		mean := sum / float64(period)

		// Calculate variance
		sumSquares := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := values[j] - mean
			sumSquares += diff * diff
		}
		// Sample standard deviation (n-1)
		result[i] = math.Sqrt(sumSquares / float64(period-1))
	}

	return result
}

// RollingMax calculates rolling maximum.
// Returns NaN for the first (period-1) values.
func RollingMax(values []float64, period int) []float64 {
	n := len(values)
	result := make([]float64, n)

	if period <= 0 || period > n {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Fill initial values with NaN
	for i := 0; i < period-1; i++ {
		result[i] = math.NaN()
	}

	// Calculate for each window
	for i := period - 1; i < n; i++ {
		maxVal := math.Inf(-1)
		for j := i - period + 1; j <= i; j++ {
			if values[j] > maxVal {
				maxVal = values[j]
			}
		}
		result[i] = maxVal
	}

	return result
}

// RollingMin calculates rolling minimum.
// Returns NaN for the first (period-1) values.
func RollingMin(values []float64, period int) []float64 {
	n := len(values)
	result := make([]float64, n)

	if period <= 0 || period > n {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Fill initial values with NaN
	for i := 0; i < period-1; i++ {
		result[i] = math.NaN()
	}

	// Calculate for each window
	for i := period - 1; i < n; i++ {
		minVal := math.Inf(1)
		for j := i - period + 1; j <= i; j++ {
			if values[j] < minVal {
				minVal = values[j]
			}
		}
		result[i] = minVal
	}

	return result
}

// RollingSum calculates rolling sum.
// Returns NaN for the first (period-1) values.
func RollingSum(values []float64, period int) []float64 {
	n := len(values)
	result := make([]float64, n)

	if period <= 0 || period > n {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Fill initial values with NaN
	for i := 0; i < period-1; i++ {
		result[i] = math.NaN()
	}

	// Calculate first sum
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[period-1] = sum

	// Calculate remaining sums using sliding window
	for i := period; i < n; i++ {
		sum = sum - values[i-period] + values[i]
		result[i] = sum
	}

	return result
}

// Diff calculates first difference (values[i] - values[i-1]).
// First value is NaN.
func Diff(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)

	if n == 0 {
		return result
	}

	result[0] = math.NaN()
	for i := 1; i < n; i++ {
		result[i] = values[i] - values[i-1]
	}

	return result
}

// Shift shifts values by n periods.
// Positive n shifts forward (adds NaN at start), negative n shifts backward.
func Shift(values []float64, periods int) []float64 {
	n := len(values)
	result := make([]float64, n)

	if periods >= n || periods <= -n {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	if periods > 0 {
		// Shift forward: NaN at start
		for i := 0; i < periods; i++ {
			result[i] = math.NaN()
		}
		for i := periods; i < n; i++ {
			result[i] = values[i-periods]
		}
	} else if periods < 0 {
		// Shift backward: NaN at end
		periods = -periods
		for i := 0; i < n-periods; i++ {
			result[i] = values[i+periods]
		}
		for i := n - periods; i < n; i++ {
			result[i] = math.NaN()
		}
	} else {
		copy(result, values)
	}

	return result
}

// FillNaN replaces NaN values with the given fill value.
func FillNaN(values []float64, fillValue float64) []float64 {
	result := make([]float64, len(values))
	for i, v := range values {
		if math.IsNaN(v) {
			result[i] = fillValue
		} else {
			result[i] = v
		}
	}
	return result
}

// FillNaNForward fills NaN values with the previous non-NaN value.
func FillNaNForward(values []float64) []float64 {
	result := make([]float64, len(values))
	lastValid := math.NaN()
	for i, v := range values {
		if !math.IsNaN(v) {
			lastValid = v
		}
		result[i] = lastValid
	}
	return result
}

// IsNaN checks if a value is NaN.
func IsNaN(value float64) bool {
	return math.IsNaN(value)
}

// CountNaN returns the count of NaN values in the slice.
func CountNaN(values []float64) int {
	count := 0
	for _, v := range values {
		if math.IsNaN(v) {
			count++
		}
	}
	return count
}

// PctChange calculates percentage change from previous value.
// First value is NaN.
func PctChange(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)

	if n == 0 {
		return result
	}

	result[0] = math.NaN()
	for i := 1; i < n; i++ {
		if values[i-1] == 0 {
			result[i] = math.NaN()
		} else {
			result[i] = (values[i] - values[i-1]) / values[i-1]
		}
	}

	return result
}

// CumSum calculates cumulative sum.
func CumSum(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)

	if n == 0 {
		return result
	}

	result[0] = values[0]
	for i := 1; i < n; i++ {
		result[i] = result[i-1] + values[i]
	}

	return result
}

// CumProd calculates cumulative product.
func CumProd(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)

	if n == 0 {
		return result
	}

	result[0] = values[0]
	for i := 1; i < n; i++ {
		result[i] = result[i-1] * values[i]
	}

	return result
}

// CumMax calculates cumulative maximum.
func CumMax(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)

	if n == 0 {
		return result
	}

	result[0] = values[0]
	for i := 1; i < n; i++ {
		if values[i] > result[i-1] {
			result[i] = values[i]
		} else {
			result[i] = result[i-1]
		}
	}

	return result
}

// CumMin calculates cumulative minimum.
func CumMin(values []float64) []float64 {
	n := len(values)
	result := make([]float64, n)

	if n == 0 {
		return result
	}

	result[0] = values[0]
	for i := 1; i < n; i++ {
		if values[i] < result[i-1] {
			result[i] = values[i]
		} else {
			result[i] = result[i-1]
		}
	}

	return result
}

// Abs returns absolute values.
func Abs(values []float64) []float64 {
	result := make([]float64, len(values))
	for i, v := range values {
		result[i] = math.Abs(v)
	}
	return result
}

// Clip clips values to be within [min, max].
func Clip(values []float64, minVal, maxVal float64) []float64 {
	result := make([]float64, len(values))
	for i, v := range values {
		if v < minVal {
			result[i] = minVal
		} else if v > maxVal {
			result[i] = maxVal
		} else {
			result[i] = v
		}
	}
	return result
}
