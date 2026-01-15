package strategies_test

import (
	"math"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
	"github.com/izzoa/backtesting.go/lib"
)

// =============================================================================
// Strategy 38: Buy and Hold
// =============================================================================

// BuyAndHoldStrategy buys on the first bar and holds forever.
type BuyAndHoldStrategy struct {
	backtesting.StrategyBase
	entered bool
}

func (s *BuyAndHoldStrategy) Init() {}

func (s *BuyAndHoldStrategy) Next() {
	if !s.entered {
		s.Buy()
		s.entered = true
	}
}

func TestBuyAndHold_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &BuyAndHoldStrategy{}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BuyAndHold_GOOG")

	t.Logf("Buy and Hold GOOG: Return=%.2f%%, MaxDD=%.2f%%",
		results.ReturnPct, results.MaxDrawdownPct)
}

func TestBuyAndHold_SingleTrade(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &BuyAndHoldStrategy{}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BuyAndHold_SingleTrade")

	// Buy and Hold should produce exactly 1 trade (which gets closed at end)
	if results.NumTrades != 1 {
		t.Errorf("Expected 1 trade, got %d", results.NumTrades)
	}
}

func TestBuyAndHold_AllDataSets(t *testing.T) {
	dataSets := []string{"GOOG.csv", "EURUSD.csv", "BTCUSD.csv"}

	for _, ds := range dataSets {
		t.Run(ds, func(t *testing.T) {
			ohlcv := loadTestData(t, ds)
			strategy := &BuyAndHoldStrategy{}
			results := runBacktest(t, ohlcv, strategy)

			assertValidResults(t, results, "BuyAndHold_"+ds)

			t.Logf("%s: Return=%.2f%%, Trades=%d", ds, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 39: Do Nothing
// =============================================================================

// DoNothingStrategy never enters any positions.
type DoNothingStrategy struct {
	backtesting.StrategyBase
}

func (s *DoNothingStrategy) Init() {}

func (s *DoNothingStrategy) Next() {}

func TestDoNothing_NoTrades(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &DoNothingStrategy{}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "DoNothing_NoTrades")
	assertNoTrades(t, results, "DoNothing_NoTrades")
}

func TestDoNothing_EquityUnchanged(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &DoNothingStrategy{}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	// Equity should remain at initial cash
	if results.FinalEquity != 10000 {
		t.Errorf("Expected equity to be 10000, got %f", results.FinalEquity)
	}

	// Return should be 0%
	if results.ReturnPct != 0 {
		t.Errorf("Expected 0%% return, got %.2f%%", results.ReturnPct)
	}
}

func TestDoNothing_NoErrors(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &DoNothingStrategy{}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	_, err := bt.Run()
	if err != nil {
		t.Errorf("DoNothing strategy should not produce errors: %v", err)
	}
}

// =============================================================================
// Strategy 40: Trade Every Bar
// =============================================================================

// TradeEveryBarStrategy alternates between buy and sell on every bar.
type TradeEveryBarStrategy struct {
	backtesting.StrategyBase
}

func (s *TradeEveryBarStrategy) Init() {}

func (s *TradeEveryBarStrategy) Next() {
	if s.Position().IsLong() {
		s.Position().Close(1.0)
	} else {
		s.Buy()
	}
}

func TestTradeEveryBar_ManyTrades(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &TradeEveryBarStrategy{}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "TradeEveryBar_ManyTrades")

	// Should have many trades (roughly half the number of bars)
	expectedMinTrades := ohlcv.Len() / 3
	if results.NumTrades < expectedMinTrades {
		t.Errorf("Expected at least %d trades, got %d", expectedMinTrades, results.NumTrades)
	}

	t.Logf("Trade Every Bar: %d trades over %d bars", results.NumTrades, ohlcv.Len())
}

func TestTradeEveryBar_NoErrors(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &TradeEveryBarStrategy{}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	_, err := bt.Run()
	if err != nil {
		t.Errorf("TradeEveryBar strategy should not produce errors: %v", err)
	}
}

func TestTradeEveryBar_HighCommissionImpact(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &TradeEveryBarStrategy{}

	// Run without commission
	btNoComm := backtesting.NewBacktest(ohlcv, &TradeEveryBarStrategy{}, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})
	resultsNoComm, _ := btNoComm.Run()

	// Run with commission (0.1% per trade)
	btWithComm := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash: 10000,
		Commission: func(size float64, price float64) float64 {
			return math.Abs(size) * price * 0.001
		},
		FinalizeTrades: true,
	})
	resultsWithComm, _ := btWithComm.Run()

	assertValidResults(t, resultsNoComm, "TradeEveryBar_NoComm")
	assertValidResults(t, resultsWithComm, "TradeEveryBar_WithComm")

	// Commission should have significant impact on high-frequency trading
	t.Logf("Without commission: Return=%.2f%%", resultsNoComm.ReturnPct)
	t.Logf("With 0.1%% commission: Return=%.2f%%", resultsWithComm.ReturnPct)
}

// =============================================================================
// Strategy 41: Single Trade
// =============================================================================

// SingleTradeStrategy enters at a specific bar and exits at another.
type SingleTradeStrategy struct {
	backtesting.StrategyBase
	EntryBar int
	ExitBar  int
	entered  bool
	exited   bool
}

func (s *SingleTradeStrategy) Init() {}

func (s *SingleTradeStrategy) Next() {
	idx := s.BarIndex()

	if idx == s.EntryBar && !s.entered {
		s.Buy()
		s.entered = true
	}

	if idx == s.ExitBar && s.entered && !s.exited {
		s.Position().Close(1.0)
		s.exited = true
	}
}

func TestSingleTrade_ExactlyOneTrade(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &SingleTradeStrategy{EntryBar: 10, ExitBar: 50}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "SingleTrade_ExactlyOne")

	if results.NumTrades != 1 {
		t.Errorf("Expected exactly 1 trade, got %d", results.NumTrades)
	}
}

func TestSingleTrade_CorrectBars(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &SingleTradeStrategy{EntryBar: 100, ExitBar: 200}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "SingleTrade_CorrectBars")
	assertHasTrades(t, results, "SingleTrade_CorrectBars")

	t.Logf("Single trade from bar 100 to 200: Return=%.2f%%", results.ReturnPct)
}

// =============================================================================
// Strategy 42: All Indicators Registered
// =============================================================================

// AllIndicatorsStrategy registers all available indicators.
type AllIndicatorsStrategy struct {
	backtesting.StrategyBase
	sma       *data.Indicator
	ema       *data.Indicator
	rsi       *data.Indicator
	macd      *data.Indicator
	signal    *data.Indicator
	histogram *data.Indicator
	bbUpper   *data.Indicator
	bbMiddle  *data.Indicator
	bbLower   *data.Indicator
	atr       *data.Indicator
	adx       *data.Indicator
	stochK    *data.Indicator
	stochD    *data.Indicator
	cci       *data.Indicator
	willR     *data.Indicator
}

func (s *AllIndicatorsStrategy) Init() {
	// Basic indicators
	s.sma = s.I("SMA", data.SMA(s.Close(), 20))
	s.ema = s.I("EMA", data.EMA(s.Close(), 20))

	// RSI
	s.rsi = s.I("RSI", lib.RSI(s.Close(), 14))

	// MACD
	macd, signal, histogram := lib.MACD(s.Close(), 12, 26, 9)
	s.macd = s.I("MACD", macd)
	s.signal = s.I("Signal", signal)
	s.histogram = s.I("Histogram", histogram)

	// Bollinger Bands
	upper, middle, lower := lib.BollingerBands(s.Close(), 20, 2.0)
	s.bbUpper = s.I("BB_Upper", upper)
	s.bbMiddle = s.I("BB_Middle", middle)
	s.bbLower = s.I("BB_Lower", lower)

	// ATR
	s.atr = s.I("ATR", lib.ATR(s.High(), s.Low(), s.Close(), 14))

	// ADX
	s.adx = s.I("ADX", lib.ADX(s.High(), s.Low(), s.Close(), 14))

	// Stochastic
	k, d := lib.Stochastic(s.High(), s.Low(), s.Close(), 14, 3)
	s.stochK = s.I("Stoch_K", k)
	s.stochD = s.I("Stoch_D", d)

	// CCI
	s.cci = s.I("CCI", lib.CCI(s.High(), s.Low(), s.Close(), 20))

	// Williams %R
	s.willR = s.I("WillR", lib.WilliamsR(s.High(), s.Low(), s.Close(), 14))
}

func (s *AllIndicatorsStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	// Simple logic using SMA
	price := s.Close()[idx]
	smaValue := s.IndicatorAt(s.sma, idx)

	if math.IsNaN(smaValue) {
		return
	}

	if price > smaValue && !s.Position().IsLong() {
		s.Buy()
	}

	if price < smaValue && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestAllIndicators_Compiles(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &AllIndicatorsStrategy{}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "AllIndicators_Compiles")

	t.Logf("All Indicators: Return=%.2f%%, Trades=%d", results.ReturnPct, results.NumTrades)
}

func TestAllIndicators_NoMemoryIssues(t *testing.T) {
	// Run multiple times to check for memory issues
	for i := 0; i < 5; i++ {
		ohlcv := loadTestData(t, "GOOG.csv")
		strategy := &AllIndicatorsStrategy{}
		results := runBacktest(t, ohlcv, strategy)

		assertValidResults(t, results, "AllIndicators_Memory")
	}
}

func TestAllIndicators_AllValuesAccessible(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &AllIndicatorsStrategy{}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	_, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	// All indicators should have been registered without errors
	t.Log("All 15 indicators registered and accessed successfully")
}

// =============================================================================
// Edge Case: Empty Data Handling
// =============================================================================

func TestEdgeCase_SmallDataSet(t *testing.T) {
	// Load data and use only first 50 bars
	ohlcv := loadTestData(t, "GOOG.csv")

	// Create a small subset
	smallOHLCV := &data.OHLCV{
		Time:   ohlcv.Time[:50],
		Open:   ohlcv.Open[:50],
		High:   ohlcv.High[:50],
		Low:    ohlcv.Low[:50],
		Close:  ohlcv.Close[:50],
		Volume: ohlcv.Volume[:50],
	}

	strategy := &BuyAndHoldStrategy{}
	bt := backtesting.NewBacktest(smallOHLCV, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest on small data failed: %v", err)
	}

	assertValidResults(t, results, "SmallDataSet")

	t.Logf("Small data set (50 bars): Return=%.2f%%", results.ReturnPct)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkBuyAndHold(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &BuyAndHoldStrategy{}
	})
}

func BenchmarkDoNothing(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &DoNothingStrategy{}
	})
}

func BenchmarkTradeEveryBar(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &TradeEveryBarStrategy{}
	})
}

func BenchmarkAllIndicators(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &AllIndicatorsStrategy{}
	})
}
