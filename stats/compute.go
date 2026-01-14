package stats

import (
	"math"
	"time"
)

// ComputeConfig contains the inputs needed to compute statistics.
type ComputeConfig struct {
	// Trade data
	Trades []*TradeRecord

	// Equity curve
	EquityCurve []float64
	Times       []time.Time

	// Initial state
	InitialCash float64

	// Benchmark data (for alpha/beta)
	BenchmarkReturns []float64

	// Configuration
	RiskFreeRate float64 // Annual risk-free rate (default 0.03)

	// Total bars in backtest (for exposure calculation)
	TotalBars int
}

// Compute calculates all statistics from the given configuration.
func Compute(cfg ComputeConfig) *Stats {
	stats := &Stats{}

	// Set defaults
	riskFreeRate := cfg.RiskFreeRate
	if riskFreeRate == 0 {
		riskFreeRate = DefaultRiskFreeRate
	}

	// Time metrics
	if len(cfg.Times) > 0 {
		stats.Start = cfg.Times[0]
		stats.End = cfg.Times[len(cfg.Times)-1]
		stats.Duration = stats.End.Sub(stats.Start)
	}

	// Build equity curve
	if len(cfg.EquityCurve) > 0 {
		stats.EquityCurve = NewEquityCurve(cfg.Times, cfg.EquityCurve)
	}

	// Return metrics
	if len(cfg.EquityCurve) > 0 {
		initialEquity := cfg.InitialCash
		if initialEquity == 0 && len(cfg.EquityCurve) > 0 {
			initialEquity = cfg.EquityCurve[0]
		}
		finalEquity := cfg.EquityCurve[len(cfg.EquityCurve)-1]

		stats.ReturnPct = CalcReturn(initialEquity, finalEquity)

		// Annualized return
		days := len(cfg.EquityCurve)
		stats.ReturnAnnPct = CalcAnnualizedReturn(stats.ReturnPct, days)

		// CAGR
		years := stats.Duration.Hours() / 24 / 365.25
		if years > 0 {
			stats.CAGR = CalcCAGR(initialEquity, finalEquity, years)
		}

		// Daily returns
		returns := CalcDailyReturns(cfg.EquityCurve)

		// Volatility
		stats.VolatilityAnnPct = CalcVolatility(returns)

		// Sharpe ratio
		stats.SharpeRatio = CalcSharpeRatio(returns, riskFreeRate)

		// Sortino ratio
		stats.SortinoRatio = CalcSortinoRatio(returns, riskFreeRate)

		// Drawdown metrics
		stats.MaxDrawdownPct, stats.MaxDrawdownValue = CalcMaxDrawdown(cfg.EquityCurve)

		// Calmar ratio
		if stats.MaxDrawdownPct > 0 {
			stats.CalmarRatio = CalcCalmarRatio(stats.ReturnAnnPct, stats.MaxDrawdownPct)
		}

		// Drawdown analysis
		drawdowns := CalcDrawdowns(cfg.EquityCurve, cfg.Times)
		if len(drawdowns) > 0 {
			stats.AvgDrawdownPct = CalcAvgDrawdown(drawdowns)
			stats.MaxDrawdownDuration, stats.AvgDrawdownDuration = CalcDrawdownDurations(drawdowns, cfg.Times)
		}

		// Beta and Alpha (if benchmark provided)
		if len(cfg.BenchmarkReturns) > 0 {
			stats.Beta = CalcBeta(returns, cfg.BenchmarkReturns)
			benchmarkReturn := 0.0
			for _, r := range cfg.BenchmarkReturns {
				if !math.IsNaN(r) {
					benchmarkReturn += r
				}
			}
			benchmarkReturn *= 100 // Convert to percentage
			stats.AlphaPct = CalcAlpha(stats.ReturnPct, benchmarkReturn, stats.Beta, riskFreeRate*100)
		}
	}

	// Trade statistics
	stats.NumTrades = len(cfg.Trades)
	stats.Trades = cfg.Trades

	if len(cfg.Trades) > 0 {
		tradeStats := CalcTradeStats(cfg.Trades)

		stats.WinRatePct = tradeStats.WinRate
		stats.BestTradePct = tradeStats.BestTradePct
		stats.WorstTradePct = tradeStats.WorstTradePct
		stats.AvgTradePct = tradeStats.AvgPLPct
		stats.MaxTradeDuration = tradeStats.MaxDuration
		stats.AvgTradeDuration = tradeStats.AvgDuration
		stats.ProfitFactor = tradeStats.ProfitFactor
		stats.ExpectancyPct = tradeStats.Expectancy
		stats.SQN = tradeStats.SQN

		// Kelly criterion
		if tradeStats.AvgLossPct != 0 {
			stats.KellyCriterion = CalcKellyCriterion(
				tradeStats.WinRate,
				tradeStats.AvgWinPct,
				math.Abs(tradeStats.AvgLossPct),
			)
		}

		// Exposure time
		if cfg.TotalBars > 0 {
			totalBarsHeld := 0
			for _, t := range cfg.Trades {
				totalBarsHeld += t.BarsHeld
			}
			stats.ExposureTimePct = float64(totalBarsHeld) / float64(cfg.TotalBars) * 100
		}
	}

	// Buy and hold return (if we have price data)
	if len(cfg.EquityCurve) > 0 && cfg.InitialCash > 0 {
		// This would require the actual price data to calculate properly
		// For now, we skip this unless benchmark returns are provided
	}

	return stats
}

// ComputeFromEquity is a convenience function to compute stats from just equity curve.
func ComputeFromEquity(equity []float64, times []time.Time) *Stats {
	return Compute(ComputeConfig{
		EquityCurve: equity,
		Times:       times,
	})
}
