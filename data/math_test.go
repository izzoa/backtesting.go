package data

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return math.Abs(a-b) <= tolerance
}

func slicesAlmostEqual(a, b []float64, tolerance float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !almostEqual(a[i], b[i], tolerance) {
			return false
		}
	}
	return true
}

func TestSMA(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		period   int
		expected []float64
	}{
		{
			name:     "simple case",
			values:   []float64{1, 2, 3, 4, 5},
			period:   3,
			expected: []float64{math.NaN(), math.NaN(), 2, 3, 4},
		},
		{
			name:     "period 1",
			values:   []float64{1, 2, 3, 4, 5},
			period:   1,
			expected: []float64{1, 2, 3, 4, 5},
		},
		{
			name:     "period equals length",
			values:   []float64{1, 2, 3, 4, 5},
			period:   5,
			expected: []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), 3},
		},
		{
			name:     "longer series",
			values:   []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			period:   4,
			expected: []float64{math.NaN(), math.NaN(), math.NaN(), 25, 35, 45, 55, 65, 75, 85},
		},
		{
			name:     "constant values",
			values:   []float64{5, 5, 5, 5, 5},
			period:   3,
			expected: []float64{math.NaN(), math.NaN(), 5, 5, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SMA(tt.values, tt.period)
			if !slicesAlmostEqual(result, tt.expected, 1e-9) {
				t.Errorf("SMA() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSMA_EdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		result := SMA([]float64{}, 3)
		if len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("period > length", func(t *testing.T) {
		result := SMA([]float64{1, 2, 3}, 5)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("result[%d] = %f, want NaN", i, v)
			}
		}
	})

	t.Run("period <= 0", func(t *testing.T) {
		result := SMA([]float64{1, 2, 3}, 0)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("result[%d] = %f, want NaN", i, v)
			}
		}
	})
}

func TestEMA(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		period int
	}{
		{
			name:   "basic EMA",
			values: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			period: 3,
		},
		{
			name:   "period 1",
			values: []float64{1, 2, 3, 4, 5},
			period: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EMA(tt.values, tt.period)

			// Verify length
			if len(result) != len(tt.values) {
				t.Errorf("EMA length = %d, want %d", len(result), len(tt.values))
			}

			// Verify first period-1 values are NaN
			for i := 0; i < tt.period-1; i++ {
				if !math.IsNaN(result[i]) {
					t.Errorf("result[%d] = %f, want NaN", i, result[i])
				}
			}

			// Verify values after warmup are not NaN
			for i := tt.period - 1; i < len(result); i++ {
				if math.IsNaN(result[i]) {
					t.Errorf("result[%d] = NaN, want non-NaN", i)
				}
			}
		})
	}
}

func TestEMA_Formula(t *testing.T) {
	// Verify EMA formula: EMA = Price * k + EMA_prev * (1-k), k = 2/(period+1)
	values := []float64{10, 12, 11, 13, 14}
	period := 3
	k := 2.0 / float64(period+1)

	result := EMA(values, period)

	// First EMA should be SMA of first 3 values
	expectedFirst := (10.0 + 12.0 + 11.0) / 3.0
	if !almostEqual(result[2], expectedFirst, 1e-9) {
		t.Errorf("First EMA = %f, want %f", result[2], expectedFirst)
	}

	// Second EMA: 13 * k + expectedFirst * (1-k)
	expectedSecond := 13.0*k + expectedFirst*(1-k)
	if !almostEqual(result[3], expectedSecond, 1e-9) {
		t.Errorf("Second EMA = %f, want %f", result[3], expectedSecond)
	}

	// Third EMA
	expectedThird := 14.0*k + expectedSecond*(1-k)
	if !almostEqual(result[4], expectedThird, 1e-9) {
		t.Errorf("Third EMA = %f, want %f", result[4], expectedThird)
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		period int
	}{
		{
			name:   "varying values",
			values: []float64{2, 4, 4, 4, 5, 5, 7, 9},
			period: 4,
		},
		{
			name:   "constant values",
			values: []float64{5, 5, 5, 5, 5},
			period: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StdDev(tt.values, tt.period)

			// Verify first period-1 values are NaN
			for i := 0; i < tt.period-1; i++ {
				if !math.IsNaN(result[i]) {
					t.Errorf("result[%d] = %f, want NaN", i, result[i])
				}
			}

			// Verify non-negative after warmup
			for i := tt.period - 1; i < len(result); i++ {
				if !math.IsNaN(result[i]) && result[i] < 0 {
					t.Errorf("result[%d] = %f, stddev cannot be negative", i, result[i])
				}
			}
		})
	}
}

func TestStdDev_ConstantValues(t *testing.T) {
	// Constant values should have stddev = 0
	values := []float64{5, 5, 5, 5, 5}
	result := StdDev(values, 3)

	for i := 2; i < len(result); i++ {
		if !almostEqual(result[i], 0, 1e-9) {
			t.Errorf("result[%d] = %f, want 0 for constant values", i, result[i])
		}
	}
}

func TestRollingMax(t *testing.T) {
	values := []float64{5, 3, 8, 1, 9, 2, 7, 4, 6, 10}
	expected := []float64{math.NaN(), math.NaN(), 8, 8, 9, 9, 9, 7, 7, 10}

	result := RollingMax(values, 3)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("RollingMax() = %v, want %v", result, expected)
	}
}

func TestRollingMin(t *testing.T) {
	values := []float64{5, 3, 8, 1, 9, 2, 7, 4, 6, 10}
	expected := []float64{math.NaN(), math.NaN(), 3, 1, 1, 1, 2, 2, 4, 4}

	result := RollingMin(values, 3)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("RollingMin() = %v, want %v", result, expected)
	}
}

func TestRollingSum(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	expected := []float64{math.NaN(), math.NaN(), 6, 9, 12}

	result := RollingSum(values, 3)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("RollingSum() = %v, want %v", result, expected)
	}
}

func TestDiff(t *testing.T) {
	values := []float64{1, 3, 6, 10, 15}
	expected := []float64{math.NaN(), 2, 3, 4, 5}

	result := Diff(values)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("Diff() = %v, want %v", result, expected)
	}
}

func TestShift(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}

	t.Run("shift forward", func(t *testing.T) {
		result := Shift(values, 2)
		expected := []float64{math.NaN(), math.NaN(), 1, 2, 3}
		if !slicesAlmostEqual(result, expected, 1e-9) {
			t.Errorf("Shift(2) = %v, want %v", result, expected)
		}
	})

	t.Run("shift backward", func(t *testing.T) {
		result := Shift(values, -2)
		expected := []float64{3, 4, 5, math.NaN(), math.NaN()}
		if !slicesAlmostEqual(result, expected, 1e-9) {
			t.Errorf("Shift(-2) = %v, want %v", result, expected)
		}
	})

	t.Run("no shift", func(t *testing.T) {
		result := Shift(values, 0)
		if !slicesAlmostEqual(result, values, 1e-9) {
			t.Errorf("Shift(0) = %v, want %v", result, values)
		}
	})
}

func TestFillNaN(t *testing.T) {
	values := []float64{1, math.NaN(), 3, math.NaN(), 5}
	result := FillNaN(values, 0)
	expected := []float64{1, 0, 3, 0, 5}

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("FillNaN() = %v, want %v", result, expected)
	}
}

func TestFillNaNForward(t *testing.T) {
	values := []float64{1, math.NaN(), math.NaN(), 4, math.NaN()}
	result := FillNaNForward(values)
	expected := []float64{1, 1, 1, 4, 4}

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("FillNaNForward() = %v, want %v", result, expected)
	}
}

func TestPctChange(t *testing.T) {
	values := []float64{100, 110, 99, 108}
	result := PctChange(values)

	// Expected: NaN, 0.1, -0.1, 0.0909...
	if !math.IsNaN(result[0]) {
		t.Errorf("result[0] = %f, want NaN", result[0])
	}
	if !almostEqual(result[1], 0.1, 1e-9) {
		t.Errorf("result[1] = %f, want 0.1", result[1])
	}
	if !almostEqual(result[2], -0.1, 1e-9) {
		t.Errorf("result[2] = %f, want -0.1", result[2])
	}
}

func TestCumSum(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	expected := []float64{1, 3, 6, 10, 15}

	result := CumSum(values)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("CumSum() = %v, want %v", result, expected)
	}
}

func TestCumProd(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	expected := []float64{1, 2, 6, 24, 120}

	result := CumProd(values)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("CumProd() = %v, want %v", result, expected)
	}
}

func TestCumMax(t *testing.T) {
	values := []float64{1, 3, 2, 5, 4}
	expected := []float64{1, 3, 3, 5, 5}

	result := CumMax(values)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("CumMax() = %v, want %v", result, expected)
	}
}

func TestCumMin(t *testing.T) {
	values := []float64{5, 3, 4, 1, 2}
	expected := []float64{5, 3, 3, 1, 1}

	result := CumMin(values)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("CumMin() = %v, want %v", result, expected)
	}
}

func TestAbs(t *testing.T) {
	values := []float64{-1, 2, -3, 4, -5}
	expected := []float64{1, 2, 3, 4, 5}

	result := Abs(values)

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("Abs() = %v, want %v", result, expected)
	}
}

func TestClip(t *testing.T) {
	values := []float64{1, 5, 10, 15, 20}
	result := Clip(values, 5, 15)
	expected := []float64{5, 5, 10, 15, 15}

	if !slicesAlmostEqual(result, expected, 1e-9) {
		t.Errorf("Clip() = %v, want %v", result, expected)
	}
}

func BenchmarkSMA(b *testing.B) {
	values := make([]float64, 10000)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SMA(values, 20)
	}
}

func BenchmarkEMA(b *testing.B) {
	values := make([]float64, 10000)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EMA(values, 20)
	}
}

func BenchmarkStdDev(b *testing.B) {
	values := make([]float64, 10000)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StdDev(values, 20)
	}
}
