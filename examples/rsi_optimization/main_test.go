package main

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
	"github.com/izzoa/backtesting.go/optimize"
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

func TestRSIStrategy_Compiles(t *testing.T) {
	// Verify the strategy struct can be created and has correct types
	strategy := &RSIStrategy{
		Period:     14,
		Overbought: 70,
		Oversold:   30,
	}

	if strategy.Period != 14 {
		t.Errorf("Period = %d, want 14", strategy.Period)
	}
	if strategy.Overbought != 70 {
		t.Errorf("Overbought = %f, want 70", strategy.Overbought)
	}
	if strategy.Oversold != 30 {
		t.Errorf("Oversold = %f, want 30", strategy.Oversold)
	}
}

func TestRSIStrategy_DefaultParams(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	strategy := &RSIStrategy{
		Period:     14,
		Overbought: 70,
		Oversold:   30,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	assertValidResults(t, results, "RSI_Default")

	t.Logf("Default (14, 70, 30): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestRSIStrategy_AggressiveParams(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// More aggressive: shorter period, wider thresholds
	strategy := &RSIStrategy{
		Period:     7,
		Overbought: 80,
		Oversold:   20,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	assertValidResults(t, results, "RSI_Aggressive")

	t.Logf("Aggressive (7, 80, 20): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestRSIStrategy_ConservativeParams(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// More conservative: longer period, tighter thresholds
	strategy := &RSIStrategy{
		Period:     21,
		Overbought: 65,
		Oversold:   35,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	assertValidResults(t, results, "RSI_Conservative")

	t.Logf("Conservative (21, 65, 35): Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestRSIStrategy_MultipleDataSets(t *testing.T) {
	dataSets := []string{"GOOG.csv", "EURUSD.csv", "BTCUSD.csv"}

	for _, ds := range dataSets {
		t.Run(ds, func(t *testing.T) {
			ohlcv := loadTestData(t, ds)

			strategy := &RSIStrategy{
				Period:     14,
				Overbought: 70,
				Oversold:   30,
			}

			bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
				Cash:           10000,
				FinalizeTrades: true,
			})

			results, err := bt.Run()
			if err != nil {
				t.Fatalf("Backtest failed: %v", err)
			}

			assertValidResults(t, results, "RSI_"+ds)

			t.Logf("%s: Return=%.2f%%, Trades=%d", ds, results.ReturnPct, results.NumTrades)
		})
	}
}

func TestRSIOptimization_FindsBestParams(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	result, err := optimize.GridSearch(
		ohlcv,
		func(params map[string]interface{}) backtesting.Strategy {
			return &RSIStrategy{
				Period:     params["Period"].(int),
				Overbought: params["Overbought"].(float64),
				Oversold:   params["Oversold"].(float64),
			}
		},
		backtesting.BacktestConfig{Cash: 10000, FinalizeTrades: true},
		optimize.GridConfig{
			Params: map[string][]interface{}{
				"Period":     optimize.Range(10, 16, 2),
				"Overbought": optimize.RangeFloat(70, 80, 5),
				"Oversold":   optimize.RangeFloat(20, 30, 5),
			},
			Maximize: "ReturnPct",
			Workers:  2,
		},
	)

	if err != nil {
		t.Fatalf("Optimization failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.BestParams == nil {
		t.Error("BestParams is nil")
	}

	if result.BestValue == 0 && len(result.AllResults) > 0 {
		t.Error("BestValue is 0 despite having combinations")
	}

	t.Logf("Best params: %v with value %.2f%%", result.BestParams, result.BestValue)
	t.Logf("Combinations tested: %d", len(result.AllResults))
}

func TestRSIOptimization_TopN(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	result, err := optimize.GridSearch(
		ohlcv,
		func(params map[string]interface{}) backtesting.Strategy {
			return &RSIStrategy{
				Period:     params["Period"].(int),
				Overbought: params["Overbought"].(float64),
				Oversold:   params["Oversold"].(float64),
			}
		},
		backtesting.BacktestConfig{Cash: 10000, FinalizeTrades: true},
		optimize.GridConfig{
			Params: map[string][]interface{}{
				"Period":     optimize.Range(10, 20, 2),
				"Overbought": optimize.RangeFloat(70, 80, 5),
				"Oversold":   optimize.RangeFloat(20, 30, 5),
			},
			Maximize: "ReturnPct",
			Workers:  2,
		},
	)

	if err != nil {
		t.Fatalf("Optimization failed: %v", err)
	}

	top5 := result.TopN(5)

	if len(top5) != 5 {
		t.Errorf("TopN(5) returned %d results, want 5", len(top5))
	}

	// Verify results are sorted in descending order
	for i := 1; i < len(top5); i++ {
		if top5[i].Value > top5[i-1].Value {
			t.Errorf("TopN results not sorted: %f > %f at index %d", top5[i].Value, top5[i-1].Value, i)
		}
	}

	t.Logf("Top 5 results:")
	for i, r := range top5 {
		t.Logf("  %d. Return=%.2f%% with %v", i+1, r.Value, r.Params)
	}
}

func TestRSIOptimization_WorkerCount(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Test with different worker counts
	workerCounts := []int{1, 2, 4}

	for _, workers := range workerCounts {
		t.Run(fmt.Sprintf("Workers_%d", workers), func(t *testing.T) {
			result, err := optimize.GridSearch(
				ohlcv,
				func(params map[string]interface{}) backtesting.Strategy {
					return &RSIStrategy{
						Period:     params["Period"].(int),
						Overbought: params["Overbought"].(float64),
						Oversold:   params["Oversold"].(float64),
					}
				},
				backtesting.BacktestConfig{Cash: 10000, FinalizeTrades: true},
				optimize.GridConfig{
					Params: map[string][]interface{}{
						"Period":     optimize.Range(10, 16, 2),
						"Overbought": optimize.RangeFloat(70, 80, 10),
						"Oversold":   optimize.RangeFloat(20, 30, 10),
					},
					Maximize: "ReturnPct",
					Workers:  workers,
				},
			)

			if err != nil {
				t.Fatalf("Optimization with %d workers failed: %v", workers, err)
			}

			if result == nil {
				t.Fatal("Result is nil")
			}

			t.Logf("Workers=%d: BestValue=%.2f%%", workers, result.BestValue)
		})
	}
}

func TestRSIOptimization_Constraints(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Test with constraint: Overbought > Oversold (which should always be true)
	result, err := optimize.GridSearch(
		ohlcv,
		func(params map[string]interface{}) backtesting.Strategy {
			return &RSIStrategy{
				Period:     params["Period"].(int),
				Overbought: params["Overbought"].(float64),
				Oversold:   params["Oversold"].(float64),
			}
		},
		backtesting.BacktestConfig{Cash: 10000, FinalizeTrades: true},
		optimize.GridConfig{
			Params: map[string][]interface{}{
				"Period":     optimize.Range(10, 16, 2),
				"Overbought": optimize.RangeFloat(60, 80, 10),
				"Oversold":   optimize.RangeFloat(20, 40, 10),
			},
			Maximize: "ReturnPct",
			Workers:  2,
			Constraint: optimize.ParamGreaterThan("Overbought", "Oversold"),
		},
	)

	if err != nil {
		t.Fatalf("Optimization with constraints failed: %v", err)
	}

	// Verify all results respect the constraint
	for _, pr := range result.AllResults {
		overbought := pr.Params["Overbought"].(float64)
		oversold := pr.Params["Oversold"].(float64)
		if overbought <= oversold {
			t.Errorf("Constraint violated: Overbought=%f <= Oversold=%f", overbought, oversold)
		}
	}

	t.Logf("Constraint test passed with %d valid combinations", len(result.AllResults))
}

func BenchmarkRSIStrategy(b *testing.B) {
	ohlcv, err := data.LoadCSV(getTestDataPath("GOOG.csv"))
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy := &RSIStrategy{
			Period:     14,
			Overbought: 70,
			Oversold:   30,
		}

		bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
			Cash:           10000,
			FinalizeTrades: true,
		})
		_, _ = bt.Run()
	}
}

func BenchmarkRSIOptimization(b *testing.B) {
	ohlcv, err := data.LoadCSV(getTestDataPath("GOOG.csv"))
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = optimize.GridSearch(
			ohlcv,
			func(params map[string]interface{}) backtesting.Strategy {
				return &RSIStrategy{
					Period:     params["Period"].(int),
					Overbought: params["Overbought"].(float64),
					Oversold:   params["Oversold"].(float64),
				}
			},
			backtesting.BacktestConfig{Cash: 10000, FinalizeTrades: true},
			optimize.GridConfig{
				Params: map[string][]interface{}{
					"Period":     optimize.Range(10, 16, 2),
					"Overbought": optimize.RangeFloat(70, 80, 10),
					"Oversold":   optimize.RangeFloat(20, 30, 10),
				},
				Maximize: "ReturnPct",
				Workers:  4,
			},
		)
	}
}
