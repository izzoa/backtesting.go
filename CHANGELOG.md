# Changelog

All notable changes to Backtesting.go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-01-14

### Added

Initial release of Backtesting.go, a Go port of [backtesting.py](https://github.com/kernc/backtesting.py).

#### Core Engine
- `Backtest` struct for running backtests on OHLCV data
- `Strategy` interface with `Init()` and `Next()` methods
- `StrategyBase` struct with helper methods for strategy development
- `Broker` for order execution and position management
- Support for market, limit, stop, and stop-limit orders
- Stop-loss (SL) and take-profit (TP) order options
- Commission and spread modeling
- Position sizing and margin calculation

#### Data Package (`data/`)
- `OHLCV` struct for time series price data
- `Bar` struct for individual OHLCV bars
- `Indicator` struct for technical indicator data
- `SMA` (Simple Moving Average) function
- `EMA` (Exponential Moving Average) function
- `LoadCSV` for loading data from CSV files

#### Library Package (`lib/`)
- `RSI` (Relative Strength Index)
- `MACD` (Moving Average Convergence Divergence)
- `BollingerBands` (Bollinger Bands with upper, middle, lower)
- `ATR` (Average True Range)
- `ADX` (Average Directional Index)
- `Stochastic` oscillator (K and D lines)
- `CCI` (Commodity Channel Index)
- `WilliamsR` (Williams %R)
- `Cross`, `Crossover`, `Crossunder` signal detection
- `BarsSince` utility function
- `Quantile`, `Median`, `Percentile`, `IQR` statistical functions
- `Resample` for multi-timeframe analysis
- `RandomOHLCData` for generating test data

#### Optimization Package (`optimize/`)
- `GridSearch` for parameter optimization
- Parallel execution with configurable workers
- Parameter constraints (`ParamLessThan`, `ParamGreaterThan`, etc.)
- Constraint combinators (`And`, `Or`)
- `Range` and `RangeFloat` for parameter value generation
- `Heatmap` generation for visualization
- `Result` struct with sorting and filtering

#### Plot Package (`plot/`)
- Interactive HTML chart generation using Chart.js
- Price chart with OHLC data
- Equity curve visualization
- Drawdown chart
- Trade markers (entry/exit points)
- Indicator overlay support
- Configurable chart options

#### Stats Package (`stats/`)
- Comprehensive `Stats` struct with 40+ metrics
- `Compute` function for calculating all statistics
- Return metrics: Total return, annualized return, CAGR
- Risk metrics: Volatility, max drawdown, average drawdown
- Risk-adjusted returns: Sharpe ratio, Sortino ratio, Calmar ratio
- Trade statistics: Win rate, profit factor, expectancy
- Advanced metrics: SQN, Kelly criterion
- `TradeRecord` and `TradeStats` for trade analysis
- Drawdown analysis with duration tracking

#### Examples
- `sma_cross` - Simple SMA crossover strategy example
- `rsi_optimization` - RSI strategy with parameter optimization

#### Testing
- Comprehensive unit tests for all packages
- Integration tests covering full backtest workflows
- Benchmark tests for performance measurement
