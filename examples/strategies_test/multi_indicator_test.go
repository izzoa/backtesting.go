package strategies_test

import (
	"math"
	"testing"

	"github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
	"github.com/izzoa/backtesting.go/lib"
)

// =============================================================================
// Strategy 17: RSI + SMA Filter
// =============================================================================

// RSISMAFilterStrategy combines RSI with SMA trend filter.
type RSISMAFilterStrategy struct {
	backtesting.StrategyBase
	RSIPeriod  int
	SMAPeriod  int
	Overbought float64
	Oversold   float64
	rsi        *data.Indicator
	sma        *data.Indicator
}

func (s *RSISMAFilterStrategy) Init() {
	s.rsi = s.I("RSI", lib.RSI(s.Close(), s.RSIPeriod))
	s.sma = s.I("SMA", data.SMA(s.Close(), s.SMAPeriod))
}

func (s *RSISMAFilterStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	price := s.Close()[idx]
	rsiValue := s.IndicatorAt(s.rsi, idx)
	smaValue := s.IndicatorAt(s.sma, idx)

	if math.IsNaN(rsiValue) || math.IsNaN(smaValue) {
		return
	}

	// Buy when RSI oversold AND price above SMA (trend filter)
	if rsiValue < s.Oversold && price > smaValue && !s.Position().IsLong() {
		s.Buy()
	}

	// Sell when RSI overbought OR price below SMA
	if (rsiValue > s.Overbought || price < smaValue) && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestRSISMAFilter_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &RSISMAFilterStrategy{
		RSIPeriod:  14,
		SMAPeriod:  50,
		Overbought: 70,
		Oversold:   30,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "RSISMAFilter_GOOG")

	t.Logf("RSI+SMA Filter GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestRSISMAFilter_FilterEffect(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Without filter (RSI only)
	rsiOnly := &RSIStrategySimple{Period: 14, Overbought: 70, Oversold: 30}
	resultsRSI := runBacktest(t, ohlcv, rsiOnly)

	// With SMA filter
	withFilter := &RSISMAFilterStrategy{
		RSIPeriod:  14,
		SMAPeriod:  50,
		Overbought: 70,
		Oversold:   30,
	}
	resultsFiltered := runBacktest(t, ohlcv, withFilter)

	assertValidResults(t, resultsRSI, "RSI_Only")
	assertValidResults(t, resultsFiltered, "RSI_Filtered")

	t.Logf("RSI Only: Return=%.2f%%, Trades=%d", resultsRSI.ReturnPct, resultsRSI.NumTrades)
	t.Logf("RSI+SMA Filter: Return=%.2f%%, Trades=%d", resultsFiltered.ReturnPct, resultsFiltered.NumTrades)
}

// RSIStrategySimple is a basic RSI strategy for comparison.
type RSIStrategySimple struct {
	backtesting.StrategyBase
	Period     int
	Overbought float64
	Oversold   float64
	rsi        *data.Indicator
}

func (s *RSIStrategySimple) Init() {
	s.rsi = s.I("RSI", lib.RSI(s.Close(), s.Period))
}

func (s *RSIStrategySimple) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	rsiValue := s.IndicatorAt(s.rsi, idx)
	if math.IsNaN(rsiValue) {
		return
	}

	if rsiValue < s.Oversold && !s.Position().IsLong() {
		s.Buy()
	}
	if rsiValue > s.Overbought && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

// =============================================================================
// Strategy 18: MACD + RSI
// =============================================================================

// MACDRSIStrategy combines MACD crossover with RSI confirmation.
type MACDRSIStrategy struct {
	backtesting.StrategyBase
	MACDFast   int
	MACDSlow   int
	MACDSignal int
	RSIPeriod  int
	RSILevel   float64
	macdLine   *data.Indicator
	signalLine *data.Indicator
	rsi        *data.Indicator
}

func (s *MACDRSIStrategy) Init() {
	macd, signal, _ := lib.MACD(s.Close(), s.MACDFast, s.MACDSlow, s.MACDSignal)
	s.macdLine = s.I("MACD", macd)
	s.signalLine = s.I("Signal", signal)
	s.rsi = s.I("RSI", lib.RSI(s.Close(), s.RSIPeriod))
}

func (s *MACDRSIStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	macdCurr := s.IndicatorAt(s.macdLine, idx)
	macdPrev := s.IndicatorAt(s.macdLine, idx-1)
	signalCurr := s.IndicatorAt(s.signalLine, idx)
	signalPrev := s.IndicatorAt(s.signalLine, idx-1)
	rsiValue := s.IndicatorAt(s.rsi, idx)

	if math.IsNaN(macdCurr) || math.IsNaN(signalCurr) || math.IsNaN(rsiValue) {
		return
	}

	// MACD crosses up AND RSI not overbought - buy
	if macdPrev <= signalPrev && macdCurr > signalCurr && rsiValue < s.RSILevel {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	// MACD crosses down - sell
	if macdPrev >= signalPrev && macdCurr < signalCurr {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestMACDRSI_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &MACDRSIStrategy{
		MACDFast:   12,
		MACDSlow:   26,
		MACDSignal: 9,
		RSIPeriod:  14,
		RSILevel:   70,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "MACDRSI_GOOG")

	t.Logf("MACD+RSI GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestMACDRSI_Confirmation(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Test different RSI confirmation levels
	levels := []float64{50, 60, 70, 80}

	for _, level := range levels {
		t.Run("", func(t *testing.T) {
			strategy := &MACDRSIStrategy{
				MACDFast:   12,
				MACDSlow:   26,
				MACDSignal: 9,
				RSIPeriod:  14,
				RSILevel:   level,
			}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "MACDRSI")

			t.Logf("RSI Level %.0f: Return=%.2f%%, Trades=%d",
				level, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 19: Stochastic + ADX
// =============================================================================

// StochasticADXStrategy combines Stochastic with ADX trend filter.
type StochasticADXStrategy struct {
	backtesting.StrategyBase
	StochK       int
	StochD       int
	ADXPeriod    int
	ADXThreshold float64
	Oversold     float64
	Overbought   float64
	k            *data.Indicator
	adx          *data.Indicator
}

func (s *StochasticADXStrategy) Init() {
	k, _ := lib.Stochastic(s.High(), s.Low(), s.Close(), s.StochK, s.StochD)
	s.k = s.I("Stoch_K", k)
	s.adx = s.I("ADX", lib.ADX(s.High(), s.Low(), s.Close(), s.ADXPeriod))
}

func (s *StochasticADXStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	kValue := s.IndicatorAt(s.k, idx)
	adxValue := s.IndicatorAt(s.adx, idx)

	if math.IsNaN(kValue) || math.IsNaN(adxValue) {
		return
	}

	// Stochastic oversold AND strong trend (ADX > threshold) - buy
	if kValue < s.Oversold && adxValue > s.ADXThreshold && !s.Position().IsLong() {
		s.Buy()
	}

	// Stochastic overbought OR weak trend - sell
	if (kValue > s.Overbought || adxValue < s.ADXThreshold) && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestStochasticADX_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &StochasticADXStrategy{
		StochK:       14,
		StochD:       3,
		ADXPeriod:    14,
		ADXThreshold: 25,
		Oversold:     20,
		Overbought:   80,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "StochasticADX_GOOG")

	t.Logf("Stochastic+ADX GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestStochasticADX_TrendStrength(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	thresholds := []float64{20, 25, 30, 35}

	for _, threshold := range thresholds {
		t.Run("", func(t *testing.T) {
			strategy := &StochasticADXStrategy{
				StochK:       14,
				StochD:       3,
				ADXPeriod:    14,
				ADXThreshold: threshold,
				Oversold:     20,
				Overbought:   80,
			}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "StochasticADX")

			t.Logf("ADX Threshold %.0f: Return=%.2f%%, Trades=%d",
				threshold, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 20: Bollinger + RSI
// =============================================================================

// BollingerRSIStrategy combines Bollinger Bands with RSI confirmation.
type BollingerRSIStrategy struct {
	backtesting.StrategyBase
	BBPeriod    int
	BBStdDev    float64
	RSIPeriod   int
	RSIOversold float64
	upper       *data.Indicator
	lower       *data.Indicator
	rsi         *data.Indicator
}

func (s *BollingerRSIStrategy) Init() {
	upper, _, lower := lib.BollingerBands(s.Close(), s.BBPeriod, s.BBStdDev)
	s.upper = s.I("BB_upper", upper)
	s.lower = s.I("BB_lower", lower)
	s.rsi = s.I("RSI", lib.RSI(s.Close(), s.RSIPeriod))
}

func (s *BollingerRSIStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	price := s.Close()[idx]
	upper := s.IndicatorAt(s.upper, idx)
	lower := s.IndicatorAt(s.lower, idx)
	rsiValue := s.IndicatorAt(s.rsi, idx)

	if math.IsNaN(upper) || math.IsNaN(lower) || math.IsNaN(rsiValue) {
		return
	}

	// Price below lower band AND RSI oversold - strong buy signal
	if price < lower && rsiValue < s.RSIOversold && !s.Position().IsLong() {
		s.Buy()
	}

	// Price above upper band - sell
	if price > upper && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestBollingerRSI_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &BollingerRSIStrategy{
		BBPeriod:    20,
		BBStdDev:    2.0,
		RSIPeriod:   14,
		RSIOversold: 30,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "BollingerRSI_GOOG")

	t.Logf("BB+RSI GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestBollingerRSI_DoubleConfirmation(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Test different RSI oversold levels
	levels := []float64{20, 30, 40}

	for _, level := range levels {
		t.Run("", func(t *testing.T) {
			strategy := &BollingerRSIStrategy{
				BBPeriod:    20,
				BBStdDev:    2.0,
				RSIPeriod:   14,
				RSIOversold: level,
			}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "BollingerRSI")

			t.Logf("RSI Oversold %.0f: Return=%.2f%%, Trades=%d",
				level, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 21: EMA + ATR
// =============================================================================

// EMAATRStrategy uses EMA with ATR-based entry/exit bands.
type EMAATRStrategy struct {
	backtesting.StrategyBase
	EMAPeriod     int
	ATRPeriod     int
	ATRMultiplier float64
	ema           *data.Indicator
	atr           *data.Indicator
}

func (s *EMAATRStrategy) Init() {
	s.ema = s.I("EMA", data.EMA(s.Close(), s.EMAPeriod))
	s.atr = s.I("ATR", lib.ATR(s.High(), s.Low(), s.Close(), s.ATRPeriod))
}

func (s *EMAATRStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	price := s.Close()[idx]
	emaValue := s.IndicatorAt(s.ema, idx)
	atrValue := s.IndicatorAt(s.atr, idx)

	if math.IsNaN(emaValue) || math.IsNaN(atrValue) {
		return
	}

	upperBand := emaValue + (s.ATRMultiplier * atrValue)
	lowerBand := emaValue - (s.ATRMultiplier * atrValue)

	// Price above EMA + ATR - buy
	if price > upperBand && !s.Position().IsLong() {
		s.Buy()
	}

	// Price below EMA - ATR - sell
	if price < lowerBand && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestEMAATR_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &EMAATRStrategy{
		EMAPeriod:     20,
		ATRPeriod:     14,
		ATRMultiplier: 1.5,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "EMAATR_GOOG")

	t.Logf("EMA+ATR GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestEMAATR_DifferentMultipliers(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	multipliers := []float64{1.0, 1.5, 2.0, 2.5}

	for _, mult := range multipliers {
		t.Run("", func(t *testing.T) {
			strategy := &EMAATRStrategy{
				EMAPeriod:     20,
				ATRPeriod:     14,
				ATRMultiplier: mult,
			}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "EMAATR")

			t.Logf("ATR Multiplier %.1f: Return=%.2f%%, Trades=%d",
				mult, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Strategy 22: Triple Confirmation
// =============================================================================

// TripleConfirmationStrategy requires SMA, RSI, and MACD alignment.
type TripleConfirmationStrategy struct {
	backtesting.StrategyBase
	SMAPeriod  int
	RSIPeriod  int
	MACDFast   int
	MACDSlow   int
	MACDSignal int
	sma        *data.Indicator
	rsi        *data.Indicator
	macdLine   *data.Indicator
	signalLine *data.Indicator
}

func (s *TripleConfirmationStrategy) Init() {
	s.sma = s.I("SMA", data.SMA(s.Close(), s.SMAPeriod))
	s.rsi = s.I("RSI", lib.RSI(s.Close(), s.RSIPeriod))
	macd, signal, _ := lib.MACD(s.Close(), s.MACDFast, s.MACDSlow, s.MACDSignal)
	s.macdLine = s.I("MACD", macd)
	s.signalLine = s.I("Signal", signal)
}

func (s *TripleConfirmationStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	price := s.Close()[idx]
	smaValue := s.IndicatorAt(s.sma, idx)
	rsiValue := s.IndicatorAt(s.rsi, idx)
	macdValue := s.IndicatorAt(s.macdLine, idx)
	signalValue := s.IndicatorAt(s.signalLine, idx)

	if math.IsNaN(smaValue) || math.IsNaN(rsiValue) || math.IsNaN(macdValue) {
		return
	}

	// All signals bullish: price > SMA, RSI < 70, MACD > Signal
	allBullish := price > smaValue && rsiValue < 70 && macdValue > signalValue

	// Any signal bearish
	anyBearish := price < smaValue || rsiValue > 80 || macdValue < signalValue

	if allBullish && !s.Position().IsLong() {
		s.Buy()
	}

	if anyBearish && s.Position().IsLong() {
		s.Position().Close(1.0)
	}
}

func TestTripleConfirmation_GOOG(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")
	strategy := &TripleConfirmationStrategy{
		SMAPeriod:  50,
		RSIPeriod:  14,
		MACDFast:   12,
		MACDSlow:   26,
		MACDSignal: 9,
	}
	results := runBacktest(t, ohlcv, strategy)

	assertValidResults(t, results, "TripleConfirmation_GOOG")

	t.Logf("Triple Confirmation GOOG: Return=%.2f%%, Trades=%d, MaxDD=%.2f%%",
		results.ReturnPct, results.NumTrades, results.MaxDrawdownPct)
}

func TestTripleConfirmation_StrictEntry(t *testing.T) {
	ohlcv := loadTestData(t, "GOOG.csv")

	// Test with different SMA periods (stricter trend filter)
	periods := []int{20, 50, 100, 200}

	for _, period := range periods {
		t.Run("", func(t *testing.T) {
			strategy := &TripleConfirmationStrategy{
				SMAPeriod:  period,
				RSIPeriod:  14,
				MACDFast:   12,
				MACDSlow:   26,
				MACDSignal: 9,
			}
			results := runBacktest(t, ohlcv, strategy)
			assertValidResults(t, results, "TripleConfirmation")

			t.Logf("SMA Period %d: Return=%.2f%%, Trades=%d",
				period, results.ReturnPct, results.NumTrades)
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRSISMAFilter(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &RSISMAFilterStrategy{
			RSIPeriod:  14,
			SMAPeriod:  50,
			Overbought: 70,
			Oversold:   30,
		}
	})
}

func BenchmarkMACDRSI(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &MACDRSIStrategy{
			MACDFast:   12,
			MACDSlow:   26,
			MACDSignal: 9,
			RSIPeriod:  14,
			RSILevel:   70,
		}
	})
}

func BenchmarkStochasticADX(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &StochasticADXStrategy{
			StochK:       14,
			StochD:       3,
			ADXPeriod:    14,
			ADXThreshold: 25,
			Oversold:     20,
			Overbought:   80,
		}
	})
}

func BenchmarkBollingerRSI(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &BollingerRSIStrategy{
			BBPeriod:    20,
			BBStdDev:    2.0,
			RSIPeriod:   14,
			RSIOversold: 30,
		}
	})
}

func BenchmarkTripleConfirmation(b *testing.B) {
	runStrategyBenchmark(b, func() backtesting.Strategy {
		return &TripleConfirmationStrategy{
			SMAPeriod:  50,
			RSIPeriod:  14,
			MACDFast:   12,
			MACDSlow:   26,
			MACDSignal: 9,
		}
	})
}
