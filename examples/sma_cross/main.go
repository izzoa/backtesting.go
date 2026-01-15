// Package main demonstrates a simple SMA crossover strategy.
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
)

// SmaCross implements a simple SMA crossover strategy.
// It buys when the fast SMA crosses above the slow SMA,
// and sells when the fast SMA crosses below.
type SmaCross struct {
	backtesting.StrategyBase
	FastPeriod int
	SlowPeriod int
	fastSMA    *data.Indicator
	slowSMA    *data.Indicator
}

func (s *SmaCross) Init() {
	s.fastSMA = s.I("SMA_fast", data.SMA(s.Close(), s.FastPeriod))
	s.slowSMA = s.I("SMA_slow", data.SMA(s.Close(), s.SlowPeriod))
}

func (s *SmaCross) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	fastCurr := s.IndicatorAt(s.fastSMA, idx)
	fastPrev := s.IndicatorAt(s.fastSMA, idx-1)
	slowCurr := s.IndicatorAt(s.slowSMA, idx)
	slowPrev := s.IndicatorAt(s.slowSMA, idx-1)

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

	// Create strategy
	strategy := &SmaCross{
		FastPeriod: 10,
		SlowPeriod: 20,
	}

	// Run backtest
	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	// Print results
	fmt.Println(results.String())
}
