package lib

import (
	"math"
	"testing"
	"time"

	"github.com/quickfixgo/backtesting/data"
)

func almostEqual(a, b, tol float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Abs(a-b) <= tol
}

func TestCross(t *testing.T) {
	s1 := []float64{1, 2, 3, 2, 1, 2}
	s2 := []float64{2, 2, 2, 2, 2, 2}

	result := Cross(s1, s2)

	// Index 2: 2->3 crosses above 2 (prevDiff=0, currDiff=1 -> cross up)
	// Index 4: 2->1 crosses below 2 (prevDiff=0, currDiff=-1 -> cross down)
	expected := []bool{false, false, true, false, true, false}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Cross at index %d = %v, want %v", i, result[i], exp)
		}
	}
}

func TestCrossover(t *testing.T) {
	s1 := []float64{1, 2, 3, 4, 3, 2}
	s2 := []float64{2, 2, 2, 2, 2, 2}

	result := Crossover(s1, s2)

	// Expected crossover at index 2 (2->3 crosses above 2)
	expected := []bool{false, false, true, false, false, false}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Crossover at index %d = %v, want %v", i, result[i], exp)
		}
	}
}

func TestCrossunder(t *testing.T) {
	s1 := []float64{3, 3, 2, 1, 1, 2}
	s2 := []float64{2, 2, 2, 2, 2, 2}

	result := Crossunder(s1, s2)

	// Expected crossunder at index 2 (3->2 at boundary, but actually 3->2 equals, not crosses under)
	// Actually at index 3: 2->1 crosses below 2
	expected := []bool{false, false, false, true, false, false}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Crossunder at index %d = %v, want %v", i, result[i], exp)
		}
	}
}

func TestBarsSince(t *testing.T) {
	cond := []bool{false, true, false, false, true, false}

	result := BarsSince(cond, -1)

	expected := []int{-1, 0, 1, 2, 0, 1}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("BarsSince at index %d = %d, want %d", i, result[i], exp)
		}
	}
}

func TestRSI(t *testing.T) {
	// Test data with known RSI values
	prices := []float64{44, 44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08,
		45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64}

	rsi := RSI(prices, 14)

	// RSI at index 14 should be around 70 for this upward trending data
	if math.IsNaN(rsi[14]) {
		t.Error("RSI at index 14 should not be NaN")
	}

	// RSI should be between 0 and 100
	for i := 14; i < len(rsi); i++ {
		if rsi[i] < 0 || rsi[i] > 100 {
			t.Errorf("RSI at index %d = %f, should be between 0 and 100", i, rsi[i])
		}
	}
}

func TestMACD(t *testing.T) {
	// Simple test data
	prices := make([]float64, 50)
	for i := range prices {
		prices[i] = 100 + float64(i)
	}

	macd, signal, hist := MACD(prices, 12, 26, 9)

	// Check that we get results
	if macd == nil || signal == nil || hist == nil {
		t.Fatal("MACD returned nil")
	}

	// MACD should be positive for upward trending data after warmup
	for i := 35; i < len(macd); i++ {
		if math.IsNaN(macd[i]) {
			t.Errorf("MACD at index %d should not be NaN", i)
		}
		if macd[i] <= 0 {
			t.Errorf("MACD at index %d = %f, should be positive for uptrend", i, macd[i])
		}
	}
}

func TestBollingerBands(t *testing.T) {
	prices := []float64{10, 11, 12, 11, 10, 11, 12, 13, 12, 11}

	upper, middle, lower := BollingerBands(prices, 5, 2.0)

	if upper == nil || middle == nil || lower == nil {
		t.Fatal("BollingerBands returned nil")
	}

	// After warmup, upper > middle > lower
	for i := 4; i < len(prices); i++ {
		if math.IsNaN(upper[i]) || math.IsNaN(middle[i]) || math.IsNaN(lower[i]) {
			t.Errorf("BollingerBands at index %d has NaN", i)
			continue
		}
		if !(upper[i] >= middle[i] && middle[i] >= lower[i]) {
			t.Errorf("BollingerBands at index %d: upper=%f, middle=%f, lower=%f - invalid order",
				i, upper[i], middle[i], lower[i])
		}
	}
}

func TestATR(t *testing.T) {
	high := []float64{10, 11, 12, 11, 10, 11, 12, 13, 12, 11}
	low := []float64{9, 10, 11, 10, 9, 10, 11, 12, 11, 10}
	close := []float64{9.5, 10.5, 11.5, 10.5, 9.5, 10.5, 11.5, 12.5, 11.5, 10.5}

	atr := ATR(high, low, close, 5)

	if atr == nil {
		t.Fatal("ATR returned nil")
	}

	// ATR should be positive after warmup
	for i := 4; i < len(atr); i++ {
		if math.IsNaN(atr[i]) {
			t.Errorf("ATR at index %d should not be NaN", i)
			continue
		}
		if atr[i] <= 0 {
			t.Errorf("ATR at index %d = %f, should be positive", i, atr[i])
		}
	}
}

func TestStochastic(t *testing.T) {
	high := []float64{10, 11, 12, 13, 14, 15, 14, 13, 12, 11}
	low := []float64{9, 10, 11, 12, 13, 14, 13, 12, 11, 10}
	close := []float64{9.5, 10.5, 11.5, 12.5, 13.5, 14.5, 13.5, 12.5, 11.5, 10.5}

	k, d := Stochastic(high, low, close, 5, 3)

	if k == nil || d == nil {
		t.Fatal("Stochastic returned nil")
	}

	// %K and %D should be between 0 and 100
	for i := 4; i < len(k); i++ {
		if !math.IsNaN(k[i]) && (k[i] < 0 || k[i] > 100) {
			t.Errorf("%%K at index %d = %f, should be between 0 and 100", i, k[i])
		}
	}

	for i := 6; i < len(d); i++ {
		if !math.IsNaN(d[i]) && (d[i] < 0 || d[i] > 100) {
			t.Errorf("%%D at index %d = %f, should be between 0 and 100", i, d[i])
		}
	}
}

func TestQuantile(t *testing.T) {
	series := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		q        float64
		expected float64
	}{
		{0.0, 1.0},
		{0.5, 5.5},
		{1.0, 10.0},
		{0.25, 3.25},
		{0.75, 7.75},
	}

	for _, tt := range tests {
		result := Quantile(series, tt.q)
		if !almostEqual(result, tt.expected, 0.01) {
			t.Errorf("Quantile(%.2f) = %f, want %f", tt.q, result, tt.expected)
		}
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name     string
		series   []float64
		expected float64
	}{
		{"odd count", []float64{1, 2, 3, 4, 5}, 3},
		{"even count", []float64{1, 2, 3, 4}, 2.5},
		{"single", []float64{5}, 5},
		{"unsorted", []float64{5, 1, 3, 2, 4}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Median(tt.series)
			if !almostEqual(result, tt.expected, 0.01) {
				t.Errorf("Median = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestResample(t *testing.T) {
	// Create hourly data
	ohlcv := data.NewOHLCV(24)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 24; i++ {
		bar := data.Bar{
			Time:   start.Add(time.Duration(i) * time.Hour),
			Open:   100 + float64(i),
			High:   105 + float64(i),
			Low:    95 + float64(i),
			Close:  102 + float64(i),
			Volume: 1000,
		}
		ohlcv.Append(bar)
	}

	// Resample to daily
	daily := Resample(ohlcv, RuleDaily)

	if daily.Len() != 1 {
		t.Errorf("Expected 1 daily bar, got %d", daily.Len())
	}

	if daily.Len() > 0 {
		// First bar should have Open = first hourly open
		if daily.Open[0] != 100 {
			t.Errorf("Daily Open = %f, want 100", daily.Open[0])
		}

		// High should be max of all hourly highs
		expectedHigh := 105.0 + 23.0 // Last hour's high
		if daily.High[0] != expectedHigh {
			t.Errorf("Daily High = %f, want %f", daily.High[0], expectedHigh)
		}

		// Low should be min of all hourly lows
		expectedLow := 95.0 // First hour's low
		if daily.Low[0] != expectedLow {
			t.Errorf("Daily Low = %f, want %f", daily.Low[0], expectedLow)
		}

		// Close should be last hourly close
		expectedClose := 102.0 + 23.0
		if daily.Close[0] != expectedClose {
			t.Errorf("Daily Close = %f, want %f", daily.Close[0], expectedClose)
		}

		// Volume should be sum
		expectedVolume := 24000.0
		if daily.Volume[0] != expectedVolume {
			t.Errorf("Daily Volume = %f, want %f", daily.Volume[0], expectedVolume)
		}
	}
}

func TestRandomOHLCData(t *testing.T) {
	// Create example data
	example := data.NewOHLCV(100)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	price := 100.0
	for i := 0; i < 100; i++ {
		bar := data.Bar{
			Time:   start.Add(time.Duration(i) * 24 * time.Hour),
			Open:   price,
			High:   price * 1.02,
			Low:    price * 0.98,
			Close:  price * 1.01,
			Volume: 1000000,
		}
		example.Append(bar)
		price = bar.Close
	}

	// Generate random data
	random := RandomOHLCData(example, 1.0, 42)

	if random.Len() != 100 {
		t.Errorf("Expected 100 bars, got %d", random.Len())
	}

	// Check that prices are positive
	for i := 0; i < random.Len(); i++ {
		if random.Close[i] <= 0 {
			t.Errorf("Random Close at %d = %f, should be positive", i, random.Close[i])
		}
	}
}

func TestGeometricBrownianMotion(t *testing.T) {
	gbm := GeometricBrownianMotion(252, 100, 0.1, 0.2, time.Now(), 24*time.Hour, 42)

	if gbm.Len() != 252 {
		t.Errorf("Expected 252 bars, got %d", gbm.Len())
	}

	// Check that prices are positive
	for i := 0; i < gbm.Len(); i++ {
		if gbm.Close[i] <= 0 {
			t.Errorf("GBM Close at %d = %f, should be positive", i, gbm.Close[i])
		}
	}
}
