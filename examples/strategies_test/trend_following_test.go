package strategies_test

import (
	"math"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
	"github.com/izzoa/backtesting.go/lib"
)

// =============================================================================
// Strategy 1: SMA Crossover
// =============================================================================

// SMACrossoverStrategy implements a simple SMA crossover strategy.
type SMACrossoverStrategy struct {
	backtesting.StrategyBase
	FastPeriod int
	SlowPeriod int
	fastSMA    *data.Indicator
	slowSMA    *data.Indicator
}

func (s *SMACrossoverStrategy) Init() {
	s.fastSMA = s.I("SMA_fast", data.SMA(s.Close(), s.FastPeriod))
	s.slowSMA = s.I("SMA_slow", data.SMA(s.Close(), s.SlowPeriod))
}

func (s *SMACrossoverStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	fastCurr := s.IndicatorAt(s.fastSMA, idx)
	fastPrev := s.IndicatorAt(s.fastSMA, idx-1)
	slowCurr := s.IndicatorAt(s.slowSMA, idx)
	slowPrev := s.IndicatorAt(s.slowSMA, idx-1)

	if math.IsNaN(fastCurr) || math.IsNaN(slowCurr) {
		return
	}

	// Fast crosses above slow - buy signal
	if fastPrev <= slowPrev && fastCurr > slowCurr {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// Fast crosses below slow - sell signal
	if fastPrev >= slowPrev && fastCurr < slowCurr {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestSMACrossover_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &SMACrossoverStrategy{FastPeriod: 10, SlowPeriod: 20}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "SMACrossover_GOOG")
	assertHasTrades(t, results, "SMACrossover_GOOG")

	t.Logf("SMACrossover GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestSMACrossover_EURUSD(t *testing.T) {
	ohlcv := loadTestData(t, "EURUSD.csv")
	strategy := &SMACrossoverStrategy{FastPeriod: 10, SlowPeriod: 20}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "SMACrossover_EURUSD")

	t.Logf("SMACrossover EURUSD: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestSMACrossover_MultiplePeriods(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	testCases := []struct {
		fast int
		slow int
	}{
		{5, 10},
		{10, 20},
		{20, 50},
		{10, 30},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			strategy := &SMACrossoverStrategy{FastPeriod: tc.fast, SlowPeriod: tc.slow}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "SMACrossover")

			t.Logf("Periods (%d,%d): Return=%.2f%%, Trades=%d",
				tc.fast, tc.slow, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 2: EMA Crossover
// =============================================================================

// EMACrossoverStrategy implements an EMA crossover strategy.
type EMACrossoverStrategy struct {
	backtesting.StrategyBase
	FastPeriod int
	SlowPeriod int
	fastEMA    *data.Indicator
	slowEMA    *data.Indicator
}

func (s *EMACrossoverStrategy) Init() {
	s.fastEMA = s.I("EMA_fast", data.EMA(s.Close(), s.FastPeriod))
	s.slowEMA = s.I("EMA_slow", data.EMA(s.Close(), s.SlowPeriod))
}

func (s *EMACrossoverStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	fastCurr := s.IndicatorAt(s.fastEMA, idx)
	fastPrev := s.IndicatorAt(s.fastEMA, idx-1)
	slowCurr := s.IndicatorAt(s.slowEMA, idx)
	slowPrev := s.IndicatorAt(s.slowEMA, idx-1)

	if math.IsNaN(fastCurr) || math.IsNaN(slowCurr) {
		return
	}

	// Fast crosses above slow - buy signal
	if fastPrev <= slowPrev && fastCurr > slowCurr {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// Fast crosses below slow - sell signal
	if fastPrev >= slowPrev && fastCurr < slowCurr {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestEMACrossover_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &EMACrossoverStrategy{FastPeriod: 12, SlowPeriod: 26}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "EMACrossover_GOOG")
	assertHasTrades(t, results, "EMACrossover_GOOG")

	t.Logf("EMACrossover GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestEMACrossover_EURUSD(t *testing.T) {
	ohlcv := loadTestData(t, "EURUSD.csv")
	strategy := &EMACrossoverStrategy{FastPeriod: 12, SlowPeriod: 26}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "EMACrossover_EURUSD")

	t.Logf("EMACrossover EURUSD: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

// =============================================================================
// Strategy 3: Triple Moving Average
// =============================================================================

// TripleMAStrategy implements a triple moving average strategy.
type TripleMAStrategy struct {
	backtesting.StrategyBase
	FastPeriod   int
	MediumPeriod int
	SlowPeriod   int
	fastSMA      *data.Indicator
	mediumSMA    *data.Indicator
	slowSMA      *data.Indicator
}

func (s *TripleMAStrategy) Init() {
	s.fastSMA = s.I("SMA_fast", data.SMA(s.Close(), s.FastPeriod))
	s.mediumSMA = s.I("SMA_medium", data.SMA(s.Close(), s.MediumPeriod))
	s.slowSMA = s.I("SMA_slow", data.SMA(s.Close(), s.SlowPeriod))
}

func (s *TripleMAStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	fast := s.IndicatorAt(s.fastSMA, idx)
	medium := s.IndicatorAt(s.mediumSMA, idx)
	slow := s.IndicatorAt(s.slowSMA, idx)

	if math.IsNaN(fast) || math.IsNaN(medium) || math.IsNaN(slow) {
		return
	}

	// Buy when fast > medium > slow (bullish alignment)
	if fast > medium && medium > slow {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// Sell when alignment breaks
	if fast < medium || medium < slow {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestTripleMA_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &TripleMAStrategy{FastPeriod: 10, MediumPeriod: 20, SlowPeriod: 50}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "TripleMA_GOOG")

	t.Logf("TripleMA GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestTripleMA_Alignment(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Test different period alignments
	configs := []struct {
		fast, medium, slow int
	}{
		{5, 10, 20},
		{10, 20, 50},
		{20, 50, 100},
	}

	for _, cfg := range configs {
		t.Run("", func(t *testing.T) {
			strategy := &TripleMAStrategy{
				FastPeriod:   cfg.fast,
				MediumPeriod: cfg.medium,
				SlowPeriod:   cfg.slow,
			}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "TripleMA")

			t.Logf("TripleMA (%d/%d/%d): Return=%.2f%%, Trades=%d",
				cfg.fast, cfg.medium, cfg.slow, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 4: MACD Crossover
// =============================================================================

// MACDCrossoverStrategy implements MACD line/signal crossover.
type MACDCrossoverStrategy struct {
	backtesting.StrategyBase
	FastPeriod   int
	SlowPeriod   int
	SignalPeriod int
	macdLine     *data.Indicator
	signalLine   *data.Indicator
}

func (s *MACDCrossoverStrategy) Init() {
	macd, signal, _ := lib.MACD(s.Close(), s.FastPeriod, s.SlowPeriod, s.SignalPeriod)
	s.macdLine = s.I("MACD", macd)
	s.signalLine = s.I("Signal", signal)
}

func (s *MACDCrossoverStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	macdCurr := s.IndicatorAt(s.macdLine, idx)
	macdPrev := s.IndicatorAt(s.macdLine, idx-1)
	signalCurr := s.IndicatorAt(s.signalLine, idx)
	signalPrev := s.IndicatorAt(s.signalLine, idx-1)

	if math.IsNaN(macdCurr) || math.IsNaN(signalCurr) {
		return
	}

	// MACD crosses above signal - buy
	if macdPrev <= signalPrev && macdCurr > signalCurr {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// MACD crosses below signal - sell
	if macdPrev >= signalPrev && macdCurr < signalCurr {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestMACDCrossover_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &MACDCrossoverStrategy{FastPeriod: 12, SlowPeriod: 26, SignalPeriod: 9}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "MACDCrossover_GOOG")
	assertHasTrades(t, results, "MACDCrossover_GOOG")

	t.Logf("MACDCrossover GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestMACDCrossover_DefaultParams(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Standard MACD parameters (12, 26, 9)
	strategy := &MACDCrossoverStrategy{FastPeriod: 12, SlowPeriod: 26, SignalPeriod: 9}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "MACDCrossover_Default")

	t.Logf("MACD (12,26,9): Return=%.2f%%, Trades=%d", results.ReturnPct, results.NumTrades)
}

// =============================================================================
// Strategy 5: MACD Histogram
// =============================================================================

// MACDHistogramStrategy trades based on MACD histogram.
type MACDHistogramStrategy struct {
	backtesting.StrategyBase
	FastPeriod   int
	SlowPeriod   int
	SignalPeriod int
	histogram    *data.Indicator
}

func (s *MACDHistogramStrategy) Init() {
	_, _, hist := lib.MACD(s.Close(), s.FastPeriod, s.SlowPeriod, s.SignalPeriod)
	s.histogram = s.I("Histogram", hist)
}

func (s *MACDHistogramStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	histCurr := s.IndicatorAt(s.histogram, idx)
	histPrev := s.IndicatorAt(s.histogram, idx-1)

	if math.IsNaN(histCurr) {
		return
	}

	// Histogram turns positive - buy
	if histPrev <= 0 && histCurr > 0 {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// Histogram turns negative - sell
	if histPrev >= 0 && histCurr < 0 {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestMACDHistogram_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &MACDHistogramStrategy{FastPeriod: 12, SlowPeriod: 26, SignalPeriod: 9}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "MACDHistogram_GOOG")
	assertHasTrades(t, results, "MACDHistogram_GOOG")

	t.Logf("MACDHistogram GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

// =============================================================================
// Strategy 6: ADX Trend
// =============================================================================

// ADXTrendStrategy uses ADX to filter trend trades.
type ADXTrendStrategy struct {
	backtesting.StrategyBase
	ADXPeriod    int
	SMAPeriod    int
	ADXThreshold float64
	adx          *data.Indicator
	sma          *data.Indicator
}

func (s *ADXTrendStrategy) Init() {
	s.adx = s.I("ADX", lib.ADX(s.High(), s.Low(), s.Close(), s.ADXPeriod))
	s.sma = s.I("SMA", data.SMA(s.Close(), s.SMAPeriod))
}

func (s *ADXTrendStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	adxValue := s.IndicatorAt(s.adx, idx)
	smaValue := s.IndicatorAt(s.sma, idx)
	price := s.Close()[idx]

	if math.IsNaN(adxValue) || math.IsNaN(smaValue) {
		return
	}

	// Strong trend (ADX > threshold) and price above SMA - buy
	if adxValue > s.ADXThreshold && price > smaValue {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// Weak trend or price below SMA - sell
	if adxValue < s.ADXThreshold || price < smaValue {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestADXTrend_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &ADXTrendStrategy{ADXPeriod: 14, SMAPeriod: 20, ADXThreshold: 25}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "ADXTrend_GOOG")

	t.Logf("ADXTrend GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestADXTrend_DifferentThresholds(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	thresholds := []float64{20, 25, 30, 35}

	for _, threshold := range thresholds {
		t.Run("", func(t *testing.T) {
			strategy := &ADXTrendStrategy{ADXPeriod: 14, SMAPeriod: 20, ADXThreshold: threshold}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "ADXTrend")

			t.Logf("ADX Threshold %.0f: Return=%.2f%%, Trades=%d",
				threshold, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkSMACrossover(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &SMACrossoverStrategy{FastPeriod: 10, SlowPeriod: 20}
	})
}

func BenchmarkEMACrossover(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &EMACrossoverStrategy{FastPeriod: 12, SlowPeriod: 26}
	})
}

func BenchmarkTripleMA(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &TripleMAStrategy{FastPeriod: 10, MediumPeriod: 20, SlowPeriod: 50}
	})
}

func BenchmarkMACDCrossover(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &MACDCrossoverStrategy{FastPeriod: 12, SlowPeriod: 26, SignalPeriod: 9}
	})
}

func BenchmarkMACDHistogram(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &MACDHistogramStrategy{FastPeriod: 12, SlowPeriod: 26, SignalPeriod: 9}
	})
}

func BenchmarkADXTrend(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &ADXTrendStrategy{ADXPeriod: 14, SMAPeriod: 20, ADXThreshold: 25}
	})
}
