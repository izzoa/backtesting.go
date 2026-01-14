package stats

import (
	"math"
	"testing"
	"time"
)

// almostEqual checks if two floats are approximately equal.
func almostEqual(a, b, tolerance float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsInf(a, 1) && math.IsInf(b, 1) {
		return true
	}
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return true
	}
	return math.Abs(a-b) <= tolerance
}

func TestCalcReturn(t *testing.T) {
	tests := []struct {
		name     string
		start    float64
		end      float64
		expected float64
	}{
		{"positive return", 100, 150, 50},
		{"negative return", 100, 80, -20},
		{"no change", 100, 100, 0},
		{"double", 100, 200, 100},
		{"zero start", 0, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcReturn(tt.start, tt.end)
			if !almostEqual(result, tt.expected, 0.01) {
				t.Errorf("CalcReturn(%f, %f) = %f, want %f", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestCalcAnnualizedReturn(t *testing.T) {
	tests := []struct {
		name      string
		returnPct float64
		days      int
		expected  float64
	}{
		{"one year 10%", 10, 252, 10},
		{"two years 21%", 21, 504, 10}, // (1.21)^0.5 - 1 ≈ 10%
		{"half year 5%", 5, 126, 10.25}, // (1.05)^2 - 1 ≈ 10.25%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcAnnualizedReturn(tt.returnPct, tt.days)
			if !almostEqual(result, tt.expected, 0.5) {
				t.Errorf("CalcAnnualizedReturn(%f, %d) = %f, want ~%f", tt.returnPct, tt.days, result, tt.expected)
			}
		})
	}
}

func TestCalcVolatility(t *testing.T) {
	// Test with known daily returns
	// If daily stddev is 0.01 (1%), annualized should be ~15.87%
	returns := make([]float64, 252)
	for i := range returns {
		if i%2 == 0 {
			returns[i] = 0.01
		} else {
			returns[i] = -0.01
		}
	}

	vol := CalcVolatility(returns)
	expectedVol := 15.87 // approximately

	if !almostEqual(vol, expectedVol, 1.0) {
		t.Errorf("CalcVolatility() = %f, want ~%f", vol, expectedVol)
	}
}

func TestCalcSharpeRatio(t *testing.T) {
	// Generate returns with positive mean
	returns := make([]float64, 252)
	for i := range returns {
		returns[i] = 0.001 // 0.1% daily return
	}

	sharpe := CalcSharpeRatio(returns, 0.03)

	// With 0.1% daily returns (~25% annual) and 3% risk-free,
	// and very low volatility (constant returns), Sharpe should be very high
	// Since all returns are identical, stddev is 0, so this should handle that case
	if math.IsNaN(sharpe) || math.IsInf(sharpe, 0) {
		// This is expected for constant returns
		t.Log("Sharpe ratio is NaN/Inf for constant returns (expected)")
	}

	// Test with varying returns
	varyingReturns := []float64{0.01, -0.005, 0.015, -0.01, 0.02, -0.005}
	sharpe2 := CalcSharpeRatio(varyingReturns, 0.03)

	if math.IsNaN(sharpe2) {
		t.Errorf("CalcSharpeRatio with varying returns should not be NaN")
	}
}

func TestCalcWinRate(t *testing.T) {
	tests := []struct {
		name     string
		pls      []float64
		expected float64
	}{
		{"all wins", []float64{100, 50, 25}, 100},
		{"all losses", []float64{-100, -50, -25}, 0},
		{"half and half", []float64{100, -100, 50, -50}, 50},
		{"mostly wins", []float64{100, 50, 25, -10}, 75},
		{"empty", []float64{}, 0},
		{"breakeven is win", []float64{0, -10, 10}, 66.67},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcWinRate(tt.pls)
			if !almostEqual(result, tt.expected, 0.1) {
				t.Errorf("CalcWinRate(%v) = %f, want %f", tt.pls, result, tt.expected)
			}
		})
	}
}

func TestCalcProfitFactor(t *testing.T) {
	tests := []struct {
		name     string
		pls      []float64
		expected float64
	}{
		{"gross profit > loss", []float64{100, -50}, 2.0},
		{"gross profit = loss", []float64{100, -100}, 1.0},
		{"gross profit < loss", []float64{50, -100}, 0.5},
		{"no losses", []float64{100, 50}, math.Inf(1)},
		{"no profits", []float64{-100, -50}, 0},
		{"empty", []float64{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcProfitFactor(tt.pls)
			if !almostEqual(result, tt.expected, 0.01) {
				t.Errorf("CalcProfitFactor(%v) = %f, want %f", tt.pls, result, tt.expected)
			}
		})
	}
}

func TestCalcSQN(t *testing.T) {
	// SQN = sqrt(N) * mean / stddev
	// With N=100, mean=1, stddev=1 -> SQN = 10
	pcts := make([]float64, 100)
	for i := range pcts {
		if i%2 == 0 {
			pcts[i] = 2 // Creates mean ~1 with stddev ~1
		} else {
			pcts[i] = 0
		}
	}

	sqn := CalcSQN(pcts)

	// Mean = 1, StdDev = 1, N = 100
	// Expected SQN = sqrt(100) * 1 / 1 = 10
	if !almostEqual(sqn, 10.0, 0.5) {
		t.Errorf("CalcSQN() = %f, want ~10", sqn)
	}
}

func TestCalcMaxDrawdown(t *testing.T) {
	tests := []struct {
		name         string
		equity       []float64
		expectedPct  float64
		expectedVal  float64
	}{
		{"no drawdown", []float64{100, 110, 120}, 0, 0},
		{"simple drawdown", []float64{100, 90, 80, 90}, 20, 20},
		{"recovery", []float64{100, 80, 100}, 20, 20},
		{"multiple drawdowns", []float64{100, 90, 100, 85, 100}, 15, 15},
		{"deeper second", []float64{100, 95, 100, 70, 100}, 30, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pct, val := CalcMaxDrawdown(tt.equity)
			if !almostEqual(pct, tt.expectedPct, 0.01) {
				t.Errorf("CalcMaxDrawdown pct = %f, want %f", pct, tt.expectedPct)
			}
			if !almostEqual(val, tt.expectedVal, 0.01) {
				t.Errorf("CalcMaxDrawdown val = %f, want %f", val, tt.expectedVal)
			}
		})
	}
}

func TestCalcDrawdowns(t *testing.T) {
	equity := []float64{100, 110, 100, 90, 100, 110, 105}
	times := make([]time.Time, len(equity))
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range times {
		times[i] = base.Add(time.Duration(i) * 24 * time.Hour)
	}

	drawdowns := CalcDrawdowns(equity, times)

	if len(drawdowns) != 2 {
		t.Errorf("Expected 2 drawdowns, got %d", len(drawdowns))
	}

	// First drawdown: 110 -> 90 (18.18%)
	if len(drawdowns) > 0 {
		dd := drawdowns[0]
		expectedPct := (110.0 - 90.0) / 110.0 * 100
		if !almostEqual(dd.DrawdownPct, expectedPct, 0.1) {
			t.Errorf("First drawdown pct = %f, want %f", dd.DrawdownPct, expectedPct)
		}
		if dd.EndBar != 5 { // Recovery at bar 5 (110 again)
			t.Errorf("First drawdown EndBar = %d, want 5", dd.EndBar)
		}
	}

	// Second drawdown: 110 -> 105 (ongoing)
	if len(drawdowns) > 1 {
		dd := drawdowns[1]
		expectedPct := (110.0 - 105.0) / 110.0 * 100
		if !almostEqual(dd.DrawdownPct, expectedPct, 0.1) {
			t.Errorf("Second drawdown pct = %f, want %f", dd.DrawdownPct, expectedPct)
		}
		if dd.EndBar != -1 { // Not recovered
			t.Errorf("Second drawdown should not be recovered, EndBar = %d", dd.EndBar)
		}
	}
}

func TestEquityCurve(t *testing.T) {
	equity := []float64{100, 110, 105, 115, 100}
	times := make([]time.Time, len(equity))
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range times {
		times[i] = base.Add(time.Duration(i) * 24 * time.Hour)
	}

	ec := NewEquityCurve(times, equity)

	if ec.Len() != 5 {
		t.Errorf("Len() = %d, want 5", ec.Len())
	}

	if ec.Initial() != 100 {
		t.Errorf("Initial() = %f, want 100", ec.Initial())
	}

	if ec.Final() != 100 {
		t.Errorf("Final() = %f, want 100", ec.Final())
	}

	if ec.Peak() != 115 {
		t.Errorf("Peak() = %f, want 115", ec.Peak())
	}

	maxDD, _ := ec.MaxDrawdown()
	expectedDD := (115.0 - 100.0) / 115.0 * 100
	if !almostEqual(maxDD, expectedDD, 0.1) {
		t.Errorf("MaxDrawdown() = %f, want %f", maxDD, expectedDD)
	}
}

func TestCalcTradeStats(t *testing.T) {
	trades := []*TradeRecord{
		{PL: 100, PLPct: 10, Duration: time.Hour, BarsHeld: 1},
		{PL: -50, PLPct: -5, Duration: 2 * time.Hour, BarsHeld: 2},
		{PL: 75, PLPct: 7.5, Duration: 3 * time.Hour, BarsHeld: 3},
		{PL: -25, PLPct: -2.5, Duration: time.Hour, BarsHeld: 1},
	}

	stats := CalcTradeStats(trades)

	if stats.TotalTrades != 4 {
		t.Errorf("TotalTrades = %d, want 4", stats.TotalTrades)
	}

	if stats.WinningTrades != 2 {
		t.Errorf("WinningTrades = %d, want 2", stats.WinningTrades)
	}

	if stats.LosingTrades != 2 {
		t.Errorf("LosingTrades = %d, want 2", stats.LosingTrades)
	}

	if !almostEqual(stats.WinRate, 50.0, 0.01) {
		t.Errorf("WinRate = %f, want 50", stats.WinRate)
	}

	expectedTotalPL := 100.0 - 50.0 + 75.0 - 25.0
	if !almostEqual(stats.TotalPL, expectedTotalPL, 0.01) {
		t.Errorf("TotalPL = %f, want %f", stats.TotalPL, expectedTotalPL)
	}

	if !almostEqual(stats.GrossProfit, 175.0, 0.01) {
		t.Errorf("GrossProfit = %f, want 175", stats.GrossProfit)
	}

	if !almostEqual(stats.GrossLoss, 75.0, 0.01) {
		t.Errorf("GrossLoss = %f, want 75", stats.GrossLoss)
	}

	// Profit factor = 175 / 75 = 2.33
	if !almostEqual(stats.ProfitFactor, 175.0/75.0, 0.01) {
		t.Errorf("ProfitFactor = %f, want %f", stats.ProfitFactor, 175.0/75.0)
	}

	if stats.MaxDuration != 3*time.Hour {
		t.Errorf("MaxDuration = %v, want 3h", stats.MaxDuration)
	}
}

func TestCompute(t *testing.T) {
	equity := []float64{10000, 10500, 10200, 11000, 10800, 11500}
	times := make([]time.Time, len(equity))
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range times {
		times[i] = base.Add(time.Duration(i) * 24 * time.Hour)
	}

	trades := []*TradeRecord{
		{PL: 500, PLPct: 5, Duration: time.Hour, BarsHeld: 1},
		{PL: -300, PLPct: -3, Duration: time.Hour, BarsHeld: 1},
		{PL: 800, PLPct: 8, Duration: time.Hour, BarsHeld: 1},
		{PL: -200, PLPct: -2, Duration: time.Hour, BarsHeld: 1},
		{PL: 700, PLPct: 7, Duration: time.Hour, BarsHeld: 1},
	}

	stats := Compute(ComputeConfig{
		Trades:       trades,
		EquityCurve:  equity,
		Times:        times,
		InitialCash:  10000,
		RiskFreeRate: 0.03,
		TotalBars:    10,
	})

	// Check basic values
	if stats.NumTrades != 5 {
		t.Errorf("NumTrades = %d, want 5", stats.NumTrades)
	}

	expectedReturn := (11500.0 - 10000.0) / 10000.0 * 100
	if !almostEqual(stats.ReturnPct, expectedReturn, 0.01) {
		t.Errorf("ReturnPct = %f, want %f", stats.ReturnPct, expectedReturn)
	}

	// Win rate should be 60% (3 wins, 2 losses)
	if !almostEqual(stats.WinRatePct, 60.0, 0.1) {
		t.Errorf("WinRatePct = %f, want 60", stats.WinRatePct)
	}

	// Exposure time = 5 bars held / 10 total bars = 50%
	if !almostEqual(stats.ExposureTimePct, 50.0, 0.1) {
		t.Errorf("ExposureTimePct = %f, want 50", stats.ExposureTimePct)
	}

	// Max drawdown should be calculated
	if stats.MaxDrawdownPct <= 0 {
		t.Errorf("MaxDrawdownPct should be > 0, got %f", stats.MaxDrawdownPct)
	}
}
