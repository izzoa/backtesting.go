package backtesting_test

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/quickfixgo/backtesting"
	"github.com/quickfixgo/backtesting/data"
	"github.com/quickfixgo/backtesting/lib"
	"github.com/quickfixgo/backtesting/optimize"
	"github.com/quickfixgo/backtesting/plot"
	"github.com/quickfixgo/backtesting/stats"
)

// =============================================================================
// Integration Test Strategies
// =============================================================================

// SmaCrossStrategy implements a simple SMA crossover strategy for testing.
type SmaCrossStrategy struct {
	backtesting.StrategyBase
	FastPeriod int
	SlowPeriod int
	fastSMA    *data.Indicator
	slowSMA    *data.Indicator
}

func (s *SmaCrossStrategy) Init() {
	s.fastSMA = s.I("SMA_fast", data.SMA(s.Close(), s.FastPeriod))
	s.slowSMA = s.I("SMA_slow", data.SMA(s.Close(), s.SlowPeriod))
}

func (s *SmaCrossStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	fastCurr := s.IndicatorAt(s.fastSMA, idx)
	fastPrev := s.IndicatorAt(s.fastSMA, idx-1)
	slowCurr := s.IndicatorAt(s.slowSMA, idx)
	slowPrev := s.IndicatorAt(s.slowSMA, idx-1)

	// Fast crosses above slow - buy
	if fastPrev <= slowPrev && fastCurr > slowCurr {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// Fast crosses below slow - sell
	if fastPrev >= slowPrev && fastCurr < slowCurr {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

// RSIStrategy implements an RSI-based mean reversion strategy.
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

// BollingerBandStrategy implements a Bollinger Band breakout strategy.
type BollingerBandStrategy struct {
	backtesting.StrategyBase
	Period int
	StdDev float64
	upper  *data.Indicator
	lower  *data.Indicator
}

func (s *BollingerBandStrategy) Init() {
	upper, _, lower := lib.BollingerBands(s.Close(), s.Period, s.StdDev)
	s.upper = s.I("BB_Upper", upper)
	s.lower = s.I("BB_Lower", lower)
}

func (s *BollingerBandStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	price := s.LastClose()
	upperValue := s.IndicatorAt(s.upper, idx)
	lowerValue := s.IndicatorAt(s.lower, idx)

	if math.IsNaN(upperValue) || math.IsNaN(lowerValue) {
		return
	}

	// Price breaks below lower band - buy
	if price < lowerValue && !s.Position().IsLong() {
		s.Buy()
	}

	// Price breaks above upper band - sell
	if price > upperValue && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestIntegration_FullBacktestCycle(t *testing.T) {
	// Load test data
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	// Create strategy
	strategy := &SmaCrossStrategy{
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
		t.Fatalf("Backtest failed: %v", err)
	}

	// Verify results structure
	if results.NumTrades == 0 {
		t.Error("Expected some trades")
	}

	if results.StartTime.IsZero() {
		t.Error("StartTime not set")
	}

	if results.EndTime.IsZero() {
		t.Error("EndTime not set")
	}

	if results.FinalEquity <= 0 {
		t.Error("FinalEquity should be positive")
	}

	if len(results.EquityCurve) == 0 {
		t.Error("EquityCurve should not be empty")
	}

	// Log results for manual verification
	t.Logf("Full backtest results:\n%s", results.String())
}

func TestIntegration_RSIStrategy(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

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

	t.Logf("RSI Strategy Results:\n")
	t.Logf("  Trades: %d\n", results.NumTrades)
	t.Logf("  Return: %.2f%%\n", results.ReturnPct)
	t.Logf("  Max Drawdown: %.2f%%\n", results.MaxDrawdownPct)
}

func TestIntegration_BollingerBandStrategy(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	strategy := &BollingerBandStrategy{
		Period: 20,
		StdDev: 2.0,
	}

	bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
		Cash:           10000,
		FinalizeTrades: true,
	})

	results, err := bt.Run()
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	t.Logf("Bollinger Band Strategy Results:\n")
	t.Logf("  Trades: %d\n", results.NumTrades)
	t.Logf("  Return: %.2f%%\n", results.ReturnPct)
	t.Logf("  Max Drawdown: %.2f%%\n", results.MaxDrawdownPct)
}

func TestIntegration_OptimizationCycle(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	// Run optimization
	result, err := optimize.GridSearch(
		ohlcv,
		optimize.StrategyFactory(func(params map[string]interface{}) backtesting.Strategy {
			return &SmaCrossStrategy{
				FastPeriod: params["FastPeriod"].(int),
				SlowPeriod: params["SlowPeriod"].(int),
			}
		}),
		backtesting.BacktestConfig{Cash: 10000, FinalizeTrades: true},
		optimize.GridConfig{
			Params: map[string][]interface{}{
				"FastPeriod": optimize.Range(5, 20, 5),
				"SlowPeriod": optimize.Range(20, 50, 10),
			},
			Constraint: optimize.ParamLessThan("FastPeriod", "SlowPeriod"),
			Maximize:   "ReturnPct",
			Workers:    2,
		},
	)

	if err != nil {
		t.Fatalf("Optimization failed: %v", err)
	}

	// Verify results
	if result.Len() == 0 {
		t.Error("Expected some optimization results")
	}

	if result.BestValue <= 0 {
		t.Error("Expected positive best value")
	}

	// Generate heatmap
	hm := result.Heatmap("FastPeriod", "SlowPeriod")
	if hm == nil {
		t.Error("Failed to generate heatmap")
	}

	t.Logf("Optimization Results:\n%s", result.String())
	t.Logf("Tested %d combinations", result.Len())
}

func TestIntegration_StatsComputation(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	strategy := &SmaCrossStrategy{
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

	// Convert trades to trade records for stats
	var tradeRecords []*stats.TradeRecord
	for _, trade := range results.Trades {
		if trade.ExitPrice == nil {
			continue
		}
		tr := &stats.TradeRecord{
			Size:       trade.Size,
			EntryTime:  trade.EntryTime,
			EntryPrice: trade.EntryPrice,
			ExitPrice:  *trade.ExitPrice,
			PL:         trade.PL(),
			PLPct:      trade.PLPct(),
			BarsHeld:   1,
		}
		if trade.ExitTime != nil {
			tr.ExitTime = *trade.ExitTime
			tr.Duration = trade.ExitTime.Sub(trade.EntryTime)
		}
		tradeRecords = append(tradeRecords, tr)
	}

	// Compute comprehensive stats
	fullStats := stats.Compute(stats.ComputeConfig{
		Trades:       tradeRecords,
		EquityCurve:  results.EquityCurve,
		Times:        ohlcv.Time,
		InitialCash:  10000,
		RiskFreeRate: 0.03,
		TotalBars:    ohlcv.Len(),
	})

	// Verify stats were computed
	if fullStats.NumTrades != len(tradeRecords) {
		t.Errorf("NumTrades mismatch: %d vs %d", fullStats.NumTrades, len(tradeRecords))
	}

	// Check that key metrics are reasonable
	if math.IsNaN(fullStats.SharpeRatio) {
		t.Error("SharpeRatio is NaN")
	}

	if math.IsNaN(fullStats.MaxDrawdownPct) {
		t.Error("MaxDrawdownPct is NaN")
	}

	t.Logf("Full Statistics:\n%s", fullStats.String())
}

func TestIntegration_PlotGeneration(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	strategy := &SmaCrossStrategy{
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

	// Convert trades to plot format
	var trades []plot.TradeInfo
	for _, trade := range results.Trades {
		ti := plot.TradeInfo{
			Size:       trade.Size,
			EntryTime:  trade.EntryTime,
			EntryPrice: trade.EntryPrice,
			PL:         trade.PL(),
			PLPct:      trade.PLPct(),
		}
		if trade.ExitTime != nil {
			ti.ExitTime = *trade.ExitTime
		}
		if trade.ExitPrice != nil {
			ti.ExitPrice = *trade.ExitPrice
		}
		trades = append(trades, ti)
	}

	// Prepare chart data
	chartData := plot.PrepareChartData(ohlcv, results.EquityCurve, trades, nil)

	// Generate plot
	tmpFile := "test_integration_chart.html"
	defer os.Remove(tmpFile)

	err = plot.Plot(chartData, plot.Config{
		Filename:     tmpFile,
		PlotEquity:   true,
		PlotDrawdown: true,
		PlotTrades:   true,
		OpenBrowser:  false,
		Title:        "Integration Test Chart",
	})

	if err != nil {
		t.Fatalf("Plot generation failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Plot file was not created")
	}
}

func TestIntegration_MultipleDatasets(t *testing.T) {
	datasets := []string{
		"testdata/csv/GOOG.csv",
		"testdata/csv/EURUSD.csv",
	}

	for _, ds := range datasets {
		t.Run(ds, func(t *testing.T) {
			ohlcv, err := data.LoadCSV(ds)
			if err != nil {
				t.Skipf("Dataset not available: %v", err)
				return
			}

			strategy := &SmaCrossStrategy{
				FastPeriod: 10,
				SlowPeriod: 20,
			}

			bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
				Cash:           10000,
				FinalizeTrades: true,
			})

			results, err := bt.Run()
			if err != nil {
				t.Errorf("Backtest failed for %s: %v", ds, err)
				return
			}

			t.Logf("%s: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
				ds, results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
		})
	}
}

func TestIntegration_CommissionAndSpread(t *testing.T) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	// Test with different commission/spread settings
	configs := []struct {
		name       string
		commission backtesting.CommissionFunc
		spread     float64
	}{
		{"No costs", nil, 0},
		{"Fixed commission", func(size, price float64) float64 { return 10 }, 0},
		{"Percent commission", func(size, price float64) float64 { return math.Abs(size) * price * 0.001 }, 0},
		{"With spread", nil, 0.0001},
		{"Full costs", func(size, price float64) float64 { return 5 }, 0.0005},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			strategy := &SmaCrossStrategy{
				FastPeriod: 10,
				SlowPeriod: 20,
			}

			bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
				Cash:           10000,
				Commission:     cfg.commission,
				Spread:         cfg.spread,
				FinalizeTrades: true,
			})

			results, err := bt.Run()
			if err != nil {
				t.Errorf("Backtest failed: %v", err)
				return
			}

			t.Logf("%s: Return=%.2f%%, Trades=%d",
				cfg.name, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkBacktest_SmaCross_2000Bars(b *testing.B) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy := &SmaCrossStrategy{
			FastPeriod: 10,
			SlowPeriod: 20,
		}
		bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{Cash: 10000})
		_, _ = bt.Run()
	}
}

func BenchmarkOptimization_GridSearch_16Combinations(b *testing.B) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	var factory optimize.StrategyFactory = func(params map[string]interface{}) backtesting.Strategy {
		return &SmaCrossStrategy{
			FastPeriod: params["FastPeriod"].(int),
			SlowPeriod: params["SlowPeriod"].(int),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = optimize.GridSearch(
			ohlcv,
			factory,
			backtesting.BacktestConfig{Cash: 10000},
			optimize.GridConfig{
				Params: map[string][]interface{}{
					"FastPeriod": {5, 10, 15, 20},
					"SlowPeriod": {20, 30, 40, 50},
				},
				Constraint: optimize.ParamLessThan("FastPeriod", "SlowPeriod"),
				Maximize:   "ReturnPct",
				Workers:    4,
			},
		)
	}
}

func BenchmarkIndicators_RSI(b *testing.B) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lib.RSI(ohlcv.Close, 14)
	}
}

func BenchmarkIndicators_MACD(b *testing.B) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lib.MACD(ohlcv.Close, 12, 26, 9)
	}
}

func BenchmarkIndicators_BollingerBands(b *testing.B) {
	ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
	if err != nil {
		b.Fatalf("Failed to load data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lib.BollingerBands(ohlcv.Close, 20, 2.0)
	}
}

func BenchmarkStatsComputation(b *testing.B) {
	// Create sample equity curve
	equity := make([]float64, 2000)
	times := make([]time.Time, 2000)
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := range equity {
		equity[i] = 10000 + float64(i)*5
		times[i] = base.Add(time.Duration(i) * 24 * time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stats.Compute(stats.ComputeConfig{
			EquityCurve:  equity,
			Times:        times,
			InitialCash:  10000,
			RiskFreeRate: 0.03,
		})
	}
}
