package lib

import (
	"math"
	"math/rand"
	"time"

	"github.com/izzoa/backtesting.go/data"
)

// RandomOHLCData generates synthetic OHLCV data based on the statistical
// properties of the example data.
func RandomOHLCData(example *data.OHLCV, frac float64, seed int64) *data.OHLCV {
	if example == nil || example.Len() == 0 {
		return data.NewOHLCV(0)
	}

	rng := rand.New(rand.NewSource(seed))

	// Calculate returns from example
	returns := make([]float64, example.Len()-1)
	for i := 1; i < example.Len(); i++ {
		if example.Close[i-1] != 0 {
			returns[i-1] = (example.Close[i] - example.Close[i-1]) / example.Close[i-1]
		}
	}

	// Calculate statistics
	meanReturn := mean(returns)
	stdReturn := stdDev(returns, meanReturn)

	// Determine number of bars to generate
	numBars := int(float64(example.Len()) * frac)
	if numBars < 1 {
		numBars = 1
	}

	result := data.NewOHLCV(numBars)

	// Start from example's starting price
	price := example.Close[0]
	startTime := example.Time[0]

	// Estimate average bar duration
	var avgDuration time.Duration
	if example.Len() > 1 {
		totalDuration := example.Time[example.Len()-1].Sub(example.Time[0])
		avgDuration = totalDuration / time.Duration(example.Len()-1)
	} else {
		avgDuration = 24 * time.Hour
	}

	for i := 0; i < numBars; i++ {
		// Generate random return using normal distribution
		ret := rng.NormFloat64()*stdReturn + meanReturn

		// Calculate OHLC for this bar
		open := price
		closePrice := price * (1 + ret)

		// Generate high and low with some randomness
		volatility := math.Abs(ret) + stdReturn*0.5
		high := math.Max(open, closePrice) * (1 + rng.Float64()*volatility)
		low := math.Min(open, closePrice) * (1 - rng.Float64()*volatility)

		// Generate volume (random around example mean)
		var volume float64
		if len(example.Volume) > 0 {
			avgVolume := mean(example.Volume)
			volume = avgVolume * (0.5 + rng.Float64())
		}

		bar := data.Bar{
			Time:   startTime.Add(time.Duration(i) * avgDuration),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		}

		result.Append(bar)
		price = closePrice
	}

	return result
}

// ShuffleReturns shuffles the daily returns of the data to create
// a random walk with the same return distribution.
func ShuffleReturns(d *data.OHLCV, seed int64) *data.OHLCV {
	if d == nil || d.Len() < 2 {
		return d
	}

	rng := rand.New(rand.NewSource(seed))

	// Calculate returns
	returns := make([]float64, d.Len()-1)
	for i := 1; i < d.Len(); i++ {
		if d.Close[i-1] != 0 {
			returns[i-1] = (d.Close[i] - d.Close[i-1]) / d.Close[i-1]
		}
	}

	// Shuffle returns
	rng.Shuffle(len(returns), func(i, j int) {
		returns[i], returns[j] = returns[j], returns[i]
	})

	// Reconstruct OHLCV with shuffled returns
	result := data.NewOHLCV(d.Len())

	// First bar is copied as-is
	result.Append(d.At(0))

	for i := 1; i < d.Len(); i++ {
		prevClose := result.Close[i-1]
		newClose := prevClose * (1 + returns[i-1])

		// Scale OHLC relative to the new close
		ratio := newClose / d.Close[i]
		if d.Close[i] == 0 {
			ratio = 1
		}

		bar := data.Bar{
			Time:   d.Time[i],
			Open:   d.Open[i] * ratio,
			High:   d.High[i] * ratio,
			Low:    d.Low[i] * ratio,
			Close:  newClose,
			Volume: d.Volume[i],
		}

		result.Append(bar)
	}

	return result
}

// GeometricBrownianMotion generates synthetic OHLCV data using GBM.
func GeometricBrownianMotion(numBars int, startPrice, annualDrift, annualVol float64, startTime time.Time, barDuration time.Duration, seed int64) *data.OHLCV {
	if numBars <= 0 {
		return data.NewOHLCV(0)
	}

	rng := rand.New(rand.NewSource(seed))

	// Convert annual parameters to per-bar
	barsPerYear := float64(365*24*time.Hour) / float64(barDuration)
	dt := 1.0 / barsPerYear
	drift := annualDrift * dt
	vol := annualVol * math.Sqrt(dt)

	result := data.NewOHLCV(numBars)
	price := startPrice

	for i := 0; i < numBars; i++ {
		// GBM: dS = mu*S*dt + sigma*S*dW
		// Discrete: S(t+dt) = S(t) * exp((mu - 0.5*sigma^2)*dt + sigma*sqrt(dt)*Z)
		z := rng.NormFloat64()
		ret := (drift - 0.5*vol*vol) + vol*z
		newPrice := price * math.Exp(ret)

		// Generate OHLC
		open := price
		closePrice := newPrice
		high := math.Max(open, closePrice) * (1 + math.Abs(rng.NormFloat64()*vol*0.5))
		low := math.Min(open, closePrice) * (1 - math.Abs(rng.NormFloat64()*vol*0.5))

		bar := data.Bar{
			Time:   startTime.Add(time.Duration(i) * barDuration),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: 1000000 * (0.5 + rng.Float64()),
		}

		result.Append(bar)
		price = newPrice
	}

	return result
}

// helper functions
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64, meanVal float64) float64 {
	if len(values) < 2 {
		return 0
	}
	variance := 0.0
	for _, v := range values {
		diff := v - meanVal
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(len(values)-1))
}
