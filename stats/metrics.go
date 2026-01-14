package stats

import (
	"math"
)

const (
	// TradingDaysPerYear is the typical number of trading days in a year.
	TradingDaysPerYear = 252
	// DefaultRiskFreeRate is the default risk-free rate (3%).
	DefaultRiskFreeRate = 0.03
)

// CalcReturn calculates total return percentage.
func CalcReturn(startEquity, endEquity float64) float64 {
	if startEquity <= 0 {
		return 0
	}
	return (endEquity - startEquity) / startEquity * 100
}

// CalcAnnualizedReturn calculates annualized return from total return and days.
func CalcAnnualizedReturn(totalReturnPct float64, days int) float64 {
	if days <= 0 {
		return 0
	}
	years := float64(days) / TradingDaysPerYear
	if years <= 0 {
		return 0
	}
	// Convert percentage to decimal, calculate, convert back
	totalReturn := totalReturnPct / 100
	annualized := math.Pow(1+totalReturn, 1/years) - 1
	return annualized * 100
}

// CalcCAGR calculates Compound Annual Growth Rate.
func CalcCAGR(startValue, endValue float64, years float64) float64 {
	if startValue <= 0 || years <= 0 {
		return 0
	}
	return (math.Pow(endValue/startValue, 1/years) - 1) * 100
}

// CalcVolatility calculates annualized volatility from daily returns.
func CalcVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	count := 0
	for _, r := range returns {
		if !math.IsNaN(r) {
			sum += r
			count++
		}
	}
	if count < 2 {
		return 0
	}
	mean := sum / float64(count)

	// Calculate variance
	variance := 0.0
	for _, r := range returns {
		if !math.IsNaN(r) {
			diff := r - mean
			variance += diff * diff
		}
	}
	variance /= float64(count - 1)

	// Standard deviation, annualized
	stdDev := math.Sqrt(variance)
	return stdDev * math.Sqrt(TradingDaysPerYear) * 100
}

// CalcSharpeRatio calculates the Sharpe ratio.
func CalcSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate mean daily return
	sum := 0.0
	count := 0
	for _, r := range returns {
		if !math.IsNaN(r) {
			sum += r
			count++
		}
	}
	if count < 2 {
		return 0
	}
	meanReturn := sum / float64(count)

	// Daily risk-free rate
	dailyRf := riskFreeRate / TradingDaysPerYear

	// Calculate excess return
	excessReturn := meanReturn - dailyRf

	// Calculate standard deviation
	variance := 0.0
	for _, r := range returns {
		if !math.IsNaN(r) {
			diff := r - meanReturn
			variance += diff * diff
		}
	}
	variance /= float64(count - 1)
	stdDev := math.Sqrt(variance)

	if stdDev <= 0 {
		return 0
	}

	// Annualize
	return (excessReturn / stdDev) * math.Sqrt(TradingDaysPerYear)
}

// CalcSortinoRatio calculates the Sortino ratio (uses downside deviation).
func CalcSortinoRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate mean daily return
	sum := 0.0
	count := 0
	for _, r := range returns {
		if !math.IsNaN(r) {
			sum += r
			count++
		}
	}
	if count < 2 {
		return 0
	}
	meanReturn := sum / float64(count)

	// Daily risk-free rate
	dailyRf := riskFreeRate / TradingDaysPerYear

	// Calculate excess return
	excessReturn := meanReturn - dailyRf

	// Calculate downside deviation (only negative returns)
	downsideVariance := 0.0
	downsideCount := 0
	for _, r := range returns {
		if !math.IsNaN(r) && r < dailyRf {
			diff := r - dailyRf
			downsideVariance += diff * diff
			downsideCount++
		}
	}

	if downsideCount == 0 {
		if excessReturn > 0 {
			return math.Inf(1) // No downside, positive return
		}
		return 0
	}

	downsideDeviation := math.Sqrt(downsideVariance / float64(downsideCount))
	if downsideDeviation <= 0 {
		return 0
	}

	// Annualize
	return (excessReturn / downsideDeviation) * math.Sqrt(TradingDaysPerYear)
}

// CalcCalmarRatio calculates the Calmar ratio (CAGR / Max Drawdown).
func CalcCalmarRatio(annualReturn, maxDrawdown float64) float64 {
	if maxDrawdown <= 0 {
		if annualReturn > 0 {
			return math.Inf(1)
		}
		return 0
	}
	return annualReturn / maxDrawdown
}

// CalcBeta calculates beta relative to benchmark.
func CalcBeta(strategyReturns, benchmarkReturns []float64) float64 {
	n := len(strategyReturns)
	if n != len(benchmarkReturns) || n < 2 {
		return 0
	}

	// Calculate means
	sumS, sumB := 0.0, 0.0
	count := 0
	for i := 0; i < n; i++ {
		if !math.IsNaN(strategyReturns[i]) && !math.IsNaN(benchmarkReturns[i]) {
			sumS += strategyReturns[i]
			sumB += benchmarkReturns[i]
			count++
		}
	}
	if count < 2 {
		return 0
	}
	meanS := sumS / float64(count)
	meanB := sumB / float64(count)

	// Calculate covariance and benchmark variance
	covariance := 0.0
	variance := 0.0
	for i := 0; i < n; i++ {
		if !math.IsNaN(strategyReturns[i]) && !math.IsNaN(benchmarkReturns[i]) {
			diffS := strategyReturns[i] - meanS
			diffB := benchmarkReturns[i] - meanB
			covariance += diffS * diffB
			variance += diffB * diffB
		}
	}

	if variance <= 0 {
		return 0
	}

	return covariance / variance
}

// CalcAlpha calculates Jensen's alpha.
func CalcAlpha(strategyReturn, benchmarkReturn, beta, riskFreeRate float64) float64 {
	return strategyReturn - (riskFreeRate + beta*(benchmarkReturn-riskFreeRate))
}

// CalcWinRate calculates the win rate percentage from trade P&Ls.
func CalcWinRate(tradePLs []float64) float64 {
	if len(tradePLs) == 0 {
		return 0
	}
	wins := 0
	for _, pl := range tradePLs {
		if pl >= 0 {
			wins++
		}
	}
	return float64(wins) / float64(len(tradePLs)) * 100
}

// CalcProfitFactor calculates gross profit / gross loss.
func CalcProfitFactor(tradePLs []float64) float64 {
	grossProfit := 0.0
	grossLoss := 0.0

	for _, pl := range tradePLs {
		if pl >= 0 {
			grossProfit += pl
		} else {
			grossLoss += math.Abs(pl)
		}
	}

	if grossLoss <= 0 {
		if grossProfit > 0 {
			return math.Inf(1)
		}
		return 0
	}

	return grossProfit / grossLoss
}

// CalcExpectancy calculates the expected value per trade.
func CalcExpectancy(tradePLPcts []float64) float64 {
	if len(tradePLPcts) == 0 {
		return 0
	}

	sum := 0.0
	for _, pct := range tradePLPcts {
		sum += pct
	}
	return sum / float64(len(tradePLPcts))
}

// CalcSQN calculates the System Quality Number.
// SQN = sqrt(N) * (mean R / stddev R) where R = R-multiples of trades.
func CalcSQN(tradePLPcts []float64) float64 {
	n := len(tradePLPcts)
	if n < 2 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, pct := range tradePLPcts {
		sum += pct
	}
	mean := sum / float64(n)

	// Calculate standard deviation
	variance := 0.0
	for _, pct := range tradePLPcts {
		diff := pct - mean
		variance += diff * diff
	}
	variance /= float64(n - 1)
	stdDev := math.Sqrt(variance)

	if stdDev <= 0 {
		return 0
	}

	return math.Sqrt(float64(n)) * (mean / stdDev)
}

// CalcKellyCriterion calculates the optimal bet size percentage.
// Kelly = W - (1-W)/R where W = win rate, R = avg win / avg loss ratio.
func CalcKellyCriterion(winRate, avgWin, avgLoss float64) float64 {
	if avgLoss <= 0 {
		return 0
	}

	w := winRate / 100 // Convert from percentage
	r := avgWin / avgLoss

	kelly := w - (1-w)/r
	return kelly * 100 // Convert to percentage
}

// CalcExposureTime calculates the percentage of time in the market.
func CalcExposureTime(tradeBars []int, totalBars int) float64 {
	if totalBars <= 0 {
		return 0
	}

	exposedBars := 0
	for _, bars := range tradeBars {
		exposedBars += bars
	}

	return float64(exposedBars) / float64(totalBars) * 100
}

// CalcDailyReturns calculates daily returns from equity curve.
func CalcDailyReturns(equity []float64) []float64 {
	if len(equity) < 2 {
		return nil
	}

	returns := make([]float64, len(equity)-1)
	for i := 1; i < len(equity); i++ {
		if equity[i-1] != 0 {
			returns[i-1] = (equity[i] - equity[i-1]) / equity[i-1]
		} else {
			returns[i-1] = math.NaN()
		}
	}
	return returns
}
