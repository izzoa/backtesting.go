package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadOption configures CSV loading behavior.
type LoadOption func(*loadConfig)

type loadConfig struct {
	timeColumn   string
	openColumn   string
	highColumn   string
	lowColumn    string
	closeColumn  string
	volumeColumn string
	timeFormat   string
	hasHeader    bool
}

func defaultLoadConfig() *loadConfig {
	return &loadConfig{
		timeColumn:   "", // Auto-detect
		openColumn:   "", // Auto-detect
		highColumn:   "", // Auto-detect
		lowColumn:    "", // Auto-detect
		closeColumn:  "", // Auto-detect
		volumeColumn: "", // Auto-detect
		timeFormat:   "", // Auto-detect
		hasHeader:    true,
	}
}

// WithTimeColumn sets the time/date column name.
func WithTimeColumn(name string) LoadOption {
	return func(c *loadConfig) {
		c.timeColumn = name
	}
}

// WithOpenColumn sets the open price column name.
func WithOpenColumn(name string) LoadOption {
	return func(c *loadConfig) {
		c.openColumn = name
	}
}

// WithHighColumn sets the high price column name.
func WithHighColumn(name string) LoadOption {
	return func(c *loadConfig) {
		c.highColumn = name
	}
}

// WithLowColumn sets the low price column name.
func WithLowColumn(name string) LoadOption {
	return func(c *loadConfig) {
		c.lowColumn = name
	}
}

// WithCloseColumn sets the close price column name.
func WithCloseColumn(name string) LoadOption {
	return func(c *loadConfig) {
		c.closeColumn = name
	}
}

// WithVolumeColumn sets the volume column name.
func WithVolumeColumn(name string) LoadOption {
	return func(c *loadConfig) {
		c.volumeColumn = name
	}
}

// WithTimeFormat sets the time format for parsing dates.
// If not set, common formats are auto-detected.
func WithTimeFormat(format string) LoadOption {
	return func(c *loadConfig) {
		c.timeFormat = format
	}
}

// WithHeader specifies whether the CSV has a header row.
func WithHeader(hasHeader bool) LoadOption {
	return func(c *loadConfig) {
		c.hasHeader = hasHeader
	}
}

// LoadCSV loads OHLCV data from a CSV file.
func LoadCSV(path string, opts ...LoadOption) (*OHLCV, error) {
	cfg := defaultLoadConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	// Parse header and find column indices
	var header []string
	dataStart := 0
	if cfg.hasHeader {
		header = records[0]
		dataStart = 1
	} else {
		// Generate numeric headers
		header = make([]string, len(records[0]))
		for i := range header {
			header[i] = strconv.Itoa(i)
		}
	}

	indices, err := findColumnIndices(header, cfg)
	if err != nil {
		return nil, fmt.Errorf("find columns: %w", err)
	}

	// Detect time format if not specified
	timeFormat := cfg.timeFormat
	if timeFormat == "" && len(records) > dataStart {
		timeFormat = detectTimeFormat(records[dataStart][indices.time])
	}

	// Parse data
	numRows := len(records) - dataStart
	ohlcv := NewOHLCV(numRows)

	for i := dataStart; i < len(records); i++ {
		row := records[i]

		t, err := parseTime(row[indices.time], timeFormat)
		if err != nil {
			return nil, fmt.Errorf("row %d: parse time %q: %w", i, row[indices.time], err)
		}

		open, err := strconv.ParseFloat(row[indices.open], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: parse open: %w", i, err)
		}

		high, err := strconv.ParseFloat(row[indices.high], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: parse high: %w", i, err)
		}

		low, err := strconv.ParseFloat(row[indices.low], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: parse low: %w", i, err)
		}

		close, err := strconv.ParseFloat(row[indices.close], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: parse close: %w", i, err)
		}

		var volume float64
		if indices.volume >= 0 {
			volume, err = strconv.ParseFloat(row[indices.volume], 64)
			if err != nil {
				return nil, fmt.Errorf("row %d: parse volume: %w", i, err)
			}
		}

		ohlcv.Append(Bar{
			Time:   t,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return ohlcv, nil
}

type columnIndices struct {
	time   int
	open   int
	high   int
	low    int
	close  int
	volume int
}

func findColumnIndices(header []string, cfg *loadConfig) (*columnIndices, error) {
	indices := &columnIndices{
		time:   -1,
		open:   -1,
		high:   -1,
		low:    -1,
		close:  -1,
		volume: -1,
	}

	// Normalize header names for matching
	normalizedHeader := make([]string, len(header))
	for i, h := range header {
		normalizedHeader[i] = strings.ToLower(strings.TrimSpace(h))
	}

	// Find time column
	if cfg.timeColumn != "" {
		indices.time = findColumn(normalizedHeader, cfg.timeColumn)
	} else {
		indices.time = findTimeColumnWithIndex(normalizedHeader)
	}
	if indices.time < 0 {
		return nil, fmt.Errorf("time column not found")
	}

	// Find OHLCV columns
	if cfg.openColumn != "" {
		indices.open = findColumn(normalizedHeader, cfg.openColumn)
	} else {
		indices.open = findColumnAny(normalizedHeader, []string{"open", "o"})
	}
	if indices.open < 0 {
		return nil, fmt.Errorf("open column not found")
	}

	if cfg.highColumn != "" {
		indices.high = findColumn(normalizedHeader, cfg.highColumn)
	} else {
		indices.high = findColumnAny(normalizedHeader, []string{"high", "h"})
	}
	if indices.high < 0 {
		return nil, fmt.Errorf("high column not found")
	}

	if cfg.lowColumn != "" {
		indices.low = findColumn(normalizedHeader, cfg.lowColumn)
	} else {
		indices.low = findColumnAny(normalizedHeader, []string{"low", "l"})
	}
	if indices.low < 0 {
		return nil, fmt.Errorf("low column not found")
	}

	if cfg.closeColumn != "" {
		indices.close = findColumn(normalizedHeader, cfg.closeColumn)
	} else {
		indices.close = findColumnAny(normalizedHeader, []string{"close", "c", "adj close"})
	}
	if indices.close < 0 {
		return nil, fmt.Errorf("close column not found")
	}

	// Volume is optional
	if cfg.volumeColumn != "" {
		indices.volume = findColumn(normalizedHeader, cfg.volumeColumn)
	} else {
		indices.volume = findColumnAny(normalizedHeader, []string{"volume", "vol", "v"})
	}

	return indices, nil
}

func findColumn(header []string, name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

func findColumnAny(header []string, names []string) int {
	for _, name := range names {
		if idx := findColumn(header, name); idx >= 0 {
			return idx
		}
	}
	return -1
}

// findTimeColumnWithIndex looks for time column, also checking for unnamed index column
func findTimeColumnWithIndex(header []string) int {
	// First try standard names
	idx := findColumnAny(header, []string{"date", "time", "datetime", "timestamp", "dt"})
	if idx >= 0 {
		return idx
	}
	// Check if first column is unnamed (empty string) - often the index
	if len(header) > 0 && header[0] == "" {
		return 0
	}
	return -1
}

var commonTimeFormats = []string{
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05Z07:00",
	"01/02/2006",
	"01/02/2006 15:04:05",
	"1/2/2006",
	"02-Jan-2006",
	"2006/01/02",
	time.RFC3339,
}

func detectTimeFormat(sample string) string {
	sample = strings.TrimSpace(sample)

	// Try parsing as Unix timestamp
	if _, err := strconv.ParseInt(sample, 10, 64); err == nil {
		return "unix"
	}

	// Try common formats
	for _, format := range commonTimeFormats {
		if _, err := time.Parse(format, sample); err == nil {
			return format
		}
	}

	return "2006-01-02" // Default fallback
}

func parseTime(s string, format string) (time.Time, error) {
	s = strings.TrimSpace(s)

	if format == "unix" {
		unix, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(unix, 0).UTC(), nil
	}

	return time.Parse(format, s)
}
