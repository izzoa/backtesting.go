package backtesting

import (
	"fmt"
	"sync/atomic"
	"time"
)

// tradeIDCounter generates unique trade IDs.
var tradeIDCounter int64

func nextTradeID() int64 {
	return atomic.AddInt64(&tradeIDCounter, 1)
}

// Trade represents an active or closed trade.
type Trade struct {
	broker *Broker

	// ID is a unique identifier for this trade.
	ID int64

	// Size is the trade size. Positive = long, negative = short.
	Size float64

	// EntryPrice is the price at which the trade was opened.
	EntryPrice float64

	// EntryBar is the bar index at which the trade was opened.
	EntryBar int

	// EntryTime is the time at which the trade was opened.
	EntryTime time.Time

	// ExitPrice is the price at which the trade was closed (nil if still open).
	ExitPrice *float64

	// ExitBar is the bar index at which the trade was closed (nil if still open).
	ExitBar *int

	// ExitTime is the time at which the trade was closed (nil if still open).
	ExitTime *time.Time

	// sl is the current stop-loss price.
	sl *float64

	// tp is the current take-profit price.
	tp *float64

	// slOrder is the stop-loss order attached to this trade.
	slOrder *Order

	// tpOrder is the take-profit order attached to this trade.
	tpOrder *Order

	// Tag is an optional user-defined tag for this trade.
	Tag interface{}
}

// NewTrade creates a new trade.
func NewTrade(broker *Broker, size float64, entryPrice float64, entryBar int, entryTime time.Time) *Trade {
	return &Trade{
		broker:     broker,
		ID:         nextTradeID(),
		Size:       size,
		EntryPrice: entryPrice,
		EntryBar:   entryBar,
		EntryTime:  entryTime,
	}
}

// IsLong returns true if this is a long trade.
func (t *Trade) IsLong() bool {
	return t.Size > 0
}

// IsShort returns true if this is a short trade.
func (t *Trade) IsShort() bool {
	return t.Size < 0
}

// IsOpen returns true if this trade is still open.
func (t *Trade) IsOpen() bool {
	return t.ExitPrice == nil
}

// IsClosed returns true if this trade has been closed.
func (t *Trade) IsClosed() bool {
	return t.ExitPrice != nil
}

// PL returns the profit/loss in cash for this trade.
// For open trades, uses the broker's last price.
func (t *Trade) PL() float64 {
	exitPrice := t.broker.lastPrice
	if t.ExitPrice != nil {
		exitPrice = *t.ExitPrice
	}
	return t.Size * (exitPrice - t.EntryPrice)
}

// PLPct returns the profit/loss as a percentage of entry value.
func (t *Trade) PLPct() float64 {
	if t.EntryPrice == 0 {
		return 0
	}
	exitPrice := t.broker.lastPrice
	if t.ExitPrice != nil {
		exitPrice = *t.ExitPrice
	}
	if t.IsLong() {
		return (exitPrice - t.EntryPrice) / t.EntryPrice * 100
	}
	// For short trades, profit when price goes down
	return (t.EntryPrice - exitPrice) / t.EntryPrice * 100
}

// Value returns the absolute value of the trade (|size| × entry price).
func (t *Trade) Value() float64 {
	return absFloat(t.Size) * t.EntryPrice
}

// SL returns the current stop-loss price.
func (t *Trade) SL() *float64 {
	return t.sl
}

// SetSL sets a stop-loss price for this trade.
// Creates or updates the stop-loss order.
func (t *Trade) SetSL(price float64) {
	t.sl = &price

	// Cancel existing SL order if any
	if t.slOrder != nil {
		t.slOrder.Cancel()
	}

	// Create new SL order (opposite direction to close the trade)
	slSize := -t.Size
	t.slOrder = NewOrder(t.broker, slSize,
		WithStop(price),
		WithParentTrade(t),
	)
	t.broker.orders = append(t.broker.orders, t.slOrder)
}

// TP returns the current take-profit price.
func (t *Trade) TP() *float64 {
	return t.tp
}

// SetTP sets a take-profit price for this trade.
// Creates or updates the take-profit order.
func (t *Trade) SetTP(price float64) {
	t.tp = &price

	// Cancel existing TP order if any
	if t.tpOrder != nil {
		t.tpOrder.Cancel()
	}

	// Create new TP order (opposite direction to close the trade)
	tpSize := -t.Size
	t.tpOrder = NewOrder(t.broker, tpSize,
		WithLimit(price),
		WithParentTrade(t),
	)
	t.broker.orders = append(t.broker.orders, t.tpOrder)
}

// Close closes this trade or a portion of it.
// portion should be between 0 and 1, where 1 closes the entire trade.
func (t *Trade) Close(portion float64) {
	if portion <= 0 || portion > 1 {
		return
	}
	if t.IsClosed() {
		return
	}

	closeSize := -t.Size * portion
	t.broker.NewOrder(closeSize, WithParentTrade(t))
}

// Duration returns the duration of the trade.
// For open trades, returns duration to current bar time.
func (t *Trade) Duration() time.Duration {
	endTime := t.broker.currentTime()
	if t.ExitTime != nil {
		endTime = *t.ExitTime
	}
	return endTime.Sub(t.EntryTime)
}

// Bars returns the number of bars the trade has been open.
func (t *Trade) Bars() int {
	endBar := t.broker.currentBar
	if t.ExitBar != nil {
		endBar = *t.ExitBar
	}
	return endBar - t.EntryBar
}

// String returns a string representation of the trade.
func (t *Trade) String() string {
	side := "LONG"
	if t.IsShort() {
		side = "SHORT"
	}

	status := "OPEN"
	if t.IsClosed() {
		status = fmt.Sprintf("CLOSED@%.4f", *t.ExitPrice)
	}

	return fmt.Sprintf("Trade#%d %s %.4f entry=%.4f %s PL=%.2f(%.2f%%)",
		t.ID, side, absFloat(t.Size), t.EntryPrice, status, t.PL(), t.PLPct())
}
