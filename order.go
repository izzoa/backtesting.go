// Package backtesting provides a backtesting framework for trading strategies.
package backtesting

import (
	"fmt"
	"sync/atomic"
)

// orderIDCounter generates unique order IDs.
var orderIDCounter int64

func nextOrderID() int64 {
	return atomic.AddInt64(&orderIDCounter, 1)
}

// OrderSide represents the side of an order.
type OrderSide int

const (
	// OrderSideBuy is a buy order.
	OrderSideBuy OrderSide = iota
	// OrderSideSell is a sell order.
	OrderSideSell
)

// Order represents a pending order in the broker.
type Order struct {
	broker *Broker

	// ID is a unique identifier for this order.
	ID int64

	// Size is the order size. Positive values indicate buy, negative indicate sell.
	Size float64

	// Limit is the limit price. Nil for market orders.
	Limit *float64

	// Stop is the stop price for stop and stop-limit orders. Nil for non-stop orders.
	Stop *float64

	// SL is the stop-loss price to set when the order fills.
	SL *float64

	// TP is the take-profit price to set when the order fills.
	TP *float64

	// ParentTrade is the trade this order is associated with (for SL/TP orders).
	ParentTrade *Trade

	// Tag is an optional user-defined tag for this order.
	Tag interface{}

	// createdBar is the bar index when this order was created.
	createdBar int

	// cancelled indicates if this order has been cancelled.
	cancelled bool

	// stopTriggered indicates if the stop condition has been triggered.
	stopTriggered bool
}

// NewOrder creates a new order with the given size and options.
func NewOrder(broker *Broker, size float64, opts ...OrderOption) *Order {
	order := &Order{
		broker:     broker,
		ID:         nextOrderID(),
		Size:       size,
		createdBar: broker.currentBar,
	}

	for _, opt := range opts {
		opt(order)
	}

	return order
}

// OrderOption configures an order.
type OrderOption func(*Order)

// WithLimit sets a limit price for the order.
func WithLimit(price float64) OrderOption {
	return func(o *Order) {
		o.Limit = &price
	}
}

// WithStop sets a stop price for the order.
func WithStop(price float64) OrderOption {
	return func(o *Order) {
		o.Stop = &price
	}
}

// WithSL sets a stop-loss price for when the order fills.
func WithSL(price float64) OrderOption {
	return func(o *Order) {
		o.SL = &price
	}
}

// WithTP sets a take-profit price for when the order fills.
func WithTP(price float64) OrderOption {
	return func(o *Order) {
		o.TP = &price
	}
}

// WithTag sets a user-defined tag for the order.
func WithTag(tag interface{}) OrderOption {
	return func(o *Order) {
		o.Tag = tag
	}
}

// WithParentTrade sets the parent trade (for contingent SL/TP orders).
func WithParentTrade(trade *Trade) OrderOption {
	return func(o *Order) {
		o.ParentTrade = trade
	}
}

// IsLong returns true if this is a long (buy) order.
func (o *Order) IsLong() bool {
	return o.Size > 0
}

// IsShort returns true if this is a short (sell) order.
func (o *Order) IsShort() bool {
	return o.Size < 0
}

// IsContingent returns true if this is a contingent order (SL/TP attached to a trade).
func (o *Order) IsContingent() bool {
	return o.ParentTrade != nil
}

// IsMarket returns true if this is a market order (no limit or stop).
func (o *Order) IsMarket() bool {
	return o.Limit == nil && o.Stop == nil
}

// IsLimit returns true if this is a limit order.
func (o *Order) IsLimit() bool {
	return o.Limit != nil && o.Stop == nil
}

// IsStop returns true if this is a stop order (stop-market or stop-limit).
func (o *Order) IsStop() bool {
	return o.Stop != nil
}

// IsStopLimit returns true if this is a stop-limit order.
func (o *Order) IsStopLimit() bool {
	return o.Stop != nil && o.Limit != nil
}

// Cancel cancels this order.
func (o *Order) Cancel() {
	o.cancelled = true
}

// IsCancelled returns true if this order has been cancelled.
func (o *Order) IsCancelled() bool {
	return o.cancelled
}

// Side returns the order side (Buy or Sell).
func (o *Order) Side() OrderSide {
	if o.Size > 0 {
		return OrderSideBuy
	}
	return OrderSideSell
}

// String returns a string representation of the order.
func (o *Order) String() string {
	side := "BUY"
	if o.IsShort() {
		side = "SELL"
	}

	orderType := "MARKET"
	if o.IsStopLimit() {
		orderType = fmt.Sprintf("STOP-LIMIT[stop=%.4f,limit=%.4f]", *o.Stop, *o.Limit)
	} else if o.IsStop() {
		orderType = fmt.Sprintf("STOP[%.4f]", *o.Stop)
	} else if o.IsLimit() {
		orderType = fmt.Sprintf("LIMIT[%.4f]", *o.Limit)
	}

	var extras string
	if o.SL != nil {
		extras += fmt.Sprintf(" SL=%.4f", *o.SL)
	}
	if o.TP != nil {
		extras += fmt.Sprintf(" TP=%.4f", *o.TP)
	}
	if o.IsContingent() {
		extras += fmt.Sprintf(" parent=%d", o.ParentTrade.ID)
	}

	return fmt.Sprintf("Order#%d %s %.4f %s%s", o.ID, side, absFloat(o.Size), orderType, extras)
}

// absFloat returns the absolute value of a float64.
func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
