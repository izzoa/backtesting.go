# Python to Go Conversion Plan: backtesting.py → backtesting-go

This document provides a complete, step-by-step conversion plan for transforming the Python backtesting.py library (~3,825 lines) into an idiomatic Go implementation. Each task is designed to be actionable by an LLM.

---

## Table of Contents

1. [Overview](#overview)
2. [Go Project Structure](#go-project-structure)
3. [Phase 1: Project Setup & Data Structures](#phase-1-project-setup--data-structures)
4. [Phase 2: Core Engine](#phase-2-core-engine)
5. [Phase 3: Statistics Module](#phase-3-statistics-module)
6. [Phase 4: Library Utilities](#phase-4-library-utilities)
7. [Phase 5: Optimization System](#phase-5-optimization-system)
8. [Phase 6: Plotting & Visualization](#phase-6-plotting--visualization)
9. [Phase 7: Testing & Validation](#phase-7-testing--validation)
10. [Phase 8: Documentation & Examples](#phase-8-documentation--examples)
11. [Architectural Decisions](#architectural-decisions)
12. [Dependency Mapping](#dependency-mapping)

---

## Overview

### Source Analysis

| Component | Python File | Lines | Go Target |
|-----------|-------------|-------|-----------|
| Core Engine | `backtesting.py` | 1,750 | `core/` package |
| Utilities | `_util.py` | 337 | `data/` package |
| Statistics | `_stats.py` | 212 | `stats/` package |
| Plotting | `_plotting.py` | 785 | `plot/` package |
| Library | `lib.py` | 646 | `lib/` package |
| Tests | `test/_test.py` | 1,200+ | `*_test.go` files |

### Key Conversion Challenges

1. **Dynamic typing → Static typing**: Python's duck typing must become Go interfaces
2. **Pandas DataFrames → Custom structs**: Need efficient time-series handling
3. **NumPy arrays → Go slices**: With gonum for numerical operations
4. **ABC/Metaclasses → Interfaces**: Strategy pattern via Go interfaces
5. **Multiprocessing → Goroutines**: Native Go concurrency
6. **Bokeh → Go templates + JS**: Or alternative charting library

---

## Go Project Structure

```
backtesting-go/
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
├── LICENSE.md
│
├── backtest.go              # Main Backtest struct and Run()
├── strategy.go              # Strategy interface and base implementations
├── broker.go                # Broker simulation engine
├── order.go                 # Order struct and methods
├── trade.go                 # Trade struct and methods
├── position.go              # Position struct and methods
│
├── data/
│   ├── ohlcv.go             # OHLCV data structure
│   ├── series.go            # Time series utilities
│   ├── indicator.go         # Indicator wrapper type
│   └── loader.go            # CSV/data loading utilities
│
├── stats/
│   ├── stats.go             # Statistics computation
│   ├── metrics.go           # Individual metric calculations
│   └── drawdown.go          # Drawdown analysis
│
├── lib/
│   ├── signals.go           # SignalStrategy implementation
│   ├── trailing.go          # TrailingStrategy implementation
│   ├── crossover.go         # Cross/crossover utilities
│   ├── resample.go          # Multi-timeframe support
│   └── random.go            # Random data generation
│
├── optimize/
│   ├── grid.go              # Grid search optimization
│   ├── bayesian.go          # Bayesian optimization (optional)
│   └── constraint.go        # Constraint handling
│
├── plot/
│   ├── plot.go              # Main plotting interface
│   ├── candlestick.go       # OHLC candlestick rendering
│   ├── equity.go            # Equity curve rendering
│   ├── templates/           # HTML templates
│   │   └── chart.html
│   └── assets/
│       └── chart.js
│
├── testdata/
│   ├── GOOG.csv
│   ├── EURUSD.csv
│   └── BTCUSD.csv
│
└── examples/
    ├── sma_cross/
    │   └── main.go
    ├── rsi_strategy/
    │   └── main.go
    └── optimization/
        └── main.go
```

---

## Phase 1: Project Setup & Data Structures

### 1.1 Initialize Go Module

- [ ] Create `go.mod` with module path `github.com/user/backtesting-go` (or appropriate path)
- [ ] Add initial dependencies:
  ```
  gonum.org/v1/gonum v0.14.0
  github.com/gocarina/gocsv v0.0.0-20231116093920-b87c2d0e983a
  ```
- [ ] Create basic directory structure as outlined above
- [ ] Add `.gitignore` for Go projects

### 1.2 Implement OHLCV Data Structure

**File: `data/ohlcv.go`**

- [ ] Define `OHLCV` struct:
  ```go
  type OHLCV struct {
      Time   []time.Time
      Open   []float64
      High   []float64
      Low    []float64
      Close  []float64
      Volume []float64
  }
  ```
- [ ] Implement `NewOHLCV(capacity int) *OHLCV` constructor
- [ ] Implement `Len() int` method
- [ ] Implement `Slice(start, end int) *OHLCV` method for creating views
- [ ] Implement `At(i int) Bar` method returning single bar
- [ ] Implement `LastN(n int) *OHLCV` method for recent data
- [ ] Define `Bar` struct for single OHLCV record:
  ```go
  type Bar struct {
      Time   time.Time
      Open   float64
      High   float64
      Low    float64
      Close  float64
      Volume float64
  }
  ```

### 1.3 Implement Data Loader

**File: `data/loader.go`**

- [ ] Implement `LoadCSV(path string, opts ...LoadOption) (*OHLCV, error)`
- [ ] Support column name mapping via options:
  ```go
  type LoadOption func(*loadConfig)
  func WithTimeColumn(name string) LoadOption
  func WithOpenColumn(name string) LoadOption
  // ... etc
  ```
- [ ] Support date parsing with configurable formats
- [ ] Implement `LoadDataFrame(data map[string][]float64, times []time.Time) (*OHLCV, error)`
- [ ] Add validation: ensure all slices have equal length
- [ ] Add validation: ensure High >= Low, High >= Open, High >= Close, etc.

### 1.4 Implement Time Series Utilities

**File: `data/series.go`**

- [ ] Define `Series` struct (named float64 slice with metadata):
  ```go
  type Series struct {
      Name   string
      Values []float64
  }
  ```
- [ ] Implement `NewSeries(name string, values []float64) *Series`
- [ ] Implement `Series.Slice(start, end int) *Series`
- [ ] Implement `Series.Last() float64`
- [ ] Implement `Series.LastN(n int) []float64`
- [ ] Implement `Series.Copy() *Series`
- [ ] Implement `Series.Apply(fn func(float64) float64) *Series`
- [ ] Implement `Series.Rolling(window int) *RollingWindow`

### 1.5 Implement Indicator Type

**File: `data/indicator.go`**

- [ ] Define `Indicator` struct extending Series with plot metadata:
  ```go
  type Indicator struct {
      Series
      PlotOptions IndicatorPlotOptions
  }

  type IndicatorPlotOptions struct {
      Overlay    bool
      Color      string
      ScatterKey string // for scatter markers
      Name       string
  }
  ```
- [ ] Implement `NewIndicator(name string, values []float64, opts ...IndicatorOption) *Indicator`
- [ ] Implement indicator option functions:
  ```go
  func WithOverlay(overlay bool) IndicatorOption
  func WithColor(color string) IndicatorOption
  func WithScatter(key string) IndicatorOption
  ```

### 1.6 Implement Basic Numerical Helpers

**File: `data/math.go`**

- [ ] Implement `SMA(values []float64, period int) []float64` - Simple Moving Average
- [ ] Implement `EMA(values []float64, period int) []float64` - Exponential Moving Average
- [ ] Implement `StdDev(values []float64, period int) []float64` - Rolling Std Dev
- [ ] Implement `Max(values []float64, period int) []float64` - Rolling Max
- [ ] Implement `Min(values []float64, period int) []float64` - Rolling Min
- [ ] Implement `Diff(values []float64) []float64` - First difference
- [ ] Implement `Shift(values []float64, periods int) []float64` - Shift with NaN fill
- [ ] Implement `FillNaN(values []float64, fillValue float64) []float64`
- [ ] Implement `IsNaN(value float64) bool` wrapper
- [ ] Use `math.NaN()` for missing values

**Acceptance Criteria for Phase 1:**
- [ ] All data structures compile without errors
- [ ] Can load GOOG.csv, EURUSD.csv, BTCUSD.csv test files
- [ ] SMA calculation matches Python ta.SMA output for test data
- [ ] Unit tests pass for all data operations

---

## Phase 2: Core Engine

### 2.1 Implement Order Type

**File: `order.go`**

- [ ] Define `OrderSide` enum:
  ```go
  type OrderSide int
  const (
      Buy OrderSide = iota
      Sell
  )
  ```
- [ ] Define `Order` struct:
  ```go
  type Order struct {
      broker      *Broker
      ID          int64
      Size        float64   // positive = buy, negative = sell
      Limit       *float64  // nil for market order
      Stop        *float64  // stop price
      SL          *float64  // stop-loss price
      TP          *float64  // take-profit price
      ParentTrade *Trade    // for contingent orders
      Tag         interface{}
      createdBar  int
  }
  ```
- [ ] Implement `Order.IsLong() bool`
- [ ] Implement `Order.IsShort() bool`
- [ ] Implement `Order.IsContingent() bool`
- [ ] Implement `Order.Cancel()`
- [ ] Implement `Order.String() string` for debugging

### 2.2 Implement Trade Type

**File: `trade.go`**

- [ ] Define `Trade` struct:
  ```go
  type Trade struct {
      broker     *Broker
      ID         int64
      Size       float64
      EntryPrice float64
      EntryBar   int
      EntryTime  time.Time
      ExitPrice  *float64
      ExitBar    *int
      ExitTime   *time.Time
      sl         *float64
      tp         *float64
      slOrder    *Order
      tpOrder    *Order
      Tag        interface{}
  }
  ```
- [ ] Implement `Trade.PL() float64` - profit/loss in cash
- [ ] Implement `Trade.PLPct() float64` - profit/loss in percent
- [ ] Implement `Trade.Value() float64` - trade value (|size| × price)
- [ ] Implement `Trade.IsLong() bool`
- [ ] Implement `Trade.IsShort() bool`
- [ ] Implement `Trade.SL() *float64` getter
- [ ] Implement `Trade.SetSL(price float64)` setter (creates contingent order)
- [ ] Implement `Trade.TP() *float64` getter
- [ ] Implement `Trade.SetTP(price float64)` setter (creates contingent order)
- [ ] Implement `Trade.Close(portion float64)` - close trade or portion
- [ ] Implement `Trade.String() string` for debugging

### 2.3 Implement Position Type

**File: `position.go`**

- [ ] Define `Position` struct:
  ```go
  type Position struct {
      broker *Broker
  }
  ```
- [ ] Implement `Position.Size() float64` - sum of all active trade sizes
- [ ] Implement `Position.PL() float64` - total P&L of active trades
- [ ] Implement `Position.PLPct() float64` - weighted average P&L %
- [ ] Implement `Position.IsLong() bool`
- [ ] Implement `Position.IsShort() bool`
- [ ] Implement `Position.Close(portion float64)` - close position by portion

### 2.4 Implement Strategy Interface

**File: `strategy.go`**

- [ ] Define `Strategy` interface:
  ```go
  type Strategy interface {
      Init()
      Next()
  }
  ```
- [ ] Define `StrategyBase` struct (embedded in user strategies):
  ```go
  type StrategyBase struct {
      broker *Broker
      data   *DataView
      indicators []*Indicator
  }
  ```
- [ ] Implement `StrategyBase.Data() *DataView` - access OHLCV data
- [ ] Implement `StrategyBase.Close() *Series` - shortcut for close prices
- [ ] Implement `StrategyBase.Open() *Series`
- [ ] Implement `StrategyBase.High() *Series`
- [ ] Implement `StrategyBase.Low() *Series`
- [ ] Implement `StrategyBase.Volume() *Series`
- [ ] Implement `StrategyBase.Index() []time.Time`
- [ ] Implement `StrategyBase.Equity() float64`
- [ ] Implement `StrategyBase.Position() *Position`
- [ ] Implement `StrategyBase.Orders() []*Order`
- [ ] Implement `StrategyBase.Trades() []*Trade`
- [ ] Implement `StrategyBase.ClosedTrades() []*Trade`
- [ ] Implement `StrategyBase.Buy(opts ...OrderOption) *Order`
- [ ] Implement `StrategyBase.Sell(opts ...OrderOption) *Order`
- [ ] Define `OrderOption` type:
  ```go
  type OrderOption func(*orderConfig)
  func WithSize(size float64) OrderOption
  func WithLimit(price float64) OrderOption
  func WithStop(price float64) OrderOption
  func WithSL(price float64) OrderOption
  func WithTP(price float64) OrderOption
  func WithTag(tag interface{}) OrderOption
  ```
- [ ] Implement `StrategyBase.I(name string, values []float64, opts ...IndicatorOption) *Indicator`
  - Validate values length matches data length
  - Auto-detect overlay (within 30% of close price range)
  - Store in indicators slice
  - Return indicator for use in strategy

### 2.5 Implement DataView Type

**File: `data/view.go`**

- [ ] Define `DataView` struct (partial view of OHLCV):
  ```go
  type DataView struct {
      source *OHLCV
      length int  // current visible length (grows during backtest)
  }
  ```
- [ ] Implement all OHLCV accessor methods with length limiting
- [ ] Implement `DataView.SetLength(n int)` for backtest progression

### 2.6 Implement Broker

**File: `broker.go`**

- [ ] Define `Broker` struct:
  ```go
  type Broker struct {
      data           *OHLCV
      cash           float64
      commission     CommissionFunc
      margin         float64
      spread         float64
      tradeOnClose   bool
      hedging        bool
      exclusiveOrders bool

      orders       []*Order
      trades       []*Trade
      closedTrades []*Trade
      position     *Position
      equity       []float64

      currentBar   int
      lastPrice    float64
  }

  type CommissionFunc func(size float64, price float64) float64
  ```
- [ ] Implement `NewBroker(opts ...BrokerOption) *Broker`
- [ ] Define broker options:
  ```go
  func WithCash(cash float64) BrokerOption
  func WithCommission(fn CommissionFunc) BrokerOption
  func WithFixedCommission(amount float64) BrokerOption
  func WithPercentCommission(pct float64) BrokerOption
  func WithMargin(margin float64) BrokerOption
  func WithSpread(spread float64) BrokerOption
  func WithTradeOnClose(b bool) BrokerOption
  func WithHedging(b bool) BrokerOption
  func WithExclusiveOrders(b bool) BrokerOption
  ```
- [ ] Implement `Broker.Equity() float64` - current equity
- [ ] Implement `Broker.MarginAvailable() float64`
- [ ] Implement `Broker.LastPrice() float64`
- [ ] Implement `Broker.NewOrder(size float64, opts ...OrderOption) *Order`
- [ ] Implement `Broker.Next()` - process one bar:
  - Update currentBar
  - Update lastPrice from data
  - Call processOrders()
  - Record equity
- [ ] Implement `Broker.processOrders()` - main order matching logic:
  - For each pending order:
    - Check stop condition (price crossed stop level)
    - Check limit condition (price crossed limit level)
    - Determine execution price with spread
    - Handle contingent orders (SL/TP) - reduce parent trade
    - Handle new orders:
      - Apply exclusive_orders if enabled
      - Reduce opposite-facing trades (FIFO, non-hedging)
      - Check margin availability
      - Open new trade
      - Create SL/TP bracket orders if specified
  - Remove filled/cancelled orders
- [ ] Implement `Broker.openTrade(order *Order, price float64) *Trade`
- [ ] Implement `Broker.closeTrade(trade *Trade, price float64)`
- [ ] Implement `Broker.reduceTrade(trade *Trade, size float64, price float64)`
- [ ] Implement `Broker.adjustMargin(size float64, price float64) bool`

### 2.7 Implement Main Backtest

**File: `backtest.go`**

- [ ] Define `Backtest` struct:
  ```go
  type Backtest struct {
      data     *OHLCV
      strategy Strategy
      broker   *Broker
  }
  ```
- [ ] Define `BacktestConfig` struct:
  ```go
  type BacktestConfig struct {
      Cash            float64
      Commission      CommissionFunc
      Margin          float64
      Spread          float64
      TradeOnClose    bool
      Hedging         bool
      ExclusiveOrders bool
      FinalizeTrades  bool
  }
  ```
- [ ] Implement `NewBacktest(data *OHLCV, strategy Strategy, cfg BacktestConfig) *Backtest`
- [ ] Implement `Backtest.Run(params map[string]interface{}) (*Stats, error)`:
  - Initialize broker with config
  - Bind strategy to broker and data view
  - Call strategy.Init()
  - Calculate warmup period (first bar where all indicators are non-NaN)
  - Main loop from warmup+1 to end:
    - Update data view length (i+1)
    - Update indicator slices
    - Call broker.Next()
    - Call strategy.Next()
    - Check for out-of-money (equity <= 0)
  - Optionally finalize open trades
  - Compute and return stats
- [ ] Implement `Backtest.SetStrategyParam(name string, value interface{})` for optimization

### 2.8 Implement Parameter Injection System

**File: `params.go`**

- [ ] Define `Param` type for strategy parameters:
  ```go
  type Param[T any] struct {
      Name    string
      Default T
      value   T
  }
  ```
- [ ] Implement parameter reflection system to inject values before Init()
- [ ] Support int, float64, bool, string parameter types

**Acceptance Criteria for Phase 2:**
- [ ] Can define a simple SMA crossover strategy implementing Strategy interface
- [ ] Can run backtest on GOOG.csv data
- [ ] Orders execute correctly (market, limit, stop)
- [ ] Position tracking is accurate
- [ ] Trade P&L calculations match Python version
- [ ] Equity curve is correctly recorded
- [ ] Unit tests pass for all core components

---

## Phase 3: Statistics Module

### 3.1 Implement Stats Structure

**File: `stats/stats.go`**

- [ ] Define `Stats` struct with all metrics:
  ```go
  type Stats struct {
      // Time
      Start           time.Time
      End             time.Time
      Duration        time.Duration
      ExposureTimePct float64

      // Returns
      ReturnPct       float64
      ReturnAnnPct    float64
      BuyHoldReturnPct float64
      CAGR            float64
      AlphaPct        float64

      // Risk
      VolatilityAnnPct float64
      Beta            float64
      MaxDrawdownPct  float64
      AvgDrawdownPct  float64
      MaxDrawdownDuration time.Duration
      AvgDrawdownDuration time.Duration

      // Ratios
      SharpeRatio     float64
      SortinoRatio    float64
      CalmarRatio     float64

      // Trade Statistics
      NumTrades       int
      WinRatePct      float64
      BestTradePct    float64
      WorstTradePct   float64
      AvgTradePct     float64
      MaxTradeDuration time.Duration
      AvgTradeDuration time.Duration
      ProfitFactor    float64
      ExpectancyPct   float64
      SQN             float64
      KellyCriterion  float64

      // Data
      EquityCurve     *EquityCurve
      Trades          []*TradeRecord
  }
  ```

### 3.2 Implement Individual Metrics

**File: `stats/metrics.go`**

- [ ] Implement `calcReturn(startEquity, endEquity float64) float64`
- [ ] Implement `calcAnnualizedReturn(totalReturn float64, days int) float64`
- [ ] Implement `calcCAGR(startValue, endValue float64, years float64) float64`
- [ ] Implement `calcVolatility(returns []float64) float64`
- [ ] Implement `calcSharpeRatio(returns []float64, riskFreeRate float64) float64`
- [ ] Implement `calcSortinoRatio(returns []float64, riskFreeRate float64) float64`
- [ ] Implement `calcCalmarRatio(annualReturn, maxDrawdown float64) float64`
- [ ] Implement `calcBeta(strategyReturns, benchmarkReturns []float64) float64`
- [ ] Implement `calcAlpha(strategyReturn, benchmarkReturn, beta, riskFreeRate float64) float64`
- [ ] Implement `calcWinRate(trades []*Trade) float64`
- [ ] Implement `calcProfitFactor(trades []*Trade) float64`
- [ ] Implement `calcExpectancy(trades []*Trade) float64`
- [ ] Implement `calcSQN(trades []*Trade) float64`
- [ ] Implement `calcKellyCriterion(winRate, avgWin, avgLoss float64) float64`
- [ ] Implement `calcExposureTime(trades []*Trade, totalBars int) float64`

### 3.3 Implement Drawdown Analysis

**File: `stats/drawdown.go`**

- [ ] Define `Drawdown` struct:
  ```go
  type Drawdown struct {
      Peak        float64
      Trough      float64
      DrawdownPct float64
      StartBar    int
      TroughBar   int
      EndBar      int  // recovery bar, or -1 if not recovered
      Duration    int  // bars
  }
  ```
- [ ] Implement `calcDrawdowns(equity []float64) []Drawdown`
- [ ] Implement `calcMaxDrawdown(equity []float64) (float64, Drawdown)`
- [ ] Implement `calcAvgDrawdown(drawdowns []Drawdown) float64`
- [ ] Implement `calcDrawdownDurations(equity []float64, times []time.Time) (max, avg time.Duration)`

### 3.4 Implement Equity Curve

**File: `stats/equity.go`**

- [ ] Define `EquityCurve` struct:
  ```go
  type EquityCurve struct {
      Time             []time.Time
      Equity           []float64
      DrawdownPct      []float64
      DrawdownDuration []int  // bars in current drawdown
  }
  ```
- [ ] Implement `NewEquityCurve(times []time.Time, equity []float64) *EquityCurve`
- [ ] Calculate drawdown series during construction

### 3.5 Implement Trade Records

**File: `stats/trades.go`**

- [ ] Define `TradeRecord` struct (serializable trade info):
  ```go
  type TradeRecord struct {
      ID         int64
      Size       float64
      EntryTime  time.Time
      ExitTime   time.Time
      EntryPrice float64
      ExitPrice  float64
      PL         float64
      PLPct      float64
      ReturnPct  float64
      Duration   time.Duration
      Tag        interface{}
      // Indicator values at entry/exit
      Indicators map[string]float64
  }
  ```
- [ ] Implement `NewTradeRecord(trade *Trade, indicators []*Indicator) *TradeRecord`

### 3.6 Implement Main Compute Function

**File: `stats/compute.go`**

- [ ] Implement `ComputeStats(cfg ComputeConfig) *Stats`:
  ```go
  type ComputeConfig struct {
      Trades       []*Trade
      ClosedTrades []*Trade
      Equity       []float64
      Data         *OHLCV
      Indicators   []*Indicator
      RiskFreeRate float64
  }
  ```
- [ ] Calculate all metrics
- [ ] Build equity curve
- [ ] Build trade records with indicator values
- [ ] Handle edge cases (no trades, single trade, etc.)

### 3.7 Implement Stats Display

**File: `stats/display.go`**

- [ ] Implement `Stats.String() string` - formatted statistics output
- [ ] Implement `Stats.ToMap() map[string]interface{}` - for JSON export
- [ ] Implement `Stats.ToCSV(path string) error` - export to CSV

**Acceptance Criteria for Phase 3:**
- [ ] All metrics calculate correctly against known values
- [ ] Stats output matches Python version for same backtest
- [ ] Drawdown calculations are accurate
- [ ] Trade records include correct indicator values
- [ ] Unit tests with golden data pass

---

## Phase 4: Library Utilities

### 4.1 Implement Crossover Functions

**File: `lib/crossover.go`**

- [ ] Implement `Cross(series1, series2 []float64) []bool` - any crossover
- [ ] Implement `Crossover(series1, series2 []float64) []bool` - upward crossover
- [ ] Implement `Crossunder(series1, series2 []float64) []bool` - downward crossover
- [ ] Implement `BarsSince(condition []bool, defaultVal int) []int`

### 4.2 Implement Quantile Function

**File: `lib/quantile.go`**

- [ ] Implement `Quantile(series []float64, q float64) float64` - single quantile
- [ ] Implement `RollingQuantile(series []float64, window int, q float64) []float64`

### 4.3 Implement Resample/Multi-Timeframe

**File: `lib/resample.go`**

- [ ] Define resampling rules:
  ```go
  var OHLCVAgg = map[string]AggFunc{
      "Open":   First,
      "High":   Max,
      "Low":    Min,
      "Close":  Last,
      "Volume": Sum,
  }
  ```
- [ ] Implement `Resample(data *OHLCV, rule string) *OHLCV`
  - Support rules: "D" (daily), "W" (weekly), "M" (monthly), "H" (hourly), etc.
- [ ] Implement `ResampleApply(rule string, fn func([]float64) float64, series []float64, times []time.Time) []float64`
- [ ] Handle alignment of resampled data back to original timeframe

### 4.4 Implement Random Data Generator

**File: `lib/random.go`**

- [ ] Implement `RandomOHLCData(example *OHLCV, frac float64, seed int64) *OHLCV`
  - Generate synthetic OHLCV data based on example distribution
  - Use bootstrap resampling of returns
- [ ] Implement `ShuffleReturns(data *OHLCV, seed int64) *OHLCV`

### 4.5 Implement SignalStrategy

**File: `lib/signals.go`**

- [ ] Define `SignalStrategy` embedding `StrategyBase`:
  ```go
  type SignalStrategy struct {
      StrategyBase
      entrySignal []float64
      exitSignal  []float64
  }
  ```
- [ ] Implement `SignalStrategy.SetSignal(entrySize []float64, exitPortion []float64)`
- [ ] Override `SignalStrategy.Next()` to execute signals

### 4.6 Implement TrailingStrategy

**File: `lib/trailing.go`**

- [ ] Define `TrailingStrategy` embedding `StrategyBase`:
  ```go
  type TrailingStrategy struct {
      StrategyBase
      atrPeriod   int
      trailingSL  float64  // in ATR units
      trailingPct float64  // as percentage
  }
  ```
- [ ] Implement `TrailingStrategy.SetATRPeriods(n int)`
- [ ] Implement `TrailingStrategy.SetTrailingSL(nATR float64)`
- [ ] Implement `TrailingStrategy.SetTrailingPct(pct float64)`
- [ ] Override `TrailingStrategy.Next()` to manage trailing stops

### 4.7 Implement Common Indicators

**File: `lib/indicators.go`**

- [ ] Implement `RSI(prices []float64, period int) []float64`
- [ ] Implement `MACD(prices []float64, fast, slow, signal int) (macd, signal, hist []float64)`
- [ ] Implement `BollingerBands(prices []float64, period int, stdDev float64) (upper, middle, lower []float64)`
- [ ] Implement `ATR(high, low, close []float64, period int) []float64`
- [ ] Implement `ADX(high, low, close []float64, period int) []float64`
- [ ] Implement `Stochastic(high, low, close []float64, kPeriod, dPeriod int) (k, d []float64)`

**Acceptance Criteria for Phase 4:**
- [ ] Crossover functions detect crossings correctly
- [ ] Resampling produces correct higher-timeframe data
- [ ] SignalStrategy executes vectorized signals
- [ ] TrailingStrategy adjusts stop-losses correctly
- [ ] All indicators match reference implementations (e.g., ta-lib)
- [ ] Unit tests pass

---

## Phase 5: Optimization System

### 5.1 Implement Grid Search

**File: `optimize/grid.go`**

- [ ] Define `GridConfig` struct:
  ```go
  type GridConfig struct {
      Params      map[string][]interface{}  // parameter -> values
      Constraint  func(params map[string]interface{}) bool
      Maximize    string  // metric to maximize
      MaxTries    int     // 0 = all combinations
      RandomState int64   // for reproducibility
  }
  ```
- [ ] Implement `GridSearch(bt *Backtest, cfg GridConfig) (*OptimizeResult, error)`
- [ ] Generate all parameter combinations using cartesian product
- [ ] Apply constraint filter
- [ ] If MaxTries < total combinations, randomly sample
- [ ] Run backtests (see 5.2 for parallelization)
- [ ] Return best parameters and all results

### 5.2 Implement Parallel Execution

**File: `optimize/parallel.go`**

- [ ] Implement worker pool for parallel backtests:
  ```go
  type WorkerPool struct {
      workers int
      jobs    chan Job
      results chan Result
  }
  ```
- [ ] Use `runtime.NumCPU()` for default worker count
- [ ] Implement `RunParallel(bt *Backtest, paramSets []map[string]interface{}) []Result`
- [ ] Handle panics in workers gracefully

### 5.3 Implement Optimization Result

**File: `optimize/result.go`**

- [ ] Define `OptimizeResult` struct:
  ```go
  type OptimizeResult struct {
      BestParams map[string]interface{}
      BestValue  float64
      AllResults []ParamResult
  }

  type ParamResult struct {
      Params map[string]interface{}
      Stats  *Stats
      Value  float64
  }
  ```
- [ ] Implement `OptimizeResult.Heatmap(param1, param2 string) [][]float64`

### 5.4 Implement Constraint System

**File: `optimize/constraint.go`**

- [ ] Implement common constraint helpers:
  ```go
  func ParamLessThan(p1, p2 string) ConstraintFunc
  func ParamGreaterThan(p1, p2 string) ConstraintFunc
  func ParamNotEqual(p1, p2 string) ConstraintFunc
  ```
- [ ] Implement constraint combinator: `And(...ConstraintFunc) ConstraintFunc`

### 5.5 Implement Bayesian Optimization (Optional)

**File: `optimize/bayesian.go`**

- [ ] Research Go libraries for Bayesian optimization (e.g., go-bayesopt)
- [ ] Define `BayesianConfig` struct:
  ```go
  type BayesianConfig struct {
      Params     map[string]ParamBounds
      Maximize   string
      MaxTries   int
      InitPoints int
  }

  type ParamBounds struct {
      Min, Max float64
      Type     ParamType  // Continuous, Integer
  }
  ```
- [ ] Implement `BayesianSearch(bt *Backtest, cfg BayesianConfig) (*OptimizeResult, error)`
- [ ] Handle integer parameters via rounding

### 5.6 Implement Heatmap Generation

**File: `optimize/heatmap.go`**

- [ ] Implement `GenerateHeatmap(results []ParamResult, xParam, yParam, metric string) *Heatmap`
- [ ] Define `Heatmap` struct:
  ```go
  type Heatmap struct {
      XLabels []string
      YLabels []string
      Values  [][]float64
  }
  ```
- [ ] Implement `Heatmap.ToCSV(path string) error`

**Acceptance Criteria for Phase 5:**
- [ ] Grid search finds optimal parameters
- [ ] Constraints filter invalid combinations
- [ ] Parallel execution speeds up optimization
- [ ] Results match Python version for same parameter grid
- [ ] Heatmap data is correct
- [ ] Unit tests pass

---

## Phase 6: Plotting & Visualization

### 6.1 Design Plotting Architecture

- [ ] Choose approach:
  - **Option A:** Generate static HTML with embedded Chart.js/Plotly.js
  - **Option B:** Serve interactive charts via local HTTP server
  - **Option C:** Generate PNG/SVG images using Go charting library
  - **Recommended:** Option A for feature parity with Python/Bokeh

### 6.2 Implement Chart Data Generation

**File: `plot/data.go`**

- [ ] Define `ChartData` struct:
  ```go
  type ChartData struct {
      OHLCV       *OHLCVData
      Equity      []float64
      Drawdown    []float64
      Trades      []TradeMarker
      Indicators  []IndicatorData
      Volume      []float64
  }

  type TradeMarker struct {
      Time       time.Time
      Price      float64
      Type       string  // "entry_long", "entry_short", "exit_long", "exit_short"
      Size       float64
      PL         float64
  }

  type IndicatorData struct {
      Name    string
      Values  []float64
      Options IndicatorPlotOptions
  }
  ```
- [ ] Implement `PrepareChartData(stats *Stats, data *OHLCV, indicators []*Indicator) *ChartData`

### 6.3 Implement HTML Template

**File: `plot/templates/chart.html`**

- [ ] Create HTML template with:
  - Chart.js or Plotly.js CDN includes
  - OHLC candlestick chart
  - Equity curve subplot
  - Drawdown subplot (optional)
  - Volume bars (optional)
  - Trade markers overlay
  - Indicator overlays
  - Zoom/pan interaction
  - Responsive layout
- [ ] Use Go's `html/template` for data injection

### 6.4 Implement Plot Function

**File: `plot/plot.go`**

- [ ] Define `PlotConfig` struct:
  ```go
  type PlotConfig struct {
      Filename       string
      PlotEquity     bool
      PlotReturn     bool
      PlotPL         bool
      PlotVolume     bool
      PlotDrawdown   bool
      PlotTrades     bool
      SmoothEquity   int   // smoothing window
      RelativeEquity bool
      Superimpose    bool  // overlay equity on price
      ShowLegend     bool
      OpenBrowser    bool
  }
  ```
- [ ] Implement `Plot(stats *Stats, data *OHLCV, indicators []*Indicator, cfg PlotConfig) error`
- [ ] Generate HTML file
- [ ] Optionally open in browser using `exec.Command`

### 6.5 Implement Chart.js Candlestick Plugin

**File: `plot/assets/candlestick.js`**

- [ ] Implement or use existing OHLC candlestick plugin for Chart.js
- [ ] Alternatively, use Plotly.js which has native OHLC support

### 6.6 Implement Interactive Features

**File: `plot/assets/chart.js`**

- [ ] Implement crosshair with price/date display
- [ ] Implement zoom/pan synchronization across subplots
- [ ] Implement hover tooltips for trades
- [ ] Implement indicator visibility toggles

**Acceptance Criteria for Phase 6:**
- [ ] Can generate interactive HTML chart
- [ ] Candlestick chart displays correctly
- [ ] Equity curve overlays properly
- [ ] Trade markers show entry/exit points
- [ ] Indicators overlay on chart
- [ ] Charts are responsive and interactive
- [ ] Browser opening works cross-platform

---

## Phase 7: Testing & Validation

### 7.1 Set Up Test Infrastructure

- [ ] Copy test data files to `testdata/`:
  - `GOOG.csv`
  - `EURUSD.csv`
  - `BTCUSD.csv`
- [ ] Create test helper functions in `test_helpers_test.go`
- [ ] Implement `almostEqual(a, b, tolerance float64) bool`
- [ ] Implement `loadTestData(name string) *OHLCV`

### 7.2 Implement Data Package Tests

**File: `data/ohlcv_test.go`**

- [ ] Test `LoadCSV` with all test data files
- [ ] Test `OHLCV.Slice` edge cases
- [ ] Test `OHLCV.At` bounds checking
- [ ] Test `Series` operations

**File: `data/math_test.go`**

- [ ] Test `SMA` against known values
- [ ] Test `EMA` against known values
- [ ] Test `StdDev` against known values
- [ ] Test edge cases (empty, single value, all NaN)

### 7.3 Implement Core Engine Tests

**File: `backtest_test.go`**

- [ ] Test basic SMA crossover strategy
- [ ] Test market orders execution
- [ ] Test limit orders execution
- [ ] Test stop orders execution
- [ ] Test stop-limit orders
- [ ] Test SL/TP bracket orders
- [ ] Test position sizing
- [ ] Test commission calculations
- [ ] Test spread application
- [ ] Test margin requirements
- [ ] Test hedging mode
- [ ] Test exclusive orders mode
- [ ] Test trade finalization

**File: `broker_test.go`**

- [ ] Test order processing logic
- [ ] Test trade opening/closing
- [ ] Test position calculations
- [ ] Test equity tracking

### 7.4 Implement Stats Tests

**File: `stats/stats_test.go`**

- [ ] Test all metrics against Python reference values
- [ ] Test drawdown calculations
- [ ] Test trade statistics
- [ ] Test edge cases (no trades, single trade)

### 7.5 Implement Integration Tests

**File: `integration_test.go`**

- [ ] Run full backtest and compare stats to Python output
- [ ] Test optimization produces same results
- [ ] Test plotting generates valid HTML

### 7.6 Implement Benchmark Tests

**File: `benchmark_test.go`**

- [ ] Benchmark `Backtest.Run` with 10 years daily data
- [ ] Benchmark `GridSearch` with 100 parameter combinations
- [ ] Benchmark indicator calculations
- [ ] Compare performance to Python version

### 7.7 Golden File Testing

- [ ] Generate golden output files from Python version
- [ ] Compare Go output to golden files
- [ ] Document any intentional differences

**Acceptance Criteria for Phase 7:**
- [ ] All unit tests pass
- [ ] Integration tests match Python output within tolerance
- [ ] Benchmark shows acceptable performance
- [ ] Test coverage > 80%
- [ ] No data races detected by race detector

---

## Phase 8: Documentation & Examples

### 8.1 API Documentation

- [ ] Add godoc comments to all exported types and functions
- [ ] Document all Strategy interface methods
- [ ] Document all Backtest options
- [ ] Document optimization configuration

### 8.2 README.md

- [ ] Write project overview
- [ ] Installation instructions
- [ ] Quick start example
- [ ] Feature list
- [ ] Link to full documentation
- [ ] Performance comparison to Python

### 8.3 Example Strategies

**File: `examples/sma_cross/main.go`**

- [ ] Implement SMA crossover example
- [ ] Include comments explaining each step

**File: `examples/rsi_strategy/main.go`**

- [ ] Implement RSI overbought/oversold strategy
- [ ] Demonstrate indicator usage

**File: `examples/optimization/main.go`**

- [ ] Demonstrate grid search optimization
- [ ] Show how to generate heatmap

**File: `examples/multi_timeframe/main.go`**

- [ ] Demonstrate `ResampleApply` usage
- [ ] Show multi-timeframe analysis

### 8.4 Migration Guide

**File: `docs/migration.md`**

- [ ] Document differences from Python API
- [ ] Provide Python → Go code examples
- [ ] List any features not implemented

**Acceptance Criteria for Phase 8:**
- [ ] All examples compile and run
- [ ] godoc generates clean documentation
- [ ] README is clear and complete
- [ ] Migration guide covers major differences

---

## Architectural Decisions

### Go Idioms to Follow

1. **Accept interfaces, return structs**: Strategy is an interface, Backtest returns concrete *Stats
2. **Functional options pattern**: Use for configurable constructors (NewBacktest, NewBroker)
3. **Error handling**: Return errors explicitly, no panic except for programmer errors
4. **Concurrency**: Use goroutines and channels for optimization parallelization
5. **Package structure**: Keep packages focused and avoid circular dependencies

### Data Handling Strategy

1. **Use []float64 slices** instead of custom array types where possible
2. **Use gonum/mat** for matrix operations if needed in statistics
3. **Avoid reflection** except for parameter injection in optimization
4. **Pre-allocate slices** when size is known

### Testing Strategy

1. **Table-driven tests** for comprehensive coverage
2. **Golden file tests** for output validation against Python
3. **Benchmark tests** for performance regression detection
4. **Example tests** that also serve as documentation

---

## Dependency Mapping

| Python Library | Go Equivalent | Notes |
|----------------|---------------|-------|
| numpy | []float64 + gonum | Native slices for simple ops, gonum for complex |
| pandas | Custom OHLCV struct | No direct equivalent needed |
| pandas.DataFrame | map[string][]float64 | For stats output |
| bokeh | Chart.js/Plotly.js | JavaScript charting via HTML |
| multiprocessing | goroutines + channels | Native Go concurrency |
| abc | Go interfaces | Strategy interface |
| functools.lru_cache | sync.Map or custom | For memoization |
| sambo (Bayesian opt) | go-bayesopt or custom | Optional dependency |

---

## Estimated Complexity by Phase

| Phase | Complexity | Key Challenge |
|-------|------------|---------------|
| Phase 1 | Low | Data structure design |
| Phase 2 | High | Order matching logic, margin handling |
| Phase 3 | Medium | Statistical calculations accuracy |
| Phase 4 | Medium | Multi-timeframe alignment |
| Phase 5 | Medium | Parallel execution, parameter reflection |
| Phase 6 | High | JavaScript charting, interactivity |
| Phase 7 | Medium | Python parity validation |
| Phase 8 | Low | Documentation effort |

---

## Success Criteria

The conversion is complete when:

1. [ ] All 8 phases are implemented
2. [ ] Test coverage exceeds 80%
3. [ ] Integration tests match Python output within 0.01% tolerance
4. [ ] Performance is equal to or better than Python version
5. [ ] All examples run successfully
6. [ ] Documentation is complete

---

## Notes for LLM Execution

1. **Work phase by phase**: Complete each phase before moving to the next
2. **Run tests frequently**: After implementing each component
3. **Validate against Python**: Use Python version to generate expected outputs
4. **Ask for clarification**: If any task is ambiguous
5. **Commit incrementally**: After each major component is working
6. **Keep code idiomatic**: Follow Go conventions, not Python patterns

---

*Generated for backtesting.py → Go conversion project*
*Source: Python backtesting.py library (~3,825 lines)*
*Target: Idiomatic Go implementation with feature parity*
