package stats

import (
	"time"
)

// TradeRecord is a serializable record of a completed trade.
type TradeRecord struct {
	ID         int64
	Size       float64
	EntryTime  time.Time
	ExitTime   time.Time
	EntryPrice float64
	ExitPrice  float64
	PL         float64
	PLPct      float64
	ReturnPct  float64
	Duration   time.Duration
	BarsHeld   int
	Tag        interface{}
	Indicators map[string]float64 // Indicator values at entry
}

// IsWin returns true if the trade was profitable.
func (tr *TradeRecord) IsWin() bool {
	return tr.PL >= 0
}

// IsLong returns true if this was a long trade.
func (tr *TradeRecord) IsLong() bool {
	return tr.Size > 0
}

// IsShort returns true if this was a short trade.
func (tr *TradeRecord) IsShort() bool {
	return tr.Size < 0
}

// TradeStats aggregates statistics from a slice of trade records.
type TradeStats struct {
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	WinRate       float64

	TotalPL       float64
	GrossProfit   float64
	GrossLoss     float64

	AvgPL         float64
	AvgPLPct      float64
	AvgWinPct     float64
	AvgLossPct    float64

	BestTradePL   float64
	BestTradePct  float64
	WorstTradePL  float64
	WorstTradePct float64

	AvgDuration   time.Duration
	MaxDuration   time.Duration
	AvgBarsHeld   float64

	ProfitFactor  float64
	Expectancy    float64
	SQN           float64
}

// CalcTradeStats computes statistics from trade records.
func CalcTradeStats(trades []*TradeRecord) *TradeStats {
	if len(trades) == 0 {
		return &TradeStats{}
	}

	stats := &TradeStats{
		TotalTrades:   len(trades),
		BestTradePL:   trades[0].PL,
		BestTradePct:  trades[0].PLPct,
		WorstTradePL:  trades[0].PL,
		WorstTradePct: trades[0].PLPct,
	}

	var totalDuration time.Duration
	var totalBars int
	var sumPLPct float64
	var sumWinPct float64
	var sumLossPct float64
	var plPcts []float64

	for _, t := range trades {
		stats.TotalPL += t.PL
		sumPLPct += t.PLPct
		plPcts = append(plPcts, t.PLPct)

		if t.PL >= 0 {
			stats.WinningTrades++
			stats.GrossProfit += t.PL
			sumWinPct += t.PLPct
		} else {
			stats.LosingTrades++
			stats.GrossLoss += -t.PL
			sumLossPct += t.PLPct
		}

		if t.PL > stats.BestTradePL {
			stats.BestTradePL = t.PL
		}
		if t.PLPct > stats.BestTradePct {
			stats.BestTradePct = t.PLPct
		}
		if t.PL < stats.WorstTradePL {
			stats.WorstTradePL = t.PL
		}
		if t.PLPct < stats.WorstTradePct {
			stats.WorstTradePct = t.PLPct
		}

		totalDuration += t.Duration
		if t.Duration > stats.MaxDuration {
			stats.MaxDuration = t.Duration
		}
		totalBars += t.BarsHeld
	}

	// Calculate averages
	n := float64(stats.TotalTrades)
	stats.AvgPL = stats.TotalPL / n
	stats.AvgPLPct = sumPLPct / n
	stats.AvgDuration = totalDuration / time.Duration(stats.TotalTrades)
	stats.AvgBarsHeld = float64(totalBars) / n

	if stats.WinningTrades > 0 {
		stats.WinRate = float64(stats.WinningTrades) / n * 100
		stats.AvgWinPct = sumWinPct / float64(stats.WinningTrades)
	}

	if stats.LosingTrades > 0 {
		stats.AvgLossPct = sumLossPct / float64(stats.LosingTrades)
	}

	// Profit factor
	stats.ProfitFactor = CalcProfitFactor(func() []float64 {
		pls := make([]float64, len(trades))
		for i, t := range trades {
			pls[i] = t.PL
		}
		return pls
	}())

	// Expectancy and SQN
	stats.Expectancy = stats.AvgPLPct
	stats.SQN = CalcSQN(plPcts)

	return stats
}
