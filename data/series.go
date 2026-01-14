package data

import (
	"fmt"
	"math"
)

// Series represents a named sequence of float64 values.
type Series struct {
	Name   string
	Values []float64
}

// NewSeries creates a new Series with the given name and values.
func NewSeries(name string, values []float64) *Series {
	return &Series{
		Name:   name,
		Values: values,
	}
}

// Len returns the number of values in the series.
func (s *Series) Len() int {
	return len(s.Values)
}

// At returns the value at the given index.
// Panics if index is out of bounds.
func (s *Series) At(i int) float64 {
	if i < 0 || i >= s.Len() {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, s.Len()))
	}
	return s.Values[i]
}

// Last returns the last value in the series.
// Panics if the series is empty.
func (s *Series) Last() float64 {
	if s.Len() == 0 {
		panic("cannot get last value of empty series")
	}
	return s.Values[s.Len()-1]
}

// LastN returns the last n values as a slice.
// If n > Len(), returns all values.
func (s *Series) LastN(n int) []float64 {
	if n >= s.Len() {
		result := make([]float64, s.Len())
		copy(result, s.Values)
		return result
	}
	result := make([]float64, n)
	copy(result, s.Values[s.Len()-n:])
	return result
}

// Slice returns a new Series containing values from start to end (exclusive).
func (s *Series) Slice(start, end int) *Series {
	if start < 0 {
		panic(fmt.Sprintf("start index %d cannot be negative", start))
	}
	if end > s.Len() {
		panic(fmt.Sprintf("end index %d exceeds length %d", end, s.Len()))
	}
	if start > end {
		panic(fmt.Sprintf("start index %d greater than end index %d", start, end))
	}
	return &Series{
		Name:   s.Name,
		Values: s.Values[start:end],
	}
}

// Copy returns a deep copy of the series.
func (s *Series) Copy() *Series {
	values := make([]float64, s.Len())
	copy(values, s.Values)
	return &Series{
		Name:   s.Name,
		Values: values,
	}
}

// Apply applies a function to each value and returns a new Series.
func (s *Series) Apply(fn func(float64) float64) *Series {
	values := make([]float64, s.Len())
	for i, v := range s.Values {
		values[i] = fn(v)
	}
	return &Series{
		Name:   s.Name,
		Values: values,
	}
}

// ApplyWithIndex applies a function that takes both value and index.
func (s *Series) ApplyWithIndex(fn func(int, float64) float64) *Series {
	values := make([]float64, s.Len())
	for i, v := range s.Values {
		values[i] = fn(i, v)
	}
	return &Series{
		Name:   s.Name,
		Values: values,
	}
}

// Rolling creates a RollingWindow for windowed calculations.
func (s *Series) Rolling(window int) *RollingWindow {
	return &RollingWindow{
		series: s,
		window: window,
	}
}

// Sum returns the sum of all non-NaN values.
func (s *Series) Sum() float64 {
	sum := 0.0
	for _, v := range s.Values {
		if !math.IsNaN(v) {
			sum += v
		}
	}
	return sum
}

// Mean returns the mean of all non-NaN values.
func (s *Series) Mean() float64 {
	sum := 0.0
	count := 0
	for _, v := range s.Values {
		if !math.IsNaN(v) {
			sum += v
			count++
		}
	}
	if count == 0 {
		return math.NaN()
	}
	return sum / float64(count)
}

// Std returns the standard deviation of all non-NaN values.
func (s *Series) Std() float64 {
	mean := s.Mean()
	if math.IsNaN(mean) {
		return math.NaN()
	}

	sumSquares := 0.0
	count := 0
	for _, v := range s.Values {
		if !math.IsNaN(v) {
			diff := v - mean
			sumSquares += diff * diff
			count++
		}
	}
	if count < 2 {
		return math.NaN()
	}
	return math.Sqrt(sumSquares / float64(count-1))
}

// Min returns the minimum non-NaN value.
func (s *Series) Min() float64 {
	minVal := math.Inf(1)
	for _, v := range s.Values {
		if !math.IsNaN(v) && v < minVal {
			minVal = v
		}
	}
	if math.IsInf(minVal, 1) {
		return math.NaN()
	}
	return minVal
}

// Max returns the maximum non-NaN value.
func (s *Series) Max() float64 {
	maxVal := math.Inf(-1)
	for _, v := range s.Values {
		if !math.IsNaN(v) && v > maxVal {
			maxVal = v
		}
	}
	if math.IsInf(maxVal, -1) {
		return math.NaN()
	}
	return maxVal
}

// RollingWindow provides windowed calculations over a series.
type RollingWindow struct {
	series *Series
	window int
}

// Mean calculates rolling mean.
func (r *RollingWindow) Mean() *Series {
	return NewSeries(r.series.Name+"_sma", SMA(r.series.Values, r.window))
}

// Std calculates rolling standard deviation.
func (r *RollingWindow) Std() *Series {
	return NewSeries(r.series.Name+"_std", StdDev(r.series.Values, r.window))
}

// Max calculates rolling maximum.
func (r *RollingWindow) Max() *Series {
	return NewSeries(r.series.Name+"_max", RollingMax(r.series.Values, r.window))
}

// Min calculates rolling minimum.
func (r *RollingWindow) Min() *Series {
	return NewSeries(r.series.Name+"_min", RollingMin(r.series.Values, r.window))
}

// Sum calculates rolling sum.
func (r *RollingWindow) Sum() *Series {
	return NewSeries(r.series.Name+"_sum", RollingSum(r.series.Values, r.window))
}
