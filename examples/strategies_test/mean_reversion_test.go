package strategies_test

import (
	"math"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
	"github.com/izzoa/backtesting.go/lib"
)

// =============================================================================
// Strategy 7: RSI Oversold/Overbought
// =============================================================================

// RSIStrategy implements a mean reversion strategy using RSI.
type RSIStrategy struct {
	backtesting.StrategyBase
	Period     int
	Overbought float64
	Oversold   float64
	rsi        *data.Indicator
}

func (s *RSIStrategy) Init() {
	s.rsi = s.I("RSI", lib.RSI(s.Close(), s.Period))
}

func (s *RSIStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	rsiValue := s.IndicatorAt(s.rsi, idx)
	if math.IsNaN(rsiValue) {
		return
	}

	// Oversold - buy signal
	if rsiValue < s.Oversold && !s.Position().IsLong() {
		s.Buy()
	}

	// Overbought - sell signal
	if rsiValue > s.Overbought && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestRSI_Standard(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &RSIStrategy{Period: 14, Overbought: 70, Oversold: 30}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "RSI_Standard")

	t.Logf("RSI (14, 70, 30): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestRSI_Aggressive(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &RSIStrategy{Period: 7, Overbought: 80, Oversold: 20}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "RSI_Aggressive")

	t.Logf("RSI (7, 80, 20): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestRSI_Conservative(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &RSIStrategy{Period: 21, Overbought: 65, Oversold: 35}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "RSI_Conservative")

	t.Logf("RSI (21, 65, 35): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

// =============================================================================
// Strategy 8: RSI Extreme
// =============================================================================

// RSIExtremeStrategy uses extreme RSI levels.
type RSIExtremeStrategy struct {
	backtesting.StrategyBase
	Period     int
	Overbought float64
	Oversold   float64
	rsi        *data.Indicator
}

func (s *RSIExtremeStrategy) Init() {
	s.rsi = s.I("RSI", lib.RSI(s.Close(), s.Period))
}

func (s *RSIExtremeStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	rsiValue := s.IndicatorAt(s.rsi, idx)
	if math.IsNaN(rsiValue) {
		return
	}

	// Extreme oversold - buy signal
	if rsiValue < s.Oversold && !s.Position().IsLong() {
		s.Buy()
	}

	// Extreme overbought - sell signal
	if rsiValue > s.Overbought && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestRSIExtreme_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &RSIExtremeStrategy{Period: 14, Overbought: 90, Oversold: 10}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "RSIExtreme_GOOG")

	t.Logf("RSI Extreme (90/10): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestRSIExtreme_RareSignals(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	// With very extreme levels, we expect fewer trades
	strategy := &RSIExtremeStrategy{Period: 14, Overbought: 95, Oversold: 5}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "RSIExtreme_Rare")

	t.Logf("RSI Extreme (95/5): Return=%.2f%%, Trades=%d (expected few trades)",
		results.ReturnPct, results.NumTrades)
}

// =============================================================================
// Strategy 9: Bollinger Band Mean Reversion
// =============================================================================

// BollingerMeanReversionStrategy trades mean reversion off Bollinger Bands.
type BollingerMeanReversionStrategy struct {
	backtesting.StrategyBase
	Period int
	StdDev float64
	upper  *data.Indicator
	middle *data.Indicator
	lower  *data.Indicator
}

func (s *BollingerMeanReversionStrategy) Init() {
	upper, middle, lower := lib.BollingerBands(s.Close(), s.Period, s.StdDev)
	s.upper = s.I("BB_upper", upper)
	s.middle = s.I("BB_middle", middle)
	s.lower = s.I("BB_lower", lower)
}

func (s *BollingerMeanReversionStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	price := s.Close()[idx]
	upper := s.IndicatorAt(s.upper, idx)
	lower := s.IndicatorAt(s.lower, idx)

	if math.IsNaN(upper) || math.IsNaN(lower) {
		return
	}

	// Price below lower band - buy (oversold)
	if price < lower && !s.Position().IsLong() {
		s.Buy()
	}

	// Price above upper band - sell (overbought)
	if price > upper && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestBollingerMeanReversion_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &BollingerMeanReversionStrategy{Period: 20, StdDev: 2.0}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BollingerMeanReversion_GOOG")

	t.Logf("BB Mean Reversion (20, 2.0): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestBollingerMeanReversion_TightBands(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &BollingerMeanReversionStrategy{Period: 20, StdDev: 1.5}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BollingerMeanReversion_Tight")

	t.Logf("BB Mean Reversion (20, 1.5): Return=%.2f%%, Trades=%d",
		results.ReturnPct, results.NumTrades)
}

func TestBollingerMeanReversion_WideBands(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &BollingerMeanReversionStrategy{Period: 20, StdDev: 2.5}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BollingerMeanReversion_Wide")

	t.Logf("BB Mean Reversion (20, 2.5): Return=%.2f%%, Trades=%d",
		results.ReturnPct, results.NumTrades)
}

// =============================================================================
// Strategy 10: Stochastic Oversold/Overbought
// =============================================================================

// StochasticStrategy implements stochastic oscillator strategy.
type StochasticStrategy struct {
	backtesting.StrategyBase
	KPeriod    int
	DPeriod    int
	Overbought float64
	Oversold   float64
	k          *data.Indicator
	d          *data.Indicator
}

func (s *StochasticStrategy) Init() {
	k, d := lib.Stochastic(s.High(), s.Low(), s.Close(), s.KPeriod, s.DPeriod)
	s.k = s.I("Stoch_K", k)
	s.d = s.I("Stoch_D", d)
}

func (s *StochasticStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	kValue := s.IndicatorAt(s.k, idx)
	if math.IsNaN(kValue) {
		return
	}

	// Oversold - buy signal
	if kValue < s.Oversold && !s.Position().IsLong() {
		s.Buy()
	}

	// Overbought - sell signal
	if kValue > s.Overbought && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestStochastic_Standard(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &StochasticStrategy{KPeriod: 14, DPeriod: 3, Overbought: 80, Oversold: 20}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "Stochastic_Standard")

	t.Logf("Stochastic (14,3,80,20): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestStochastic_WithDLine(t *testing.T) {
	// Test that uses D line for confirmation
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &StochasticDConfirmStrategy{KPeriod: 14, DPeriod: 3, Overbought: 80, Oversold: 20}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "Stochastic_DLine")

	t.Logf("Stochastic with D confirmation: Return=%.2f%%, Trades=%d",
		results.ReturnPct, results.NumTrades)
}

// StochasticDConfirmStrategy uses D line for confirmation.
type StochasticDConfirmStrategy struct {
	backtesting.StrategyBase
	KPeriod    int
	DPeriod    int
	Overbought float64
	Oversold   float64
	k          *data.Indicator
	d          *data.Indicator
}

func (s *StochasticDConfirmStrategy) Init() {
	k, d := lib.Stochastic(s.High(), s.Low(), s.Close(), s.KPeriod, s.DPeriod)
	s.k = s.I("Stoch_K", k)
	s.d = s.I("Stoch_D", d)
}

func (s *StochasticDConfirmStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	kValue := s.IndicatorAt(s.k, idx)
	dValue := s.IndicatorAt(s.d, idx)
	if math.IsNaN(kValue) || math.IsNaN(dValue) {
		return
	}

	// Oversold with K crossing above D - buy signal
	if kValue < s.Oversold && kValue > dValue && !s.Position().IsLong() {
		s.Buy()
	}

	// Overbought with K crossing below D - sell signal
	if kValue > s.Overbought && kValue < dValue && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

// =============================================================================
// Strategy 11: CCI Oversold/Overbought
// =============================================================================

// CCIStrategy implements CCI-based mean reversion.
type CCIStrategy struct {
	backtesting.StrategyBase
	Period     int
	Overbought float64
	Oversold   float64
	cci        *data.Indicator
}

func (s *CCIStrategy) Init() {
	s.cci = s.I("CCI", lib.CCI(s.High(), s.Low(), s.Close(), s.Period))
}

func (s *CCIStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	cciValue := s.IndicatorAt(s.cci, idx)
	if math.IsNaN(cciValue) {
		return
	}

	// Oversold - buy signal
	if cciValue < s.Oversold && !s.Position().IsLong() {
		s.Buy()
	}

	// Overbought - sell signal
	if cciValue > s.Overbought && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestCCI_Standard(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &CCIStrategy{Period: 20, Overbought: 100, Oversold: -100}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "CCI_Standard")

	t.Logf("CCI (20, +/-100): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestCCI_ExtremeLevels(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &CCIStrategy{Period: 20, Overbought: 200, Oversold: -200}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "CCI_Extreme")

	t.Logf("CCI (20, +/-200): Return=%.2f%%, Trades=%d",
		results.ReturnPct, results.NumTrades)
}

// =============================================================================
// Strategy 12: Williams %R
// =============================================================================

// WilliamsRStrategy implements Williams %R strategy.
type WilliamsRStrategy struct {
	backtesting.StrategyBase
	Period     int
	Overbought float64
	Oversold   float64
	willR      *data.Indicator
}

func (s *WilliamsRStrategy) Init() {
	s.willR = s.I("WilliamsR", lib.WilliamsR(s.High(), s.Low(), s.Close(), s.Period))
}

func (s *WilliamsRStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	willRValue := s.IndicatorAt(s.willR, idx)
	if math.IsNaN(willRValue) {
		return
	}

	// Oversold (< -80) - buy signal
	if willRValue < s.Oversold && !s.Position().IsLong() {
		s.Buy()
	}

	// Overbought (> -20) - sell signal
	if willRValue > s.Overbought && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestWilliamsR_Standard(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &WilliamsRStrategy{Period: 14, Overbought: -20, Oversold: -80}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "WilliamsR_Standard")

	t.Logf("Williams %%R (14, -20/-80): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestWilliamsR_EURUSD(t *testing.T) {
	ohlcv := loadTestData(t, "EURUSD.csv")
	strategy := &WilliamsRStrategy{Period: 14, Overbought: -20, Oversold: -80}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "WilliamsR_EURUSD")

	t.Logf("Williams %%R EURUSD: Return=%.2f%%, Trades=%d",
		results.ReturnPct, results.NumTrades)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRSI(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &RSIStrategy{Period: 14, Overbought: 70, Oversold: 30}
	})
}

func BenchmarkBollingerMeanReversion(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &BollingerMeanReversionStrategy{Period: 20, StdDev: 2.0}
	})
}

func BenchmarkStochastic(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &StochasticStrategy{KPeriod: 14, DPeriod: 3, Overbought: 80, Oversold: 20}
	})
}

func BenchmarkCCI(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &CCIStrategy{Period: 20, Overbought: 100, Oversold: -100}
	})
}

func BenchmarkWilliamsR(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &WilliamsRStrategy{Period: 14, Overbought: -20, Oversold: -80}
	})
}
