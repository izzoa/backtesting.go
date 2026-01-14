// Package testutil provides testing utilities for the backtesting package.
package testutil

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/quickfixgo/backtesting/data"
)

// Tolerance constants for comparisons.
const (
	FloatTolerance   = 1e-9   // For exact calculations
	PercentTolerance = 1e-6   // For percentage comparisons
	PriceTolerance   = 1e-4   // For price comparisons
	StatsTolerance   = 1e-4   // For statistical metrics
)

// AlmostEqual compares two float64 values within a tolerance.
// Handles NaN and Inf values correctly.
func AlmostEqual(a, b, tolerance float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		return a == b
	}
	return math.Abs(a-b) <= tolerance
}

// SlicesAlmostEqual compares two float64 slices within a tolerance.
func SlicesAlmostEqual(a, b []float64, tolerance float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !AlmostEqual(a[i], b[i], tolerance) {
			return false
		}
	}
	return true
}

// RequireAlmostEqual fails the test if values are not almost equal.
func RequireAlmostEqual(t *testing.T, expected, actual, tolerance float64, msg string) {
	t.Helper()
	if !AlmostEqual(expected, actual, tolerance) {
		t.Errorf("%s: expected %v, got %v (tolerance: %v)", msg, expected, actual, tolerance)
	}
}

// RequireSlicesAlmostEqual fails the test if slices are not almost equal.
func RequireSlicesAlmostEqual(t *testing.T, expected, actual []float64, tolerance float64, msg string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("%s: length mismatch: expected %d, got %d", msg, len(expected), len(actual))
		return
	}
	for i := range expected {
		if !AlmostEqual(expected[i], actual[i], tolerance) {
			t.Errorf("%s[%d]: expected %v, got %v", msg, i, expected[i], actual[i])
		}
	}
}

// projectRoot returns the root directory of the project.
func projectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine project root")
	}
	// internal/testutil/helpers.go -> project root
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// TestDataPath returns the path to a test data file.
func TestDataPath(name string) string {
	return filepath.Join(projectRoot(), "testdata", "csv", name)
}

// GoldenPath returns the path to a golden file.
func GoldenPath(name string) string {
	return filepath.Join(projectRoot(), "testdata", "golden", name+".json")
}

// LoadTestData loads OHLCV data from testdata/csv.
func LoadTestData(t *testing.T, name string) *data.OHLCV {
	t.Helper()
	path := TestDataPath(name + ".csv")
	ohlcv, err := data.LoadCSV(path)
	if err != nil {
		t.Fatalf("failed to load test data %s: %v", name, err)
	}
	return ohlcv
}

// LoadGolden loads golden file data into the provided struct.
func LoadGolden(t *testing.T, name string, v interface{}) {
	t.Helper()
	path := GoldenPath(name)
	fileData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load golden file %s: %v", name, err)
	}
	if err := json.Unmarshal(fileData, v); err != nil {
		t.Fatalf("failed to parse golden file %s: %v", name, err)
	}
}

// UpdateGolden updates a golden file with new data.
// Only call this when intentionally updating test expectations.
func UpdateGolden(t *testing.T, name string, v interface{}) {
	t.Helper()
	path := GoldenPath(name)
	fileData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal golden data: %v", err)
	}
	if err := os.WriteFile(path, fileData, 0644); err != nil {
		t.Fatalf("failed to write golden file: %v", err)
	}
}

// MakeTestOHLCV creates synthetic OHLCV data for testing.
func MakeTestOHLCV(length int, startPrice float64) *data.OHLCV {
	return MakeTestOHLCVWithSeed(length, startPrice, time.Now().UnixNano())
}

// MakeTestOHLCVWithSeed creates synthetic OHLCV data with a specific random seed.
func MakeTestOHLCVWithSeed(length int, startPrice float64, seed int64) *data.OHLCV {
	rng := rand.New(rand.NewSource(seed))
	ohlcv := data.NewOHLCV(length)
	price := startPrice
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < length; i++ {
		// Generate random price movement (-1% to +1%)
		change := (rng.Float64() - 0.5) * 2.0 / 100.0
		open := price
		close := price * (1 + change)

		// High is max of open/close plus some random upward movement
		high := math.Max(open, close) * (1 + rng.Float64()*0.005)
		// Low is min of open/close minus some random downward movement
		low := math.Min(open, close) * (1 - rng.Float64()*0.005)

		ohlcv.Append(data.Bar{
			Time:   baseTime.AddDate(0, 0, i),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: float64(rng.Intn(1000000) + 100000),
		})

		price = close
	}

	return ohlcv
}

// MakeTrendingOHLCV creates OHLCV data with a specific trend.
// trend > 0 creates upward trend, trend < 0 creates downward trend.
func MakeTrendingOHLCV(length int, startPrice float64, trendPct float64, seed int64) *data.OHLCV {
	rng := rand.New(rand.NewSource(seed))
	ohlcv := data.NewOHLCV(length)
	price := startPrice
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	dailyTrend := trendPct / 100.0 / float64(length)

	for i := 0; i < length; i++ {
		// Add trend plus random noise
		noise := (rng.Float64() - 0.5) * 0.01
		change := dailyTrend + noise
		open := price
		close := price * (1 + change)

		high := math.Max(open, close) * (1 + rng.Float64()*0.005)
		low := math.Min(open, close) * (1 - rng.Float64()*0.005)

		ohlcv.Append(data.Bar{
			Time:   baseTime.AddDate(0, 0, i),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: float64(rng.Intn(1000000) + 100000),
		})

		price = close
	}

	return ohlcv
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}
