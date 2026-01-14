package lib

import (
	"math"
)

// RSI calculates the Relative Strength Index.
func RSI(prices []float64, period int) []float64 {
	if len(prices) < period+1 || period <= 0 {
		result := make([]float64, len(prices))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	result := make([]float64, len(prices))
	for i := range result {
		result[i] = math.NaN()
	}

	// Calculate price changes
	changes := make([]float64, len(prices))
	for i := 1; i < len(prices); i++ {
		changes[i] = prices[i] - prices[i-1]
	}

	// Calculate initial average gain and loss
	avgGain := 0.0
	avgLoss := 0.0

	for i := 1; i <= period; i++ {
		if changes[i] > 0 {
			avgGain += changes[i]
		} else {
			avgLoss -= changes[i]
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// First RSI value
	if avgLoss == 0 {
		result[period] = 100
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - (100 / (1 + rs))
	}

	// Calculate subsequent RSI values using smoothed averages
	for i := period + 1; i < len(prices); i++ {
		change := changes[i]
		gain := 0.0
		loss := 0.0

		if change > 0 {
			gain = change
		} else {
			loss = -change
		}

		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}

	return result
}

// MACD calculates the Moving Average Convergence Divergence.
// Returns macd line, signal line, and histogram.
func MACD(prices []float64, fast, slow, signal int) (macd, signalLine, hist []float64) {
	if len(prices) == 0 {
		return nil, nil, nil
	}

	macd = make([]float64, len(prices))
	signalLine = make([]float64, len(prices))
	hist = make([]float64, len(prices))

	for i := range macd {
		macd[i] = math.NaN()
		signalLine[i] = math.NaN()
		hist[i] = math.NaN()
	}

	// Calculate EMAs
	fastEMA := ema(prices, fast)
	slowEMA := ema(prices, slow)

	// Calculate MACD line
	for i := 0; i < len(prices); i++ {
		if !math.IsNaN(fastEMA[i]) && !math.IsNaN(slowEMA[i]) {
			macd[i] = fastEMA[i] - slowEMA[i]
		}
	}

	// Calculate signal line (EMA of MACD)
	signalLine = ema(macd, signal)

	// Calculate histogram
	for i := 0; i < len(prices); i++ {
		if !math.IsNaN(macd[i]) && !math.IsNaN(signalLine[i]) {
			hist[i] = macd[i] - signalLine[i]
		}
	}

	return macd, signalLine, hist
}

// BollingerBands calculates Bollinger Bands.
// Returns upper band, middle band (SMA), and lower band.
func BollingerBands(prices []float64, period int, stdDev float64) (upper, middle, lower []float64) {
	if len(prices) == 0 || period <= 0 {
		return nil, nil, nil
	}

	upper = make([]float64, len(prices))
	middle = make([]float64, len(prices))
	lower = make([]float64, len(prices))

	for i := range upper {
		upper[i] = math.NaN()
		middle[i] = math.NaN()
		lower[i] = math.NaN()
	}

	// Calculate SMA and standard deviation
	for i := period - 1; i < len(prices); i++ {
		// Calculate mean
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += prices[j]
		}
		mean := sum / float64(period)
		middle[i] = mean

		// Calculate standard deviation
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := prices[j] - mean
			variance += diff * diff
		}
		std := math.Sqrt(variance / float64(period))

		upper[i] = mean + stdDev*std
		lower[i] = mean - stdDev*std
	}

	return upper, middle, lower
}

// ATR calculates the Average True Range.
func ATR(high, low, close []float64, period int) []float64 {
	n := len(close)
	if n == 0 || len(high) != n || len(low) != n || period <= 0 {
		return nil
	}

	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	// Calculate True Range
	tr := make([]float64, n)
	tr[0] = high[0] - low[0] // First TR is just high-low

	for i := 1; i < n; i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// Calculate ATR using Wilder's smoothing (EMA)
	// First ATR is simple average of first 'period' TR values
	if period <= n {
		sum := 0.0
		for i := 0; i < period; i++ {
			sum += tr[i]
		}
		result[period-1] = sum / float64(period)

		// Subsequent ATR values use smoothing
		multiplier := 1.0 / float64(period)
		for i := period; i < n; i++ {
			result[i] = (result[i-1] * float64(period-1) + tr[i]) * multiplier
		}
	}

	return result
}

// ADX calculates the Average Directional Index.
func ADX(high, low, close []float64, period int) []float64 {
	n := len(close)
	if n == 0 || len(high) != n || len(low) != n || period <= 0 {
		return nil
	}

	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	// Calculate +DM, -DM, and TR
	plusDM := make([]float64, n)
	minusDM := make([]float64, n)
	tr := make([]float64, n)

	for i := 1; i < n; i++ {
		upMove := high[i] - high[i-1]
		downMove := low[i-1] - low[i]

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}

		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// Smooth the values
	smoothedPlusDM := wilderSmooth(plusDM, period)
	smoothedMinusDM := wilderSmooth(minusDM, period)
	smoothedTR := wilderSmooth(tr, period)

	// Calculate +DI and -DI
	plusDI := make([]float64, n)
	minusDI := make([]float64, n)
	dx := make([]float64, n)

	for i := period; i < n; i++ {
		if smoothedTR[i] != 0 {
			plusDI[i] = 100 * smoothedPlusDM[i] / smoothedTR[i]
			minusDI[i] = 100 * smoothedMinusDM[i] / smoothedTR[i]
		}

		diSum := plusDI[i] + minusDI[i]
		if diSum != 0 {
			dx[i] = 100 * math.Abs(plusDI[i]-minusDI[i]) / diSum
		}
	}

	// Calculate ADX (smoothed DX)
	if 2*period <= n {
		// First ADX is average of first 'period' DX values
		sum := 0.0
		for i := period; i < 2*period; i++ {
			sum += dx[i]
		}
		result[2*period-1] = sum / float64(period)

		// Subsequent ADX values
		for i := 2 * period; i < n; i++ {
			result[i] = (result[i-1]*float64(period-1) + dx[i]) / float64(period)
		}
	}

	return result
}

// Stochastic calculates the Stochastic Oscillator.
// Returns %K and %D lines.
func Stochastic(high, low, close []float64, kPeriod, dPeriod int) (k, d []float64) {
	n := len(close)
	if n == 0 || len(high) != n || len(low) != n || kPeriod <= 0 || dPeriod <= 0 {
		return nil, nil
	}

	k = make([]float64, n)
	d = make([]float64, n)

	for i := range k {
		k[i] = math.NaN()
		d[i] = math.NaN()
	}

	// Calculate %K
	for i := kPeriod - 1; i < n; i++ {
		highest := high[i]
		lowest := low[i]

		for j := i - kPeriod + 1; j <= i; j++ {
			if high[j] > highest {
				highest = high[j]
			}
			if low[j] < lowest {
				lowest = low[j]
			}
		}

		if highest != lowest {
			k[i] = 100 * (close[i] - lowest) / (highest - lowest)
		} else {
			k[i] = 50 // When range is 0
		}
	}

	// Calculate %D (SMA of %K)
	for i := kPeriod + dPeriod - 2; i < n; i++ {
		sum := 0.0
		count := 0
		for j := i - dPeriod + 1; j <= i; j++ {
			if !math.IsNaN(k[j]) {
				sum += k[j]
				count++
			}
		}
		if count == dPeriod {
			d[i] = sum / float64(dPeriod)
		}
	}

	return k, d
}

// CCI calculates the Commodity Channel Index.
func CCI(high, low, close []float64, period int) []float64 {
	n := len(close)
	if n == 0 || len(high) != n || len(low) != n || period <= 0 {
		return nil
	}

	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	// Calculate Typical Price
	tp := make([]float64, n)
	for i := 0; i < n; i++ {
		tp[i] = (high[i] + low[i] + close[i]) / 3
	}

	for i := period - 1; i < n; i++ {
		// Calculate SMA of TP
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += tp[j]
		}
		smaTP := sum / float64(period)

		// Calculate Mean Deviation
		mdSum := 0.0
		for j := i - period + 1; j <= i; j++ {
			mdSum += math.Abs(tp[j] - smaTP)
		}
		meanDev := mdSum / float64(period)

		if meanDev != 0 {
			result[i] = (tp[i] - smaTP) / (0.015 * meanDev)
		}
	}

	return result
}

// WilliamsR calculates Williams %R.
func WilliamsR(high, low, close []float64, period int) []float64 {
	n := len(close)
	if n == 0 || len(high) != n || len(low) != n || period <= 0 {
		return nil
	}

	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	for i := period - 1; i < n; i++ {
		highest := high[i]
		lowest := low[i]

		for j := i - period + 1; j <= i; j++ {
			if high[j] > highest {
				highest = high[j]
			}
			if low[j] < lowest {
				lowest = low[j]
			}
		}

		if highest != lowest {
			result[i] = -100 * (highest - close[i]) / (highest - lowest)
		}
	}

	return result
}

// Helper function: EMA calculation
func ema(values []float64, period int) []float64 {
	if len(values) == 0 || period <= 0 {
		return nil
	}

	result := make([]float64, len(values))
	for i := range result {
		result[i] = math.NaN()
	}

	// Find first non-NaN values for initial SMA
	startIdx := -1
	sum := 0.0
	count := 0

	for i := 0; i < len(values); i++ {
		if !math.IsNaN(values[i]) {
			sum += values[i]
			count++
			if count == period {
				startIdx = i
				break
			}
		}
	}

	if startIdx < 0 {
		return result
	}

	// First EMA value is SMA
	result[startIdx] = sum / float64(period)

	// Calculate subsequent EMA values
	multiplier := 2.0 / float64(period+1)
	for i := startIdx + 1; i < len(values); i++ {
		if !math.IsNaN(values[i]) {
			result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
		} else {
			result[i] = result[i-1]
		}
	}

	return result
}

// Helper function: Wilder's smoothing method
func wilderSmooth(values []float64, period int) []float64 {
	if len(values) == 0 || period <= 0 {
		return nil
	}

	result := make([]float64, len(values))
	for i := range result {
		result[i] = 0
	}

	if period > len(values) {
		return result
	}

	// First value is sum of first 'period' values
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[period-1] = sum

	// Subsequent values use Wilder's smoothing
	for i := period; i < len(values); i++ {
		result[i] = result[i-1] - (result[i-1] / float64(period)) + values[i]
	}

	return result
}
