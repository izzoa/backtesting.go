// Package data provides data structures and utilities for financial time series data.
package data

import (
	"fmt"
	"time"
)

// OHLCV represents Open-High-Low-Close-Volume financial data.
type OHLCV struct {
	Time   []time.Time
	Open   []float64
	High   []float64
	Low    []float64
	Close  []float64
	Volume []float64
}

// Bar represents a single OHLCV record.
type Bar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// NewOHLCV creates a new OHLCV with the given capacity.
func NewOHLCV(capacity int) *OHLCV {
	return &OHLCV{
		Time:   make([]time.Time, 0, capacity),
		Open:   make([]float64, 0, capacity),
		High:   make([]float64, 0, capacity),
		Low:    make([]float64, 0, capacity),
		Close:  make([]float64, 0, capacity),
		Volume: make([]float64, 0, capacity),
	}
}

// NewOHLCVWithData creates an OHLCV from existing slices.
// All slices must have the same length.
func NewOHLCVWithData(times []time.Time, open, high, low, close, volume []float64) (*OHLCV, error) {
	n := len(times)
	if len(open) != n || len(high) != n || len(low) != n || len(close) != n || len(volume) != n {
		return nil, fmt.Errorf("all slices must have equal length")
	}
	return &OHLCV{
		Time:   times,
		Open:   open,
		High:   high,
		Low:    low,
		Close:  close,
		Volume: volume,
	}, nil
}

// Len returns the number of bars in the OHLCV data.
func (o *OHLCV) Len() int {
	return len(o.Time)
}

// At returns the bar at the given index.
// Panics if index is out of bounds.
func (o *OHLCV) At(i int) Bar {
	if i < 0 || i >= o.Len() {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, o.Len()))
	}
	return Bar{
		Time:   o.Time[i],
		Open:   o.Open[i],
		High:   o.High[i],
		Low:    o.Low[i],
		Close:  o.Close[i],
		Volume: o.Volume[i],
	}
}

// Slice returns a new OHLCV containing bars from start to end (exclusive).
// The returned OHLCV shares the underlying data with the original.
func (o *OHLCV) Slice(start, end int) *OHLCV {
	if start < 0 {
		panic(fmt.Sprintf("start index %d cannot be negative", start))
	}
	if end > o.Len() {
		panic(fmt.Sprintf("end index %d exceeds length %d", end, o.Len()))
	}
	if start > end {
		panic(fmt.Sprintf("start index %d greater than end index %d", start, end))
	}
	return &OHLCV{
		Time:   o.Time[start:end],
		Open:   o.Open[start:end],
		High:   o.High[start:end],
		Low:    o.Low[start:end],
		Close:  o.Close[start:end],
		Volume: o.Volume[start:end],
	}
}

// LastN returns a new OHLCV containing the last n bars.
// If n > Len(), returns all bars.
func (o *OHLCV) LastN(n int) *OHLCV {
	if n >= o.Len() {
		return o.Slice(0, o.Len())
	}
	return o.Slice(o.Len()-n, o.Len())
}

// Append adds a new bar to the OHLCV data.
func (o *OHLCV) Append(bar Bar) {
	o.Time = append(o.Time, bar.Time)
	o.Open = append(o.Open, bar.Open)
	o.High = append(o.High, bar.High)
	o.Low = append(o.Low, bar.Low)
	o.Close = append(o.Close, bar.Close)
	o.Volume = append(o.Volume, bar.Volume)
}

// Copy returns a deep copy of the OHLCV data.
func (o *OHLCV) Copy() *OHLCV {
	n := o.Len()
	return &OHLCV{
		Time:   append(make([]time.Time, 0, n), o.Time...),
		Open:   append(make([]float64, 0, n), o.Open...),
		High:   append(make([]float64, 0, n), o.High...),
		Low:    append(make([]float64, 0, n), o.Low...),
		Close:  append(make([]float64, 0, n), o.Close...),
		Volume: append(make([]float64, 0, n), o.Volume...),
	}
}

// Validate checks that the OHLCV data is valid:
// - All slices have equal length
// - High >= Low for all bars
// - High >= Open and High >= Close for all bars
// - Low <= Open and Low <= Close for all bars
// - Volume >= 0 for all bars
func (o *OHLCV) Validate() error {
	n := o.Len()
	if len(o.Open) != n || len(o.High) != n || len(o.Low) != n || len(o.Close) != n || len(o.Volume) != n {
		return fmt.Errorf("slice length mismatch")
	}

	for i := 0; i < n; i++ {
		if o.High[i] < o.Low[i] {
			return fmt.Errorf("bar %d: High (%f) < Low (%f)", i, o.High[i], o.Low[i])
		}
		if o.High[i] < o.Open[i] {
			return fmt.Errorf("bar %d: High (%f) < Open (%f)", i, o.High[i], o.Open[i])
		}
		if o.High[i] < o.Close[i] {
			return fmt.Errorf("bar %d: High (%f) < Close (%f)", i, o.High[i], o.Close[i])
		}
		if o.Low[i] > o.Open[i] {
			return fmt.Errorf("bar %d: Low (%f) > Open (%f)", i, o.Low[i], o.Open[i])
		}
		if o.Low[i] > o.Close[i] {
			return fmt.Errorf("bar %d: Low (%f) > Close (%f)", i, o.Low[i], o.Close[i])
		}
		if o.Volume[i] < 0 {
			return fmt.Errorf("bar %d: negative Volume (%f)", i, o.Volume[i])
		}
	}

	return nil
}

// String returns a summary string representation.
func (o *OHLCV) String() string {
	if o.Len() == 0 {
		return "OHLCV(empty)"
	}
	return fmt.Sprintf("OHLCV(%d bars, %s to %s)",
		o.Len(),
		o.Time[0].Format("2006-01-02"),
		o.Time[o.Len()-1].Format("2006-01-02"))
}
