package plot

import (
	"os"
	"testing"
	"time"

	"github.com/izzoa/backtesting.go/data"
)

func TestPrepareChartData(t *testing.T) {
	// Create test OHLCV data
	ohlcv := data.NewOHLCV(10)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 10; i++ {
		bar := data.Bar{
			Time:   start.Add(time.Duration(i) * 24 * time.Hour),
			Open:   100 + float64(i),
			High:   105 + float64(i),
			Low:    95 + float64(i),
			Close:  102 + float64(i),
			Volume: 1000,
		}
		ohlcv.Append(bar)
	}

	// Create equity curve
	equity := make([]float64, 10)
	for i := range equity {
		equity[i] = 10000 + float64(i)*100
	}

	// Create trades
	trades := []TradeInfo{
		{
			Size:       10,
			EntryTime:  start.Add(2 * 24 * time.Hour),
			ExitTime:   start.Add(5 * 24 * time.Hour),
			EntryPrice: 102,
			ExitPrice:  107,
			PL:         50,
			PLPct:      5,
		},
	}

	chartData := PrepareChartData(ohlcv, equity, trades, nil)

	if chartData == nil {
		t.Fatal("PrepareChartData returned nil")
	}

	if len(chartData.Times) != 10 {
		t.Errorf("Expected 10 times, got %d", len(chartData.Times))
	}

	if len(chartData.Equity) != 10 {
		t.Errorf("Expected 10 equity values, got %d", len(chartData.Equity))
	}

	// Should have 2 trade markers (entry and exit)
	if len(chartData.Trades) != 2 {
		t.Errorf("Expected 2 trade markers, got %d", len(chartData.Trades))
	}

	// Check trade marker types
	hasEntryLong := false
	hasExitLong := false
	for _, tm := range chartData.Trades {
		if tm.Type == "entry_long" {
			hasEntryLong = true
		}
		if tm.Type == "exit_long" {
			hasExitLong = true
		}
	}

	if !hasEntryLong {
		t.Error("Missing entry_long marker")
	}
	if !hasExitLong {
		t.Error("Missing exit_long marker")
	}

	// Check drawdown calculation
	if len(chartData.DrawdownPct) != 10 {
		t.Errorf("Expected 10 drawdown values, got %d", len(chartData.DrawdownPct))
	}

	// With monotonically increasing equity, drawdown should be 0
	for i, dd := range chartData.DrawdownPct {
		if dd != 0 {
			t.Errorf("Expected 0 drawdown at index %d, got %f", i, dd)
		}
	}
}

func TestCalcDrawdownPct(t *testing.T) {
	tests := []struct {
		name     string
		equity   []float64
		expected []float64
	}{
		{
			name:     "no drawdown",
			equity:   []float64{100, 110, 120},
			expected: []float64{0, 0, 0},
		},
		{
			name:     "simple drawdown",
			equity:   []float64{100, 110, 100},
			expected: []float64{0, 0, 9.09}, // (110-100)/110 * 100
		},
		{
			name:     "recovery",
			equity:   []float64{100, 110, 100, 110},
			expected: []float64{0, 0, 9.09, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcDrawdownPct(tt.equity)
			if len(result) != len(tt.expected) {
				t.Fatalf("Length mismatch: got %d, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				diff := result[i] - tt.expected[i]
				if diff > 0.1 || diff < -0.1 {
					t.Errorf("At index %d: got %.2f, want %.2f", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestPlot(t *testing.T) {
	// Create test data
	times := make([]time.Time, 100)
	close := make([]float64, 100)
	equity := make([]float64, 100)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 100; i++ {
		times[i] = start.Add(time.Duration(i) * 24 * time.Hour)
		close[i] = 100 + float64(i)*0.5
		equity[i] = 10000 + float64(i)*50
	}

	data := &ChartData{
		Times:  times,
		Close:  close,
		Open:   close,
		High:   close,
		Low:    close,
		Equity: equity,
		Trades: []TradeMarker{
			{Time: times[10], Price: close[10], Type: "entry_long"},
			{Time: times[50], Price: close[50], Type: "exit_long", PL: 500},
		},
	}

	// Create temp file
	tmpFile := "test_backtest.html"
	defer os.Remove(tmpFile)

	cfg := Config{
		Filename:     tmpFile,
		PlotEquity:   true,
		PlotDrawdown: true,
		PlotTrades:   true,
		ShowLegend:   true,
		OpenBrowser:  false, // Don't open browser in tests
		Title:        "Test Chart",
	}

	err := Plot(data, cfg)
	if err != nil {
		t.Fatalf("Plot failed: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("HTML file was not created")
	}

	// Read file and check basic content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check for expected content
	checks := []string{
		"<title>Test Chart</title>",
		"chart.js",
		"priceChart",
		"equityChart",
	}

	for _, check := range checks {
		if !contains(string(content), check) {
			t.Errorf("HTML missing expected content: %s", check)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Filename != "backtest.html" {
		t.Errorf("Filename = %s, want backtest.html", cfg.Filename)
	}

	if !cfg.PlotEquity {
		t.Error("PlotEquity should be true by default")
	}

	if !cfg.PlotDrawdown {
		t.Error("PlotDrawdown should be true by default")
	}

	if !cfg.PlotTrades {
		t.Error("PlotTrades should be true by default")
	}

	if !cfg.OpenBrowser {
		t.Error("OpenBrowser should be true by default")
	}
}
