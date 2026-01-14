package backtesting

import (
	"time"

	"github.com/quickfixgo/backtesting/data"
)

// CommissionFunc calculates commission for a trade.
// size is the trade size (positive or negative), price is the execution price.
type CommissionFunc func(size float64, price float64) float64

// Broker simulates order execution and manages trades.
type Broker struct {
	// Configuration
	data            *data.OHLCV
	cash            float64
	initialCash     float64
	commission      CommissionFunc
	margin          float64 // 1.0 = no leverage, 0.1 = 10x leverage
	spread          float64 // Spread as a fraction (e.g., 0.0001 for 1 pip)
	tradeOnClose    bool    // If true, orders fill at close price; otherwise at next open
	hedging         bool    // If true, allow simultaneous long and short positions
	exclusiveOrders bool    // If true, new orders close previous trades first

	// State
	orders       []*Order
	trades       []*Trade
	closedTrades []*Trade
	position     *Position

	// Equity tracking
	equity     []float64
	maxEquity  float64
	currentBar int
	lastPrice  float64
}

// BrokerOption configures the broker.
type BrokerOption func(*Broker)

// NewBroker creates a new broker with the given options.
func NewBroker(opts ...BrokerOption) *Broker {
	b := &Broker{
		cash:        10000,
		margin:      1.0, // No leverage by default
		spread:      0,
		tradeOnClose: false,
		hedging:      false,
		exclusiveOrders: false,
		orders:       make([]*Order, 0),
		trades:       make([]*Trade, 0),
		closedTrades: make([]*Trade, 0),
		equity:       make([]float64, 0),
	}

	for _, opt := range opts {
		opt(b)
	}

	b.initialCash = b.cash
	b.maxEquity = b.cash
	b.position = NewPosition(b)

	return b
}

// WithData sets the OHLCV data for the broker.
func WithData(d *data.OHLCV) BrokerOption {
	return func(b *Broker) {
		b.data = d
	}
}

// WithCash sets the initial cash amount.
func WithCash(cash float64) BrokerOption {
	return func(b *Broker) {
		b.cash = cash
	}
}

// WithCommission sets the commission function.
func WithCommission(fn CommissionFunc) BrokerOption {
	return func(b *Broker) {
		b.commission = fn
	}
}

// WithFixedCommission sets a fixed commission per trade.
func WithFixedCommission(amount float64) BrokerOption {
	return func(b *Broker) {
		b.commission = func(size, price float64) float64 {
			return amount
		}
	}
}

// WithPercentCommission sets a percentage-based commission.
func WithPercentCommission(pct float64) BrokerOption {
	return func(b *Broker) {
		b.commission = func(size, price float64) float64 {
			return absFloat(size) * price * pct
		}
	}
}

// WithMargin sets the margin requirement.
// 1.0 = 100% margin (no leverage), 0.1 = 10% margin (10x leverage).
func WithMargin(margin float64) BrokerOption {
	return func(b *Broker) {
		b.margin = margin
	}
}

// WithSpread sets the bid-ask spread as a fraction of price.
func WithSpread(spread float64) BrokerOption {
	return func(b *Broker) {
		b.spread = spread
	}
}

// WithTradeOnClose sets whether to fill orders at close price.
func WithTradeOnClose(tradeOnClose bool) BrokerOption {
	return func(b *Broker) {
		b.tradeOnClose = tradeOnClose
	}
}

// WithHedging enables simultaneous long and short positions.
func WithHedging(hedging bool) BrokerOption {
	return func(b *Broker) {
		b.hedging = hedging
	}
}

// WithExclusiveOrders enables closing previous trades when new orders are placed.
func WithExclusiveOrders(exclusive bool) BrokerOption {
	return func(b *Broker) {
		b.exclusiveOrders = exclusive
	}
}

// Equity returns the current account equity.
// Equity = Cash + Market value of all open positions
func (b *Broker) Equity() float64 {
	equity := b.cash
	for _, trade := range b.trades {
		// Add back the margin we reserved, plus unrealized P&L
		equity += absFloat(trade.Size) * trade.EntryPrice * b.margin
		equity += trade.PL()
	}
	return equity
}

// MarginAvailable returns the available margin for new trades.
func (b *Broker) MarginAvailable() float64 {
	usedMargin := 0.0
	for _, trade := range b.trades {
		usedMargin += trade.Value() * b.margin
	}
	return b.Equity() - usedMargin
}

// LastPrice returns the last known price.
func (b *Broker) LastPrice() float64 {
	return b.lastPrice
}

// Position returns the current position.
func (b *Broker) Position() *Position {
	return b.position
}

// Orders returns a copy of pending orders.
func (b *Broker) Orders() []*Order {
	orders := make([]*Order, len(b.orders))
	copy(orders, b.orders)
	return orders
}

// Trades returns a copy of active trades.
func (b *Broker) Trades() []*Trade {
	trades := make([]*Trade, len(b.trades))
	copy(trades, b.trades)
	return trades
}

// ClosedTrades returns a copy of closed trades.
func (b *Broker) ClosedTrades() []*Trade {
	trades := make([]*Trade, len(b.closedTrades))
	copy(trades, b.closedTrades)
	return trades
}

// NewOrder creates and queues a new order.
func (b *Broker) NewOrder(size float64, opts ...OrderOption) *Order {
	order := NewOrder(b, size, opts...)
	b.orders = append(b.orders, order)
	return order
}

// Next processes the next bar: updates prices and processes pending orders.
func (b *Broker) Next() {
	if b.data == nil || b.currentBar >= b.data.Len() {
		return
	}

	bar := b.data.At(b.currentBar)
	b.lastPrice = bar.Close

	// Process orders
	b.processOrders(bar)

	// Record equity
	b.equity = append(b.equity, b.Equity())
	if b.Equity() > b.maxEquity {
		b.maxEquity = b.Equity()
	}
}

// processOrders processes all pending orders against the current bar.
func (b *Broker) processOrders(bar data.Bar) {
	// Process orders in a stable order
	remainingOrders := make([]*Order, 0, len(b.orders))

	for _, order := range b.orders {
		if order.IsCancelled() {
			continue
		}

		filled := b.tryFillOrder(order, bar)
		if !filled {
			remainingOrders = append(remainingOrders, order)
		}
	}

	b.orders = remainingOrders
}

// tryFillOrder attempts to fill an order. Returns true if filled.
func (b *Broker) tryFillOrder(order *Order, bar data.Bar) bool {
	// Determine execution price
	execPrice, canFill := b.checkOrderConditions(order, bar)
	if !canFill {
		return false
	}

	// Apply spread
	if order.IsLong() {
		execPrice *= (1 + b.spread)
	} else {
		execPrice *= (1 - b.spread)
	}

	// Handle contingent orders (SL/TP)
	if order.IsContingent() {
		return b.fillContingentOrder(order, execPrice, bar)
	}

	// Handle regular orders
	return b.fillOrder(order, execPrice, bar)
}

// checkOrderConditions checks if an order can be filled and returns the execution price.
func (b *Broker) checkOrderConditions(order *Order, bar data.Bar) (float64, bool) {
	// Market orders fill immediately
	if order.IsMarket() {
		if b.tradeOnClose {
			return bar.Close, true
		}
		return bar.Open, true
	}

	// Stop orders: check if stop is triggered
	if order.IsStop() && !order.stopTriggered {
		stopTriggered := false
		if order.IsLong() {
			// Long stop: triggers when price rises to/above stop
			stopTriggered = bar.High >= *order.Stop
		} else {
			// Short stop: triggers when price falls to/below stop
			stopTriggered = bar.Low <= *order.Stop
		}

		if !stopTriggered {
			return 0, false
		}
		order.stopTriggered = true

		// If stop-limit, now wait for limit
		if order.IsStopLimit() {
			// Check if limit is also hit in this bar
			return b.checkLimitCondition(order, bar)
		}

		// Stop-market: fill at stop price
		return *order.Stop, true
	}

	// Stop-limit after stop triggered
	if order.IsStopLimit() && order.stopTriggered {
		return b.checkLimitCondition(order, bar)
	}

	// Limit orders
	if order.IsLimit() {
		return b.checkLimitCondition(order, bar)
	}

	return 0, false
}

// checkLimitCondition checks if a limit order can be filled.
func (b *Broker) checkLimitCondition(order *Order, bar data.Bar) (float64, bool) {
	if order.Limit == nil {
		return 0, false
	}

	if order.IsLong() {
		// Long limit: fill when price falls to/below limit
		if bar.Low <= *order.Limit {
			// Fill at limit price or open if gap down
			if bar.Open <= *order.Limit {
				return bar.Open, true
			}
			return *order.Limit, true
		}
	} else {
		// Short limit: fill when price rises to/above limit
		if bar.High >= *order.Limit {
			// Fill at limit price or open if gap up
			if bar.Open >= *order.Limit {
				return bar.Open, true
			}
			return *order.Limit, true
		}
	}

	return 0, false
}

// fillContingentOrder handles SL/TP orders that reduce a parent trade.
func (b *Broker) fillContingentOrder(order *Order, execPrice float64, bar data.Bar) bool {
	parent := order.ParentTrade
	if parent == nil || parent.IsClosed() {
		return true // Order is no longer valid
	}

	// Calculate the portion to close
	closeSize := absFloat(order.Size)
	tradeSize := absFloat(parent.Size)

	if closeSize >= tradeSize {
		// Close entire trade
		b.closeTrade(parent, execPrice, bar)
	} else {
		// Partial close
		portion := closeSize / tradeSize
		b.reduceTrade(parent, portion, execPrice, bar)
	}

	return true
}

// fillOrder fills a regular (non-contingent) order.
func (b *Broker) fillOrder(order *Order, execPrice float64, bar data.Bar) bool {
	// Check if we should close existing trades first
	if b.exclusiveOrders && len(b.trades) > 0 {
		// Close all existing trades
		for _, trade := range b.trades {
			b.closeTrade(trade, execPrice, bar)
		}
		b.trades = nil
	}

	// If not hedging, reduce opposite-facing trades first
	if !b.hedging {
		remainingSize := order.Size
		remainingSize = b.reduceOppositeTrades(remainingSize, execPrice, bar)

		if remainingSize == 0 {
			return true // Order fully absorbed by closing opposite trades
		}
		order.Size = remainingSize
	}

	// Check margin
	requiredMargin := absFloat(order.Size) * execPrice * b.margin
	if requiredMargin > b.MarginAvailable() {
		return true // Cancel order due to insufficient margin
	}

	// Open new trade
	b.openTrade(order, execPrice, bar)

	return true
}

// reduceOppositeTrades reduces opposite-facing trades (FIFO).
// Returns the remaining order size after reductions.
func (b *Broker) reduceOppositeTrades(orderSize float64, execPrice float64, bar data.Bar) float64 {
	if len(b.trades) == 0 {
		return orderSize
	}

	// Find opposite-facing trades
	var oppositeTrades []*Trade
	for _, trade := range b.trades {
		if (orderSize > 0 && trade.IsShort()) || (orderSize < 0 && trade.IsLong()) {
			oppositeTrades = append(oppositeTrades, trade)
		}
	}

	remainingSize := orderSize
	for _, trade := range oppositeTrades {
		if remainingSize == 0 {
			break
		}

		tradeSize := absFloat(trade.Size)
		orderAbsSize := absFloat(remainingSize)

		if orderAbsSize >= tradeSize {
			// Close entire trade
			b.closeTrade(trade, execPrice, bar)
			if remainingSize > 0 {
				remainingSize -= tradeSize
			} else {
				remainingSize += tradeSize
			}
		} else {
			// Partial close
			portion := orderAbsSize / tradeSize
			b.reduceTrade(trade, portion, execPrice, bar)
			remainingSize = 0
		}
	}

	return remainingSize
}

// openTrade opens a new trade from an order.
func (b *Broker) openTrade(order *Order, execPrice float64, bar data.Bar) *Trade {
	trade := NewTrade(b, order.Size, execPrice, b.currentBar, bar.Time)
	trade.Tag = order.Tag

	// Apply commission
	if b.commission != nil {
		commission := b.commission(order.Size, execPrice)
		b.cash -= commission
	}

	// Deduct cost from cash (for margin calculation)
	tradeCost := absFloat(order.Size) * execPrice * b.margin
	b.cash -= tradeCost

	b.trades = append(b.trades, trade)

	// Set SL/TP if specified
	if order.SL != nil {
		trade.SetSL(*order.SL)
	}
	if order.TP != nil {
		trade.SetTP(*order.TP)
	}

	return trade
}

// closeTrade closes a trade completely.
func (b *Broker) closeTrade(trade *Trade, execPrice float64, bar data.Bar) {
	if trade.IsClosed() {
		return
	}

	trade.ExitPrice = &execPrice
	exitBar := b.currentBar
	trade.ExitBar = &exitBar
	exitTime := bar.Time
	trade.ExitTime = &exitTime

	// Calculate P&L
	pl := trade.Size * (execPrice - trade.EntryPrice)

	// Apply commission
	if b.commission != nil {
		commission := b.commission(trade.Size, execPrice)
		b.cash -= commission
	}

	// Return margin and add P&L
	tradeCost := absFloat(trade.Size) * trade.EntryPrice * b.margin
	b.cash += tradeCost + pl

	// Cancel any attached SL/TP orders
	if trade.slOrder != nil {
		trade.slOrder.Cancel()
	}
	if trade.tpOrder != nil {
		trade.tpOrder.Cancel()
	}

	// Move to closed trades
	b.closedTrades = append(b.closedTrades, trade)
	b.removeTrade(trade)
}

// reduceTrade partially closes a trade.
func (b *Broker) reduceTrade(trade *Trade, portion float64, execPrice float64, bar data.Bar) {
	if portion <= 0 || portion >= 1 {
		if portion >= 1 {
			b.closeTrade(trade, execPrice, bar)
		}
		return
	}

	// Calculate the size to close
	closeSize := trade.Size * portion
	remainingSize := trade.Size - closeSize

	// Create a new closed trade record for the closed portion
	closedPortion := NewTrade(b, closeSize, trade.EntryPrice, trade.EntryBar, trade.EntryTime)
	closedPortion.ExitPrice = &execPrice
	exitBar := b.currentBar
	closedPortion.ExitBar = &exitBar
	exitTime := bar.Time
	closedPortion.ExitTime = &exitTime
	closedPortion.Tag = trade.Tag

	// Calculate P&L for closed portion
	pl := closeSize * (execPrice - trade.EntryPrice)

	// Apply commission for partial close
	if b.commission != nil {
		commission := b.commission(closeSize, execPrice)
		b.cash -= commission
	}

	// Return margin and add P&L for closed portion
	closedCost := absFloat(closeSize) * trade.EntryPrice * b.margin
	b.cash += closedCost + pl

	// Update original trade with remaining size
	trade.Size = remainingSize

	b.closedTrades = append(b.closedTrades, closedPortion)
}

// removeTrade removes a trade from the active trades list.
func (b *Broker) removeTrade(trade *Trade) {
	for i, t := range b.trades {
		if t == trade {
			b.trades = append(b.trades[:i], b.trades[i+1:]...)
			return
		}
	}
}

// currentTime returns the current bar's time.
func (b *Broker) currentTime() time.Time {
	if b.data == nil || b.currentBar >= b.data.Len() {
		return time.Time{}
	}
	return b.data.Time[b.currentBar]
}

// SetBar sets the current bar index.
func (b *Broker) SetBar(bar int) {
	b.currentBar = bar
}

// EquityCurve returns the recorded equity curve.
func (b *Broker) EquityCurve() []float64 {
	return b.equity
}
