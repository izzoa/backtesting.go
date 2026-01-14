package plot

import (
	"time"

	"github.com/quickfixgo/backtesting/data"
)

// ChartData contains all data needed for chart rendering.
type ChartData struct {
	// OHLCV data
	Times  []time.Time
	Open   []float64
	High   []float64
	Low    []float64
	Close  []float64
	Volume []float64

	// Equity curve
	Equity      []float64
	DrawdownPct []float64

	// Trades
	Trades []TradeMarker

	// Indicators
	Indicators []IndicatorData

	// Metadata
	StartTime   time.Time
	EndTime     time.Time
	NumTrades   int
	ReturnPct   float64
	MaxDrawdown float64
}

// TradeMarker represents a trade entry or exit point.
type TradeMarker struct {
	Time       time.Time
	Price      float64
	Type       string  // "entry_long", "entry_short", "exit_long", "exit_short"
	Size       float64
	PL         float64
	PLPct      float64
	TradeIndex int // Index of the trade in the trades list
}

// IndicatorData contains data for a single indicator.
type IndicatorData struct {
	Name    string
	Values  []float64
	Overlay bool   // If true, overlay on price chart
	Color   string // CSS color
}

// PrepareChartData converts backtest results to chart data.
func PrepareChartData(
	ohlcv *data.OHLCV,
	equity []float64,
	trades []TradeInfo,
	indicators []*data.Indicator,
) *ChartData {
	if ohlcv == nil || ohlcv.Len() == 0 {
		return &ChartData{}
	}

	chartData := &ChartData{
		Times:  ohlcv.Time,
		Open:   ohlcv.Open,
		High:   ohlcv.High,
		Low:    ohlcv.Low,
		Close:  ohlcv.Close,
		Volume: ohlcv.Volume,
		Equity: equity,
	}

	// Calculate drawdown
	if len(equity) > 0 {
		chartData.DrawdownPct = calcDrawdownPct(equity)
	}

	// Set metadata
	if len(ohlcv.Time) > 0 {
		chartData.StartTime = ohlcv.Time[0]
		chartData.EndTime = ohlcv.Time[len(ohlcv.Time)-1]
	}

	// Calculate return
	if len(equity) > 0 {
		initial := equity[0]
		final := equity[len(equity)-1]
		if initial > 0 {
			chartData.ReturnPct = (final - initial) / initial * 100
		}
	}

	// Calculate max drawdown
	if len(chartData.DrawdownPct) > 0 {
		maxDD := 0.0
		for _, dd := range chartData.DrawdownPct {
			if dd > maxDD {
				maxDD = dd
			}
		}
		chartData.MaxDrawdown = maxDD
	}

	// Convert trades to markers
	chartData.NumTrades = len(trades)
	for i, trade := range trades {
		// Entry marker
		entryType := "entry_long"
		if trade.Size < 0 {
			entryType = "entry_short"
		}
		chartData.Trades = append(chartData.Trades, TradeMarker{
			Time:       trade.EntryTime,
			Price:      trade.EntryPrice,
			Type:       entryType,
			Size:       trade.Size,
			TradeIndex: i,
		})

		// Exit marker (if trade is closed)
		if !trade.ExitTime.IsZero() {
			exitType := "exit_long"
			if trade.Size < 0 {
				exitType = "exit_short"
			}
			chartData.Trades = append(chartData.Trades, TradeMarker{
				Time:       trade.ExitTime,
				Price:      trade.ExitPrice,
				Type:       exitType,
				Size:       trade.Size,
				PL:         trade.PL,
				PLPct:      trade.PLPct,
				TradeIndex: i,
			})
		}
	}

	// Convert indicators
	for _, ind := range indicators {
		if ind == nil {
			continue
		}
		chartData.Indicators = append(chartData.Indicators, IndicatorData{
			Name:    ind.Name,
			Values:  ind.Values,
			Overlay: ind.PlotOptions.Overlay,
			Color:   ind.PlotOptions.Color,
		})
	}

	return chartData
}

// TradeInfo contains trade information for chart rendering.
type TradeInfo struct {
	Size       float64
	EntryTime  time.Time
	ExitTime   time.Time
	EntryPrice float64
	ExitPrice  float64
	PL         float64
	PLPct      float64
}

// calcDrawdownPct calculates drawdown percentage at each point.
func calcDrawdownPct(equity []float64) []float64 {
	if len(equity) == 0 {
		return nil
	}

	drawdown := make([]float64, len(equity))
	peak := equity[0]

	for i, eq := range equity {
		if eq > peak {
			peak = eq
		}
		if peak > 0 {
			drawdown[i] = (peak - eq) / peak * 100
		}
	}

	return drawdown
}
