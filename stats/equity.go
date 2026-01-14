package stats

import (
	"time"
)

// EquityCurve contains equity data over time with derived series.
type EquityCurve struct {
	Time             []time.Time // Timestamps
	Equity           []float64   // Equity values
	DrawdownPct      []float64   // Drawdown percentage at each point
	DrawdownDuration []int       // Bars in current drawdown at each point
}

// NewEquityCurve creates a new equity curve with derived calculations.
func NewEquityCurve(times []time.Time, equity []float64) *EquityCurve {
	if len(equity) == 0 {
		return &EquityCurve{}
	}

	ec := &EquityCurve{
		Time:             times,
		Equity:           equity,
		DrawdownPct:      CalcDrawdownSeries(equity),
		DrawdownDuration: CalcDrawdownDurationSeries(equity),
	}

	return ec
}

// Len returns the length of the equity curve.
func (ec *EquityCurve) Len() int {
	return len(ec.Equity)
}

// Returns calculates daily returns from the equity curve.
func (ec *EquityCurve) Returns() []float64 {
	return CalcDailyReturns(ec.Equity)
}

// Peak returns the peak equity value.
func (ec *EquityCurve) Peak() float64 {
	if len(ec.Equity) == 0 {
		return 0
	}

	peak := ec.Equity[0]
	for _, eq := range ec.Equity {
		if eq > peak {
			peak = eq
		}
	}
	return peak
}

// Final returns the final equity value.
func (ec *EquityCurve) Final() float64 {
	if len(ec.Equity) == 0 {
		return 0
	}
	return ec.Equity[len(ec.Equity)-1]
}

// Initial returns the initial equity value.
func (ec *EquityCurve) Initial() float64 {
	if len(ec.Equity) == 0 {
		return 0
	}
	return ec.Equity[0]
}

// MaxDrawdown returns the maximum drawdown percentage and value.
func (ec *EquityCurve) MaxDrawdown() (float64, float64) {
	return CalcMaxDrawdown(ec.Equity)
}

// Drawdowns returns all drawdown periods.
func (ec *EquityCurve) Drawdowns() []Drawdown {
	return CalcDrawdowns(ec.Equity, ec.Time)
}
