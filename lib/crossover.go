package lib

import "math"

// Cross returns true at indices where series1 crosses series2 (either direction).
func Cross(series1, series2 []float64) []bool {
	if len(series1) != len(series2) || len(series1) < 2 {
		return make([]bool, len(series1))
	}

	result := make([]bool, len(series1))

	for i := 1; i < len(series1); i++ {
		prev1, curr1 := series1[i-1], series1[i]
		prev2, curr2 := series2[i-1], series2[i]

		if math.IsNaN(prev1) || math.IsNaN(curr1) || math.IsNaN(prev2) || math.IsNaN(curr2) {
			continue
		}

		// Cross occurs when sign of difference changes
		prevDiff := prev1 - prev2
		currDiff := curr1 - curr2

		result[i] = (prevDiff <= 0 && currDiff > 0) || (prevDiff >= 0 && currDiff < 0)
	}

	return result
}

// Crossover returns true at indices where series1 crosses above series2.
func Crossover(series1, series2 []float64) []bool {
	if len(series1) != len(series2) || len(series1) < 2 {
		return make([]bool, len(series1))
	}

	result := make([]bool, len(series1))

	for i := 1; i < len(series1); i++ {
		prev1, curr1 := series1[i-1], series1[i]
		prev2, curr2 := series2[i-1], series2[i]

		if math.IsNaN(prev1) || math.IsNaN(curr1) || math.IsNaN(prev2) || math.IsNaN(curr2) {
			continue
		}

		// Crossover: was at or below, now above
		result[i] = prev1 <= prev2 && curr1 > curr2
	}

	return result
}

// Crossunder returns true at indices where series1 crosses below series2.
func Crossunder(series1, series2 []float64) []bool {
	if len(series1) != len(series2) || len(series1) < 2 {
		return make([]bool, len(series1))
	}

	result := make([]bool, len(series1))

	for i := 1; i < len(series1); i++ {
		prev1, curr1 := series1[i-1], series1[i]
		prev2, curr2 := series2[i-1], series2[i]

		if math.IsNaN(prev1) || math.IsNaN(curr1) || math.IsNaN(prev2) || math.IsNaN(curr2) {
			continue
		}

		// Crossunder: was at or above, now below
		result[i] = prev1 >= prev2 && curr1 < curr2
	}

	return result
}

// BarsSince returns the number of bars since the condition was last true.
// defaultVal is returned if the condition was never true.
func BarsSince(condition []bool, defaultVal int) []int {
	result := make([]int, len(condition))

	lastTrue := -1
	for i, cond := range condition {
		if cond {
			lastTrue = i
		}

		if lastTrue < 0 {
			result[i] = defaultVal
		} else {
			result[i] = i - lastTrue
		}
	}

	return result
}

// CrossoverValue returns true if series1 crosses above the given value.
func CrossoverValue(series []float64, value float64) []bool {
	if len(series) < 2 {
		return make([]bool, len(series))
	}

	result := make([]bool, len(series))

	for i := 1; i < len(series); i++ {
		prev, curr := series[i-1], series[i]

		if math.IsNaN(prev) || math.IsNaN(curr) {
			continue
		}

		result[i] = prev <= value && curr > value
	}

	return result
}

// CrossunderValue returns true if series crosses below the given value.
func CrossunderValue(series []float64, value float64) []bool {
	if len(series) < 2 {
		return make([]bool, len(series))
	}

	result := make([]bool, len(series))

	for i := 1; i < len(series); i++ {
		prev, curr := series[i-1], series[i]

		if math.IsNaN(prev) || math.IsNaN(curr) {
			continue
		}

		result[i] = prev >= value && curr < value
	}

	return result
}
