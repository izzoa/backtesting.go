// Package main demonstrates RSI strategy with parameter optimization.
package main

import (
	"fmt"
	"log"
	"math"
	"path/filepath"
	"runtime"

	"github.com/quickfixgo/backtesting"
	"github.com/quickfixgo/backtesting/data"
	"github.com/quickfixgo/backtesting/lib"
	"github.com/quickfixgo/backtesting/optimize"
)

// RSIStrategy implements a mean reversion strategy using RSI.
// It buys when RSI is oversold and sells when RSI is overbought.
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

func main() {
	// Get the path to the testdata directory
	_, filename, _, _ := runtime.Caller(0)
	testdataPath := filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "csv", "GOOG.csv")

	// Load OHLCV data
	ohlcv, err := data.LoadCSV(testdataPath)
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	fmt.Printf("Loaded %d bars of GOOG data\n", ohlcv.Len())
	fmt.Printf("Period: %s to %s\n\n", ohlcv.Time[0].Format("2006-01-02"), ohlcv.Time[ohlcv.Len()-1].Format("2006-01-02"))

	// First, run with default parameters
	fmt.Println("=== Running with default parameters ===")
	defaultStrategy := &RSIStrategy{
		Period:     14,
		Overbought: 70,
		Oversold:   30,
	}

	bt := backtesting.NewBacktest(ohlcv, defaultStrategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	fmt.Printf("Return: %.2f%%\n", results.ReturnPct)
	fmt.Printf("Trades: %d\n", results.NumTrades)
	fmt.Printf("Max Drawdown: %.2f%%\n\n", results.MaxDrawdownPct)

	// Now run optimization
	fmt.Println("=== Running parameter optimization ===")

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
			Workers:  4,
		},
	)

	if err != nil {
		log.Fatalf("Optimization failed: %v", err)
	}

	fmt.Println(result.String())

	// Show top 5 results
	fmt.Println("\n=== Top 5 Parameter Combinations ===")
	top5 := result.TopN(5)
	for i, r := range top5 {
		fmt.Printf("%d. Period=%v, Overbought=%.0f, Oversold=%.0f => Return=%.2f%%\n",
			i+1,
			r.Params["Period"],
			r.Params["Overbought"],
			r.Params["Oversold"],
			r.Value)
	}
}
