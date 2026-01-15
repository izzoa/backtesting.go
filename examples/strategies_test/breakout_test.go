package strategies_test

import (
	"math"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
	"github.com/izzoa/backtesting.go/lib"
)

// =============================================================================
// Strategy 13: Bollinger Band Breakout
// =============================================================================

// BollingerBreakoutStrategy trades breakouts from Bollinger Bands.
type BollingerBreakoutStrategy struct {
	backtesting.StrategyBase
	Period int
	StdDev float64
	upper  *data.Indicator
	middle *data.Indicator
	lower  *data.Indicator
}

func (s *BollingerBreakoutStrategy) Init() {
	upper, middle, lower := lib.BollingerBands(s.Close(), s.Period, s.StdDev)
	s.upper = s.I("BB_upper", upper)
	s.middle = s.I("BB_middle", middle)
	s.lower = s.I("BB_lower", lower)
}

func (s *BollingerBreakoutStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	price := s.Close()[idx]
	upper := s.IndicatorAt(s.upper, idx)
	middle := s.IndicatorAt(s.middle, idx)

	if math.IsNaN(upper) || math.IsNaN(middle) {
		return
	}

	// Price breaks above upper band - buy (momentum breakout)
	if price > upper && !s.Position().IsLong() {
		s.Buy()
	}

	// Price drops below middle band - sell
	if price < middle && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestBollingerBreakout_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &BollingerBreakoutStrategy{Period: 20, StdDev: 2.0}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BollingerBreakout_GOOG")

	t.Logf("BB Breakout GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestBollingerBreakout_BTCUSD(t *testing.T) {
	ohlcv := loadTestData(t, "BTCUSD.csv")
	strategy := &BollingerBreakoutStrategy{Period: 20, StdDev: 2.0}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BollingerBreakout_BTCUSD")

	t.Logf("BB Breakout BTCUSD (volatile): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

// =============================================================================
// Strategy 14: ATR Breakout
// =============================================================================

// ATRBreakoutStrategy trades breakouts based on ATR.
type ATRBreakoutStrategy struct {
	backtesting.StrategyBase
	Period     int
	Multiplier float64
	atr        *data.Indicator
}

func (s *ATRBreakoutStrategy) Init() {
	s.atr = s.I("ATR", lib.ATR(s.High(), s.Low(), s.Close(), s.Period))
}

func (s *ATRBreakoutStrategy) Next() {
	idx := s.BarIndex()
	if idx < 2 {
		return
	}

	price := s.Close()[idx]
	prevClose := s.Close()[idx-1]
	atrValue := s.IndicatorAt(s.atr, idx)

	if math.IsNaN(atrValue) {
		return
	}

	breakoutLevel := prevClose + (s.Multiplier * atrValue)
	breakdownLevel := prevClose - (s.Multiplier * atrValue)

	// Price breaks above ATR level - buy
	if price > breakoutLevel && !s.Position().IsLong() {
		s.Buy()
	}

	// Price breaks below ATR level - sell
	if price < breakdownLevel && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestATRBreakout_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &ATRBreakoutStrategy{Period: 14, Multiplier: 2.0}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "ATRBreakout_GOOG")

	t.Logf("ATR Breakout GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestATRBreakout_DifferentMultipliers(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	multipliers := []float64{1.5, 2.0, 2.5, 3.0}

	for _, mult := range multipliers {
		t.Run("", func(t *testing.T) {
			strategy := &ATRBreakoutStrategy{Period: 14, Multiplier: mult}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "ATRBreakout")

			t.Logf("ATR Multiplier %.1f: Return=%.2f%%, Trades=%d",
				mult, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 15: Donchian Channel (High/Low Breakout)
// =============================================================================

// DonchianChannelStrategy implements a Donchian channel breakout.
type DonchianChannelStrategy struct {
	backtesting.StrategyBase
	EntryPeriod int
	ExitPeriod  int
}

func (s *DonchianChannelStrategy) Init() {
	// No indicators needed - we use raw price data
}

func (s *DonchianChannelStrategy) Next() {
	idx := s.BarIndex()
	if idx < s.EntryPeriod {
		return
	}

	price := s.Close()[idx]

	// Calculate N-bar high for entry
	entryHigh := math.Inf(-1)
	for i := idx - s.EntryPeriod; i < idx; i++ {
		if s.High()[i] > entryHigh {
			entryHigh = s.High()[i]
		}
	}

	// Calculate M-bar low for exit
	exitLow := math.Inf(1)
	exitStart := idx - s.ExitPeriod
	if exitStart < 0 {
		exitStart = 0
	}
	for i := exitStart; i < idx; i++ {
		if s.Low()[i] < exitLow {
			exitLow = s.Low()[i]
		}
	}

	// Price breaks N-bar high - buy
	if price > entryHigh && !s.Position().IsLong() {
		s.Buy()
	}

	// Price breaks M-bar low - sell
	if price < exitLow && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestDonchian_20_10(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &DonchianChannelStrategy{EntryPeriod: 20, ExitPeriod: 10}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "Donchian_20_10")

	t.Logf("Donchian (20/10): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestDonchian_55_20(t *testing.T) {
	// Turtle trading style
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &DonchianChannelStrategy{EntryPeriod: 55, ExitPeriod: 20}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "Donchian_55_20")

	t.Logf("Donchian (55/20 - Turtle): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

// =============================================================================
// Strategy 16: Volatility Expansion
// =============================================================================

// VolatilityExpansionStrategy trades volatility expansion.
type VolatilityExpansionStrategy struct {
	backtesting.StrategyBase
	BBPeriod       int
	BBStdDev       float64
	ATRPeriod      int
	ExpansionRatio float64
	upper          *data.Indicator
	lower          *data.Indicator
	atr            *data.Indicator
}

func (s *VolatilityExpansionStrategy) Init() {
	upper, _, lower := lib.BollingerBands(s.Close(), s.BBPeriod, s.BBStdDev)
	s.upper = s.I("BB_upper", upper)
	s.lower = s.I("BB_lower", lower)
	s.atr = s.I("ATR", lib.ATR(s.High(), s.Low(), s.Close(), s.ATRPeriod))
}

func (s *VolatilityExpansionStrategy) Next() {
	idx := s.BarIndex()
	if idx < 2 {
		return
	}

	upper := s.IndicatorAt(s.upper, idx)
	lower := s.IndicatorAt(s.lower, idx)
	upperPrev := s.IndicatorAt(s.upper, idx-1)
	lowerPrev := s.IndicatorAt(s.lower, idx-1)
	atr := s.IndicatorAt(s.atr, idx)

	if math.IsNaN(upper) || math.IsNaN(lower) || math.IsNaN(atr) {
		return
	}

	// Calculate BB width
	bbWidth := upper - lower
	bbWidthPrev := upperPrev - lowerPrev

	if bbWidthPrev == 0 {
		return
	}

	// Volatility expansion ratio
	expansion := bbWidth / bbWidthPrev

	price := s.Close()[idx]

	// Volatility expanding and price above middle - buy
	if expansion > s.ExpansionRatio && price > (upper+lower)/2 && !s.Position().IsLong() {
		s.Buy()
	}

	// Volatility contracting or price falling - sell
	if (expansion < 1/s.ExpansionRatio || price < (upper+lower)/2) && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestVolatilityExpansion_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &VolatilityExpansionStrategy{
		BBPeriod:       20,
		BBStdDev:       2.0,
		ATRPeriod:      14,
		ExpansionRatio: 1.1,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "VolatilityExpansion_GOOG")

	t.Logf("Volatility Expansion GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestVolatilityExpansion_BTCUSD(t *testing.T) {
	ohlcv := loadTestData(t, "BTCUSD.csv")
	strategy := &VolatilityExpansionStrategy{
		BBPeriod:       20,
		BBStdDev:       2.0,
		ATRPeriod:      14,
		ExpansionRatio: 1.1,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "VolatilityExpansion_BTCUSD")

	t.Logf("Volatility Expansion BTCUSD: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkBollingerBreakout(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &BollingerBreakoutStrategy{Period: 20, StdDev: 2.0}
	})
}

func BenchmarkATRBreakout(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &ATRBreakoutStrategy{Period: 14, Multiplier: 2.0}
	})
}

func BenchmarkDonchianChannel(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &DonchianChannelStrategy{EntryPeriod: 20, ExitPeriod: 10}
	})
}

func BenchmarkVolatilityExpansion(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &VolatilityExpansionStrategy{
			BBPeriod:       20,
			BBStdDev:       2.0,
			ATRPeriod:      14,
			ExpansionRatio: 1.1,
		}
	})
}
