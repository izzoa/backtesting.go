package plot

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Config configures chart generation.
type Config struct {
	// Filename is the output HTML file path.
	Filename string

	// Chart options
	PlotEquity     bool
	PlotDrawdown   bool
	PlotVolume     bool
	PlotTrades     bool
	SmoothEquity   int  // Smoothing window for equity curve (0 = no smoothing)
	RelativeEquity bool // Show equity as percentage change

	// Display options
	ShowLegend  bool
	OpenBrowser bool
	Title       string
	Width       int // Chart width in pixels (0 = responsive)
	Height      int // Chart height in pixels (0 = auto)
}

// DefaultConfig returns the default plot configuration.
func DefaultConfig() Config {
	return Config{
		Filename:     "backtest.html",
		PlotEquity:   true,
		PlotDrawdown: true,
		PlotVolume:   true,
		PlotTrades:   true,
		ShowLegend:   true,
		OpenBrowser:  true,
		Title:        "Backtest Results",
	}
}

// Plot generates an interactive HTML chart.
func Plot(data *ChartData, cfg Config) error {
	if data == nil {
		return fmt.Errorf("no data to plot")
	}

	// Convert data to JSON for JavaScript
	chartJSON, err := prepareChartJSON(data, cfg)
	if err != nil {
		return fmt.Errorf("failed to prepare chart data: %w", err)
	}

	// Generate HTML
	html, err := generateHTML(chartJSON, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cfg.Filename, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Open in browser
	if cfg.OpenBrowser {
		openBrowser(cfg.Filename)
	}

	return nil
}

// chartJSONData is the structure passed to JavaScript.
type chartJSONData struct {
	Labels     []string          `json:"labels"`
	OHLC       [][]float64       `json:"ohlc"`
	Volume     []float64         `json:"volume"`
	Equity     []float64         `json:"equity"`
	Drawdown   []float64         `json:"drawdown"`
	Trades     []tradeJSON       `json:"trades"`
	Indicators []indicatorJSON   `json:"indicators"`
	Config     chartConfigJSON   `json:"config"`
	Stats      map[string]string `json:"stats"`
}

type tradeJSON struct {
	Time  string  `json:"time"`
	Price float64 `json:"price"`
	Type  string  `json:"type"`
	PL    float64 `json:"pl"`
}

type indicatorJSON struct {
	Name    string    `json:"name"`
	Values  []float64 `json:"values"`
	Overlay bool      `json:"overlay"`
	Color   string    `json:"color"`
}

type chartConfigJSON struct {
	PlotEquity   bool   `json:"plotEquity"`
	PlotDrawdown bool   `json:"plotDrawdown"`
	PlotVolume   bool   `json:"plotVolume"`
	PlotTrades   bool   `json:"plotTrades"`
	ShowLegend   bool   `json:"showLegend"`
	Title        string `json:"title"`
}

func prepareChartJSON(data *ChartData, cfg Config) (string, error) {
	chartData := chartJSONData{
		Labels:   make([]string, len(data.Times)),
		OHLC:     make([][]float64, len(data.Times)),
		Volume:   data.Volume,
		Equity:   data.Equity,
		Drawdown: data.DrawdownPct,
		Config: chartConfigJSON{
			PlotEquity:   cfg.PlotEquity,
			PlotDrawdown: cfg.PlotDrawdown,
			PlotVolume:   cfg.PlotVolume,
			PlotTrades:   cfg.PlotTrades,
			ShowLegend:   cfg.ShowLegend,
			Title:        cfg.Title,
		},
		Stats: map[string]string{
			"startDate":   data.StartTime.Format("2006-01-02"),
			"endDate":     data.EndTime.Format("2006-01-02"),
			"numTrades":   fmt.Sprintf("%d", data.NumTrades),
			"returnPct":   fmt.Sprintf("%.2f%%", data.ReturnPct),
			"maxDrawdown": fmt.Sprintf("%.2f%%", data.MaxDrawdown),
		},
	}

	// Convert times to labels
	for i, t := range data.Times {
		chartData.Labels[i] = t.Format("2006-01-02")
		chartData.OHLC[i] = []float64{data.Open[i], data.High[i], data.Low[i], data.Close[i]}
	}

	// Convert trades
	for _, t := range data.Trades {
		chartData.Trades = append(chartData.Trades, tradeJSON{
			Time:  t.Time.Format("2006-01-02"),
			Price: t.Price,
			Type:  t.Type,
			PL:    t.PL,
		})
	}

	// Convert indicators
	for _, ind := range data.Indicators {
		color := ind.Color
		if color == "" {
			color = getDefaultColor(len(chartData.Indicators))
		}
		chartData.Indicators = append(chartData.Indicators, indicatorJSON{
			Name:    ind.Name,
			Values:  ind.Values,
			Overlay: ind.Overlay,
			Color:   color,
		})
	}

	jsonBytes, err := json.Marshal(chartData)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func getDefaultColor(index int) string {
	colors := []string{
		"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0",
		"#9966FF", "#FF9F40", "#FF6384", "#C9CBCF",
	}
	return colors[index%len(colors)]
}

func generateHTML(chartJSON string, cfg Config) (string, error) {
	tmpl, err := template.New("chart").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf []byte
	writer := &byteWriter{buf: &buf}

	err = tmpl.Execute(writer, map[string]interface{}{
		"ChartData": template.JS(chartJSON),
		"Title":     cfg.Title,
	})
	if err != nil {
		return "", err
	}

	return string(*writer.buf), nil
}

type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // Linux and others
		cmd = "xdg-open"
		args = []string{url}
	}

	exec.Command(cmd, args...).Start()
}

// QuickPlot is a convenience function for simple plotting.
func QuickPlot(times []time.Time, close, equity []float64, filename string) error {
	data := &ChartData{
		Times:  times,
		Close:  close,
		Equity: equity,
	}

	cfg := DefaultConfig()
	cfg.Filename = filename
	cfg.PlotVolume = false

	return Plot(data, cfg)
}

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-chart-financial"></script>
    <script src="https://cdn.jsdelivr.net/npm/luxon@3/build/global/luxon.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-luxon@1"></script>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #1a1a2e; color: #eee; }
        .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
        .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
        .header h1 { font-size: 24px; }
        .stats { display: flex; gap: 20px; }
        .stat { background: #16213e; padding: 10px 20px; border-radius: 8px; }
        .stat-label { font-size: 12px; color: #888; }
        .stat-value { font-size: 18px; font-weight: bold; }
        .stat-value.positive { color: #00ff88; }
        .stat-value.negative { color: #ff4444; }
        .chart-container { background: #16213e; border-radius: 8px; padding: 20px; margin-bottom: 20px; }
        .chart-wrapper { position: relative; height: 400px; }
        .chart-wrapper.small { height: 150px; }
        canvas { width: 100% !important; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
            <div class="stats" id="stats"></div>
        </div>

        <div class="chart-container">
            <div class="chart-wrapper">
                <canvas id="priceChart"></canvas>
            </div>
        </div>

        <div class="chart-container" id="equityContainer" style="display: none;">
            <div class="chart-wrapper small">
                <canvas id="equityChart"></canvas>
            </div>
        </div>

        <div class="chart-container" id="drawdownContainer" style="display: none;">
            <div class="chart-wrapper small">
                <canvas id="drawdownChart"></canvas>
            </div>
        </div>
    </div>

    <script>
        const chartData = {{.ChartData}};

        // Display stats
        const statsEl = document.getElementById('stats');
        Object.entries(chartData.stats).forEach(([key, value]) => {
            const label = key.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase());
            const isPositive = value.includes('%') && !value.includes('-');
            const isNegative = value.includes('-');
            statsEl.innerHTML += ` + "`" + `
                <div class="stat">
                    <div class="stat-label">${label}</div>
                    <div class="stat-value ${isPositive ? 'positive' : ''} ${isNegative ? 'negative' : ''}">${value}</div>
                </div>
            ` + "`" + `;
        });

        // Price chart with candlesticks
        const priceCtx = document.getElementById('priceChart').getContext('2d');

        const datasets = [{
            label: 'Price',
            data: chartData.labels.map((label, i) => chartData.ohlc[i] ? chartData.ohlc[i][3] : null),
            borderColor: '#36A2EB',
            backgroundColor: 'rgba(54, 162, 235, 0.1)',
            fill: true,
            tension: 0.1,
            pointRadius: 0
        }];

        // Add indicators that overlay on price
        chartData.indicators.filter(ind => ind.overlay).forEach((ind, idx) => {
            datasets.push({
                label: ind.name,
                data: ind.values,
                borderColor: ind.color || '#FF6384',
                backgroundColor: 'transparent',
                borderWidth: 1,
                pointRadius: 0
            });
        });

        // Add trade markers
        if (chartData.config.plotTrades && chartData.trades.length > 0) {
            const entryLong = chartData.trades.filter(t => t.type === 'entry_long').map(t => ({
                x: t.time,
                y: t.price
            }));
            const exitLong = chartData.trades.filter(t => t.type === 'exit_long').map(t => ({
                x: t.time,
                y: t.price
            }));
            const entryShort = chartData.trades.filter(t => t.type === 'entry_short').map(t => ({
                x: t.time,
                y: t.price
            }));
            const exitShort = chartData.trades.filter(t => t.type === 'exit_short').map(t => ({
                x: t.time,
                y: t.price
            }));

            if (entryLong.length > 0) {
                datasets.push({
                    label: 'Buy',
                    data: entryLong.map(e => chartData.labels.indexOf(e.x) >= 0 ? e.y : null),
                    pointStyle: 'triangle',
                    pointRadius: 8,
                    pointBackgroundColor: '#00ff88',
                    showLine: false
                });
            }
            if (exitLong.length > 0) {
                datasets.push({
                    label: 'Sell',
                    data: exitLong.map(e => chartData.labels.indexOf(e.x) >= 0 ? e.y : null),
                    pointStyle: 'rectRot',
                    pointRadius: 8,
                    pointBackgroundColor: '#ff4444',
                    showLine: false
                });
            }
        }

        new Chart(priceCtx, {
            type: 'line',
            data: {
                labels: chartData.labels,
                datasets: datasets
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    intersect: false,
                    mode: 'index'
                },
                plugins: {
                    legend: {
                        display: chartData.config.showLegend,
                        labels: { color: '#eee' }
                    },
                    tooltip: {
                        backgroundColor: 'rgba(0,0,0,0.8)'
                    }
                },
                scales: {
                    x: {
                        ticks: { color: '#888', maxTicksLimit: 10 },
                        grid: { color: '#333' }
                    },
                    y: {
                        ticks: { color: '#888' },
                        grid: { color: '#333' }
                    }
                }
            }
        });

        // Equity chart
        if (chartData.config.plotEquity && chartData.equity && chartData.equity.length > 0) {
            document.getElementById('equityContainer').style.display = 'block';
            const equityCtx = document.getElementById('equityChart').getContext('2d');
            new Chart(equityCtx, {
                type: 'line',
                data: {
                    labels: chartData.labels,
                    datasets: [{
                        label: 'Equity',
                        data: chartData.equity,
                        borderColor: '#00ff88',
                        backgroundColor: 'rgba(0, 255, 136, 0.1)',
                        fill: true,
                        tension: 0.1,
                        pointRadius: 0
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { display: false }
                    },
                    scales: {
                        x: {
                            ticks: { color: '#888', maxTicksLimit: 10 },
                            grid: { color: '#333' }
                        },
                        y: {
                            ticks: { color: '#888' },
                            grid: { color: '#333' }
                        }
                    }
                }
            });
        }

        // Drawdown chart
        if (chartData.config.plotDrawdown && chartData.drawdown && chartData.drawdown.length > 0) {
            document.getElementById('drawdownContainer').style.display = 'block';
            const ddCtx = document.getElementById('drawdownChart').getContext('2d');
            new Chart(ddCtx, {
                type: 'line',
                data: {
                    labels: chartData.labels,
                    datasets: [{
                        label: 'Drawdown %',
                        data: chartData.drawdown.map(d => -d),
                        borderColor: '#ff4444',
                        backgroundColor: 'rgba(255, 68, 68, 0.3)',
                        fill: true,
                        tension: 0.1,
                        pointRadius: 0
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { display: false }
                    },
                    scales: {
                        x: {
                            ticks: { color: '#888', maxTicksLimit: 10 },
                            grid: { color: '#333' }
                        },
                        y: {
                            ticks: { color: '#888' },
                            grid: { color: '#333' },
                            max: 0
                        }
                    }
                }
            });
        }
    </script>
</body>
</html>`
