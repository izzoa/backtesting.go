package stats

import (
	"fmt"
	"time"
)

// Stats contains comprehensive backtest statistics.
type Stats struct {
	// Time
	Start           time.Time
	End             time.Time
	Duration        time.Duration
	ExposureTimePct float64

	// Returns
	ReturnPct        float64
	ReturnAnnPct     float64
	BuyHoldReturnPct float64
	CAGR             float64
	AlphaPct         float64

	// Risk
	VolatilityAnnPct    float64
	Beta                float64
	MaxDrawdownPct      float64
	MaxDrawdownValue    float64
	AvgDrawdownPct      float64
	MaxDrawdownDuration time.Duration
	AvgDrawdownDuration time.Duration

	// Ratios
	SharpeRatio  float64
	SortinoRatio float64
	CalmarRatio  float64

	// Trade Statistics
	NumTrades        int
	WinRatePct       float64
	BestTradePct     float64
	WorstTradePct    float64
	AvgTradePct      float64
	MaxTradeDuration time.Duration
	AvgTradeDuration time.Duration
	ProfitFactor     float64
	ExpectancyPct    float64
	SQN              float64
	KellyCriterion   float64

	// Data
	EquityCurve *EquityCurve
	Trades      []*TradeRecord
}

// String returns a formatted summary of the statistics.
func (s *Stats) String() string {
	return fmt.Sprintf(`Backtest Statistics
===================
Period: %s to %s (%v)
Exposure Time: %.2f%%

RETURNS
-------
Total Return: %.2f%%
Annualized Return: %.2f%%
Buy & Hold Return: %.2f%%
CAGR: %.2f%%
Alpha: %.2f%%

RISK METRICS
------------
Volatility (Ann.): %.2f%%
Max Drawdown: %.2f%%
Avg Drawdown: %.2f%%
Max DD Duration: %v
Avg DD Duration: %v

RISK-ADJUSTED RETURNS
---------------------
Sharpe Ratio: %.2f
Sortino Ratio: %.2f
Calmar Ratio: %.2f

TRADE STATISTICS
----------------
Total Trades: %d
Win Rate: %.2f%%
Best Trade: %.2f%%
Worst Trade: %.2f%%
Avg Trade: %.2f%%
Max Trade Duration: %v
Avg Trade Duration: %v
Profit Factor: %.2f
Expectancy: %.2f%%
SQN: %.2f
Kelly Criterion: %.2f%%`,
		s.Start.Format("2006-01-02"),
		s.End.Format("2006-01-02"),
		s.Duration,
		s.ExposureTimePct,
		s.ReturnPct,
		s.ReturnAnnPct,
		s.BuyHoldReturnPct,
		s.CAGR,
		s.AlphaPct,
		s.VolatilityAnnPct,
		s.MaxDrawdownPct,
		s.AvgDrawdownPct,
		s.MaxDrawdownDuration,
		s.AvgDrawdownDuration,
		s.SharpeRatio,
		s.SortinoRatio,
		s.CalmarRatio,
		s.NumTrades,
		s.WinRatePct,
		s.BestTradePct,
		s.WorstTradePct,
		s.AvgTradePct,
		s.MaxTradeDuration,
		s.AvgTradeDuration,
		s.ProfitFactor,
		s.ExpectancyPct,
		s.SQN,
		s.KellyCriterion)
}

// ToMap returns the statistics as a map for JSON export.
func (s *Stats) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"start":                s.Start,
		"end":                  s.End,
		"duration":             s.Duration.String(),
		"exposure_time_pct":    s.ExposureTimePct,
		"return_pct":           s.ReturnPct,
		"return_ann_pct":       s.ReturnAnnPct,
		"buy_hold_return_pct":  s.BuyHoldReturnPct,
		"cagr":                 s.CAGR,
		"alpha_pct":            s.AlphaPct,
		"volatility_ann_pct":   s.VolatilityAnnPct,
		"beta":                 s.Beta,
		"max_drawdown_pct":     s.MaxDrawdownPct,
		"max_drawdown_value":   s.MaxDrawdownValue,
		"avg_drawdown_pct":     s.AvgDrawdownPct,
		"max_dd_duration":      s.MaxDrawdownDuration.String(),
		"avg_dd_duration":      s.AvgDrawdownDuration.String(),
		"sharpe_ratio":         s.SharpeRatio,
		"sortino_ratio":        s.SortinoRatio,
		"calmar_ratio":         s.CalmarRatio,
		"num_trades":           s.NumTrades,
		"win_rate_pct":         s.WinRatePct,
		"best_trade_pct":       s.BestTradePct,
		"worst_trade_pct":      s.WorstTradePct,
		"avg_trade_pct":        s.AvgTradePct,
		"max_trade_duration":   s.MaxTradeDuration.String(),
		"avg_trade_duration":   s.AvgTradeDuration.String(),
		"profit_factor":        s.ProfitFactor,
		"expectancy_pct":       s.ExpectancyPct,
		"sqn":                  s.SQN,
		"kelly_criterion":      s.KellyCriterion,
	}
}
