package backtesting

import "fmt"

// Position represents the current position (aggregate of all active trades).
type Position struct {
	broker *Broker
}

// NewPosition creates a new position view for the broker.
func NewPosition(broker *Broker) *Position {
	return &Position{broker: broker}
}

// Size returns the net position size (sum of all active trade sizes).
// Positive = net long, negative = net short.
func (p *Position) Size() float64 {
	size := 0.0
	for _, trade := range p.broker.trades {
		size += trade.Size
	}
	return size
}

// PL returns the total unrealized profit/loss of all active trades.
func (p *Position) PL() float64 {
	pl := 0.0
	for _, trade := range p.broker.trades {
		pl += trade.PL()
	}
	return pl
}

// PLPct returns the weighted average P&L percentage of all active trades.
func (p *Position) PLPct() float64 {
	if len(p.broker.trades) == 0 {
		return 0
	}

	totalValue := 0.0
	weightedPL := 0.0
	for _, trade := range p.broker.trades {
		value := trade.Value()
		totalValue += value
		weightedPL += trade.PLPct() * value
	}

	if totalValue == 0 {
		return 0
	}
	return weightedPL / totalValue
}

// IsLong returns true if the net position is long.
func (p *Position) IsLong() bool {
	return p.Size() > 0
}

// IsShort returns true if the net position is short.
func (p *Position) IsShort() bool {
	return p.Size() < 0
}

// IsFlat returns true if there is no position.
func (p *Position) IsFlat() bool {
	return p.Size() == 0
}

// Close closes the entire position or a portion of it.
// portion should be between 0 and 1, where 1 closes the entire position.
func (p *Position) Close(portion float64) {
	if portion <= 0 || portion > 1 {
		return
	}

	// Close portion of each trade (FIFO order)
	for _, trade := range p.broker.trades {
		trade.Close(portion)
	}
}

// NumTrades returns the number of active trades.
func (p *Position) NumTrades() int {
	return len(p.broker.trades)
}

// Trades returns a copy of the active trades slice.
func (p *Position) Trades() []*Trade {
	trades := make([]*Trade, len(p.broker.trades))
	copy(trades, p.broker.trades)
	return trades
}

// String returns a string representation of the position.
func (p *Position) String() string {
	size := p.Size()
	if size == 0 {
		return "Position: FLAT"
	}

	side := "LONG"
	if size < 0 {
		side = "SHORT"
	}

	return fmt.Sprintf("Position: %s %.4f PL=%.2f(%.2f%%) trades=%d",
		side, absFloat(size), p.PL(), p.PLPct(), p.NumTrades())
}
