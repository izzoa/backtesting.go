package data

import (
	"fmt"
	"time"
)

// DataView provides a partial view of OHLCV data.
// It simulates the gradual revelation of data during a backtest.
type DataView struct {
	source *OHLCV
	length int // Current visible length (grows during backtest)
}

// NewDataView creates a new DataView for the given OHLCV data.
// Initially, the view shows all data. Use SetLength to simulate backtest progression.
func NewDataView(source *OHLCV) *DataView {
	return &DataView{
		source: source,
		length: source.Len(),
	}
}

// SetLength sets the number of visible bars.
// Used during backtesting to simulate data revelation.
func (v *DataView) SetLength(n int) {
	if n < 0 {
		panic(fmt.Sprintf("length %d cannot be negative", n))
	}
	if n > v.source.Len() {
		panic(fmt.Sprintf("length %d exceeds source length %d", n, v.source.Len()))
	}
	v.length = n
}

// Len returns the current visible length.
func (v *DataView) Len() int {
	return v.length
}

// FullLen returns the full source length.
func (v *DataView) FullLen() int {
	return v.source.Len()
}

// At returns the bar at the given index.
// Panics if index is out of the visible range.
func (v *DataView) At(i int) Bar {
	if i < 0 || i >= v.length {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, v.length))
	}
	return v.source.At(i)
}

// Time returns the visible time slice.
func (v *DataView) Time() []time.Time {
	return v.source.Time[:v.length]
}

// Open returns the visible open prices.
func (v *DataView) Open() []float64 {
	return v.source.Open[:v.length]
}

// High returns the visible high prices.
func (v *DataView) High() []float64 {
	return v.source.High[:v.length]
}

// Low returns the visible low prices.
func (v *DataView) Low() []float64 {
	return v.source.Low[:v.length]
}

// Close returns the visible close prices.
func (v *DataView) Close() []float64 {
	return v.source.Close[:v.length]
}

// Volume returns the visible volume.
func (v *DataView) Volume() []float64 {
	return v.source.Volume[:v.length]
}

// LastTime returns the time of the last visible bar.
func (v *DataView) LastTime() time.Time {
	if v.length == 0 {
		return time.Time{}
	}
	return v.source.Time[v.length-1]
}

// LastOpen returns the open of the last visible bar.
func (v *DataView) LastOpen() float64 {
	if v.length == 0 {
		return 0
	}
	return v.source.Open[v.length-1]
}

// LastHigh returns the high of the last visible bar.
func (v *DataView) LastHigh() float64 {
	if v.length == 0 {
		return 0
	}
	return v.source.High[v.length-1]
}

// LastLow returns the low of the last visible bar.
func (v *DataView) LastLow() float64 {
	if v.length == 0 {
		return 0
	}
	return v.source.Low[v.length-1]
}

// LastClose returns the close of the last visible bar.
func (v *DataView) LastClose() float64 {
	if v.length == 0 {
		return 0
	}
	return v.source.Close[v.length-1]
}

// LastVolume returns the volume of the last visible bar.
func (v *DataView) LastVolume() float64 {
	if v.length == 0 {
		return 0
	}
	return v.source.Volume[v.length-1]
}

// Index returns the current bar index (length - 1).
func (v *DataView) Index() int {
	return v.length - 1
}

// OpenSeries returns a Series of visible open prices.
func (v *DataView) OpenSeries() *Series {
	return NewSeries("Open", v.Open())
}

// HighSeries returns a Series of visible high prices.
func (v *DataView) HighSeries() *Series {
	return NewSeries("High", v.High())
}

// LowSeries returns a Series of visible low prices.
func (v *DataView) LowSeries() *Series {
	return NewSeries("Low", v.Low())
}

// CloseSeries returns a Series of visible close prices.
func (v *DataView) CloseSeries() *Series {
	return NewSeries("Close", v.Close())
}

// VolumeSeries returns a Series of visible volume.
func (v *DataView) VolumeSeries() *Series {
	return NewSeries("Volume", v.Volume())
}

// OHLCV returns a new OHLCV containing only the visible data.
func (v *DataView) OHLCV() *OHLCV {
	return v.source.Slice(0, v.length)
}

// Source returns the underlying OHLCV data.
func (v *DataView) Source() *OHLCV {
	return v.source
}

// IndicatorView provides a partial view of an indicator.
type IndicatorView struct {
	source *Indicator
	length int
}

// NewIndicatorView creates a new IndicatorView.
func NewIndicatorView(source *Indicator) *IndicatorView {
	return &IndicatorView{
		source: source,
		length: source.Len(),
	}
}

// SetLength sets the visible length.
func (v *IndicatorView) SetLength(n int) {
	if n < 0 {
		panic(fmt.Sprintf("length %d cannot be negative", n))
	}
	if n > v.source.Len() {
		panic(fmt.Sprintf("length %d exceeds source length %d", n, v.source.Len()))
	}
	v.length = n
}

// Len returns the visible length.
func (v *IndicatorView) Len() int {
	return v.length
}

// Values returns the visible indicator values.
func (v *IndicatorView) Values() []float64 {
	return v.source.Values[:v.length]
}

// Last returns the last visible indicator value.
func (v *IndicatorView) Last() float64 {
	if v.length == 0 {
		return 0
	}
	return v.source.Values[v.length-1]
}

// At returns the indicator value at the given index.
func (v *IndicatorView) At(i int) float64 {
	if i < 0 || i >= v.length {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, v.length))
	}
	return v.source.Values[i]
}

// Name returns the indicator name.
func (v *IndicatorView) Name() string {
	return v.source.Name
}

// PlotOptions returns the indicator's plot options.
func (v *IndicatorView) PlotOptions() IndicatorPlotOptions {
	return v.source.PlotOptions
}

// Source returns the underlying indicator.
func (v *IndicatorView) Source() *Indicator {
	return v.source
}
