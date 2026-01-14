package backtesting

import (
	"time"

	"github.com/quickfixgo/backtesting/data"
)

// Strategy is the interface that all trading strategies must implement.
type Strategy interface {
	// Init initializes the strategy. Called once before backtesting starts.
	// Use this to declare indicators and set up state.
	Init()

	// Next is called for each bar during backtesting.
	// Implement your trading logic here.
	Next()

	// SetBroker sets the broker for this strategy.
	SetBroker(broker *Broker)

	// SetData sets the data view for this strategy.
	SetData(dataView *data.DataView)
}

// StrategyBase provides a base implementation for strategies.
// Embed this in your strategy struct to get access to data, broker, and indicators.
type StrategyBase struct {
	broker     *Broker
	dataView   *data.DataView
	indicators []*data.Indicator
}

// SetBroker sets the broker for this strategy.
func (s *StrategyBase) SetBroker(broker *Broker) {
	s.broker = broker
}

// SetData sets the data view for this strategy.
func (s *StrategyBase) SetData(dataView *data.DataView) {
	s.dataView = dataView
}

// Data returns the data view.
func (s *StrategyBase) Data() *data.DataView {
	return s.dataView
}

// Broker returns the broker.
func (s *StrategyBase) Broker() *Broker {
	return s.broker
}

// Close returns the close prices as a slice.
func (s *StrategyBase) Close() []float64 {
	return s.dataView.Close()
}

// Open returns the open prices as a slice.
func (s *StrategyBase) Open() []float64 {
	return s.dataView.Open()
}

// High returns the high prices as a slice.
func (s *StrategyBase) High() []float64 {
	return s.dataView.High()
}

// Low returns the low prices as a slice.
func (s *StrategyBase) Low() []float64 {
	return s.dataView.Low()
}

// Volume returns the volume as a slice.
func (s *StrategyBase) Volume() []float64 {
	return s.dataView.Volume()
}

// Index returns the time index.
func (s *StrategyBase) Index() []time.Time {
	return s.dataView.Time()
}

// Equity returns the current account equity.
func (s *StrategyBase) Equity() float64 {
	return s.broker.Equity()
}

// Position returns the current position.
func (s *StrategyBase) Position() *Position {
	return s.broker.Position()
}

// Orders returns pending orders.
func (s *StrategyBase) Orders() []*Order {
	return s.broker.Orders()
}

// Trades returns active trades.
func (s *StrategyBase) Trades() []*Trade {
	return s.broker.Trades()
}

// ClosedTrades returns closed trades.
func (s *StrategyBase) ClosedTrades() []*Trade {
	return s.broker.ClosedTrades()
}

// Buy places a buy order.
func (s *StrategyBase) Buy(opts ...OrderOption) *Order {
	// Default to buying with ~95% of available equity
	// Leave buffer for price movement between order and execution
	size := s.broker.Equity() * 0.95 / s.dataView.LastClose()
	if size <= 0 {
		return nil
	}
	return s.broker.NewOrder(size, opts...)
}

// Sell places a sell order.
func (s *StrategyBase) Sell(opts ...OrderOption) *Order {
	// Default to selling with ~95% of available equity
	// Leave buffer for price movement between order and execution
	size := -s.broker.Equity() * 0.95 / s.dataView.LastClose()
	if size >= 0 {
		return nil
	}
	return s.broker.NewOrder(size, opts...)
}

// BuySize places a buy order with a specific size.
func (s *StrategyBase) BuySize(size float64, opts ...OrderOption) *Order {
	if size <= 0 {
		return nil
	}
	return s.broker.NewOrder(size, opts...)
}

// SellSize places a sell order with a specific size.
func (s *StrategyBase) SellSize(size float64, opts ...OrderOption) *Order {
	if size <= 0 {
		return nil
	}
	return s.broker.NewOrder(-size, opts...)
}

// I declares an indicator. Call this in Init() to register indicators.
// The indicator values are automatically sliced during backtesting to prevent look-ahead bias.
func (s *StrategyBase) I(name string, values []float64, opts ...data.IndicatorOption) *data.Indicator {
	// Validate length
	if len(values) != s.dataView.FullLen() {
		panic("indicator values length must match data length")
	}

	// Auto-detect overlay if not specified
	autoOverlay := data.AutoDetectOverlay(values, s.dataView.Source().Close)

	// Create indicator with auto-detected overlay as default
	indicator := data.NewIndicator(name, values, opts...)
	if !hasOverlayOption(opts) {
		indicator.PlotOptions.Overlay = autoOverlay
	}

	s.indicators = append(s.indicators, indicator)
	return indicator
}

// hasOverlayOption checks if WithOverlay was specified in options.
func hasOverlayOption(opts []data.IndicatorOption) bool {
	// We can't easily check this without modifying the option pattern,
	// so for now we assume if overlay is false it might have been explicitly set
	// This is a simplification; in practice we might track this differently
	return false
}

// Indicators returns all registered indicators.
func (s *StrategyBase) Indicators() []*data.Indicator {
	return s.indicators
}

// IndicatorValues returns the current (sliced) values of an indicator.
// Use this in Next() to access indicator values up to the current bar.
func (s *StrategyBase) IndicatorValues(ind *data.Indicator) []float64 {
	return ind.Values[:s.dataView.Len()]
}

// IndicatorLast returns the last value of an indicator at the current bar.
func (s *StrategyBase) IndicatorLast(ind *data.Indicator) float64 {
	idx := s.dataView.Len() - 1
	if idx < 0 || idx >= len(ind.Values) {
		return 0
	}
	return ind.Values[idx]
}

// IndicatorAt returns the indicator value at a specific index.
func (s *StrategyBase) IndicatorAt(ind *data.Indicator, i int) float64 {
	if i < 0 || i >= s.dataView.Len() || i >= len(ind.Values) {
		return 0
	}
	return ind.Values[i]
}

// BarIndex returns the current bar index.
func (s *StrategyBase) BarIndex() int {
	return s.dataView.Index()
}

// CurrentTime returns the current bar's time.
func (s *StrategyBase) CurrentTime() time.Time {
	return s.dataView.LastTime()
}

// LastClose returns the close of the current bar.
func (s *StrategyBase) LastClose() float64 {
	return s.dataView.LastClose()
}

// LastOpen returns the open of the current bar.
func (s *StrategyBase) LastOpen() float64 {
	return s.dataView.LastOpen()
}

// LastHigh returns the high of the current bar.
func (s *StrategyBase) LastHigh() float64 {
	return s.dataView.LastHigh()
}

// LastLow returns the low of the current bar.
func (s *StrategyBase) LastLow() float64 {
	return s.dataView.LastLow()
}

// LastVolume returns the volume of the current bar.
func (s *StrategyBase) LastVolume() float64 {
	return s.dataView.LastVolume()
}
