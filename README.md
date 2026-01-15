# Backtesting.go

[![Go Reference](https://pkg.go.dev/badge/github.com/izzoa/backtesting.go.svg)](https://pkg.go.dev/github.com/izzoa/backtesting.go)
[![Go Report Card](https://goreportcard.com/badge/github.com/izzoa/backtesting.go)](https://goreportcard.com/report/github.com/izzoa/backtesting.go)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

Backtesting.go is a high-performance backtesting framework for trading strategies written in Go. It provides a simple, idiomatic API for testing trading strategies on historical OHLCV (Open-High-Low-Close-Volume) data.

## Features

- **Simple API** — Define strategies by implementing just two methods: `Init()` and `Next()`
- **High Performance** — Written in Go for speed; backtests run in milliseconds
- **Comprehensive Statistics** — 40+ performance metrics including Sharpe ratio, Sortino ratio, max drawdown, and more
- **Parameter Optimization** — Built-in grid search with parallel execution
- **Technical Indicators** — Library of common indicators (SMA, EMA, RSI, MACD, Bollinger Bands, ATR, etc.)
- **Interactive Charts** — Generate HTML visualizations with Chart.js
- **Flexible Order Types** — Market, limit, stop, and stop-limit orders with SL/TP support
- **Commission & Spread** — Model realistic trading costs
- **Multi-Timeframe** — Resample data to different timeframes

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Strategy Development](#strategy-development)
  - [The Strategy Interface](#the-strategy-interface)
  - [Using StrategyBase](#using-strategybase)
  - [Accessing Price Data](#accessing-price-data)
  - [Using Indicators](#using-indicators)
- [Order Management](#order-management)
  - [Market Orders](#market-orders)
  - [Limit and Stop Orders](#limit-and-stop-orders)
  - [Stop Loss and Take Profit](#stop-loss-and-take-profit)
- [Position Management](#position-management)
- [Backtest Configuration](#backtest-configuration)
- [Results and Statistics](#results-and-statistics)
- [Parameter Optimization](#parameter-optimization)
- [Technical Indicators](#technical-indicators)
- [Plotting and Visualization](#plotting-and-visualization)
- [Data Loading](#data-loading)
- [Multi-Timeframe Analysis](#multi-timeframe-analysis)
- [Examples](#examples)
- [Performance](#performance)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [License](#license)

## Installation

```bash
go get github.com/izzoa/backtesting.go
```

Requires Go 1.21 or later.

## Quick Start

Here's a complete example of a simple SMA crossover strategy:

```go
package main

import (
    "fmt"
    "log"

    "github.com/izzoa/backtesting.go"
    "github.com/izzoa/backtesting.go/data"
)

// SmaCross buys when fast SMA crosses above slow SMA,
// and sells when fast SMA crosses below slow SMA.
type SmaCross struct {
    backtesting.StrategyBase
    FastPeriod int
    SlowPeriod int
    fastSMA    *data.Indicator
    slowSMA    *data.Indicator
}

func (s *SmaCross) Init() {
    // Create indicators during initialization
    s.fastSMA = s.I("SMA_fast", data.SMA(s.Close(), s.FastPeriod))
    s.slowSMA = s.I("SMA_slow", data.SMA(s.Close(), s.SlowPeriod))
}

func (s *SmaCross) Next() {
    idx := s.BarIndex()
    if idx < 1 {
        return // Need at least 2 bars for crossover detection
    }

    // Get current and previous indicator values
    fastCurr := s.IndicatorAt(s.fastSMA, idx)
    fastPrev := s.IndicatorAt(s.fastSMA, idx-1)
    slowCurr := s.IndicatorAt(s.slowSMA, idx)
    slowPrev := s.IndicatorAt(s.slowSMA, idx-1)

    // Buy signal: fast crosses above slow
    if fastPrev <= slowPrev && fastCurr > slowCurr {
        if !s.Position().IsLong() {
            s.Buy()
        }
    }

    // Sell signal: fast crosses below slow
    if fastPrev >= slowPrev && fastCurr < slowCurr {
        if s.Position().IsLong() {
            s.Position().Close(1.0)
        }
    }
}

func main() {
    // Load historical data
    ohlcv, err := data.LoadCSV("GOOG.csv")
    if err != nil {
        log.Fatal(err)
    }

    // Create strategy instance
    strategy := &SmaCross{
        FastPeriod: 10,
        SlowPeriod: 20,
    }

    // Configure and run backtest
    bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
        Cash:           10000.0,
        FinalizeTrades: true,
    })

    results, err := bt.Run()
    if err != nil {
        log.Fatal(err)
    }

    // Print results
    fmt.Println(results.String())
}
```

**Output:**
```
Backtest Results
================
Period: 2004-08-19 to 2013-03-01 (74784h0m0s)
Initial Cash: $10000.00
Final Equity: $70377.47
Return: 603.77%

Trades: 47
Win Rate: 63.83%
Profit Factor: 3.51
Avg Trade: 5.08%
Best Trade: 57.43%
Worst Trade: -12.27%

Max Drawdown: 18.62%
```

## Strategy Development

### The Strategy Interface

Every strategy must implement the `Strategy` interface:

```go
type Strategy interface {
    Init()                    // Called once before backtesting starts
    Next()                    // Called for each bar during backtesting
    SetBroker(broker *Broker) // Sets the broker (called automatically)
    SetDataView(dv *DataView) // Sets the data view (called automatically)
}
```

In practice, you embed `StrategyBase` which provides default implementations and helper methods.

### Using StrategyBase

`StrategyBase` provides everything you need to build strategies:

```go
type MyStrategy struct {
    backtesting.StrategyBase  // Embed this
    // Your custom fields
    Period int
}

func (s *MyStrategy) Init() {
    // Initialize indicators here
}

func (s *MyStrategy) Next() {
    // Trading logic here
}
```

### Accessing Price Data

`StrategyBase` provides methods to access price data:

```go
// Current bar index (0-based)
idx := s.BarIndex()

// Price slices for indicator calculation
closes := s.Close()   // []float64
opens := s.Open()     // []float64
highs := s.High()     // []float64
lows := s.Low()       // []float64
volumes := s.Volume() // []float64

// Current bar values
currentClose := s.LastClose()
currentOpen := s.LastOpen()
currentHigh := s.LastHigh()
currentLow := s.LastLow()
currentVolume := s.LastVolume()
currentTime := s.LastTime()

// Access specific bar
bar := s.BarAt(idx)
```

### Using Indicators

Create indicators in `Init()` using the `I()` method:

```go
func (s *MyStrategy) Init() {
    // Register an indicator for use in Next()
    s.sma = s.I("SMA_20", data.SMA(s.Close(), 20))

    // Multiple indicators
    s.ema = s.I("EMA_10", data.EMA(s.Close(), 10))
}

func (s *MyStrategy) Next() {
    idx := s.BarIndex()

    // Get indicator value at current bar
    smaValue := s.IndicatorAt(s.sma, idx)

    // Check for NaN (indicator may not have enough data yet)
    if math.IsNaN(smaValue) {
        return
    }

    // Use the value
    if s.LastClose() > smaValue {
        // Price above SMA
    }
}
```

## Order Management

### Market Orders

Market orders execute at the next bar's open price:

```go
// Buy with all available equity
s.Buy()

// Sell (open short position) with all available equity
s.Sell()
```

### Limit and Stop Orders

```go
// Limit order: buy at specified price or better
s.Buy(backtesting.WithLimit(100.0))

// Stop order: buy when price reaches specified level
s.Buy(backtesting.WithStop(105.0))

// Stop-limit order: stop triggers limit order
s.Buy(
    backtesting.WithStop(105.0),
    backtesting.WithLimit(106.0),
)
```

### Stop Loss and Take Profit

```go
// Set stop loss
s.Buy(backtesting.WithSL(95.0))

// Set take profit
s.Buy(backtesting.WithTP(110.0))

// Both SL and TP
s.Buy(
    backtesting.WithSL(95.0),
    backtesting.WithTP(110.0),
)

// Percentage-based (relative to entry price)
// These are applied after order fills
```

### Order Options Summary

| Option | Description |
|--------|-------------|
| `WithLimit(price)` | Limit price for entry |
| `WithStop(price)` | Stop price for entry |
| `WithSL(price)` | Stop-loss price |
| `WithTP(price)` | Take-profit price |
| `WithSize(size)` | Specific position size |
| `WithTag(tag)` | Tag for order tracking |

## Position Management

```go
// Check position status
if s.Position().IsLong() {
    // Currently long
}

if s.Position().IsShort() {
    // Currently short
}

if s.Position().Size() == 0 {
    // No position
}

// Get position details
size := s.Position().Size()        // Positive = long, negative = short
pl := s.Position().PL()            // Unrealized P&L in dollars
plPct := s.Position().PLPct()      // Unrealized P&L as percentage
entryPrice := s.Position().EntryPrice()

// Close position
s.Position().Close(1.0)   // Close 100%
s.Position().Close(0.5)   // Close 50%
s.Position().Close(0.25)  // Close 25%
```

## Backtest Configuration

```go
bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
    // Starting capital
    Cash: 10000.0,

    // Commission per trade (function of size and price)
    Commission: func(size, price float64) float64 {
        return math.Abs(size) * price * 0.001 // 0.1%
    },

    // Bid-ask spread as fraction of price
    Spread: 0.0001, // 1 basis point

    // Margin requirement (1.0 = no margin)
    Margin: 1.0,

    // Close open positions at end of backtest
    FinalizeTrades: true,

    // Execute orders on close instead of next open
    TradeOnClose: false,

    // Allow only one position at a time
    ExclusiveOrders: false,

    // Allow hedging (long and short simultaneously)
    Hedging: false,
})
```

## Results and Statistics

The `BacktestResults` struct contains comprehensive performance metrics:

```go
results, _ := bt.Run()

// Basic metrics
results.StartTime        // Backtest start time
results.EndTime          // Backtest end time
results.Duration         // Total duration
results.InitialCash      // Starting capital
results.FinalEquity      // Ending equity
results.ReturnPct        // Total return percentage

// Trade metrics
results.NumTrades        // Number of trades
results.WinRate          // Win rate (0-100%)
results.ProfitFactor     // Gross profit / gross loss
results.AvgTradePct      // Average trade return
results.BestTradePct     // Best trade return
results.WorstTradePct    // Worst trade return

// Risk metrics
results.MaxDrawdownPct   // Maximum drawdown percentage

// Equity curve
results.EquityCurve      // []float64 - equity at each bar

// Individual trades
for _, trade := range results.Trades {
    trade.Size           // Position size
    trade.EntryPrice     // Entry price
    trade.EntryTime      // Entry time
    trade.ExitPrice      // Exit price (if closed)
    trade.ExitTime       // Exit time (if closed)
    trade.PL()           // Profit/loss
    trade.PLPct()        // P&L percentage
}
```

### Detailed Statistics

For comprehensive statistics, use the `stats` package:

```go
import "github.com/izzoa/backtesting.go/stats"

// Convert trades to records
var records []*stats.TradeRecord
for _, trade := range results.Trades {
    if trade.ExitPrice != nil {
        records = append(records, &stats.TradeRecord{
            Size:       trade.Size,
            EntryTime:  trade.EntryTime,
            EntryPrice: trade.EntryPrice,
            ExitTime:   *trade.ExitTime,
            ExitPrice:  *trade.ExitPrice,
            PL:         trade.PL(),
            PLPct:      trade.PLPct(),
        })
    }
}

// Compute full statistics
fullStats := stats.Compute(stats.ComputeConfig{
    Trades:       records,
    EquityCurve:  results.EquityCurve,
    Times:        ohlcv.Time,
    InitialCash:  10000,
    RiskFreeRate: 0.03,  // 3% annual
    TotalBars:    ohlcv.Len(),
})

// Access all metrics
fullStats.SharpeRatio      // Risk-adjusted return
fullStats.SortinoRatio     // Downside risk-adjusted return
fullStats.CalmarRatio      // Return / max drawdown
fullStats.AnnualizedReturn // Annualized return
fullStats.CAGR             // Compound annual growth rate
fullStats.Volatility       // Annualized volatility
fullStats.MaxDrawdownPct   // Maximum drawdown
fullStats.AvgDrawdownPct   // Average drawdown
fullStats.MaxDrawdownDuration  // Longest drawdown period
fullStats.SQN              // System Quality Number
fullStats.KellyCriterion   // Optimal bet fraction
fullStats.Expectancy       // Expected return per trade
```

## Parameter Optimization

Find optimal strategy parameters using grid search:

```go
import "github.com/izzoa/backtesting.go/optimize"

result, err := optimize.GridSearch(
    ohlcv,
    // Strategy factory function
    func(params map[string]interface{}) backtesting.Strategy {
        return &SmaCross{
            FastPeriod: params["FastPeriod"].(int),
            SlowPeriod: params["SlowPeriod"].(int),
        }
    },
    backtesting.BacktestConfig{
        Cash:           10000,
        FinalizeTrades: true,
    },
    optimize.GridConfig{
        // Parameter ranges
        Params: map[string][]interface{}{
            "FastPeriod": optimize.Range(5, 20, 5),    // 5, 10, 15, 20
            "SlowPeriod": optimize.Range(20, 50, 10),  // 20, 30, 40, 50
        },

        // Ensure fast < slow
        Constraint: optimize.ParamLessThan("FastPeriod", "SlowPeriod"),

        // Metric to optimize
        Maximize: "ReturnPct",  // or "SharpeRatio", "SortinoRatio", etc.

        // Parallel workers (0 = auto)
        Workers: 4,

        // Random sampling (0 = test all)
        MaxTries: 0,
    },
)

// Results
fmt.Printf("Best Parameters: %v\n", result.BestParams)
fmt.Printf("Best Return: %.2f%%\n", result.BestValue)
fmt.Printf("Combinations Tested: %d\n", result.Len())

// Get top N results
top5 := result.TopN(5)
for i, r := range top5 {
    fmt.Printf("%d. %v => %.2f%%\n", i+1, r.Params, r.Value)
}

// Generate heatmap for visualization
heatmap := result.Heatmap("FastPeriod", "SlowPeriod")
maxVal, xLabel, yLabel := heatmap.Max()
```

### Constraint Functions

```go
// Basic constraints
optimize.ParamLessThan("a", "b")      // a < b
optimize.ParamGreaterThan("a", "b")   // a > b
optimize.ParamNotEqual("a", "b")      // a != b
optimize.ParamMinValue("a", 10)       // a >= 10
optimize.ParamMaxValue("a", 100)      // a <= 100
optimize.ParamRange("a", 10, 100)     // 10 <= a <= 100

// Combine constraints
optimize.And(constraint1, constraint2)  // Both must be true
optimize.Or(constraint1, constraint2)   // Either can be true
```

### Parameter Range Helpers

```go
// Integer range
optimize.Range(5, 20, 5)     // [5, 10, 15, 20]

// Float range
optimize.RangeFloat(0.1, 0.5, 0.1)  // [0.1, 0.2, 0.3, 0.4, 0.5]

// Custom values
optimize.Values(10, 20, 50, 100)
```

## Technical Indicators

### Basic Indicators (data package)

```go
import "github.com/izzoa/backtesting.go/data"

// Simple Moving Average
sma := data.SMA(prices, period)

// Exponential Moving Average
ema := data.EMA(prices, period)
```

### Extended Indicators (lib package)

```go
import "github.com/izzoa/backtesting.go/lib"

// RSI (Relative Strength Index)
rsi := lib.RSI(closes, 14)

// MACD (Moving Average Convergence Divergence)
macdLine, signalLine, histogram := lib.MACD(closes, 12, 26, 9)

// Bollinger Bands
upper, middle, lower := lib.BollingerBands(closes, 20, 2.0)

// ATR (Average True Range)
atr := lib.ATR(highs, lows, closes, 14)

// ADX (Average Directional Index)
adx := lib.ADX(highs, lows, closes, 14)

// Stochastic Oscillator
k, d := lib.Stochastic(highs, lows, closes, 14, 3)

// CCI (Commodity Channel Index)
cci := lib.CCI(highs, lows, closes, 20)

// Williams %R
willR := lib.WilliamsR(highs, lows, closes, 14)
```

### Signal Detection

```go
import "github.com/izzoa/backtesting.go/lib"

// Check if series1 crossed series2 at index
crossed := lib.Cross(series1, series2, idx)

// Check if series1 crossed above series2
crossedOver := lib.Crossover(series1, series2, idx)

// Check if series1 crossed below series2
crossedUnder := lib.Crossunder(series1, series2, idx)

// Bars since condition was true
barsSince := lib.BarsSince(conditions, idx)  // conditions is []bool
```

### Statistical Functions

```go
import "github.com/izzoa/backtesting.go/lib"

// Quantile (0.0 to 1.0)
q := lib.Quantile(data, 0.75)

// Median (50th percentile)
med := lib.Median(data)

// Percentile (0 to 100)
p90 := lib.Percentile(data, 90)

// Interquartile Range
iqr := lib.IQR(data)
```

## Plotting and Visualization

Generate interactive HTML charts:

```go
import "github.com/izzoa/backtesting.go/plot"

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
        ti.ExitPrice = *trade.ExitPrice
    }
    trades = append(trades, ti)
}

// Prepare chart data
chartData := plot.PrepareChartData(ohlcv, results.EquityCurve, trades, nil)

// Generate HTML chart
err := plot.Plot(chartData, plot.Config{
    Filename:     "backtest_results.html",
    Title:        "SMA Crossover Strategy",
    PlotEquity:   true,   // Show equity curve
    PlotDrawdown: true,   // Show drawdown chart
    PlotVolume:   true,   // Show volume bars
    PlotTrades:   true,   // Show entry/exit markers
    ShowLegend:   true,   // Show legend
    OpenBrowser:  true,   // Open in browser automatically
})
```

### Chart Configuration

```go
plot.Config{
    Filename:       "chart.html",   // Output file
    Title:          "My Strategy",   // Chart title
    PlotEquity:     true,            // Equity curve
    PlotDrawdown:   true,            // Drawdown chart
    PlotVolume:     true,            // Volume bars
    PlotTrades:     true,            // Trade markers
    SmoothEquity:   0,               // Smoothing window (0 = none)
    RelativeEquity: false,           // Show as percentage change
    ShowLegend:     true,            // Show legend
    OpenBrowser:    true,            // Auto-open in browser
    Width:          0,               // Width in pixels (0 = responsive)
    Height:         0,               // Height in pixels (0 = auto)
}
```

## Data Loading

### From CSV

```go
import "github.com/izzoa/backtesting.go/data"

// Load from CSV file
// Expected columns: Date, Open, High, Low, Close, Volume
ohlcv, err := data.LoadCSV("data.csv")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Loaded %d bars\n", ohlcv.Len())
fmt.Printf("Date range: %s to %s\n",
    ohlcv.Time[0].Format("2006-01-02"),
    ohlcv.Time[ohlcv.Len()-1].Format("2006-01-02"))
```

### Programmatic Creation

```go
import "github.com/izzoa/backtesting.go/data"

// Create empty OHLCV with capacity
ohlcv := data.NewOHLCV(1000)

// Append bars
ohlcv.Append(data.Bar{
    Time:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    Open:   100.0,
    High:   105.0,
    Low:    99.0,
    Close:  103.0,
    Volume: 1000000,
})

// Access data
ohlcv.Time     // []time.Time
ohlcv.Open     // []float64
ohlcv.High     // []float64
ohlcv.Low      // []float64
ohlcv.Close    // []float64
ohlcv.Volume   // []float64
```

### Generate Random Data

```go
import "github.com/izzoa/backtesting.go/lib"

// Generate random OHLCV data for testing
ohlcv := lib.RandomOHLCData(lib.RandomDataConfig{
    Bars:       1000,
    StartPrice: 100.0,
    Volatility: 0.02,      // 2% daily volatility
    Drift:      0.0001,    // Slight upward drift
    StartTime:  time.Now().AddDate(-3, 0, 0),
    Interval:   24 * time.Hour,
})
```

## Multi-Timeframe Analysis

Resample data to higher timeframes:

```go
import "github.com/izzoa/backtesting.go/lib"

// Resample to daily
daily := lib.Resample(ohlcv, lib.ResampleDaily)

// Resample to weekly
weekly := lib.Resample(ohlcv, lib.ResampleWeekly)

// Resample to monthly
monthly := lib.Resample(ohlcv, lib.ResampleMonthly)

// Custom resampling function
custom := lib.ResampleApply(ohlcv, func(t time.Time) time.Time {
    // Group by hour
    return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
})
```

## Examples

See the [examples](./examples/) directory:

- **[sma_cross](./examples/sma_cross/)** — Simple SMA crossover strategy
- **[rsi_optimization](./examples/rsi_optimization/)** — RSI strategy with parameter optimization

Run an example:

```bash
go run ./examples/sma_cross
```

## Performance

Benchmarks on Apple M2:

| Benchmark | Operations | Time per Op |
|-----------|------------|-------------|
| Backtest (2000 bars, SMA cross) | 2,600 | 1.16 ms |
| Grid Search (16 combinations) | 100 | 39.05 ms |
| RSI Indicator (2000 bars) | 4,464 | 0.31 ms |
| MACD Indicator (2000 bars) | 6,896 | 1.19 ms |
| Bollinger Bands (2000 bars) | 936 | 2.22 ms |
| Stats Computation | 5,349 | 1.85 ms |

Run benchmarks:

```bash
go test -bench=. ./...
```

## API Reference

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/izzoa/backtesting.go).

### Package Overview

| Package | Description |
|---------|-------------|
| `backtesting` | Core backtesting engine |
| `backtesting/data` | OHLCV data structures and basic indicators |
| `backtesting/lib` | Extended indicators and utilities |
| `backtesting/optimize` | Parameter optimization |
| `backtesting/plot` | HTML chart generation |
| `backtesting/stats` | Trading statistics |

### Key Types

| Type | Package | Description |
|------|---------|-------------|
| `Backtest` | backtesting | Main backtest runner |
| `Strategy` | backtesting | Strategy interface |
| `StrategyBase` | backtesting | Base struct for strategies |
| `BacktestConfig` | backtesting | Configuration options |
| `BacktestResults` | backtesting | Results and metrics |
| `OHLCV` | data | Price data container |
| `Indicator` | data | Indicator data |
| `GridConfig` | optimize | Grid search configuration |
| `Result` | optimize | Optimization results |
| `Stats` | stats | Comprehensive statistics |

## Contributing

Contributions are welcome! Please read the [Contributing Guidelines](./CONTRIBUTING.md) before submitting pull requests.

### Development

```bash
# Clone the repository
git clone https://github.com/izzoa/backtesting.go.git
cd backtesting

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

## License

This project is licensed under the AGPL-3.0 License. See [LICENSE.md](./LICENSE.md) for details.

## Acknowledgments

This project is a Go port of [backtesting.py](https://github.com/kernc/backtesting.py) by Žiga Avsec. The original project provided the conceptual foundation and API design inspiration for this implementation.
