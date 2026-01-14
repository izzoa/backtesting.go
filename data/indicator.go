package data

// Indicator represents a calculated indicator series with plotting metadata.
type Indicator struct {
	*Series
	PlotOptions IndicatorPlotOptions
}

// IndicatorPlotOptions configures how an indicator is plotted.
type IndicatorPlotOptions struct {
	// Overlay indicates whether the indicator should be overlaid on the price chart
	// or plotted in a separate subplot.
	Overlay bool

	// Color is the line color for the indicator (e.g., "#FF0000", "red").
	Color string

	// ScatterKey specifies a column in the data to use for scatter markers.
	// Empty string means no scatter markers.
	ScatterKey string

	// Name is the display name for the indicator in the legend.
	// If empty, uses the Series.Name.
	Name string

	// LineWidth is the width of the indicator line.
	LineWidth float64

	// Opacity is the line opacity (0.0 to 1.0).
	Opacity float64
}

// IndicatorOption configures indicator creation.
type IndicatorOption func(*indicatorConfig)

type indicatorConfig struct {
	overlay    *bool
	color      string
	scatterKey string
	name       string
	lineWidth  float64
	opacity    float64
}

// WithOverlay sets whether the indicator overlays on price chart.
func WithOverlay(overlay bool) IndicatorOption {
	return func(c *indicatorConfig) {
		c.overlay = &overlay
	}
}

// WithColor sets the indicator line color.
func WithColor(color string) IndicatorOption {
	return func(c *indicatorConfig) {
		c.color = color
	}
}

// WithScatter sets the scatter marker key.
func WithScatter(key string) IndicatorOption {
	return func(c *indicatorConfig) {
		c.scatterKey = key
	}
}

// WithIndicatorName sets the display name.
func WithIndicatorName(name string) IndicatorOption {
	return func(c *indicatorConfig) {
		c.name = name
	}
}

// WithLineWidth sets the line width.
func WithLineWidth(width float64) IndicatorOption {
	return func(c *indicatorConfig) {
		c.lineWidth = width
	}
}

// WithOpacity sets the line opacity.
func WithOpacity(opacity float64) IndicatorOption {
	return func(c *indicatorConfig) {
		c.opacity = opacity
	}
}

// NewIndicator creates a new indicator from values with options.
// The overlay option is auto-detected if not specified: indicators within
// 30% of the price range are considered overlays.
func NewIndicator(name string, values []float64, opts ...IndicatorOption) *Indicator {
	cfg := &indicatorConfig{
		lineWidth: 1.0,
		opacity:   1.0,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	displayName := cfg.name
	if displayName == "" {
		displayName = name
	}

	overlay := false
	if cfg.overlay != nil {
		overlay = *cfg.overlay
	}

	return &Indicator{
		Series: NewSeries(name, values),
		PlotOptions: IndicatorPlotOptions{
			Overlay:    overlay,
			Color:      cfg.color,
			ScatterKey: cfg.scatterKey,
			Name:       displayName,
			LineWidth:  cfg.lineWidth,
			Opacity:    cfg.opacity,
		},
	}
}

// AutoDetectOverlay determines if an indicator should overlay on the price chart
// based on whether its values are within 30% of the close price range.
func AutoDetectOverlay(indicatorValues []float64, closeValues []float64) bool {
	if len(indicatorValues) == 0 || len(closeValues) == 0 {
		return false
	}

	// Get close price range
	closeMin, closeMax := closeValues[0], closeValues[0]
	for _, v := range closeValues {
		if v < closeMin {
			closeMin = v
		}
		if v > closeMax {
			closeMax = v
		}
	}

	if closeMax == closeMin {
		return false
	}

	// Check if indicator values are within 30% of close range
	closeRange := closeMax - closeMin
	threshold := closeRange * 0.3

	expandedMin := closeMin - threshold
	expandedMax := closeMax + threshold

	// Count how many indicator values fall within the expanded range
	inRange := 0
	total := 0
	for _, v := range indicatorValues {
		if IsNaN(v) {
			continue
		}
		total++
		if v >= expandedMin && v <= expandedMax {
			inRange++
		}
	}

	if total == 0 {
		return false
	}

	// If more than 80% of values are in range, it's an overlay
	return float64(inRange)/float64(total) >= 0.8
}

// Copy returns a deep copy of the indicator.
func (ind *Indicator) Copy() *Indicator {
	return &Indicator{
		Series:      ind.Series.Copy(),
		PlotOptions: ind.PlotOptions,
	}
}

// Slice returns a new indicator with a sliced series.
func (ind *Indicator) Slice(start, end int) *Indicator {
	return &Indicator{
		Series:      ind.Series.Slice(start, end),
		PlotOptions: ind.PlotOptions,
	}
}
