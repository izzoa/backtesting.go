// Package strategies_test provides comprehensive strategy tests for backtesting.go.
package strategies_test

import (
	"math"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
)

// getTestDataPath returns the path to testdata directory
func getTestDataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "testdata", "csv", filename)
}

// loadTestData loads OHLCV data for testing
func loadTestData(t testing.TB, filename string) *data.OHLCV {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	ohlcv, err := data.LoadCSV(getTestDataPath(filename))
	if err != nil {
		t.Fatalf("Failed to load %s: %v", filename, err)
	}
	return ohlcv
}

// runBacktest runs a backtest and returns results
func runBacktest(t testing.TB, ohlcv *data.OHLCV, strategy backtesting.Strategy) *backtesting.Results {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})
	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}
	return results
}

// runBacktestWithConfig runs a backtest with custom config
func runBacktestWithConfig(t testing.TB, ohlcv *data.OHLCV, strategy backtesting.Strategy, cfg backtesting.BacktestConfig) *backtesting.Results {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	bt := backtesting.NewBacktest(ohlcv, strategy, cfg)
	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}
	return results
}

// assertValidResults checks common result validity
func assertValidResults(t testing.TB, results *backtesting.Results, strategyName string) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results == nil {
		t.Fatalf("%s: results is nil", strategyName)
	}
	if results.FinalEquity <= 0 {
		t.Errorf("%s: FinalEquity should be positive, got %f", strategyName, results.FinalEquity)
	}
	if math.IsNaN(results.ReturnPct) {
		t.Errorf("%s: ReturnPct is NaN", strategyName)
	}
	if math.IsNaN(results.MaxDrawdownPct) {
		t.Errorf("%s: MaxDrawdownPct is NaN", strategyName)
	}
	if math.IsNaN(results.WinRate) && results.NumTrades > 0 {
		t.Errorf("%s: WinRate is NaN with %d trades", strategyName, results.NumTrades)
	}
}

// assertHasTrades checks that strategy produced trades
func assertHasTrades(t testing.TB, results *backtesting.Results, strategyName string) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results.NumTrades == 0 {
		t.Errorf("%s: expected trades but got 0", strategyName)
	}
}

// assertNoTrades checks that strategy produced no trades (for edge cases)
func assertNoTrades(t testing.TB, results *backtesting.Results, strategyName string) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results.NumTrades != 0 {
		t.Errorf("%s: expected 0 trades but got %d", strategyName, results.NumTrades)
	}
}

// assertMinTrades checks minimum number of trades
func assertMinTrades(t testing.TB, results *backtesting.Results, strategyName string, minTrades int) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results.NumTrades < minTrades {
		t.Errorf("%s: expected at least %d trades but got %d", strategyName, minTrades, results.NumTrades)
	}
}

// assertMaxTrades checks maximum number of trades
func assertMaxTrades(t testing.TB, results *backtesting.Results, strategyName string, maxTrades int) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results.NumTrades > maxTrades {
		t.Errorf("%s: expected at most %d trades but got %d", strategyName, maxTrades, results.NumTrades)
	}
}

// assertPositiveReturn checks for positive returns
func assertPositiveReturn(t testing.TB, results *backtesting.Results, strategyName string) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results.ReturnPct <= 0 {
		t.Errorf("%s: expected positive return but got %.2f%%", strategyName, results.ReturnPct)
	}
}

// assertWinRateAbove checks win rate is above threshold
func assertWinRateAbove(t testing.TB, results *backtesting.Results, strategyName string, minWinRate float64) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results.NumTrades > 0 && results.WinRate < minWinRate {
		t.Errorf("%s: expected win rate above %.2f%% but got %.2f%%", strategyName, minWinRate, results.WinRate)
	}
}

// assertDrawdownBelow checks max drawdown is below threshold
func assertDrawdownBelow(t testing.TB, results *backtesting.Results, strategyName string, maxDrawdown float64) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if results.MaxDrawdownPct > maxDrawdown {
		t.Errorf("%s: expected max drawdown below %.2f%% but got %.2f%%", strategyName, maxDrawdown, results.MaxDrawdownPct)
	}
}

// testAllDataSets runs a test across all available data sets
func testAllDataSets(t *testing.T, createStrategy func() backtesting.Strategy, strategyName string) {
	t.Helper()
	dataSets := []string{"GOOG.csv", "EURUSD.csv", "BTCUSD.csv"}

	for _, ds := range dataSets {
		t.Run(ds, func(t *testing.T) {
			ohlcv := loadTestData(t, ds)
			strategy := createStrategy()
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, strategyName+"_"+ds)

			t.Logf("%s on %s: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
				strategyName, ds, results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
		})
	}
}

// loadBenchData loads OHLCV data for benchmarking
func loadBenchData(b *testing.B) *data.OHLCV {
	b.Helper()
	ohlcv, err := data.LoadCSV(getTestDataPath("GOOG.csv"))
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}
	return ohlcv
}

// runStrategyBenchmark is a helper for benchmarking strategies
func runStrategyBenchmark(b *testing.B, createStrategy func() backtesting.Strategy) {
	b.Helper()
	ohlcv := loadBenchData(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		strategy := createStrategy()
		bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
			Cash:           10000,
			FinalizeTrades: true,
		})
		_, _ = bt.Run()
	}
}
