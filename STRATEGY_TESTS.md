# Strategy Tests Implementation Plan

This document outlines a phased approach to implementing comprehensive strategy tests for the backtesting.go library. Each phase builds on the previous one, ensuring all library capabilities are thoroughly tested.

---

## Overview

**Goal:** Test every possible strategy type supported by the library to ensure correctness and reliability.

**Test Location:** `examples/strategies_test/`

**Test Data:** Use existing CSV files in `testdata/csv/` (GOOG.csv, EURUSD.csv, BTCUSD.csv)

**Module Path:** `github.com/izzoa/backtesting.go`

---

## File Structure

```
examples/
├── sma_cross/
│   ├── main.go
│   └── main_test.go           # Phase 1
├── rsi_optimization/
│   ├── main.go
│   └── main_test.go           # Phase 1
└── strategies_test/           # Phase 2-6
    ├── common_test.go         # Shared test utilities
    ├── trend_following_test.go
    ├── mean_reversion_test.go
    ├── breakout_test.go
    ├── multi_indicator_test.go
    ├── risk_management_test.go
    ├── order_types_test.go
    ├── position_management_test.go
    ├── short_strategies_test.go
    └── edge_cases_test.go
```

---

## Phase 1: Example Tests

### Task 1.1: Create `examples/sma_cross/main_test.go`

```go
// File: examples/sma_cross/main_test.go
package main

import (
    "testing"
    // imports...
)
```

**Tests to implement:**

- [ ] `TestSmaCross_Compiles` - Verify the strategy struct is valid
- [ ] `TestSmaCross_GOOG` - Run on GOOG data, verify:
  - Returns no error
  - NumTrades > 0
  - FinalEquity > 0
  - ReturnPct is not NaN
- [ ] `TestSmaCross_EURUSD` - Run on EURUSD data with same checks
- [ ] `TestSmaCross_BTCUSD` - Run on BTCUSD data with same checks
- [ ] `TestSmaCross_DifferentPeriods` - Test with periods (5,10), (10,20), (20,50)
- [ ] `BenchmarkSmaCross` - Benchmark execution time

**Assertions for each test:**
```go
func assertValidResults(t *testing.T, results *backtesting.BacktestResults, name string) {
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
```

---

### Task 1.2: Create `examples/rsi_optimization/main_test.go`

```go
// File: examples/rsi_optimization/main_test.go
package main

import (
    "testing"
    // imports...
)
```

**Tests to implement:**

- [ ] `TestRSIStrategy_Compiles` - Verify the strategy struct is valid
- [ ] `TestRSIStrategy_DefaultParams` - Run with Period=14, Overbought=70, Oversold=30
- [ ] `TestRSIStrategy_AggressiveParams` - Run with Period=7, Overbought=80, Oversold=20
- [ ] `TestRSIStrategy_ConservativeParams` - Run with Period=21, Overbought=65, Oversold=35
- [ ] `TestRSIOptimization_FindsBestParams` - Verify optimization returns results
- [ ] `TestRSIOptimization_TopN` - Verify TopN(5) returns 5 results
- [ ] `TestRSIOptimization_Constraints` - Verify constraints are applied
- [ ] `BenchmarkRSIStrategy` - Benchmark single run
- [ ] `BenchmarkRSIOptimization` - Benchmark optimization

---

## Phase 2: Trend Following Strategies

### Task 2.1: Create `examples/strategies_test/common_test.go`

**Shared utilities for all strategy tests:**

```go
// File: examples/strategies_test/common_test.go
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
func loadTestData(t *testing.T, filename string) *data.OHLCV {
    t.Helper()
    ohlcv, err := data.LoadCSV(getTestDataPath(filename))
    if err != nil {
        t.Fatalf("Failed to load %s: %v", filename, err)
    }
    return ohlcv
}

// runBacktest runs a backtest and returns results
func runBacktest(t *testing.T, ohlcv *data.OHLCV, strategy backtesting.Strategy) *backtesting.BacktestResults {
    t.Helper()
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

// assertValidResults checks common result validity
func assertValidResults(t *testing.T, results *backtesting.BacktestResults, strategyName string) {
    t.Helper()
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
func assertHasTrades(t *testing.T, results *backtesting.BacktestResults, strategyName string) {
    t.Helper()
    if results.NumTrades == 0 {
        t.Errorf("%s: expected trades but got 0", strategyName)
    }
}

// assertNoTrades checks that strategy produced no trades (for edge cases)
func assertNoTrades(t *testing.T, results *backtesting.BacktestResults, strategyName string) {
    t.Helper()
    if results.NumTrades != 0 {
        t.Errorf("%s: expected 0 trades but got %d", strategyName, results.NumTrades)
    }
}
```

---

### Task 2.2: Create `examples/strategies_test/trend_following_test.go`

**Strategies to implement and test:**

#### Strategy 1: SMA Crossover
```go
type SMACrossoverStrategy struct {
    backtesting.StrategyBase
    FastPeriod int
    SlowPeriod int
    fastSMA    *data.Indicator
    slowSMA    *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicators
- [ ] Implement `Next()` - Buy when fast > slow, sell when fast < slow
- [ ] Test: `TestSMACrossover_GOOG`
- [ ] Test: `TestSMACrossover_EURUSD`
- [ ] Test: `TestSMACrossover_MultiplePeriods`

#### Strategy 2: EMA Crossover
```go
type EMACrossoverStrategy struct {
    backtesting.StrategyBase
    FastPeriod int
    SlowPeriod int
    fastEMA    *data.Indicator
    slowEMA    *data.Indicator
}
```
- [ ] Implement `Init()` - Create EMA indicators
- [ ] Implement `Next()` - Buy when fast > slow, sell when fast < slow
- [ ] Test: `TestEMACrossover_GOOG`
- [ ] Test: `TestEMACrossover_EURUSD`

#### Strategy 3: Triple Moving Average
```go
type TripleMAStrategy struct {
    backtesting.StrategyBase
    FastPeriod   int
    MediumPeriod int
    SlowPeriod   int
    fastSMA      *data.Indicator
    mediumSMA    *data.Indicator
    slowSMA      *data.Indicator
}
```
- [ ] Implement `Init()` - Create 3 SMA indicators
- [ ] Implement `Next()` - Buy when fast > medium > slow, sell when any cross down
- [ ] Test: `TestTripleMA_GOOG`
- [ ] Test: `TestTripleMA_Alignment`

#### Strategy 4: MACD Crossover
```go
type MACDCrossoverStrategy struct {
    backtesting.StrategyBase
    FastPeriod   int
    SlowPeriod   int
    SignalPeriod int
    macdLine     *data.Indicator
    signalLine   *data.Indicator
}
```
- [ ] Implement `Init()` - Create MACD indicators using `lib.MACD()`
- [ ] Implement `Next()` - Buy when MACD > Signal, sell when MACD < Signal
- [ ] Test: `TestMACDCrossover_GOOG`
- [ ] Test: `TestMACDCrossover_DefaultParams`

#### Strategy 5: MACD Histogram
```go
type MACDHistogramStrategy struct {
    backtesting.StrategyBase
    FastPeriod   int
    SlowPeriod   int
    SignalPeriod int
    histogram    *data.Indicator
}
```
- [ ] Implement `Init()` - Create MACD histogram indicator
- [ ] Implement `Next()` - Buy when histogram > 0, sell when histogram < 0
- [ ] Test: `TestMACDHistogram_GOOG`

#### Strategy 6: ADX Trend
```go
type ADXTrendStrategy struct {
    backtesting.StrategyBase
    ADXPeriod    int
    SMAPeriod    int
    ADXThreshold float64
    adx          *data.Indicator
    sma          *data.Indicator
}
```
- [ ] Implement `Init()` - Create ADX and SMA indicators
- [ ] Implement `Next()` - Buy when ADX > threshold AND price > SMA, sell when ADX < threshold
- [ ] Test: `TestADXTrend_GOOG`
- [ ] Test: `TestADXTrend_DifferentThresholds`

---

## Phase 3: Mean Reversion Strategies

### Task 3.1: Create `examples/strategies_test/mean_reversion_test.go`

#### Strategy 7: RSI Oversold/Overbought
```go
type RSIStrategy struct {
    backtesting.StrategyBase
    Period     int
    Overbought float64
    Oversold   float64
    rsi        *data.Indicator
}
```
- [ ] Implement `Init()` - Create RSI indicator using `lib.RSI()`
- [ ] Implement `Next()` - Buy when RSI < Oversold, sell when RSI > Overbought
- [ ] Test: `TestRSI_Standard` (14, 70, 30)
- [ ] Test: `TestRSI_Aggressive` (7, 80, 20)
- [ ] Test: `TestRSI_Conservative` (21, 65, 35)

#### Strategy 8: RSI Extreme
```go
type RSIExtremeStrategy struct {
    backtesting.StrategyBase
    Period     int
    Overbought float64
    Oversold   float64
    rsi        *data.Indicator
}
```
- [ ] Implement with extreme levels (RSI < 10, RSI > 90)
- [ ] Test: `TestRSIExtreme_GOOG`
- [ ] Test: `TestRSIExtreme_RareSignals`

#### Strategy 9: Bollinger Band Mean Reversion
```go
type BollingerMeanReversionStrategy struct {
    backtesting.StrategyBase
    Period int
    StdDev float64
    upper  *data.Indicator
    middle *data.Indicator
    lower  *data.Indicator
}
```
- [ ] Implement `Init()` - Create Bollinger Bands using `lib.BollingerBands()`
- [ ] Implement `Next()` - Buy when price < lower, sell when price > upper
- [ ] Test: `TestBollingerMeanReversion_GOOG`
- [ ] Test: `TestBollingerMeanReversion_TightBands` (1.5 stddev)
- [ ] Test: `TestBollingerMeanReversion_WideBands` (2.5 stddev)

#### Strategy 10: Stochastic Oversold/Overbought
```go
type StochasticStrategy struct {
    backtesting.StrategyBase
    KPeriod    int
    DPeriod    int
    Overbought float64
    Oversold   float64
    k          *data.Indicator
    d          *data.Indicator
}
```
- [ ] Implement `Init()` - Create Stochastic indicators using `lib.Stochastic()`
- [ ] Implement `Next()` - Buy when K < Oversold, sell when K > Overbought
- [ ] Test: `TestStochastic_Standard`
- [ ] Test: `TestStochastic_WithDLine` (use D for confirmation)

#### Strategy 11: CCI Oversold/Overbought
```go
type CCIStrategy struct {
    backtesting.StrategyBase
    Period     int
    Overbought float64
    Oversold   float64
    cci        *data.Indicator
}
```
- [ ] Implement `Init()` - Create CCI indicator using `lib.CCI()`
- [ ] Implement `Next()` - Buy when CCI < -100, sell when CCI > 100
- [ ] Test: `TestCCI_Standard`
- [ ] Test: `TestCCI_ExtremeLevels` (-200, 200)

#### Strategy 12: Williams %R
```go
type WilliamsRStrategy struct {
    backtesting.StrategyBase
    Period     int
    Overbought float64
    Oversold   float64
    willR      *data.Indicator
}
```
- [ ] Implement `Init()` - Create Williams %R indicator using `lib.WilliamsR()`
- [ ] Implement `Next()` - Buy when %R < -80, sell when %R > -20
- [ ] Test: `TestWilliamsR_Standard`
- [ ] Test: `TestWilliamsR_EURUSD`

---

## Phase 4: Breakout Strategies

### Task 4.1: Create `examples/strategies_test/breakout_test.go`

#### Strategy 13: Bollinger Band Breakout
```go
type BollingerBreakoutStrategy struct {
    backtesting.StrategyBase
    Period int
    StdDev float64
    upper  *data.Indicator
    middle *data.Indicator
    lower  *data.Indicator
}
```
- [ ] Implement `Init()` - Create Bollinger Bands
- [ ] Implement `Next()` - Buy when price > upper, sell when price < middle
- [ ] Test: `TestBollingerBreakout_GOOG`
- [ ] Test: `TestBollingerBreakout_BTCUSD` (volatile asset)

#### Strategy 14: ATR Breakout
```go
type ATRBreakoutStrategy struct {
    backtesting.StrategyBase
    Period     int
    Multiplier float64
    atr        *data.Indicator
    prevClose  float64
}
```
- [ ] Implement `Init()` - Create ATR indicator using `lib.ATR()`
- [ ] Implement `Next()` - Buy when price > prevClose + (Multiplier * ATR)
- [ ] Test: `TestATRBreakout_GOOG`
- [ ] Test: `TestATRBreakout_DifferentMultipliers` (1.5, 2.0, 2.5)

#### Strategy 15: Donchian Channel (High/Low Breakout)
```go
type DonchianChannelStrategy struct {
    backtesting.StrategyBase
    EntryPeriod int
    ExitPeriod  int
}
```
- [ ] Implement `Init()` - No indicators needed, use raw price data
- [ ] Implement `Next()` - Buy at N-bar high, sell at M-bar low
- [ ] Test: `TestDonchian_20_10` (20-bar entry, 10-bar exit)
- [ ] Test: `TestDonchian_55_20` (turtle trading style)

#### Strategy 16: Volatility Expansion
```go
type VolatilityExpansionStrategy struct {
    backtesting.StrategyBase
    BBPeriod       int
    BBStdDev       float64
    ATRPeriod      int
    ExpansionRatio float64
    upper          *data.Indicator
    lower          *data.Indicator
    atr            *data.Indicator
}
```
- [ ] Implement `Init()` - Create BB and ATR indicators
- [ ] Implement `Next()` - Buy when BB width expanding rapidly
- [ ] Test: `TestVolatilityExpansion_GOOG`
- [ ] Test: `TestVolatilityExpansion_BTCUSD`

---

## Phase 5: Multi-Indicator Strategies

### Task 5.1: Create `examples/strategies_test/multi_indicator_test.go`

#### Strategy 17: RSI + SMA Filter
```go
type RSISMAFilterStrategy struct {
    backtesting.StrategyBase
    RSIPeriod  int
    SMAPeriod  int
    Overbought float64
    Oversold   float64
    rsi        *data.Indicator
    sma        *data.Indicator
}
```
- [ ] Implement `Init()` - Create RSI and SMA indicators
- [ ] Implement `Next()` - Buy when RSI < 30 AND price > SMA (trend filter)
- [ ] Test: `TestRSISMAFilter_GOOG`
- [ ] Test: `TestRSISMAFilter_FilterEffect` (compare with/without filter)

#### Strategy 18: MACD + RSI
```go
type MACDRSIStrategy struct {
    backtesting.StrategyBase
    MACDFast   int
    MACDSlow   int
    MACDSignal int
    RSIPeriod  int
    RSILevel   float64
    macdLine   *data.Indicator
    signalLine *data.Indicator
    rsi        *data.Indicator
}
```
- [ ] Implement `Init()` - Create MACD and RSI indicators
- [ ] Implement `Next()` - Buy when MACD crosses up AND RSI < 50
- [ ] Test: `TestMACDRSI_GOOG`
- [ ] Test: `TestMACDRSI_Confirmation`

#### Strategy 19: Stochastic + ADX
```go
type StochasticADXStrategy struct {
    backtesting.StrategyBase
    StochK       int
    StochD       int
    ADXPeriod    int
    ADXThreshold float64
    Oversold     float64
    k            *data.Indicator
    adx          *data.Indicator
}
```
- [ ] Implement `Init()` - Create Stochastic and ADX indicators
- [ ] Implement `Next()` - Buy when K < 20 AND ADX > 25
- [ ] Test: `TestStochasticADX_GOOG`
- [ ] Test: `TestStochasticADX_TrendStrength`

#### Strategy 20: Bollinger + RSI
```go
type BollingerRSIStrategy struct {
    backtesting.StrategyBase
    BBPeriod   int
    BBStdDev   float64
    RSIPeriod  int
    RSIOversold float64
    upper      *data.Indicator
    lower      *data.Indicator
    rsi        *data.Indicator
}
```
- [ ] Implement `Init()` - Create BB and RSI indicators
- [ ] Implement `Next()` - Buy when price < lower AND RSI < 30
- [ ] Test: `TestBollingerRSI_GOOG`
- [ ] Test: `TestBollingerRSI_DoubleConfirmation`

#### Strategy 21: EMA + ATR
```go
type EMAATRStrategy struct {
    backtesting.StrategyBase
    EMAPeriod     int
    ATRPeriod     int
    ATRMultiplier float64
    ema           *data.Indicator
    atr           *data.Indicator
}
```
- [ ] Implement `Init()` - Create EMA and ATR indicators
- [ ] Implement `Next()` - Buy when price > EMA + ATR, sell when price < EMA - ATR
- [ ] Test: `TestEMAATR_GOOG`
- [ ] Test: `TestEMAATR_DifferentMultipliers`

#### Strategy 22: Triple Confirmation
```go
type TripleConfirmationStrategy struct {
    backtesting.StrategyBase
    SMAPeriod    int
    RSIPeriod    int
    MACDFast     int
    MACDSlow     int
    MACDSignal   int
    sma          *data.Indicator
    rsi          *data.Indicator
    macdLine     *data.Indicator
    signalLine   *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA, RSI, MACD indicators
- [ ] Implement `Next()` - Buy when ALL signals bullish, sell when ANY bearish
- [ ] Test: `TestTripleConfirmation_GOOG`
- [ ] Test: `TestTripleConfirmation_StrictEntry`

---

## Phase 6: Risk Management & Order Types

### Task 6.1: Create `examples/strategies_test/risk_management_test.go`

#### Strategy 23: Fixed Stop Loss
```go
type FixedStopLossStrategy struct {
    backtesting.StrategyBase
    SMAPeriod    int
    StopLossPct  float64
    sma          *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Use `s.Buy(backtesting.WithSL(stopPrice))`
- [ ] Test: `TestFixedStopLoss_Triggered`
- [ ] Test: `TestFixedStopLoss_NotTriggered`
- [ ] Test: `TestFixedStopLoss_DifferentLevels` (1%, 2%, 5%)

#### Strategy 24: ATR Stop Loss
```go
type ATRStopLossStrategy struct {
    backtesting.StrategyBase
    SMAPeriod     int
    ATRPeriod     int
    ATRMultiplier float64
    sma           *data.Indicator
    atr           *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA and ATR indicators
- [ ] Implement `Next()` - Calculate stop as entry - (ATR * multiplier)
- [ ] Test: `TestATRStopLoss_GOOG`
- [ ] Test: `TestATRStopLoss_VolatileAsset` (BTCUSD)

#### Strategy 25: Fixed Take Profit
```go
type FixedTakeProfitStrategy struct {
    backtesting.StrategyBase
    SMAPeriod      int
    TakeProfitPct  float64
    sma            *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Use `s.Buy(backtesting.WithTP(targetPrice))`
- [ ] Test: `TestFixedTakeProfit_Triggered`
- [ ] Test: `TestFixedTakeProfit_DifferentLevels` (2%, 5%, 10%)

#### Strategy 26: Risk/Reward Ratio
```go
type RiskRewardStrategy struct {
    backtesting.StrategyBase
    SMAPeriod   int
    RiskPct     float64
    RewardRatio float64  // e.g., 2.0 for 1:2 risk/reward
    sma         *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Use `s.Buy(backtesting.WithSL(sl), backtesting.WithTP(tp))`
- [ ] Test: `TestRiskReward_1to2`
- [ ] Test: `TestRiskReward_1to3`
- [ ] Test: `TestRiskReward_WinRateVsRatio`

#### Strategy 27: Trailing Stop (Manual Implementation)
```go
type TrailingStopStrategy struct {
    backtesting.StrategyBase
    SMAPeriod     int
    TrailPct      float64
    sma           *data.Indicator
    highestPrice  float64
    inPosition    bool
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Track highest price, exit if price drops by TrailPct
- [ ] Test: `TestTrailingStop_GOOG`
- [ ] Test: `TestTrailingStop_LocksInProfits`

---

### Task 6.2: Create `examples/strategies_test/order_types_test.go`

#### Strategy 28: Limit Entry
```go
type LimitEntryStrategy struct {
    backtesting.StrategyBase
    SMAPeriod    int
    LimitOffset  float64  // Buy X% below current price
    sma          *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Use `s.Buy(backtesting.WithLimit(limitPrice))`
- [ ] Test: `TestLimitEntry_OrderPlaced`
- [ ] Test: `TestLimitEntry_OrderFilled`
- [ ] Test: `TestLimitEntry_OrderNotFilled`

#### Strategy 29: Stop Entry (Breakout)
```go
type StopEntryStrategy struct {
    backtesting.StrategyBase
    LookbackPeriod int
    stopPrice      float64
}
```
- [ ] Implement `Init()` - No indicators needed
- [ ] Implement `Next()` - Use `s.Buy(backtesting.WithStop(breakoutPrice))`
- [ ] Test: `TestStopEntry_BreakoutTriggered`
- [ ] Test: `TestStopEntry_NoBreakout`

#### Strategy 30: Stop-Limit Order
```go
type StopLimitStrategy struct {
    backtesting.StrategyBase
    LookbackPeriod int
    LimitOffset    float64
}
```
- [ ] Implement `Init()` - No indicators needed
- [ ] Implement `Next()` - Use `s.Buy(backtesting.WithStop(stop), backtesting.WithLimit(limit))`
- [ ] Test: `TestStopLimit_BothTriggered`
- [ ] Test: `TestStopLimit_StopOnly`

#### Strategy 31: Bracket Order
```go
type BracketOrderStrategy struct {
    backtesting.StrategyBase
    SMAPeriod   int
    StopLossPct float64
    TakeProfitPct float64
    sma         *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Use `s.Buy(backtesting.WithSL(sl), backtesting.WithTP(tp))`
- [ ] Test: `TestBracketOrder_SLTriggered`
- [ ] Test: `TestBracketOrder_TPTriggered`
- [ ] Test: `TestBracketOrder_NeitherTriggered`

---

### Task 6.3: Create `examples/strategies_test/position_management_test.go`

#### Strategy 32: Scale Out
```go
type ScaleOutStrategy struct {
    backtesting.StrategyBase
    SMAPeriod     int
    FirstTarget   float64  // Close 50% at this profit %
    sma           *data.Indicator
    scaledOut     bool
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Use `s.Position().Close(0.5)` at first target
- [ ] Test: `TestScaleOut_PartialClose`
- [ ] Test: `TestScaleOut_FullExit`

#### Strategy 33: Add to Winners (Pyramiding)
```go
type PyramidingStrategy struct {
    backtesting.StrategyBase
    SMAPeriod     int
    AddThreshold  float64  // Add when profit exceeds this %
    MaxPositions  int
    sma           *data.Indicator
    positionCount int
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Add to position when winning
- [ ] Test: `TestPyramiding_AddsPosition`
- [ ] Test: `TestPyramiding_MaxPositions`

#### Strategy 34: Scale In
```go
type ScaleInStrategy struct {
    backtesting.StrategyBase
    SMAPeriod  int
    Tranches   int
    sma        *data.Indicator
    entryCount int
}
```
- [ ] Implement `Init()` - Create SMA indicator
- [ ] Implement `Next()` - Enter in multiple tranches
- [ ] Test: `TestScaleIn_MultipleEntries`

---

### Task 6.4: Create `examples/strategies_test/short_strategies_test.go`

#### Strategy 35: Short RSI
```go
type ShortRSIStrategy struct {
    backtesting.StrategyBase
    Period     int
    Overbought float64
    Oversold   float64
    rsi        *data.Indicator
}
```
- [ ] Implement `Init()` - Create RSI indicator
- [ ] Implement `Next()` - Short when RSI > Overbought, cover when RSI < Oversold
- [ ] Test: `TestShortRSI_GOOG`
- [ ] Test: `TestShortRSI_OpensShort`

#### Strategy 36: Short SMA
```go
type ShortSMAStrategy struct {
    backtesting.StrategyBase
    FastPeriod int
    SlowPeriod int
    fastSMA    *data.Indicator
    slowSMA    *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicators
- [ ] Implement `Next()` - Short when fast < slow
- [ ] Test: `TestShortSMA_GOOG`
- [ ] Test: `TestShortSMA_ShortPosition`

#### Strategy 37: Long/Short Switching
```go
type LongShortSwitchStrategy struct {
    backtesting.StrategyBase
    FastPeriod int
    SlowPeriod int
    fastSMA    *data.Indicator
    slowSMA    *data.Indicator
}
```
- [ ] Implement `Init()` - Create SMA indicators
- [ ] Implement `Next()` - Always in market, either long or short
- [ ] Test: `TestLongShortSwitch_GOOG`
- [ ] Test: `TestLongShortSwitch_AlwaysInMarket`
- [ ] Test: `TestLongShortSwitch_SwitchesCorrectly`

---

## Phase 7: Edge Cases

### Task 7.1: Create `examples/strategies_test/edge_cases_test.go`

#### Strategy 38: Buy and Hold
```go
type BuyAndHoldStrategy struct {
    backtesting.StrategyBase
    entered bool
}
```
- [ ] Implement `Init()` - No indicators
- [ ] Implement `Next()` - Buy on first bar, never sell
- [ ] Test: `TestBuyAndHold_GOOG`
- [ ] Test: `TestBuyAndHold_SingleTrade`
- [ ] Test: `TestBuyAndHold_MatchesBenchmark`

#### Strategy 39: Do Nothing
```go
type DoNothingStrategy struct {
    backtesting.StrategyBase
}
```
- [ ] Implement `Init()` - Empty
- [ ] Implement `Next()` - Empty
- [ ] Test: `TestDoNothing_NoTrades`
- [ ] Test: `TestDoNothing_EquityUnchanged`
- [ ] Test: `TestDoNothing_NoErrors`

#### Strategy 40: Trade Every Bar
```go
type TradeEveryBarStrategy struct {
    backtesting.StrategyBase
}
```
- [ ] Implement `Init()` - Empty
- [ ] Implement `Next()` - Alternate buy/sell every bar
- [ ] Test: `TestTradeEveryBar_ManyTrades`
- [ ] Test: `TestTradeEveryBar_NoErrors`
- [ ] Test: `TestTradeEveryBar_HighCommissionImpact`

#### Strategy 41: Single Trade
```go
type SingleTradeStrategy struct {
    backtesting.StrategyBase
    EntryBar int
    ExitBar  int
    entered  bool
    exited   bool
}
```
- [ ] Implement `Init()` - Empty
- [ ] Implement `Next()` - Enter at bar N, exit at bar M
- [ ] Test: `TestSingleTrade_ExactlyOneTrade`
- [ ] Test: `TestSingleTrade_CorrectBars`

#### Strategy 42: All Indicators Registered
```go
type AllIndicatorsStrategy struct {
    backtesting.StrategyBase
    // All indicator types
    sma        *data.Indicator
    ema        *data.Indicator
    rsi        *data.Indicator
    macd       *data.Indicator
    signal     *data.Indicator
    histogram  *data.Indicator
    bbUpper    *data.Indicator
    bbMiddle   *data.Indicator
    bbLower    *data.Indicator
    atr        *data.Indicator
    adx        *data.Indicator
    stochK     *data.Indicator
    stochD     *data.Indicator
    cci        *data.Indicator
    willR      *data.Indicator
}
```
- [ ] Implement `Init()` - Register ALL indicator types
- [ ] Implement `Next()` - Simple logic using one indicator
- [ ] Test: `TestAllIndicators_Compiles`
- [ ] Test: `TestAllIndicators_NoMemoryIssues`
- [ ] Test: `TestAllIndicators_AllValuesAccessible`

---

## Phase 8: Benchmarks

### Task 8.1: Add benchmarks to each test file

**Add to each `*_test.go` file:**

```go
func BenchmarkStrategyName(b *testing.B) {
    ohlcv := loadBenchData(b)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        strategy := &StrategyName{...}
        bt := backtesting.NewBacktest(ohlcv, strategy, backtesting.BacktestConfig{
            Cash: 10000,
        })
        _, _ = bt.Run()
    }
}
```

- [ ] `BenchmarkSMACrossover`
- [ ] `BenchmarkEMACrossover`
- [ ] `BenchmarkRSI`
- [ ] `BenchmarkMACD`
- [ ] `BenchmarkBollingerBands`
- [ ] `BenchmarkMultiIndicator`
- [ ] `BenchmarkAllIndicators`

---

## Verification Checklist

After implementing all phases, verify:

- [ ] `go test ./examples/...` passes
- [ ] `go test -race ./examples/...` passes (no race conditions)
- [ ] `go test -bench=. ./examples/...` completes
- [ ] Each strategy type has at least one passing test
- [ ] Each indicator is used in at least one strategy
- [ ] All order types (limit, stop, SL, TP) are tested
- [ ] Both long and short positions are tested
- [ ] Edge cases (no trades, many trades) are tested
- [ ] Code coverage > 80% for strategy logic

---

## Summary

| Phase | Strategies | Tests | Priority |
|-------|------------|-------|----------|
| 1 | 2 (existing examples) | ~15 | High |
| 2 | 6 (trend following) | ~20 | High |
| 3 | 6 (mean reversion) | ~18 | High |
| 4 | 4 (breakout) | ~12 | Medium |
| 5 | 6 (multi-indicator) | ~18 | Medium |
| 6 | 10 (risk/orders/position) | ~30 | Medium |
| 7 | 5 (edge cases) | ~15 | High |
| 8 | Benchmarks | ~10 | Low |

**Total: 39 strategies, ~138 tests**
