# Backtesting.go Documentation

## API Documentation

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/izzoa/backtesting.go).

## Examples

See the [examples](../examples/) directory for runnable examples:

| Example | Description |
|---------|-------------|
| [sma_cross](../examples/sma_cross/) | Simple SMA crossover strategy |
| [rsi_optimization](../examples/rsi_optimization/) | RSI strategy with parameter optimization |

Run an example:

```bash
go run ./examples/sma_cross
```

## Building Documentation Locally

Generate and view documentation locally using `godoc`:

```bash
# Install godoc (if not already installed)
go install golang.org/x/tools/cmd/godoc@latest

# Start godoc server
godoc -http=:6060

# Open in browser
open http://localhost:6060/pkg/github.com/izzoa/backtesting.go/
```

## Package Overview

### Main Package (`backtesting`)

The core backtesting engine:

| Type | Description |
|------|-------------|
| `Backtest` | Main backtest runner |
| `BacktestConfig` | Configuration options |
| `BacktestResults` | Results and performance metrics |
| `Strategy` | Interface for trading strategies |
| `StrategyBase` | Base struct to embed in strategies |
| `Broker` | Order execution and position management |
| `Order` | Order representation |
| `Trade` | Completed trade record |
| `Position` | Current position state |

### Data Package (`data/`)

OHLCV data structures and basic indicators:

| Type/Function | Description |
|---------------|-------------|
| `OHLCV` | Time series of price data |
| `Bar` | Single OHLCV bar |
| `Indicator` | Indicator data structure |
| `DataView` | Windowed view of data during backtest |
| `SMA()` | Simple Moving Average |
| `EMA()` | Exponential Moving Average |
| `LoadCSV()` | Load data from CSV files |

### Library Package (`lib/`)

Extended indicators and utilities:

**Indicators:**
| Function | Description |
|----------|-------------|
| `RSI()` | Relative Strength Index |
| `MACD()` | Moving Average Convergence Divergence |
| `BollingerBands()` | Bollinger Bands (upper, middle, lower) |
| `ATR()` | Average True Range |
| `ADX()` | Average Directional Index |
| `Stochastic()` | Stochastic Oscillator (K, D) |
| `CCI()` | Commodity Channel Index |
| `WilliamsR()` | Williams %R |

**Signal Detection:**
| Function | Description |
|----------|-------------|
| `Cross()` | Detect crossover at index |
| `Crossover()` | Detect cross above |
| `Crossunder()` | Detect cross below |
| `BarsSince()` | Bars since condition was true |

**Statistics:**
| Function | Description |
|----------|-------------|
| `Quantile()` | Calculate quantile (0.0-1.0) |
| `Median()` | Calculate median |
| `Percentile()` | Calculate percentile (0-100) |
| `IQR()` | Interquartile range |

**Data Generation:**
| Function | Description |
|----------|-------------|
| `RandomOHLCData()` | Generate random OHLCV data |
| `Resample()` | Resample to higher timeframe |
| `ResampleApply()` | Custom resampling |

### Optimize Package (`optimize/`)

Parameter optimization:

| Type/Function | Description |
|---------------|-------------|
| `GridSearch()` | Grid search optimization |
| `GridConfig` | Grid search configuration |
| `Result` | Optimization results |
| `ParamResult` | Single parameter combination result |
| `Heatmap` | 2D heatmap data |
| `ConstraintFunc` | Parameter constraint function |
| `Range()` | Integer range generator |
| `RangeFloat()` | Float range generator |

**Built-in Constraints:**
| Function | Description |
|----------|-------------|
| `ParamLessThan()` | a < b |
| `ParamGreaterThan()` | a > b |
| `ParamNotEqual()` | a != b |
| `ParamMinValue()` | a >= value |
| `ParamMaxValue()` | a <= value |
| `ParamRange()` | min <= a <= max |
| `And()` | Combine constraints (all must pass) |
| `Or()` | Combine constraints (any can pass) |

### Plot Package (`plot/`)

HTML chart visualization:

| Type/Function | Description |
|---------------|-------------|
| `Plot()` | Generate HTML chart |
| `PrepareChartData()` | Prepare data for plotting |
| `Config` | Plot configuration |
| `ChartData` | Data structure for charts |
| `TradeInfo` | Trade data for markers |
| `TradeMarker` | Entry/exit marker |
| `IndicatorData` | Indicator overlay data |

### Stats Package (`stats/`)

Trading statistics:

| Type/Function | Description |
|---------------|-------------|
| `Stats` | Comprehensive statistics struct |
| `Compute()` | Calculate all statistics |
| `ComputeConfig` | Configuration for computation |
| `TradeRecord` | Individual trade data |
| `TradeStats` | Aggregated trade statistics |
| `EquityCurve` | Equity curve with derived metrics |
| `DrawdownInfo` | Drawdown period information |

**Available Metrics:**
- Return metrics: Total return, annualized return, CAGR, alpha
- Risk metrics: Volatility, max drawdown, average drawdown, drawdown duration
- Risk-adjusted: Sharpe ratio, Sortino ratio, Calmar ratio
- Trade stats: Win rate, profit factor, expectancy, best/worst trade
- Advanced: SQN, Kelly criterion, exposure time

## Testing

Run all tests:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

Run benchmarks:

```bash
go test -bench=. ./...
```

Run with race detector:

```bash
go test -race ./...
```
