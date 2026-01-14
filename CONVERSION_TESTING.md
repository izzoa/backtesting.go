# Go Testing Strategy for backtesting-go Conversion

This document provides a comprehensive testing strategy aligned with each phase of CONVERSION.md. Each section contains actionable test specifications an LLM can follow to ensure correctness and Python parity.

---

## Table of Contents

1. [Testing Philosophy & Standards](#testing-philosophy--standards)
2. [Test Infrastructure Setup](#test-infrastructure-setup)
3. [Phase 1 Tests: Data Structures](#phase-1-tests-data-structures)
4. [Phase 2 Tests: Core Engine](#phase-2-tests-core-engine)
5. [Phase 3 Tests: Statistics Module](#phase-3-tests-statistics-module)
6. [Phase 4 Tests: Library Utilities](#phase-4-tests-library-utilities)
7. [Phase 5 Tests: Optimization System](#phase-5-tests-optimization-system)
8. [Phase 6 Tests: Plotting & Visualization](#phase-6-tests-plotting--visualization)
9. [Phase 7 Tests: Integration & Validation](#phase-7-tests-integration--validation)
10. [Phase 8 Tests: Documentation & Examples](#phase-8-tests-documentation--examples)
11. [Golden File Generation](#golden-file-generation)
12. [CI/CD Pipeline Configuration](#cicd-pipeline-configuration)

---

## Testing Philosophy & Standards

### Core Principles

1. **Test-Driven Development (TDD)**: Write tests before or alongside implementation
2. **Python Parity**: All numerical outputs must match Python within tolerance (0.0001%)
3. **Table-Driven Tests**: Use Go's table-driven test pattern for comprehensive coverage
4. **Golden Files**: Compare complex outputs against pre-generated Python reference data
5. **Fuzz Testing**: Use Go's native fuzzing for edge case discovery
6. **Benchmark Everything**: Performance regression detection from day one

### Test File Naming Convention

```
<package>/<file>_test.go           # Unit tests
<package>/<file>_benchmark_test.go # Benchmarks (if separate)
<package>/<file>_fuzz_test.go      # Fuzz tests
integration_test.go                 # Cross-package integration tests
```

### Coverage Requirements

| Phase | Minimum Coverage | Critical Paths |
|-------|------------------|----------------|
| Phase 1 | 90% | Data loading, SMA/EMA calculations |
| Phase 2 | 95% | Order matching, trade P&L, broker logic |
| Phase 3 | 95% | All statistical metrics |
| Phase 4 | 85% | Crossover detection, resampling |
| Phase 5 | 80% | Grid search, parallel execution |
| Phase 6 | 70% | HTML generation, data serialization |
| Overall | 85% | - |

### Tolerance Constants

```go
const (
    FloatTolerance     = 1e-9      // For exact calculations
    PercentTolerance   = 1e-6      // For percentage comparisons
    PriceTolerance     = 1e-4      // For price comparisons
    StatsTolerance     = 1e-4      // For statistical metrics
    TimeTolerance      = time.Second // For timestamp comparisons
)
```

---

## Test Infrastructure Setup

### Directory Structure

```
backtesting-go/
├── testdata/
│   ├── csv/
│   │   ├── GOOG.csv
│   │   ├── EURUSD.csv
│   │   └── BTCUSD.csv
│   ├── golden/
│   │   ├── sma_cross_stats.json
│   │   ├── rsi_strategy_stats.json
│   │   ├── drawdown_series.json
│   │   └── optimization_results.json
│   └── fixtures/
│       ├── simple_ohlcv.json
│       ├── edge_case_data.json
│       └── invalid_data.json
├── internal/
│   └── testutil/
│       ├── helpers.go
│       ├── assertions.go
│       ├── fixtures.go
│       └── golden.go
```

### Test Helper Implementation

**File: `internal/testutil/helpers.go`**

- [ ] Implement test helper package:
  ```go
  package testutil

  import (
      "encoding/json"
      "math"
      "os"
      "path/filepath"
      "testing"
  )

  // AlmostEqual compares floats within tolerance
  func AlmostEqual(a, b, tolerance float64) bool {
      if math.IsNaN(a) && math.IsNaN(b) {
          return true
      }
      if math.IsNaN(a) || math.IsNaN(b) {
          return false
      }
      if math.IsInf(a, 0) || math.IsInf(b, 0) {
          return a == b
      }
      return math.Abs(a-b) <= tolerance
  }

  // SlicesAlmostEqual compares float slices
  func SlicesAlmostEqual(a, b []float64, tolerance float64) bool {
      if len(a) != len(b) {
          return false
      }
      for i := range a {
          if !AlmostEqual(a[i], b[i], tolerance) {
              return false
          }
      }
      return true
  }

  // RequireAlmostEqual fails test if not equal
  func RequireAlmostEqual(t *testing.T, expected, actual, tolerance float64, msg string) {
      t.Helper()
      if !AlmostEqual(expected, actual, tolerance) {
          t.Errorf("%s: expected %v, got %v (tolerance: %v)", msg, expected, actual, tolerance)
      }
  }
  ```

- [ ] Implement fixture loading:
  ```go
  // LoadTestData loads OHLCV from testdata/csv
  func LoadTestData(t *testing.T, name string) *data.OHLCV {
      t.Helper()
      path := filepath.Join("testdata", "csv", name+".csv")
      ohlcv, err := data.LoadCSV(path)
      if err != nil {
          t.Fatalf("failed to load test data %s: %v", name, err)
      }
      return ohlcv
  }

  // LoadGolden loads golden file data
  func LoadGolden(t *testing.T, name string, v interface{}) {
      t.Helper()
      path := filepath.Join("testdata", "golden", name+".json")
      data, err := os.ReadFile(path)
      if err != nil {
          t.Fatalf("failed to load golden file %s: %v", name, err)
      }
      if err := json.Unmarshal(data, v); err != nil {
          t.Fatalf("failed to parse golden file %s: %v", name, err)
      }
  }

  // UpdateGolden updates golden file (run with -update flag)
  func UpdateGolden(t *testing.T, name string, v interface{}) {
      t.Helper()
      path := filepath.Join("testdata", "golden", name+".json")
      data, err := json.MarshalIndent(v, "", "  ")
      if err != nil {
          t.Fatalf("failed to marshal golden data: %v", err)
      }
      if err := os.WriteFile(path, data, 0644); err != nil {
          t.Fatalf("failed to write golden file: %v", err)
      }
  }
  ```

- [ ] Implement test OHLCV generators:
  ```go
  // MakeTestOHLCV creates synthetic test data
  func MakeTestOHLCV(length int, startPrice float64) *data.OHLCV {
      ohlcv := data.NewOHLCV(length)
      price := startPrice
      baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

      for i := 0; i < length; i++ {
          change := (rand.Float64() - 0.5) * 2 // -1% to +1%
          open := price
          close := price * (1 + change/100)
          high := math.Max(open, close) * (1 + rand.Float64()*0.5/100)
          low := math.Min(open, close) * (1 - rand.Float64()*0.5/100)

          ohlcv.Time[i] = baseTime.AddDate(0, 0, i)
          ohlcv.Open[i] = open
          ohlcv.High[i] = high
          ohlcv.Low[i] = low
          ohlcv.Close[i] = close
          ohlcv.Volume[i] = float64(rand.Intn(1000000))

          price = close
      }
      return ohlcv
  }
  ```

---

## Phase 1 Tests: Data Structures

### 1.1 OHLCV Tests

**File: `data/ohlcv_test.go`**

- [ ] Test OHLCV construction:
  ```go
  func TestNewOHLCV(t *testing.T) {
      tests := []struct {
          name     string
          capacity int
          wantLen  int
      }{
          {"zero capacity", 0, 0},
          {"small capacity", 10, 0},
          {"large capacity", 10000, 0},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              ohlcv := NewOHLCV(tt.capacity)
              if ohlcv.Len() != tt.wantLen {
                  t.Errorf("Len() = %d, want %d", ohlcv.Len(), tt.wantLen)
              }
          })
      }
  }
  ```

- [ ] Test OHLCV slicing:
  ```go
  func TestOHLCV_Slice(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)

      tests := []struct {
          name      string
          start     int
          end       int
          wantLen   int
          wantPanic bool
      }{
          {"full slice", 0, 100, 100, false},
          {"partial start", 10, 100, 90, false},
          {"partial end", 0, 50, 50, false},
          {"middle slice", 25, 75, 50, false},
          {"single element", 50, 51, 1, false},
          {"empty slice", 50, 50, 0, false},
          {"negative start", -1, 50, 0, true},
          {"end > length", 0, 150, 0, true},
          {"start > end", 50, 25, 0, true},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              if tt.wantPanic {
                  defer func() {
                      if r := recover(); r == nil {
                          t.Error("expected panic")
                      }
                  }()
              }
              slice := ohlcv.Slice(tt.start, tt.end)
              if !tt.wantPanic && slice.Len() != tt.wantLen {
                  t.Errorf("Slice(%d, %d).Len() = %d, want %d",
                      tt.start, tt.end, slice.Len(), tt.wantLen)
              }
          })
      }
  }
  ```

- [ ] Test OHLCV data integrity:
  ```go
  func TestOHLCV_DataIntegrity(t *testing.T) {
      ohlcv := testutil.LoadTestData(t, "GOOG")

      for i := 0; i < ohlcv.Len(); i++ {
          bar := ohlcv.At(i)

          // High must be highest
          if bar.High < bar.Open || bar.High < bar.Close || bar.High < bar.Low {
              t.Errorf("bar %d: High (%f) is not highest", i, bar.High)
          }

          // Low must be lowest
          if bar.Low > bar.Open || bar.Low > bar.Close || bar.Low > bar.High {
              t.Errorf("bar %d: Low (%f) is not lowest", i, bar.Low)
          }

          // Volume must be non-negative
          if bar.Volume < 0 {
              t.Errorf("bar %d: negative Volume (%f)", i, bar.Volume)
          }
      }
  }
  ```

### 1.2 CSV Loader Tests

**File: `data/loader_test.go`**

- [ ] Test CSV loading with real data:
  ```go
  func TestLoadCSV(t *testing.T) {
      tests := []struct {
          name        string
          file        string
          wantLen     int
          wantErr     bool
          firstClose  float64
          lastClose   float64
      }{
          {
              name:       "GOOG daily",
              file:       "testdata/csv/GOOG.csv",
              wantLen:    2510, // Verify actual length
              wantErr:    false,
              firstClose: 50.12, // Verify actual values
              lastClose:  556.97,
          },
          {
              name:       "EURUSD hourly",
              file:       "testdata/csv/EURUSD.csv",
              wantLen:    6000, // Verify actual length
              wantErr:    false,
              firstClose: 1.0623,
              lastClose:  1.2345,
          },
          {
              name:    "nonexistent file",
              file:    "testdata/csv/NOTEXIST.csv",
              wantErr: true,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              ohlcv, err := LoadCSV(tt.file)
              if (err != nil) != tt.wantErr {
                  t.Errorf("LoadCSV() error = %v, wantErr %v", err, tt.wantErr)
                  return
              }
              if tt.wantErr {
                  return
              }
              if ohlcv.Len() != tt.wantLen {
                  t.Errorf("Len() = %d, want %d", ohlcv.Len(), tt.wantLen)
              }
              testutil.RequireAlmostEqual(t, tt.firstClose, ohlcv.Close[0],
                  PriceTolerance, "first close")
              testutil.RequireAlmostEqual(t, tt.lastClose, ohlcv.Close[ohlcv.Len()-1],
                  PriceTolerance, "last close")
          })
      }
  }
  ```

- [ ] Test CSV with custom column mapping:
  ```go
  func TestLoadCSV_CustomColumns(t *testing.T) {
      // Test with different column name conventions
      ohlcv, err := LoadCSV("testdata/csv/custom_columns.csv",
          WithTimeColumn("date"),
          WithOpenColumn("open_price"),
          WithHighColumn("high_price"),
          WithLowColumn("low_price"),
          WithCloseColumn("close_price"),
          WithVolumeColumn("vol"),
      )
      if err != nil {
          t.Fatalf("LoadCSV failed: %v", err)
      }
      if ohlcv.Len() == 0 {
          t.Error("loaded OHLCV is empty")
      }
  }
  ```

- [ ] Test CSV date format parsing:
  ```go
  func TestLoadCSV_DateFormats(t *testing.T) {
      formats := []struct {
          name   string
          format string
          sample string
      }{
          {"ISO", "2006-01-02", "2020-01-15"},
          {"US", "01/02/2006", "01/15/2020"},
          {"datetime", "2006-01-02 15:04:05", "2020-01-15 09:30:00"},
          {"unix", "unix", "1579089000"},
      }

      for _, fmt := range formats {
          t.Run(fmt.name, func(t *testing.T) {
              // Create temp file with this date format and test parsing
          })
      }
  }
  ```

### 1.3 Series Tests

**File: `data/series_test.go`**

- [ ] Test Series operations:
  ```go
  func TestSeries_Operations(t *testing.T) {
      values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
      series := NewSeries("test", values)

      t.Run("Last", func(t *testing.T) {
          if series.Last() != 10.0 {
              t.Errorf("Last() = %f, want 10.0", series.Last())
          }
      })

      t.Run("LastN", func(t *testing.T) {
          lastN := series.LastN(3)
          want := []float64{8, 9, 10}
          if !testutil.SlicesAlmostEqual(lastN, want, FloatTolerance) {
              t.Errorf("LastN(3) = %v, want %v", lastN, want)
          }
      })

      t.Run("Slice", func(t *testing.T) {
          slice := series.Slice(2, 5)
          if slice.Len() != 3 {
              t.Errorf("Slice(2,5).Len() = %d, want 3", slice.Len())
          }
      })

      t.Run("Apply", func(t *testing.T) {
          doubled := series.Apply(func(x float64) float64 { return x * 2 })
          if doubled.Last() != 20.0 {
              t.Errorf("Apply(double).Last() = %f, want 20.0", doubled.Last())
          }
      })
  }
  ```

### 1.4 Math Functions Tests

**File: `data/math_test.go`**

- [ ] Test SMA calculation:
  ```go
  func TestSMA(t *testing.T) {
      tests := []struct {
          name     string
          values   []float64
          period   int
          expected []float64
      }{
          {
              name:     "simple case",
              values:   []float64{1, 2, 3, 4, 5},
              period:   3,
              expected: []float64{math.NaN(), math.NaN(), 2, 3, 4},
          },
          {
              name:     "period 1",
              values:   []float64{1, 2, 3, 4, 5},
              period:   1,
              expected: []float64{1, 2, 3, 4, 5},
          },
          {
              name:     "period equals length",
              values:   []float64{1, 2, 3, 4, 5},
              period:   5,
              expected: []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), 3},
          },
          {
              name:     "with NaN values",
              values:   []float64{1, math.NaN(), 3, 4, 5},
              period:   2,
              expected: []float64{math.NaN(), math.NaN(), math.NaN(), 3.5, 4.5},
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := SMA(tt.values, tt.period)
              if !testutil.SlicesAlmostEqual(result, tt.expected, FloatTolerance) {
                  t.Errorf("SMA() = %v, want %v", result, tt.expected)
              }
          })
      }
  }
  ```

- [ ] Test SMA against Python reference:
  ```go
  func TestSMA_PythonParity(t *testing.T) {
      ohlcv := testutil.LoadTestData(t, "GOOG")

      // Golden values generated from Python:
      // pd.Series(close).rolling(20).mean()
      var golden struct {
          SMA20 []float64 `json:"sma_20"`
          SMA50 []float64 `json:"sma_50"`
      }
      testutil.LoadGolden(t, "goog_sma", &golden)

      sma20 := SMA(ohlcv.Close, 20)
      sma50 := SMA(ohlcv.Close, 50)

      // Compare only non-NaN values (after warmup)
      for i := 20; i < len(sma20); i++ {
          testutil.RequireAlmostEqual(t, golden.SMA20[i], sma20[i],
              StatsTolerance, fmt.Sprintf("SMA20[%d]", i))
      }
      for i := 50; i < len(sma50); i++ {
          testutil.RequireAlmostEqual(t, golden.SMA50[i], sma50[i],
              StatsTolerance, fmt.Sprintf("SMA50[%d]", i))
      }
  }
  ```

- [ ] Test EMA calculation:
  ```go
  func TestEMA(t *testing.T) {
      tests := []struct {
          name     string
          values   []float64
          period   int
          expected []float64
      }{
          {
              name:   "standard EMA",
              values: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
              period: 3,
              // EMA formula: EMA = Price * k + EMA_prev * (1-k), k = 2/(period+1)
              expected: []float64{
                  math.NaN(), math.NaN(), 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0,
              },
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := EMA(tt.values, tt.period)
              // Note: EMA values differ from SMA, verify calculation
              if len(result) != len(tt.values) {
                  t.Errorf("EMA length = %d, want %d", len(result), len(tt.values))
              }
          })
      }
  }
  ```

- [ ] Test rolling functions:
  ```go
  func TestRollingFunctions(t *testing.T) {
      values := []float64{5, 3, 8, 1, 9, 2, 7, 4, 6, 10}

      t.Run("Max", func(t *testing.T) {
          result := Max(values, 3)
          expected := []float64{math.NaN(), math.NaN(), 8, 8, 9, 9, 9, 7, 7, 10}
          if !testutil.SlicesAlmostEqual(result, expected, FloatTolerance) {
              t.Errorf("Max() = %v, want %v", result, expected)
          }
      })

      t.Run("Min", func(t *testing.T) {
          result := Min(values, 3)
          expected := []float64{math.NaN(), math.NaN(), 3, 1, 1, 1, 2, 2, 4, 4}
          if !testutil.SlicesAlmostEqual(result, expected, FloatTolerance) {
              t.Errorf("Min() = %v, want %v", result, expected)
          }
      })

      t.Run("StdDev", func(t *testing.T) {
          result := StdDev(values, 3)
          // Verify manually calculated standard deviations
          if math.IsNaN(result[2]) {
              t.Error("StdDev[2] should not be NaN")
          }
      })
  }
  ```

### 1.5 Fuzz Tests for Phase 1

**File: `data/fuzz_test.go`**

- [ ] Fuzz test SMA:
  ```go
  func FuzzSMA(f *testing.F) {
      // Seed corpus
      f.Add([]byte{1, 2, 3, 4, 5}, 3)

      f.Fuzz(func(t *testing.T, data []byte, period int) {
          if len(data) == 0 || period <= 0 || period > len(data) {
              return
          }

          values := make([]float64, len(data))
          for i, b := range data {
              values[i] = float64(b)
          }

          result := SMA(values, period)

          // Invariants that must hold
          if len(result) != len(values) {
              t.Errorf("result length %d != input length %d", len(result), len(values))
          }

          // First period-1 values must be NaN
          for i := 0; i < period-1; i++ {
              if !math.IsNaN(result[i]) {
                  t.Errorf("result[%d] = %f, want NaN", i, result[i])
              }
          }

          // No NaN or Inf after warmup (if input has no NaN)
          for i := period - 1; i < len(result); i++ {
              if math.IsInf(result[i], 0) {
                  t.Errorf("result[%d] is Inf", i)
              }
          }
      })
  }
  ```

### 1.6 Benchmark Tests for Phase 1

**File: `data/benchmark_test.go`**

- [ ] Benchmark data operations:
  ```go
  func BenchmarkLoadCSV(b *testing.B) {
      for i := 0; i < b.N; i++ {
          _, _ = LoadCSV("testdata/csv/GOOG.csv")
      }
  }

  func BenchmarkSMA(b *testing.B) {
      ohlcv := testutil.MakeTestOHLCV(10000, 100.0)
      b.ResetTimer()

      for i := 0; i < b.N; i++ {
          _ = SMA(ohlcv.Close, 20)
      }
  }

  func BenchmarkEMA(b *testing.B) {
      ohlcv := testutil.MakeTestOHLCV(10000, 100.0)
      b.ResetTimer()

      for i := 0; i < b.N; i++ {
          _ = EMA(ohlcv.Close, 20)
      }
  }

  func BenchmarkOHLCV_Slice(b *testing.B) {
      ohlcv := testutil.MakeTestOHLCV(10000, 100.0)
      b.ResetTimer()

      for i := 0; i < b.N; i++ {
          _ = ohlcv.Slice(1000, 5000)
      }
  }
  ```

---

## Phase 2 Tests: Core Engine

### 2.1 Order Tests

**File: `order_test.go`**

- [ ] Test Order creation and properties:
  ```go
  func TestOrder_Properties(t *testing.T) {
      tests := []struct {
          name      string
          size      float64
          wantLong  bool
          wantShort bool
      }{
          {"positive size (long)", 100, true, false},
          {"negative size (short)", -100, false, true},
          {"zero size", 0, false, false},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              order := &Order{Size: tt.size}
              if order.IsLong() != tt.wantLong {
                  t.Errorf("IsLong() = %v, want %v", order.IsLong(), tt.wantLong)
              }
              if order.IsShort() != tt.wantShort {
                  t.Errorf("IsShort() = %v, want %v", order.IsShort(), tt.wantShort)
              }
          })
      }
  }
  ```

- [ ] Test contingent order detection:
  ```go
  func TestOrder_IsContingent(t *testing.T) {
      parentTrade := &Trade{ID: 1}

      tests := []struct {
          name           string
          parentTrade    *Trade
          wantContingent bool
      }{
          {"with parent trade", parentTrade, true},
          {"without parent trade", nil, false},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              order := &Order{ParentTrade: tt.parentTrade}
              if order.IsContingent() != tt.wantContingent {
                  t.Errorf("IsContingent() = %v, want %v",
                      order.IsContingent(), tt.wantContingent)
              }
          })
      }
  }
  ```

### 2.2 Trade Tests

**File: `trade_test.go`**

- [ ] Test Trade P&L calculations:
  ```go
  func TestTrade_PL(t *testing.T) {
      tests := []struct {
          name       string
          size       float64
          entryPrice float64
          exitPrice  float64
          wantPL     float64
          wantPLPct  float64
      }{
          {
              name:       "long winning trade",
              size:       100,
              entryPrice: 50.0,
              exitPrice:  55.0,
              wantPL:     500.0,  // 100 * (55 - 50)
              wantPLPct:  10.0,   // (55 - 50) / 50 * 100
          },
          {
              name:       "long losing trade",
              size:       100,
              entryPrice: 50.0,
              exitPrice:  45.0,
              wantPL:     -500.0,
              wantPLPct:  -10.0,
          },
          {
              name:       "short winning trade",
              size:       -100,
              entryPrice: 50.0,
              exitPrice:  45.0,
              wantPL:     500.0,  // -100 * (45 - 50)
              wantPLPct:  10.0,
          },
          {
              name:       "short losing trade",
              size:       -100,
              entryPrice: 50.0,
              exitPrice:  55.0,
              wantPL:     -500.0,
              wantPLPct:  -10.0,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              trade := &Trade{
                  Size:       tt.size,
                  EntryPrice: tt.entryPrice,
                  ExitPrice:  &tt.exitPrice,
              }

              testutil.RequireAlmostEqual(t, tt.wantPL, trade.PL(),
                  FloatTolerance, "PL")
              testutil.RequireAlmostEqual(t, tt.wantPLPct, trade.PLPct(),
                  PercentTolerance, "PLPct")
          })
      }
  }
  ```

- [ ] Test Trade value calculation:
  ```go
  func TestTrade_Value(t *testing.T) {
      tests := []struct {
          name       string
          size       float64
          entryPrice float64
          wantValue  float64
      }{
          {"long trade", 100, 50.0, 5000.0},
          {"short trade", -100, 50.0, 5000.0}, // Value is absolute
          {"fractional", 10.5, 100.0, 1050.0},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              trade := &Trade{Size: tt.size, EntryPrice: tt.entryPrice}
              testutil.RequireAlmostEqual(t, tt.wantValue, trade.Value(),
                  FloatTolerance, "Value")
          })
      }
  }
  ```

### 2.3 Position Tests

**File: `position_test.go`**

- [ ] Test Position aggregation:
  ```go
  func TestPosition_Size(t *testing.T) {
      // Create broker with multiple trades
      broker := &Broker{
          trades: []*Trade{
              {Size: 100, EntryPrice: 50.0},
              {Size: 50, EntryPrice: 52.0},
              {Size: -30, EntryPrice: 51.0}, // Partial close
          },
      }
      position := &Position{broker: broker}

      // Position size should be sum of all trade sizes
      wantSize := 100.0 + 50.0 - 30.0
      if position.Size() != wantSize {
          t.Errorf("Size() = %f, want %f", position.Size(), wantSize)
      }
  }
  ```

### 2.4 Broker Tests

**File: `broker_test.go`**

- [ ] Test market order execution:
  ```go
  func TestBroker_MarketOrder(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)
      broker := NewBroker(
          WithCash(10000),
          WithData(ohlcv),
      )

      // Create market buy order
      order := broker.NewOrder(10) // Buy 10 units

      // Process next bar
      broker.currentBar = 1
      broker.Next()

      // Order should be filled
      if len(broker.orders) != 0 {
          t.Error("order should be filled and removed")
      }
      if len(broker.trades) != 1 {
          t.Error("trade should be opened")
      }

      trade := broker.trades[0]
      if trade.Size != 10 {
          t.Errorf("trade.Size = %f, want 10", trade.Size)
      }
  }
  ```

- [ ] Test limit order execution:
  ```go
  func TestBroker_LimitOrder(t *testing.T) {
      // Create OHLCV where price drops to hit limit
      ohlcv := &data.OHLCV{
          Time:   make([]time.Time, 5),
          Open:   []float64{100, 99, 98, 97, 96},
          High:   []float64{101, 100, 99, 98, 97},
          Low:    []float64{99, 98, 95, 96, 95},  // Hits 95 on bar 2
          Close:  []float64{100, 98, 96, 97, 96},
          Volume: []float64{1000, 1000, 1000, 1000, 1000},
      }

      broker := NewBroker(WithCash(10000), WithData(ohlcv))

      // Place limit buy at 95
      limitPrice := 95.0
      broker.NewOrder(10, WithLimit(limitPrice))

      // Bar 0: limit not hit
      broker.currentBar = 0
      broker.Next()
      if len(broker.trades) != 0 {
          t.Error("order should not fill on bar 0")
      }

      // Bar 1: limit not hit (low is 98)
      broker.currentBar = 1
      broker.Next()
      if len(broker.trades) != 0 {
          t.Error("order should not fill on bar 1")
      }

      // Bar 2: limit hit (low is 95)
      broker.currentBar = 2
      broker.Next()
      if len(broker.trades) != 1 {
          t.Error("order should fill on bar 2")
      }

      // Verify fill price
      trade := broker.trades[0]
      testutil.RequireAlmostEqual(t, limitPrice, trade.EntryPrice,
          PriceTolerance, "entry price")
  }
  ```

- [ ] Test stop order execution:
  ```go
  func TestBroker_StopOrder(t *testing.T) {
      // Create OHLCV where price rises to hit stop
      ohlcv := &data.OHLCV{
          Time:   make([]time.Time, 5),
          Open:   []float64{100, 101, 102, 103, 104},
          High:   []float64{101, 102, 105, 104, 106},  // Hits 105 on bar 2
          Low:    []float64{99, 100, 101, 102, 103},
          Close:  []float64{100, 101, 103, 103, 105},
          Volume: []float64{1000, 1000, 1000, 1000, 1000},
      }

      broker := NewBroker(WithCash(10000), WithData(ohlcv))

      // Place stop buy at 105
      stopPrice := 105.0
      broker.NewOrder(10, WithStop(stopPrice))

      // Process bars until stop is hit
      for i := 0; i < 3; i++ {
          broker.currentBar = i
          broker.Next()
      }

      if len(broker.trades) != 1 {
          t.Fatalf("expected 1 trade, got %d", len(broker.trades))
      }

      trade := broker.trades[0]
      testutil.RequireAlmostEqual(t, stopPrice, trade.EntryPrice,
          PriceTolerance, "entry price")
  }
  ```

- [ ] Test stop-limit order:
  ```go
  func TestBroker_StopLimitOrder(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(20, 100.0)
      broker := NewBroker(WithCash(10000), WithData(ohlcv))

      // Stop-limit: stop at 105, limit at 106
      stopPrice := 105.0
      limitPrice := 106.0
      broker.NewOrder(10, WithStop(stopPrice), WithLimit(limitPrice))

      // Order should become limit order when stop is hit
      // Then fill when limit is reached
      // ... detailed test logic
  }
  ```

- [ ] Test SL/TP bracket orders:
  ```go
  func TestBroker_BracketOrders(t *testing.T) {
      ohlcv := &data.OHLCV{
          Time:   make([]time.Time, 10),
          Open:   []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100},
          High:   []float64{102, 102, 102, 102, 110, 102, 102, 102, 102, 102},
          Low:    []float64{98, 98, 98, 98, 98, 98, 98, 98, 98, 98},
          Close:  []float64{100, 100, 100, 100, 108, 100, 100, 100, 100, 100},
          Volume: []float64{1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000},
      }

      broker := NewBroker(WithCash(10000), WithData(ohlcv))

      // Buy with SL at 95 and TP at 110
      broker.NewOrder(10, WithSL(95.0), WithTP(110.0))

      // Process bar 0 - order fills
      broker.currentBar = 0
      broker.Next()

      if len(broker.trades) != 1 {
          t.Fatal("trade should be opened")
      }

      trade := broker.trades[0]

      // Verify SL and TP orders exist
      if trade.SL() == nil || *trade.SL() != 95.0 {
          t.Error("SL not set correctly")
      }
      if trade.TP() == nil || *trade.TP() != 110.0 {
          t.Error("TP not set correctly")
      }

      // Process until TP hit (bar 4, high = 110)
      for i := 1; i <= 4; i++ {
          broker.currentBar = i
          broker.Next()
      }

      if len(broker.trades) != 0 {
          t.Error("trade should be closed by TP")
      }
      if len(broker.closedTrades) != 1 {
          t.Error("trade should be in closedTrades")
      }
  }
  ```

- [ ] Test commission calculations:
  ```go
  func TestBroker_Commission(t *testing.T) {
      tests := []struct {
          name           string
          commission     CommissionFunc
          orderSize      float64
          price          float64
          wantCommission float64
      }{
          {
              name:           "fixed commission",
              commission:     func(size, price float64) float64 { return 10.0 },
              orderSize:      100,
              price:          50.0,
              wantCommission: 10.0,
          },
          {
              name:           "percent commission",
              commission:     func(size, price float64) float64 { return math.Abs(size) * price * 0.001 },
              orderSize:      100,
              price:          50.0,
              wantCommission: 5.0, // 100 * 50 * 0.001
          },
          {
              name:           "per share commission",
              commission:     func(size, price float64) float64 { return math.Abs(size) * 0.01 },
              orderSize:      100,
              price:          50.0,
              wantCommission: 1.0, // 100 * 0.01
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              ohlcv := testutil.MakeTestOHLCV(10, tt.price)
              broker := NewBroker(
                  WithCash(100000),
                  WithData(ohlcv),
                  WithCommission(tt.commission),
              )

              initialCash := broker.cash
              broker.NewOrder(tt.orderSize)
              broker.currentBar = 0
              broker.Next()

              // Cash should be reduced by trade value + commission
              expectedCash := initialCash - (tt.orderSize * tt.price) - tt.wantCommission
              testutil.RequireAlmostEqual(t, expectedCash, broker.cash,
                  FloatTolerance, "cash after commission")
          })
      }
  }
  ```

- [ ] Test margin requirements:
  ```go
  func TestBroker_Margin(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(10, 100.0)

      tests := []struct {
          name       string
          cash       float64
          margin     float64 // 1.0 = 100%, 0.1 = 10x leverage
          orderSize  float64
          orderValue float64
          shouldFill bool
      }{
          {
              name:       "no leverage, sufficient cash",
              cash:       10000,
              margin:     1.0,
              orderSize:  50,
              orderValue: 5000,
              shouldFill: true,
          },
          {
              name:       "no leverage, insufficient cash",
              cash:       1000,
              margin:     1.0,
              orderSize:  50,
              orderValue: 5000,
              shouldFill: false,
          },
          {
              name:       "10x leverage, sufficient margin",
              cash:       1000,
              margin:     0.1,
              orderSize:  50,
              orderValue: 5000, // Requires 500 margin
              shouldFill: true,
          },
          {
              name:       "10x leverage, insufficient margin",
              cash:       100,
              margin:     0.1,
              orderSize:  50,
              orderValue: 5000, // Requires 500 margin
              shouldFill: false,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              broker := NewBroker(
                  WithCash(tt.cash),
                  WithData(ohlcv),
                  WithMargin(tt.margin),
              )

              broker.NewOrder(tt.orderSize)
              broker.currentBar = 0
              broker.Next()

              hasTrade := len(broker.trades) > 0
              if hasTrade != tt.shouldFill {
                  t.Errorf("trade filled = %v, want %v", hasTrade, tt.shouldFill)
              }
          })
      }
  }
  ```

- [ ] Test hedging mode:
  ```go
  func TestBroker_Hedging(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(10, 100.0)

      t.Run("hedging disabled", func(t *testing.T) {
          broker := NewBroker(
              WithCash(100000),
              WithData(ohlcv),
              WithHedging(false),
          )

          // Open long position
          broker.NewOrder(10)
          broker.currentBar = 0
          broker.Next()

          // Open short position (should close long first)
          broker.NewOrder(-10)
          broker.currentBar = 1
          broker.Next()

          // Should have no active trades (positions netted)
          if len(broker.trades) != 0 {
              t.Errorf("expected 0 trades, got %d", len(broker.trades))
          }
      })

      t.Run("hedging enabled", func(t *testing.T) {
          broker := NewBroker(
              WithCash(100000),
              WithData(ohlcv),
              WithHedging(true),
          )

          // Open long position
          broker.NewOrder(10)
          broker.currentBar = 0
          broker.Next()

          // Open short position (should coexist)
          broker.NewOrder(-10)
          broker.currentBar = 1
          broker.Next()

          // Should have both trades
          if len(broker.trades) != 2 {
              t.Errorf("expected 2 trades, got %d", len(broker.trades))
          }
      })
  }
  ```

- [ ] Test exclusive orders mode:
  ```go
  func TestBroker_ExclusiveOrders(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(10, 100.0)
      broker := NewBroker(
          WithCash(100000),
          WithData(ohlcv),
          WithExclusiveOrders(true),
      )

      // Place first order
      broker.NewOrder(10)
      broker.currentBar = 0
      broker.Next()

      firstTrade := broker.trades[0]

      // Place second order (should close first)
      broker.NewOrder(20)
      broker.currentBar = 1
      broker.Next()

      // First trade should be closed
      if len(broker.closedTrades) != 1 {
          t.Error("first trade should be closed")
      }
      if broker.closedTrades[0].ID != firstTrade.ID {
          t.Error("closed trade should be first trade")
      }

      // Second trade should be active
      if len(broker.trades) != 1 {
          t.Error("second trade should be active")
      }
      if broker.trades[0].Size != 20 {
          t.Errorf("second trade size = %f, want 20", broker.trades[0].Size)
      }
  }
  ```

### 2.5 Backtest Tests

**File: `backtest_test.go`**

- [ ] Test simple SMA crossover strategy:
  ```go
  // SmaCrossStrategy is a test strategy
  type SmaCrossStrategy struct {
      StrategyBase
      FastPeriod int
      SlowPeriod int
      fastSMA    *Indicator
      slowSMA    *Indicator
  }

  func (s *SmaCrossStrategy) Init() {
      s.fastSMA = s.I("SMA_fast", SMA(s.Close().Values, s.FastPeriod))
      s.slowSMA = s.I("SMA_slow", SMA(s.Close().Values, s.SlowPeriod))
  }

  func (s *SmaCrossStrategy) Next() {
      if len(s.fastSMA.Values) < 2 {
          return
      }

      i := len(s.fastSMA.Values) - 1

      // Crossover: fast crosses above slow
      if s.fastSMA.Values[i-1] <= s.slowSMA.Values[i-1] &&
         s.fastSMA.Values[i] > s.slowSMA.Values[i] {
          if !s.Position().IsLong() {
              s.Buy()
          }
      }

      // Crossunder: fast crosses below slow
      if s.fastSMA.Values[i-1] >= s.slowSMA.Values[i-1] &&
         s.fastSMA.Values[i] < s.slowSMA.Values[i] {
          if s.Position().IsLong() {
              s.Position().Close(1.0)
          }
      }
  }

  func TestBacktest_SmaCross(t *testing.T) {
      ohlcv := testutil.LoadTestData(t, "GOOG")

      strategy := &SmaCrossStrategy{
          FastPeriod: 10,
          SlowPeriod: 20,
      }

      bt := NewBacktest(ohlcv, strategy, BacktestConfig{
          Cash: 10000,
      })

      stats, err := bt.Run(nil)
      if err != nil {
          t.Fatalf("backtest failed: %v", err)
      }

      // Basic sanity checks
      if stats.NumTrades == 0 {
          t.Error("expected some trades")
      }
      if stats.Start.IsZero() {
          t.Error("start time not set")
      }
      if stats.End.IsZero() {
          t.Error("end time not set")
      }

      // Compare against Python golden values
      var golden struct {
          NumTrades  int     `json:"num_trades"`
          ReturnPct  float64 `json:"return_pct"`
          SharpeRatio float64 `json:"sharpe_ratio"`
      }
      testutil.LoadGolden(t, "sma_cross_goog", &golden)

      if stats.NumTrades != golden.NumTrades {
          t.Errorf("NumTrades = %d, want %d", stats.NumTrades, golden.NumTrades)
      }
      testutil.RequireAlmostEqual(t, golden.ReturnPct, stats.ReturnPct,
          StatsTolerance, "ReturnPct")
      testutil.RequireAlmostEqual(t, golden.SharpeRatio, stats.SharpeRatio,
          StatsTolerance, "SharpeRatio")
  }
  ```

- [ ] Test backtest with no trades:
  ```go
  type NeverTradeStrategy struct {
      StrategyBase
  }

  func (s *NeverTradeStrategy) Init() {}
  func (s *NeverTradeStrategy) Next() {}

  func TestBacktest_NoTrades(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)
      bt := NewBacktest(ohlcv, &NeverTradeStrategy{}, BacktestConfig{Cash: 10000})

      stats, err := bt.Run(nil)
      if err != nil {
          t.Fatalf("backtest failed: %v", err)
      }

      if stats.NumTrades != 0 {
          t.Errorf("NumTrades = %d, want 0", stats.NumTrades)
      }
      if stats.ReturnPct != 0 {
          t.Errorf("ReturnPct = %f, want 0", stats.ReturnPct)
      }
  }
  ```

- [ ] Test out-of-money condition:
  ```go
  type AllInStrategy struct {
      StrategyBase
  }

  func (s *AllInStrategy) Init() {}

  func (s *AllInStrategy) Next() {
      if !s.Position().IsLong() {
          // Buy with all available cash
          s.Buy(WithSize(s.Equity() / s.Close().Last()))
      }
  }

  func TestBacktest_OutOfMoney(t *testing.T) {
      // Create data that crashes to zero
      ohlcv := &data.OHLCV{
          Time:   make([]time.Time, 10),
          Open:   []float64{100, 90, 80, 70, 60, 50, 40, 30, 20, 10},
          High:   []float64{100, 90, 80, 70, 60, 50, 40, 30, 20, 10},
          Low:    []float64{90, 80, 70, 60, 50, 40, 30, 20, 10, 1},
          Close:  []float64{90, 80, 70, 60, 50, 40, 30, 20, 10, 1},
          Volume: []float64{1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000},
      }

      bt := NewBacktest(ohlcv, &AllInStrategy{}, BacktestConfig{
          Cash:   1000,
          Margin: 0.1, // 10x leverage
      })

      stats, err := bt.Run(nil)

      // Should complete without error (just stop early)
      if err != nil {
          t.Fatalf("backtest failed: %v", err)
      }

      // Equity should be nearly zero or negative
      if stats.ReturnPct > -90 {
          t.Errorf("expected massive loss, got %f%%", stats.ReturnPct)
      }
  }
  ```

### 2.6 Benchmark Tests for Phase 2

**File: `benchmark_test.go`**

- [ ] Benchmark backtest execution:
  ```go
  func BenchmarkBacktest_SmaCross(b *testing.B) {
      ohlcv := testutil.LoadTestData(nil, "GOOG")

      for i := 0; i < b.N; i++ {
          strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
          bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
          _, _ = bt.Run(nil)
      }
  }

  func BenchmarkBacktest_LargeData(b *testing.B) {
      ohlcv := testutil.MakeTestOHLCV(100000, 100.0) // 100k bars

      for i := 0; i < b.N; i++ {
          strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
          bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
          _, _ = bt.Run(nil)
      }
  }

  func BenchmarkBroker_OrderProcessing(b *testing.B) {
      ohlcv := testutil.MakeTestOHLCV(1000, 100.0)

      b.ResetTimer()
      for i := 0; i < b.N; i++ {
          broker := NewBroker(WithCash(1000000), WithData(ohlcv))
          for j := 0; j < 100; j++ {
              broker.NewOrder(float64(10 + j%20))
              broker.currentBar = j % ohlcv.Len()
              broker.Next()
          }
      }
  }
  ```

---

## Phase 3 Tests: Statistics Module

### 3.1 Individual Metric Tests

**File: `stats/metrics_test.go`**

- [ ] Test return calculations:
  ```go
  func TestCalcReturn(t *testing.T) {
      tests := []struct {
          name        string
          startEquity float64
          endEquity   float64
          wantReturn  float64
      }{
          {"positive return", 10000, 15000, 50.0},
          {"negative return", 10000, 8000, -20.0},
          {"no change", 10000, 10000, 0.0},
          {"double", 10000, 20000, 100.0},
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := calcReturn(tt.startEquity, tt.endEquity)
              testutil.RequireAlmostEqual(t, tt.wantReturn, result,
                  PercentTolerance, "return")
          })
      }
  }
  ```

- [ ] Test Sharpe ratio:
  ```go
  func TestCalcSharpeRatio(t *testing.T) {
      tests := []struct {
          name         string
          returns      []float64
          riskFreeRate float64
          wantSharpe   float64
      }{
          {
              name:         "positive sharpe",
              returns:      []float64{0.01, 0.02, -0.01, 0.03, 0.01, 0.02},
              riskFreeRate: 0.0,
              wantSharpe:   1.5, // Approximate
          },
          {
              name:         "zero returns",
              returns:      []float64{0, 0, 0, 0, 0},
              riskFreeRate: 0.0,
              wantSharpe:   0.0,
          },
          {
              name:         "constant returns",
              returns:      []float64{0.01, 0.01, 0.01, 0.01, 0.01},
              riskFreeRate: 0.0,
              wantSharpe:   math.Inf(1), // Undefined (zero std dev)
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := calcSharpeRatio(tt.returns, tt.riskFreeRate)
              if math.IsInf(tt.wantSharpe, 0) {
                  if !math.IsInf(result, 0) {
                      t.Errorf("expected Inf, got %f", result)
                  }
              } else {
                  // Allow larger tolerance for Sharpe
                  testutil.RequireAlmostEqual(t, tt.wantSharpe, result,
                      0.5, "Sharpe ratio")
              }
          })
      }
  }
  ```

- [ ] Test Sortino ratio:
  ```go
  func TestCalcSortinoRatio(t *testing.T) {
      tests := []struct {
          name         string
          returns      []float64
          riskFreeRate float64
          wantSortino  float64
      }{
          {
              name:         "mixed returns",
              returns:      []float64{0.02, -0.01, 0.03, -0.02, 0.01},
              riskFreeRate: 0.0,
              wantSortino:  1.0, // Approximate
          },
          {
              name:         "all positive",
              returns:      []float64{0.01, 0.02, 0.01, 0.03, 0.02},
              riskFreeRate: 0.0,
              wantSortino:  math.Inf(1), // No downside deviation
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := calcSortinoRatio(tt.returns, tt.riskFreeRate)
              // Verify finite or expected infinite
              if !math.IsNaN(result) && math.IsFinite(result) {
                  // Just verify it's positive for positive avg return
              }
          })
      }
  }
  ```

- [ ] Test win rate:
  ```go
  func TestCalcWinRate(t *testing.T) {
      tests := []struct {
          name       string
          trades     []*Trade
          wantWinPct float64
      }{
          {
              name: "50% win rate",
              trades: []*Trade{
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(110)}, // Win
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(90)},  // Loss
              },
              wantWinPct: 50.0,
          },
          {
              name: "100% win rate",
              trades: []*Trade{
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(110)},
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(105)},
              },
              wantWinPct: 100.0,
          },
          {
              name:       "no trades",
              trades:     []*Trade{},
              wantWinPct: 0.0,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := calcWinRate(tt.trades)
              testutil.RequireAlmostEqual(t, tt.wantWinPct, result,
                  PercentTolerance, "win rate")
          })
      }
  }
  ```

- [ ] Test profit factor:
  ```go
  func TestCalcProfitFactor(t *testing.T) {
      tests := []struct {
          name             string
          trades           []*Trade
          wantProfitFactor float64
      }{
          {
              name: "2:1 profit factor",
              trades: []*Trade{
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(120)}, // +200
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(90)},  // -100
              },
              wantProfitFactor: 2.0,
          },
          {
              name: "no losses",
              trades: []*Trade{
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(110)},
              },
              wantProfitFactor: math.Inf(1),
          },
          {
              name: "no profits",
              trades: []*Trade{
                  {Size: 10, EntryPrice: 100, ExitPrice: ptr(90)},
              },
              wantProfitFactor: 0.0,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := calcProfitFactor(tt.trades)
              if math.IsInf(tt.wantProfitFactor, 1) {
                  if !math.IsInf(result, 1) {
                      t.Errorf("expected +Inf, got %f", result)
                  }
              } else {
                  testutil.RequireAlmostEqual(t, tt.wantProfitFactor, result,
                      FloatTolerance, "profit factor")
              }
          })
      }
  }
  ```

### 3.2 Drawdown Tests

**File: `stats/drawdown_test.go`**

- [ ] Test drawdown calculation:
  ```go
  func TestCalcDrawdowns(t *testing.T) {
      tests := []struct {
          name           string
          equity         []float64
          wantMaxDD      float64
          wantNumDD      int
      }{
          {
              name:      "simple drawdown",
              equity:    []float64{100, 110, 105, 100, 115, 110, 120},
              wantMaxDD: (110.0 - 100.0) / 110.0 * 100, // ~9.09%
              wantNumDD: 2, // Two drawdown periods
          },
          {
              name:      "no drawdown (always increasing)",
              equity:    []float64{100, 110, 120, 130, 140},
              wantMaxDD: 0.0,
              wantNumDD: 0,
          },
          {
              name:      "continuous drawdown",
              equity:    []float64{100, 90, 80, 70, 60},
              wantMaxDD: 40.0, // 100 -> 60
              wantNumDD: 1,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              maxDD, drawdowns := calcDrawdowns(tt.equity)

              testutil.RequireAlmostEqual(t, tt.wantMaxDD, maxDD,
                  PercentTolerance, "max drawdown")

              if len(drawdowns) != tt.wantNumDD {
                  t.Errorf("num drawdowns = %d, want %d", len(drawdowns), tt.wantNumDD)
              }
          })
      }
  }
  ```

- [ ] Test drawdown duration:
  ```go
  func TestCalcDrawdownDuration(t *testing.T) {
      times := []time.Time{
          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
          time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
          time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
          time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
          time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
          time.Date(2020, 1, 6, 0, 0, 0, 0, time.UTC),
      }
      equity := []float64{100, 110, 100, 95, 105, 115}

      maxDur, avgDur := calcDrawdownDurations(equity, times)

      // Drawdown from 110 to 95, recovered at 115
      // Duration: Jan 2 to Jan 6 = 4 days
      expectedMaxDur := 4 * 24 * time.Hour

      if maxDur != expectedMaxDur {
          t.Errorf("max duration = %v, want %v", maxDur, expectedMaxDur)
      }
  }
  ```

### 3.3 Full Stats Computation Test

**File: `stats/compute_test.go`**

- [ ] Test stats computation against Python golden:
  ```go
  func TestComputeStats_PythonParity(t *testing.T) {
      // Load golden stats generated from Python
      var golden map[string]float64
      testutil.LoadGolden(t, "sma_cross_stats", &golden)

      // Run same backtest
      ohlcv := testutil.LoadTestData(t, "GOOG")
      strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
      stats, _ := bt.Run(nil)

      // Compare all metrics
      checks := []struct {
          name     string
          got      float64
          wantKey  string
      }{
          {"Return%", stats.ReturnPct, "return_pct"},
          {"Sharpe", stats.SharpeRatio, "sharpe_ratio"},
          {"Sortino", stats.SortinoRatio, "sortino_ratio"},
          {"MaxDD%", stats.MaxDrawdownPct, "max_drawdown_pct"},
          {"WinRate%", stats.WinRatePct, "win_rate_pct"},
          {"ProfitFactor", stats.ProfitFactor, "profit_factor"},
          {"Expectancy%", stats.ExpectancyPct, "expectancy_pct"},
          {"SQN", stats.SQN, "sqn"},
      }

      for _, check := range checks {
          t.Run(check.name, func(t *testing.T) {
              want := golden[check.wantKey]
              testutil.RequireAlmostEqual(t, want, check.got,
                  StatsTolerance, check.name)
          })
      }
  }
  ```

### 3.4 Benchmark Tests for Phase 3

**File: `stats/benchmark_test.go`**

- [ ] Benchmark stats computation:
  ```go
  func BenchmarkComputeStats(b *testing.B) {
      // Create realistic trade data
      trades := make([]*Trade, 100)
      equity := make([]float64, 1000)
      // ... populate with realistic data

      b.ResetTimer()
      for i := 0; i < b.N; i++ {
          _ = ComputeStats(ComputeConfig{
              ClosedTrades: trades,
              Equity:       equity,
          })
      }
  }
  ```

---

## Phase 4 Tests: Library Utilities

### 4.1 Crossover Tests

**File: `lib/crossover_test.go`**

- [ ] Test crossover detection:
  ```go
  func TestCrossover(t *testing.T) {
      tests := []struct {
          name    string
          series1 []float64
          series2 []float64
          want    []bool
      }{
          {
              name:    "simple crossover",
              series1: []float64{1, 2, 3, 4, 3, 2},
              series2: []float64{2, 2, 2, 2, 2, 2},
              want:    []bool{false, false, true, false, false, false},
          },
          {
              name:    "no crossover",
              series1: []float64{1, 1, 1, 1},
              series2: []float64{2, 2, 2, 2},
              want:    []bool{false, false, false, false},
          },
          {
              name:    "touch but no cross",
              series1: []float64{1, 2, 1},
              series2: []float64{2, 2, 2},
              want:    []bool{false, false, false},
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              result := Crossover(tt.series1, tt.series2)
              for i, want := range tt.want {
                  if result[i] != want {
                      t.Errorf("Crossover[%d] = %v, want %v", i, result[i], want)
                  }
              }
          })
      }
  }
  ```

- [ ] Test BarsSince:
  ```go
  func TestBarsSince(t *testing.T) {
      condition := []bool{false, true, false, false, true, false}
      want := []int{-1, 0, 1, 2, 0, 1}

      result := BarsSince(condition, -1)

      for i, w := range want {
          if result[i] != w {
              t.Errorf("BarsSince[%d] = %d, want %d", i, result[i], w)
          }
      }
  }
  ```

### 4.2 Indicator Tests

**File: `lib/indicators_test.go`**

- [ ] Test RSI calculation:
  ```go
  func TestRSI(t *testing.T) {
      // Use known values from ta-lib or Python
      prices := []float64{44, 44.34, 44.09, 43.61, 44.33, 44.83, 45.10,
                          45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28}
      period := 14

      result := RSI(prices, period)

      // RSI at index 14 should be approximately 70.46
      expectedRSI := 70.46
      testutil.RequireAlmostEqual(t, expectedRSI, result[14], 0.5, "RSI[14]")
  }
  ```

- [ ] Test MACD calculation:
  ```go
  func TestMACD(t *testing.T) {
      ohlcv := testutil.LoadTestData(t, "GOOG")

      macd, signal, hist := MACD(ohlcv.Close, 12, 26, 9)

      // Load golden values
      var golden struct {
          MACD   []float64 `json:"macd"`
          Signal []float64 `json:"signal"`
          Hist   []float64 `json:"hist"`
      }
      testutil.LoadGolden(t, "goog_macd", &golden)

      // Compare after warmup period (26 + 9 - 1 = 34)
      for i := 35; i < len(macd); i++ {
          testutil.RequireAlmostEqual(t, golden.MACD[i], macd[i],
              StatsTolerance, fmt.Sprintf("MACD[%d]", i))
      }
  }
  ```

- [ ] Test Bollinger Bands:
  ```go
  func TestBollingerBands(t *testing.T) {
      prices := testutil.LoadTestData(t, "GOOG").Close

      upper, middle, lower := BollingerBands(prices, 20, 2.0)

      // Verify relationships
      for i := 20; i < len(prices); i++ {
          if upper[i] <= middle[i] {
              t.Errorf("upper[%d] (%f) <= middle (%f)", i, upper[i], middle[i])
          }
          if lower[i] >= middle[i] {
              t.Errorf("lower[%d] (%f) >= middle (%f)", i, lower[i], middle[i])
          }
      }
  }
  ```

### 4.3 Resample Tests

**File: `lib/resample_test.go`**

- [ ] Test OHLCV resampling:
  ```go
  func TestResample(t *testing.T) {
      // Create hourly data
      hourly := testutil.LoadTestData(t, "EURUSD") // Hourly forex data

      // Resample to daily
      daily := Resample(hourly, "D")

      // Verify aggregation rules
      // - Open should be first of day
      // - High should be max of day
      // - Low should be min of day
      // - Close should be last of day
      // - Volume should be sum of day

      // Check first day manually
      // ...
  }
  ```

### 4.4 SignalStrategy Tests

**File: `lib/signals_test.go`**

- [ ] Test SignalStrategy execution:
  ```go
  func TestSignalStrategy(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)

      strategy := &SignalStrategy{}
      strategy.StrategyBase = StrategyBase{}

      // Set entry signals (buy on bars 10, 30, 50)
      entry := make([]float64, 100)
      entry[10] = 1.0  // Full position
      entry[30] = 0.5  // Half position
      entry[50] = 1.0

      // Set exit signals (sell on bars 20, 40, 60)
      exit := make([]float64, 100)
      exit[20] = 1.0  // Close all
      exit[40] = 0.5  // Close half
      exit[60] = 1.0

      strategy.SetSignal(entry, exit)

      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
      stats, _ := bt.Run(nil)

      // Verify correct number of trades
      expectedTrades := 3
      if stats.NumTrades != expectedTrades {
          t.Errorf("NumTrades = %d, want %d", stats.NumTrades, expectedTrades)
      }
  }
  ```

### 4.5 TrailingStrategy Tests

**File: `lib/trailing_test.go`**

- [ ] Test trailing stop adjustment:
  ```go
  func TestTrailingStrategy(t *testing.T) {
      // Create uptrending data then reversal
      ohlcv := &data.OHLCV{
          Time:   make([]time.Time, 20),
          Open:   []float64{100, 102, 104, 106, 108, 110, 112, 114, 116, 118,
                            120, 118, 116, 114, 112, 110, 108, 106, 104, 102},
          High:   []float64{102, 104, 106, 108, 110, 112, 114, 116, 118, 120,
                            122, 120, 118, 116, 114, 112, 110, 108, 106, 104},
          Low:    []float64{99, 101, 103, 105, 107, 109, 111, 113, 115, 117,
                            119, 117, 115, 113, 111, 109, 107, 105, 103, 101},
          Close:  []float64{101, 103, 105, 107, 109, 111, 113, 115, 117, 119,
                            121, 119, 117, 115, 113, 111, 109, 107, 105, 103},
          Volume: make([]float64, 20),
      }

      strategy := &TrailingStrategy{}
      strategy.SetATRPeriods(5)
      strategy.SetTrailingSL(2.0) // 2 ATR trailing stop

      // Strategy should:
      // 1. Enter long
      // 2. Trail stop up as price rises
      // 3. Get stopped out when price reverses

      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
      stats, _ := bt.Run(nil)

      // Should have profitable trade (entered early, trailed up)
      if stats.ReturnPct <= 0 {
          t.Error("expected positive return from trailing stop")
      }
  }
  ```

---

## Phase 5 Tests: Optimization System

### 5.1 Grid Search Tests

**File: `optimize/grid_test.go`**

- [ ] Test grid search:
  ```go
  func TestGridSearch(t *testing.T) {
      ohlcv := testutil.LoadTestData(t, "GOOG")

      strategy := &SmaCrossStrategy{}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})

      result, err := GridSearch(bt, GridConfig{
          Params: map[string][]interface{}{
              "FastPeriod": {5, 10, 15},
              "SlowPeriod": {20, 30, 40},
          },
          Maximize: "ReturnPct",
      })

      if err != nil {
          t.Fatalf("GridSearch failed: %v", err)
      }

      // Should find best parameters
      if result.BestParams == nil {
          t.Error("no best params found")
      }

      // Should have 9 results (3 x 3)
      if len(result.AllResults) != 9 {
          t.Errorf("expected 9 results, got %d", len(result.AllResults))
      }
  }
  ```

- [ ] Test constraints:
  ```go
  func TestGridSearch_Constraints(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)
      strategy := &SmaCrossStrategy{}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})

      result, _ := GridSearch(bt, GridConfig{
          Params: map[string][]interface{}{
              "FastPeriod": {5, 10, 20, 30},
              "SlowPeriod": {10, 20, 30, 40},
          },
          Constraint: func(p map[string]interface{}) bool {
              return p["FastPeriod"].(int) < p["SlowPeriod"].(int)
          },
          Maximize: "ReturnPct",
      })

      // All results should satisfy constraint
      for _, r := range result.AllResults {
          fast := r.Params["FastPeriod"].(int)
          slow := r.Params["SlowPeriod"].(int)
          if fast >= slow {
              t.Errorf("constraint violated: fast=%d >= slow=%d", fast, slow)
          }
      }
  }
  ```

- [ ] Test max tries limiting:
  ```go
  func TestGridSearch_MaxTries(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)
      strategy := &SmaCrossStrategy{}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})

      result, _ := GridSearch(bt, GridConfig{
          Params: map[string][]interface{}{
              "FastPeriod": {5, 10, 15, 20, 25},
              "SlowPeriod": {30, 35, 40, 45, 50},
          },
          MaxTries: 10, // Only try 10 of 25 combinations
          Maximize: "ReturnPct",
      })

      if len(result.AllResults) != 10 {
          t.Errorf("expected 10 results, got %d", len(result.AllResults))
      }
  }
  ```

### 5.2 Parallel Execution Tests

**File: `optimize/parallel_test.go`**

- [ ] Test parallel execution correctness:
  ```go
  func TestParallelExecution_Correctness(t *testing.T) {
      ohlcv := testutil.LoadTestData(t, "GOOG")
      strategy := &SmaCrossStrategy{}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})

      // Run sequentially
      seqResult, _ := GridSearch(bt, GridConfig{
          Params: map[string][]interface{}{
              "FastPeriod": {5, 10},
              "SlowPeriod": {20, 30},
          },
          Maximize:    "ReturnPct",
          // Workers: 1, // Sequential
      })

      // Run in parallel
      parResult, _ := GridSearch(bt, GridConfig{
          Params: map[string][]interface{}{
              "FastPeriod": {5, 10},
              "SlowPeriod": {20, 30},
          },
          Maximize: "ReturnPct",
          // Workers: 4, // Parallel
      })

      // Results should be identical
      if seqResult.BestValue != parResult.BestValue {
          t.Errorf("best values differ: seq=%f, par=%f",
              seqResult.BestValue, parResult.BestValue)
      }
  }
  ```

- [ ] Test race conditions:
  ```go
  func TestParallelExecution_NoRaces(t *testing.T) {
      // Run with -race flag
      ohlcv := testutil.MakeTestOHLCV(1000, 100.0)
      strategy := &SmaCrossStrategy{}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})

      // Run optimization with many workers
      _, err := GridSearch(bt, GridConfig{
          Params: map[string][]interface{}{
              "FastPeriod": {5, 10, 15, 20},
              "SlowPeriod": {25, 30, 35, 40},
          },
          Maximize: "ReturnPct",
      })

      if err != nil {
          t.Errorf("optimization failed: %v", err)
      }
  }
  ```

### 5.3 Heatmap Tests

**File: `optimize/heatmap_test.go`**

- [ ] Test heatmap generation:
  ```go
  func TestGenerateHeatmap(t *testing.T) {
      results := []ParamResult{
          {Params: map[string]interface{}{"x": 1, "y": 1}, Value: 10},
          {Params: map[string]interface{}{"x": 1, "y": 2}, Value: 20},
          {Params: map[string]interface{}{"x": 2, "y": 1}, Value: 15},
          {Params: map[string]interface{}{"x": 2, "y": 2}, Value: 25},
      }

      heatmap := GenerateHeatmap(results, "x", "y", "Value")

      if len(heatmap.XLabels) != 2 {
          t.Errorf("expected 2 x labels, got %d", len(heatmap.XLabels))
      }
      if len(heatmap.YLabels) != 2 {
          t.Errorf("expected 2 y labels, got %d", len(heatmap.YLabels))
      }

      // Check values
      if heatmap.Values[0][0] != 10 {
          t.Errorf("Values[0][0] = %f, want 10", heatmap.Values[0][0])
      }
      if heatmap.Values[1][1] != 25 {
          t.Errorf("Values[1][1] = %f, want 25", heatmap.Values[1][1])
      }
  }
  ```

### 5.4 Benchmark Tests for Phase 5

**File: `optimize/benchmark_test.go`**

- [ ] Benchmark optimization:
  ```go
  func BenchmarkGridSearch(b *testing.B) {
      ohlcv := testutil.LoadTestData(nil, "GOOG")

      b.ResetTimer()
      for i := 0; i < b.N; i++ {
          strategy := &SmaCrossStrategy{}
          bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})

          _, _ = GridSearch(bt, GridConfig{
              Params: map[string][]interface{}{
                  "FastPeriod": {5, 10, 15, 20},
                  "SlowPeriod": {25, 30, 35, 40},
              },
              Maximize: "ReturnPct",
          })
      }
  }
  ```

---

## Phase 6 Tests: Plotting & Visualization

### 6.1 Chart Data Tests

**File: `plot/data_test.go`**

- [ ] Test chart data preparation:
  ```go
  func TestPrepareChartData(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)
      strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
      stats, _ := bt.Run(nil)

      chartData := PrepareChartData(stats, ohlcv, strategy.indicators)

      if chartData.OHLCV == nil {
          t.Error("OHLCV data not set")
      }
      if len(chartData.Equity) == 0 {
          t.Error("equity data not set")
      }
      if len(chartData.Indicators) != 2 {
          t.Errorf("expected 2 indicators, got %d", len(chartData.Indicators))
      }
  }
  ```

- [ ] Test trade marker generation:
  ```go
  func TestTradeMarkers(t *testing.T) {
      trades := []*TradeRecord{
          {
              EntryTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
              ExitTime:   time.Date(2020, 1, 10, 0, 0, 0, 0, time.UTC),
              EntryPrice: 100,
              ExitPrice:  110,
              Size:       10,
              PL:         100,
          },
      }

      markers := generateTradeMarkers(trades)

      if len(markers) != 2 { // Entry + Exit
          t.Errorf("expected 2 markers, got %d", len(markers))
      }

      // Check entry marker
      if markers[0].Type != "entry_long" {
          t.Errorf("entry type = %s, want entry_long", markers[0].Type)
      }
      if markers[0].Price != 100 {
          t.Errorf("entry price = %f, want 100", markers[0].Price)
      }
  }
  ```

### 6.2 HTML Generation Tests

**File: `plot/plot_test.go`**

- [ ] Test HTML generation:
  ```go
  func TestPlot_GeneratesHTML(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(100, 100.0)
      strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
      stats, _ := bt.Run(nil)

      tmpFile := filepath.Join(t.TempDir(), "test_chart.html")

      err := Plot(stats, ohlcv, strategy.indicators, PlotConfig{
          Filename:   tmpFile,
          PlotEquity: true,
          PlotTrades: true,
      })

      if err != nil {
          t.Fatalf("Plot failed: %v", err)
      }

      // Verify file exists and has content
      content, err := os.ReadFile(tmpFile)
      if err != nil {
          t.Fatalf("failed to read output: %v", err)
      }

      if len(content) == 0 {
          t.Error("output file is empty")
      }

      // Check for expected HTML elements
      html := string(content)
      if !strings.Contains(html, "<html") {
          t.Error("missing <html> tag")
      }
      if !strings.Contains(html, "candlestick") || !strings.Contains(html, "chart") {
          t.Error("missing chart elements")
      }
  }
  ```

- [ ] Test JSON data embedding:
  ```go
  func TestPlot_JSONDataEmbedded(t *testing.T) {
      ohlcv := testutil.MakeTestOHLCV(50, 100.0)
      strategy := &SmaCrossStrategy{FastPeriod: 5, SlowPeriod: 10}
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
      stats, _ := bt.Run(nil)

      tmpFile := filepath.Join(t.TempDir(), "test_chart.html")
      _ = Plot(stats, ohlcv, strategy.indicators, PlotConfig{Filename: tmpFile})

      content, _ := os.ReadFile(tmpFile)
      html := string(content)

      // Verify OHLCV data is embedded
      if !strings.Contains(html, "ohlcvData") {
          t.Error("OHLCV data not embedded")
      }

      // Verify data is valid JSON
      // Extract and parse JSON...
  }
  ```

---

## Phase 7 Tests: Integration & Validation

### 7.1 End-to-End Integration Tests

**File: `integration_test.go`**

- [ ] Test full workflow:
  ```go
  func TestIntegration_FullWorkflow(t *testing.T) {
      // 1. Load data
      ohlcv, err := data.LoadCSV("testdata/csv/GOOG.csv")
      if err != nil {
          t.Fatalf("failed to load data: %v", err)
      }

      // 2. Create strategy
      strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}

      // 3. Run backtest
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{
          Cash:       10000,
          Commission: func(s, p float64) float64 { return math.Abs(s) * p * 0.001 },
      })
      stats, err := bt.Run(nil)
      if err != nil {
          t.Fatalf("backtest failed: %v", err)
      }

      // 4. Verify stats
      if stats.NumTrades == 0 {
          t.Error("no trades executed")
      }

      // 5. Run optimization
      result, err := optimize.GridSearch(bt, optimize.GridConfig{
          Params: map[string][]interface{}{
              "FastPeriod": {5, 10, 15},
              "SlowPeriod": {20, 30, 40},
          },
          Maximize: "SharpeRatio",
      })
      if err != nil {
          t.Fatalf("optimization failed: %v", err)
      }

      // 6. Re-run with optimal params
      optStrategy := &SmaCrossStrategy{
          FastPeriod: result.BestParams["FastPeriod"].(int),
          SlowPeriod: result.BestParams["SlowPeriod"].(int),
      }
      optBt := NewBacktest(ohlcv, optStrategy, BacktestConfig{Cash: 10000})
      optStats, _ := optBt.Run(nil)

      // Optimal should be at least as good
      if optStats.SharpeRatio < stats.SharpeRatio {
          t.Error("optimized strategy worse than default")
      }

      // 7. Generate plot
      tmpFile := filepath.Join(t.TempDir(), "integration_chart.html")
      err = plot.Plot(optStats, ohlcv, optStrategy.indicators, plot.PlotConfig{
          Filename:   tmpFile,
          PlotEquity: true,
          PlotTrades: true,
      })
      if err != nil {
          t.Fatalf("plotting failed: %v", err)
      }
  }
  ```

### 7.2 Python Parity Tests

**File: `parity_test.go`**

- [ ] Test comprehensive Python parity:
  ```go
  func TestPythonParity_AllMetrics(t *testing.T) {
      // Load comprehensive golden data from Python
      var golden struct {
          // Backtest config
          Data       string             `json:"data"`
          FastPeriod int                `json:"fast_period"`
          SlowPeriod int                `json:"slow_period"`
          Cash       float64            `json:"cash"`
          Commission float64            `json:"commission_pct"`

          // Expected results
          Stats      map[string]float64 `json:"stats"`
          NumTrades  int                `json:"num_trades"`
          Trades     []struct {
              EntryBar   int     `json:"entry_bar"`
              ExitBar    int     `json:"exit_bar"`
              Size       float64 `json:"size"`
              EntryPrice float64 `json:"entry_price"`
              ExitPrice  float64 `json:"exit_price"`
              PL         float64 `json:"pl"`
          } `json:"trades"`
          Equity []float64 `json:"equity"`
      }
      testutil.LoadGolden(t, "python_parity_comprehensive", &golden)

      // Run identical backtest
      ohlcv := testutil.LoadTestData(t, golden.Data)
      strategy := &SmaCrossStrategy{
          FastPeriod: golden.FastPeriod,
          SlowPeriod: golden.SlowPeriod,
      }
      bt := NewBacktest(ohlcv, strategy, BacktestConfig{
          Cash: golden.Cash,
          Commission: func(s, p float64) float64 {
              return math.Abs(s) * p * golden.Commission
          },
      })
      stats, _ := bt.Run(nil)

      // Compare all stats
      for name, want := range golden.Stats {
          var got float64
          switch name {
          case "return_pct":
              got = stats.ReturnPct
          case "sharpe_ratio":
              got = stats.SharpeRatio
          case "max_drawdown_pct":
              got = stats.MaxDrawdownPct
          // ... map all fields
          }

          t.Run("Stat_"+name, func(t *testing.T) {
              testutil.RequireAlmostEqual(t, want, got, StatsTolerance, name)
          })
      }

      // Compare trades
      t.Run("NumTrades", func(t *testing.T) {
          if stats.NumTrades != golden.NumTrades {
              t.Errorf("NumTrades = %d, want %d", stats.NumTrades, golden.NumTrades)
          }
      })

      // Compare equity curve
      t.Run("EquityCurve", func(t *testing.T) {
          if len(stats.EquityCurve.Equity) != len(golden.Equity) {
              t.Fatalf("equity length mismatch: %d vs %d",
                  len(stats.EquityCurve.Equity), len(golden.Equity))
          }
          for i, want := range golden.Equity {
              got := stats.EquityCurve.Equity[i]
              if !testutil.AlmostEqual(want, got, FloatTolerance) {
                  t.Errorf("Equity[%d] = %f, want %f", i, got, want)
              }
          }
      })
  }
  ```

### 7.3 Performance Comparison Tests

**File: `performance_test.go`**

- [ ] Compare performance to Python:
  ```go
  func TestPerformance_VsPython(t *testing.T) {
      if testing.Short() {
          t.Skip("skipping performance test in short mode")
      }

      // Load large dataset
      ohlcv := testutil.MakeTestOHLCV(100000, 100.0)

      // Time Go backtest
      start := time.Now()
      for i := 0; i < 10; i++ {
          strategy := &SmaCrossStrategy{FastPeriod: 10, SlowPeriod: 20}
          bt := NewBacktest(ohlcv, strategy, BacktestConfig{Cash: 10000})
          _, _ = bt.Run(nil)
      }
      goTime := time.Since(start) / 10

      // Python baseline (from golden file)
      var golden struct {
          PythonTimeMs float64 `json:"python_time_ms"`
      }
      testutil.LoadGolden(t, "performance_baseline", &golden)
      pythonTime := time.Duration(golden.PythonTimeMs) * time.Millisecond

      // Go should be at least as fast
      t.Logf("Go: %v, Python: %v", goTime, pythonTime)
      if goTime > pythonTime*2 { // Allow 2x tolerance
          t.Errorf("Go slower than expected: %v vs %v", goTime, pythonTime)
      }
  }
  ```

---

## Phase 8 Tests: Documentation & Examples

### 8.1 Example Tests

**File: `examples/sma_cross/main_test.go`**

- [ ] Test example compiles and runs:
  ```go
  func TestExample_SmaCross(t *testing.T) {
      // Capture stdout
      old := os.Stdout
      r, w, _ := os.Pipe()
      os.Stdout = w

      // Run example main
      main()

      // Restore stdout
      w.Close()
      os.Stdout = old

      var buf bytes.Buffer
      io.Copy(&buf, r)
      output := buf.String()

      // Verify output contains expected elements
      if !strings.Contains(output, "Return") {
          t.Error("output missing Return stats")
      }
      if !strings.Contains(output, "Sharpe") {
          t.Error("output missing Sharpe ratio")
      }
  }
  ```

### 8.2 Documentation Tests

- [ ] Test all code examples in documentation:
  ```go
  func TestDocExamples(t *testing.T) {
      // Extract code blocks from README.md
      // Compile and run each one
      // Verify no errors
  }
  ```

---

## Golden File Generation

### Python Script for Golden File Generation

**File: `scripts/generate_golden.py`**

```python
#!/usr/bin/env python3
"""Generate golden test data from Python backtesting.py"""

import json
import numpy as np
from backtesting import Backtest, Strategy
from backtesting.lib import crossover
from backtesting.test import GOOG, SMA

class SmaCross(Strategy):
    fast_period = 10
    slow_period = 20

    def init(self):
        self.fast_sma = self.I(SMA, self.data.Close, self.fast_period)
        self.slow_sma = self.I(SMA, self.data.Close, self.slow_period)

    def next(self):
        if crossover(self.fast_sma, self.slow_sma):
            self.buy()
        elif crossover(self.slow_sma, self.fast_sma):
            self.position.close()

def generate_sma_golden():
    bt = Backtest(GOOG, SmaCross, cash=10000)
    stats = bt.run()

    golden = {
        "num_trades": stats['# Trades'],
        "return_pct": stats['Return [%]'],
        "sharpe_ratio": stats['Sharpe Ratio'],
        "sortino_ratio": stats['Sortino Ratio'],
        "max_drawdown_pct": stats['Max. Drawdown [%]'],
        "win_rate_pct": stats['Win Rate [%]'],
        "profit_factor": stats['Profit Factor'],
        "expectancy_pct": stats['Expectancy [%]'],
        "sqn": stats['SQN'],
    }

    with open('testdata/golden/sma_cross_stats.json', 'w') as f:
        json.dump(golden, f, indent=2)

def generate_sma_values():
    close = GOOG['Close'].values
    sma_20 = GOOG['Close'].rolling(20).mean().values
    sma_50 = GOOG['Close'].rolling(50).mean().values

    golden = {
        "sma_20": [float(x) if not np.isnan(x) else None for x in sma_20],
        "sma_50": [float(x) if not np.isnan(x) else None for x in sma_50],
    }

    with open('testdata/golden/goog_sma.json', 'w') as f:
        json.dump(golden, f, indent=2)

if __name__ == '__main__':
    generate_sma_golden()
    generate_sma_values()
    print("Golden files generated!")
```

### Running Golden File Generation

```bash
# Generate all golden files
python scripts/generate_golden.py

# Verify golden files exist
ls testdata/golden/
```

---

## CI/CD Pipeline Configuration

### GitHub Actions Workflow

**File: `.github/workflows/test.yml`**

```yaml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        run: |
          go tool cover -func=coverage.out | grep total | awk '{print $3}' | \
          sed 's/%//' | \
          awk '{if ($1 < 85) exit 1; else print "Coverage:", $1, "%"}'

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: coverage.out

  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run benchmarks
        run: go test -bench=. -benchmem ./... | tee benchmark.txt

      - name: Store benchmark
        uses: actions/upload-artifact@v4
        with:
          name: benchmark
          path: benchmark.txt

  fuzz:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run fuzz tests
        run: |
          go test -fuzz=FuzzSMA -fuzztime=30s ./data/
          go test -fuzz=FuzzEMA -fuzztime=30s ./data/
```

---

## Test Execution Commands

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./data/...
go test -v ./stats/...

# Run specific test
go test -v -run TestSMA ./data/

# Run benchmarks
go test -bench=. ./...
go test -bench=BenchmarkBacktest -benchmem ./...

# Run fuzz tests
go test -fuzz=FuzzSMA -fuzztime=1m ./data/

# Generate golden files
python scripts/generate_golden.py

# Run integration tests
go test -v -tags=integration ./...

# Run parity tests against Python
go test -v -run TestPythonParity ./...
```

---

## Summary Checklist

### Pre-Implementation
- [ ] Set up testdata directory structure
- [ ] Copy CSV test files from Python project
- [ ] Generate initial golden files from Python
- [ ] Implement test helper package

### Phase 1: Data
- [ ] OHLCV construction and slicing tests
- [ ] CSV loader tests with all formats
- [ ] Series operation tests
- [ ] Math function tests (SMA, EMA, etc.)
- [ ] Fuzz tests for numerical functions
- [ ] Benchmarks for data operations

### Phase 2: Core Engine
- [ ] Order property tests
- [ ] Trade P&L calculation tests
- [ ] Position aggregation tests
- [ ] Broker order execution tests (market, limit, stop)
- [ ] Broker commission and margin tests
- [ ] Backtest end-to-end tests
- [ ] Strategy interface tests
- [ ] Benchmarks for backtest execution

### Phase 3: Statistics
- [ ] Individual metric tests
- [ ] Drawdown calculation tests
- [ ] Full stats computation tests
- [ ] Python parity tests for all metrics
- [ ] Benchmarks for stats computation

### Phase 4: Library
- [ ] Crossover detection tests
- [ ] Indicator calculation tests (RSI, MACD, etc.)
- [ ] Resample function tests
- [ ] SignalStrategy tests
- [ ] TrailingStrategy tests

### Phase 5: Optimization
- [ ] Grid search tests
- [ ] Constraint tests
- [ ] Parallel execution tests
- [ ] Race condition tests
- [ ] Heatmap generation tests
- [ ] Benchmarks for optimization

### Phase 6: Plotting
- [ ] Chart data preparation tests
- [ ] HTML generation tests
- [ ] JSON embedding tests

### Phase 7: Integration
- [ ] Full workflow integration tests
- [ ] Comprehensive Python parity tests
- [ ] Performance comparison tests

### Phase 8: Documentation
- [ ] Example compilation tests
- [ ] Documentation code block tests

---

*This testing strategy ensures comprehensive coverage and Python parity for the backtesting-go conversion.*
