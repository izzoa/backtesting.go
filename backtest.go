package backtesting

import (
	"fmt"
	"math"
	"time"

	"github.com/izzoa/backtesting.go/data"
)

// Backtest orchestrates the backtesting process.
type Backtest struct {
	data     *data.OHLCV
	strategy Strategy
	config   BacktestConfig
}

// BacktestConfig contains configuration for the backtest.
type BacktestConfig struct {
	// Cash is the initial cash amount.
	Cash float64

	// Commission is the commission function. Nil means no commission.
	Commission CommissionFunc

	// Margin is the margin requirement (1.0 = 100%, 0.1 = 10x leverage).
	Margin float64

	// Spread is the bid-ask spread as a fraction.
	Spread float64

	// TradeOnClose fills orders at close price instead of next open.
	TradeOnClose bool

	// Hedging allows simultaneous long and short positions.
	Hedging bool

	// ExclusiveOrders closes previous trades when new orders are placed.
	ExclusiveOrders bool

	// FinalizeTrades closes all open trades at the end of the backtest.
	FinalizeTrades bool
}

// DefaultConfig returns a default backtest configuration.
func DefaultConfig() BacktestConfig {
	return BacktestConfig{
		Cash:           10000,
		Margin:         1.0,
		FinalizeTrades: true,
	}
}

// NewBacktest creates a new backtest with the given data, strategy, and configuration.
func NewBacktest(ohlcv *data.OHLCV, strategy Strategy, cfg BacktestConfig) *Backtest {
	if cfg.Cash == 0 {
		cfg.Cash = 10000
	}
	if cfg.Margin == 0 {
		cfg.Margin = 1.0
	}

	return &Backtest{
		data:     ohlcv,
		strategy: strategy,
		config:   cfg,
	}
}

// Run executes the backtest and returns the results.
func (bt *Backtest) Run() (*Results, error) {
	if bt.data == nil || bt.data.Len() == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	// Create broker
	broker := NewBroker(
		WithData(bt.data),
		WithCash(bt.config.Cash),
		WithMargin(bt.config.Margin),
		WithSpread(bt.config.Spread),
		WithTradeOnClose(bt.config.TradeOnClose),
		WithHedging(bt.config.Hedging),
		WithExclusiveOrders(bt.config.ExclusiveOrders),
	)

	if bt.config.Commission != nil {
		broker.commission = bt.config.Commission
	}

	// Create data view
	dataView := data.NewDataView(bt.data)

	// Initialize strategy
	bt.strategy.SetBroker(broker)
	bt.strategy.SetData(dataView)
	bt.strategy.Init()

	// Calculate warmup period (first bar where all indicators have valid values)
	warmup := bt.calculateWarmup()

	// Main backtesting loop
	for i := warmup; i < bt.data.Len(); i++ {
		// Update data view to show data up to current bar
		dataView.SetLength(i + 1)

		// Update broker state
		broker.SetBar(i)
		broker.Next()

		// Call strategy
		bt.strategy.Next()

		// Check for out-of-money condition
		if broker.Equity() <= 0 {
			break
		}
	}

	// Finalize trades if requested
	if bt.config.FinalizeTrades && len(broker.trades) > 0 {
		lastBar := bt.data.At(bt.data.Len() - 1)
		// Make a copy to avoid modifying slice while iterating
		tradesToClose := make([]*Trade, len(broker.trades))
		copy(tradesToClose, broker.trades)
		for _, trade := range tradesToClose {
			broker.closeTrade(trade, lastBar.Close, lastBar)
		}
	}

	// Compute results
	return bt.computeResults(broker, dataView)
}

// calculateWarmup determines the number of bars to skip for indicator warmup.
func (bt *Backtest) calculateWarmup() int {
	// Try to get indicators from strategy via interface
	if s, ok := bt.strategy.(interface{ Indicators() []*data.Indicator }); ok {
		indicators := s.Indicators()
		return calculateIndicatorWarmup(indicators)
	}
	return 1 // Default: start from bar 1
}

// calculateIndicatorWarmup finds the first bar where all indicators are non-NaN.
func calculateIndicatorWarmup(indicators []*data.Indicator) int {
	if len(indicators) == 0 {
		return 1
	}

	maxWarmup := 1
	for _, ind := range indicators {
		for i, v := range ind.Values {
			if !math.IsNaN(v) {
				if i+1 > maxWarmup {
					maxWarmup = i + 1
				}
				break
			}
		}
	}

	return maxWarmup
}

// computeResults calculates backtest statistics.
func (bt *Backtest) computeResults(broker *Broker, dataView *data.DataView) (*Results, error) {
	results := &Results{
		StartTime:    bt.data.Time[0],
		EndTime:      bt.data.Time[bt.data.Len()-1],
		InitialCash:  broker.initialCash,
		FinalEquity:  broker.Equity(),
		EquityCurve:  broker.EquityCurve(),
		Trades:       broker.ClosedTrades(),
		NumTrades:    len(broker.closedTrades),
		Data:         bt.data,
	}

	// Calculate returns
	if results.InitialCash > 0 {
		results.ReturnPct = (results.FinalEquity - results.InitialCash) / results.InitialCash * 100
	}

	// Calculate trade statistics
	if results.NumTrades > 0 {
		results.calculateTradeStats()
	}

	// Calculate drawdown
	results.calculateDrawdown()

	// Calculate duration
	results.Duration = results.EndTime.Sub(results.StartTime)

	return results, nil
}

// Results contains the results of a backtest.
type Results struct {
	// Time
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	// Equity
	InitialCash float64
	FinalEquity float64
	EquityCurve []float64

	// Returns
	ReturnPct    float64
	ReturnAnnPct float64
	CAGR         float64

	// Risk
	MaxDrawdownPct    float64
	MaxDrawdownValue  float64
	VolatilityAnnPct  float64

	// Trade Statistics
	NumTrades       int
	WinRate         float64
	ProfitFactor    float64
	Expectancy      float64
	AvgTradePct     float64
	AvgWinPct       float64
	AvgLossPct      float64
	BestTradePct    float64
	WorstTradePct   float64

	// Data
	Trades []*Trade
	Data   *data.OHLCV
}

// calculateTradeStats computes trade-related statistics.
func (r *Results) calculateTradeStats() {
	wins := 0
	totalProfit := 0.0
	totalLoss := 0.0
	sumPct := 0.0

	r.BestTradePct = math.Inf(-1)
	r.WorstTradePct = math.Inf(1)

	for _, trade := range r.Trades {
		pct := trade.PLPct()
		sumPct += pct

		if pct > r.BestTradePct {
			r.BestTradePct = pct
		}
		if pct < r.WorstTradePct {
			r.WorstTradePct = pct
		}

		if trade.PL() >= 0 {
			wins++
			totalProfit += trade.PL()
		} else {
			totalLoss += math.Abs(trade.PL())
		}
	}

	r.WinRate = float64(wins) / float64(r.NumTrades) * 100
	r.AvgTradePct = sumPct / float64(r.NumTrades)

	if totalLoss > 0 {
		r.ProfitFactor = totalProfit / totalLoss
	} else if totalProfit > 0 {
		r.ProfitFactor = math.Inf(1)
	}

	// Calculate average win/loss
	if wins > 0 {
		winSum := 0.0
		for _, trade := range r.Trades {
			if trade.PL() >= 0 {
				winSum += trade.PLPct()
			}
		}
		r.AvgWinPct = winSum / float64(wins)
	}

	losses := r.NumTrades - wins
	if losses > 0 {
		lossSum := 0.0
		for _, trade := range r.Trades {
			if trade.PL() < 0 {
				lossSum += trade.PLPct()
			}
		}
		r.AvgLossPct = lossSum / float64(losses)
	}

	// Expectancy
	r.Expectancy = r.AvgTradePct
}

// calculateDrawdown computes maximum drawdown from equity curve.
func (r *Results) calculateDrawdown() {
	if len(r.EquityCurve) == 0 {
		return
	}

	peak := r.EquityCurve[0]
	maxDrawdown := 0.0
	maxDrawdownValue := 0.0

	for _, equity := range r.EquityCurve {
		if equity > peak {
			peak = equity
		}

		drawdown := peak - equity
		drawdownPct := drawdown / peak * 100

		if drawdownPct > maxDrawdown {
			maxDrawdown = drawdownPct
			maxDrawdownValue = drawdown
		}
	}

	r.MaxDrawdownPct = maxDrawdown
	r.MaxDrawdownValue = maxDrawdownValue
}

// String returns a formatted summary of the results.
func (r *Results) String() string {
	return fmt.Sprintf(`Backtest Results
================
Period: %s to %s (%v)
Initial Cash: $%.2f
Final Equity: $%.2f
Return: %.2f%%

Trades: %d
Win Rate: %.2f%%
Profit Factor: %.2f
Avg Trade: %.2f%%
Best Trade: %.2f%%
Worst Trade: %.2f%%

Max Drawdown: %.2f%%`,
		r.StartTime.Format("2006-01-02"),
		r.EndTime.Format("2006-01-02"),
		r.Duration,
		r.InitialCash,
		r.FinalEquity,
		r.ReturnPct,
		r.NumTrades,
		r.WinRate,
		r.ProfitFactor,
		r.AvgTradePct,
		r.BestTradePct,
		r.WorstTradePct,
		r.MaxDrawdownPct)
}
