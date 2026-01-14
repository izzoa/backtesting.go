package backtesting

import (
	"testing"

	"github.com/quickfixgo/backtesting/data"
)

// SmaCrossStrategy implements a simple SMA crossover strategy.
type SmaCrossStrategy struct {
	StrategyBase
	FastPeriod int
	SlowPeriod int
	fastSMA    *data.Indicator
	slowSMA    *data.Indicator
}

func (s *SmaCrossStrategy) Init() {
	// Calculate SMAs on the full data
	s.fastSMA = s.I("SMA_fast", data.SMA(s.Close(), s.FastPeriod))
	s.slowSMA = s.I("SMA_slow", data.SMA(s.Close(), s.SlowPeriod))
}

func (s *SmaCrossStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	// Get current and previous SMA values
	fastCurr := s.IndicatorAt(s.fastSMA, idx)
	fastPrev := s.IndicatorAt(s.fastSMA, idx-1)
	slowCurr := s.IndicatorAt(s.slowSMA, idx)
	slowPrev := s.IndicatorAt(s.slowSMA, idx-1)

	// Check for crossover (fast crosses above slow)
	if fastPrev <= slowPrev && fastCurr > slowCurr {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// Check for crossunder (fast crosses below slow)
	if fastPrev >= slowPrev && fastCurr < slowCurr {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestBacktest_SmaCross_GOOG(t *testing.T) {
	// Load test data
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("failed to load data: %v", err)
	}

	strategy := &SmaCrossStrategy{
		FastPeriod: 10,
		SlowPeriod: 20,
	}

	bt := NewBacktest(ohlcv, strategy, BacktestConfig{
		Cash: 10000,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	// Log results
	t.Logf("Results:\n%s", results.String())

	// Basic sanity checks
	if results.NumTrades == 0 {
		t.Error("expected some trades")
	}
	if results.StartTime.IsZero() {
		t.Error("start time not set")
	}
	if results.EndTime.IsZero() {
		t.Error("end time not set")
	}
	if results.FinalEquity <= 0 {
		t.Error("final equity should be positive")
	}
}

func TestBacktest_NoData(t *testing.T) {
	strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
	bt := NewBacktest(nil, strategy, BacktestConfig{Cash: 10000})

	_, err := bt.Run()
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestBacktest_EmptyData(t *testing.T) {
	strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
	bt := NewBacktest(data.NewOHLCV(0), strategy, BacktestConfig{Cash: 10000})

	_, err := bt.Run()
	if err == nil {
		t.Error("expected error for empty data")
	}
}

// NeverTradeStrategy is a strategy that never trades.
type NeverTradeStrategy struct {
	StrategyBase
}

func (s *NeverTradeStrategy) Init() {}
func (s *NeverTradeStrategy) Next() {}

func TestBacktest_NoTrades(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("failed to load data: %v", err)
	}

	strategy := &NeverTradeStrategy{}
	bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	if results.NumTrades != 0 {
		t.Errorf("NumTrades = %d, want 0", results.NumTrades)
	}
	if results.ReturnPct != 0 {
		t.Errorf("ReturnPct = %f, want 0", results.ReturnPct)
	}
	if results.FinalEquity != 10000 {
		t.Errorf("FinalEquity = %f, want 10000", results.FinalEquity)
	}
}

// BuyAndHoldStrategy buys on the first bar and holds.
type BuyAndHoldStrategy struct {
	StrategyBase
	bought bool
}

func (s *BuyAndHoldStrategy) Init() {}

func (s *BuyAndHoldStrategy) Next() {
	if !s.bought && !s.Position().IsLong() {
		s.Buy()
		s.bought = true
	}
}

func TestBacktest_BuyAndHold(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("failed to load data: %v", err)
	}

	strategy := &BuyAndHoldStrategy{}
	bt := NewBacktest(ohlcv, strategy, BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	t.Logf("Buy and Hold Results:\n%s", results.String())

	if results.NumTrades != 1 {
		t.Errorf("NumTrades = %d, want 1", results.NumTrades)
	}

	// GOOG went up significantly from 2004 to 2013, so we expect positive returns
	if results.ReturnPct <= 0 {
		t.Error("expected positive returns for GOOG buy and hold")
	}
}

func TestBacktest_Commission(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("failed to load data: %v", err)
	}

	// Test with 1% commission
	strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
	bt := NewBacktest(ohlcv, strategy, BacktestConfig{
		Cash:       10000,
		Commission: func(size, price float64) float64 { return absFloat(size) * price * 0.01 },
	})

	resultsWithComm, err := bt.Run()
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	// Test without commission
	strategy2 := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
	bt2 := NewBacktest(ohlcv, strategy2, BacktestConfig{Cash: 10000})

	resultsNoComm, err := bt2.Run()
	if err != nil {
		t.Fatalf("backtest failed: %v", err)
	}

	// Results with commission should be worse
	if resultsWithComm.ReturnPct >= resultsNoComm.ReturnPct {
		t.Error("commission should reduce returns")
	}

	t.Logf("With commission: %.2f%%, Without: %.2f%%",
		resultsWithComm.ReturnPct, resultsNoComm.ReturnPct)
}

func BenchmarkBacktest_SmaCross(b *testing.B) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		b.Fatalf("failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
		bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
		_, _ = bt.Run()
	}
}
