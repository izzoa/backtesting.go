package main

import (
	"math"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
)

// getTestDataPath returns the path to a testdata file
func getTestDataPath(filename string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "testdata", "csv", filename)
}

// loadTestData loads OHLCV data for testing
func loadTestData(t *testing.T, filename string) *data.OHLCV {
	t.Helper()
	ohlcv, err := data.LoadCSV(getTestDataPath(filename))
	if err != nil {
		t.Fatalf("Failed to load %s: %v", filename, err)
	}
	return ohlcv
}

// assertValidResults checks common result validity
func assertValidResults(t *testing.T, results *backtesting.Results, name string) {
	t.Helper()
	if results == nil {
		t.Fatalf("%s: results is nil", name)
	}
	if results.FinalEquity <= 0 {
		t.Errorf("%s: FinalEquity should be positive, got %f", name, results.FinalEquity)
	}
	if math.IsNaN(results.ReturnPct) {
		t.Errorf("%s: ReturnPct is NaN", name)
	}
	if math.IsNaN(results.MaxDrawdownPct) {
		t.Errorf("%s: MaxDrawdownPct is NaN", name)
	}
}

func TestSmaCross_Compiles(t *testing.T) {
	// Verify the strategy struct can be created and has correct types
	strategy := &SmaCross{
		FastPeriod: 10,
		SlowPeriod: 20,
	}

	if strategy.FastPeriod != 10 {
		t.Errorf("FastPeriod = %d, want 10", strategy.FastPeriod)
	}
	if strategy.SlowPeriod != 20 {
		t.Errorf("SlowPeriod = %d, want 20", strategy.SlowPeriod)
	}
}

func TestSmaCross_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	strategy := &SmaCross{
		FastPeriod: 10,
		SlowPeriod: 20,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	assertValidResults(t, results, "SmaCross_GOOG")

	if results.NumTrades == 0 {
		t.Error("Expected at least one trade")
	}

	t.Logf("GOOG Results: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestSmaCross_EURUSD(t *testing.T) {
	ohlcv := loadTestData(t, "EURUSD.csv")

	strategy := &SmaCross{
		FastPeriod: 10,
		SlowPeriod: 20,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	assertValidResults(t, results, "SmaCross_EURUSD")

	t.Logf("EURUSD Results: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestSmaCross_BTCUSD(t *testing.T) {
	ohlcv := loadTestData(t, "BTCUSD.csv")

	strategy := &SmaCross{
		FastPeriod: 10,
		SlowPeriod: 20,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	assertValidResults(t, results, "SmaCross_BTCUSD")

	t.Logf("BTCUSD Results: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestSmaCross_DifferentPeriods(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	testCases := []struct {
		name       string
		fastPeriod int
		slowPeriod int
	}{
		{"5_10", 5, 10},
		{"10_20", 10, 20},
		{"20_50", 20, 50},
		{"10_30", 10, 30},
		{"15_45", 15, 45},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := &SmaCross{
				FastPeriod: tc.fastPeriod,
				SlowPeriod: tc.slowPeriod,
			}

			bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
				Cash:           10000,
				FinalizeTrades: true,
			})

			results, err := bt.Run()
			if err != nil {
				t.Fatalf("Backtest failed: %v", err)
			}

			assertValidResults(t, results, "SmaCross_"+tc.name)

			t.Logf("Periods (%d,%d): Return=%.2f%%, Trades=%d",
				tc.fastPeriod, tc.slowPeriod, results.ReturnPct, results.NumTrades)
		})
	}
}

func TestSmaCross_ValidateCrossoverLogic(t *testing.T) {
	// Test that crossover detection works correctly
	ohlcv := loadTestData(t, "GOOG.csv")

	strategy := &SmaCross{
		FastPeriod: 5,
		SlowPeriod: 20,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	// With a fast 5-period SMA vs slow 20-period SMA, we should get trades
	if results.NumTrades == 0 {
		t.Error("Expected trades with 5/20 SMA crossover on GOOG data")
	}

	// Win rate should be between 0 and 100
	if results.NumTrades > 0 {
		if results.WinRate < 0 || results.WinRate > 100 {
			t.Errorf("WinRate should be between 0 and 100, got %f", results.WinRate)
		}
	}
}

func BenchmarkSmaCross(b *testing.B) {
	ohlcv, err := data.LoadCSV(getTestDataPath("GOOG.csv"))
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy := &SmaCross{
			FastPeriod: 10,
			SlowPeriod: 20,
		}

		bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
			Cash:           10000,
			FinalizeTrades: true,
		})
		_, _ = bt.Run()
	}
}

func BenchmarkSmaCross_LargeData(b *testing.B) {
	// Use BTCUSD which typically has more data
	ohlcv, err := data.LoadCSV(getTestDataPath("BTCUSD.csv"))
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy := &SmaCross{
			FastPeriod: 10,
			SlowPeriod: 20,
		}

		bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
			Cash:           10000,
			FinalizeTrades: true,
		})
		_, _ = bt.Run()
	}
}
