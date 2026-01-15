package optimize

import (
	"math"
	"testing"

	backtesting "github.com/izzoa/backtesting.go"
	"github.com/izzoa/backtesting.go/data"
)

func TestGenerateCombinations(t *testing.T) {
	params := map[string][]interface{}{
		"a": {1, 2},
		"b": {10, 20, 30},
	}

	combinations := generateCombinations(params)

	// Should have 2 * 3 = 6 combinations
	if len(combinations) != 6 {
		t.Errorf("Expected 6 combinations, got %d", len(combinations))
	}

	// Check that all combinations are unique
	seen := make(map[string]bool)
	for _, combo := range combinations {
		key := ""
		for k, v := range combo {
			key += k + ":" + string(rune(v.(int))) + ","
		}
		if seen[key] {
			t.Errorf("Duplicate combination: %v", combo)
		}
		seen[key] = true
	}
}

func TestConstraints(t *testing.T) {
	tests := []struct {
		name       string
		constraint ConstraintFunc
		params     map[string]interface{}
		expected   bool
	}{
		{
			name:       "ParamLessThan true",
			constraint: ParamLessThan("a", "b"),
			params:     map[string]interface{}{"a": 5, "b": 10},
			expected:   true,
		},
		{
			name:       "ParamLessThan false",
			constraint: ParamLessThan("a", "b"),
			params:     map[string]interface{}{"a": 10, "b": 5},
			expected:   false,
		},
		{
			name:       "ParamGreaterThan true",
			constraint: ParamGreaterThan("a", "b"),
			params:     map[string]interface{}{"a": 10, "b": 5},
			expected:   true,
		},
		{
			name:       "ParamNotEqual true",
			constraint: ParamNotEqual("a", "b"),
			params:     map[string]interface{}{"a": 5, "b": 10},
			expected:   true,
		},
		{
			name:       "ParamNotEqual false",
			constraint: ParamNotEqual("a", "b"),
			params:     map[string]interface{}{"a": 10, "b": 10},
			expected:   false,
		},
		{
			name:       "ParamRange true",
			constraint: ParamRange("a", 0, 100),
			params:     map[string]interface{}{"a": 50.0},
			expected:   true,
		},
		{
			name:       "ParamRange false",
			constraint: ParamRange("a", 0, 100),
			params:     map[string]interface{}{"a": 150.0},
			expected:   false,
		},
		{
			name:       "And both true",
			constraint: And(ParamLessThan("a", "b"), ParamMinValue("a", 0)),
			params:     map[string]interface{}{"a": 5, "b": 10},
			expected:   true,
		},
		{
			name:       "And one false",
			constraint: And(ParamLessThan("a", "b"), ParamMinValue("a", 10)),
			params:     map[string]interface{}{"a": 5, "b": 10},
			expected:   false,
		},
		{
			name:       "Or one true",
			constraint: Or(ParamLessThan("a", "b"), ParamGreaterThan("a", "b")),
			params:     map[string]interface{}{"a": 5, "b": 10},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.constraint(tt.params)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name     string
		start    int
		end      int
		step     int
		expected []interface{}
	}{
		{"basic", 1, 5, 1, []interface{}{1, 2, 3, 4, 5}},
		{"step 2", 0, 10, 2, []interface{}{0, 2, 4, 6, 8, 10}},
		{"single", 5, 5, 1, []interface{}{5}},
		{"negative step", 10, 5, -1, []interface{}{10, 9, 8, 7, 6, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Range(tt.start, tt.end, tt.step)
			if len(result) != len(tt.expected) {
				t.Errorf("Length mismatch: got %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Value at %d: got %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestHeatmap(t *testing.T) {
	results := []ParamResult{
		{Params: map[string]interface{}{"x": 1, "y": 10}, Value: 100},
		{Params: map[string]interface{}{"x": 1, "y": 20}, Value: 110},
		{Params: map[string]interface{}{"x": 2, "y": 10}, Value: 90},
		{Params: map[string]interface{}{"x": 2, "y": 20}, Value: 120},
	}

	hm := GenerateHeatmap(results, "x", "y", "Value")

	if hm == nil {
		t.Fatal("Heatmap is nil")
	}

	if len(hm.XLabels) != 2 {
		t.Errorf("Expected 2 x labels, got %d", len(hm.XLabels))
	}

	if len(hm.YLabels) != 2 {
		t.Errorf("Expected 2 y labels, got %d", len(hm.YLabels))
	}

	// Check max value
	maxVal, xLabel, yLabel := hm.Max()
	if maxVal != 120 {
		t.Errorf("Max value = %f, want 120", maxVal)
	}
	if xLabel != "2" || yLabel != "20" {
		t.Errorf("Max at (%s, %s), want (2, 20)", xLabel, yLabel)
	}
}

func TestResult(t *testing.T) {
	results := []ParamResult{
		{Params: map[string]interface{}{"a": 1}, Value: 50},
		{Params: map[string]interface{}{"a": 2}, Value: 100},
		{Params: map[string]interface{}{"a": 3}, Value: 75},
	}

	r := &Result{
		Metric:     "ReturnPct",
		Maximize:   true,
		AllResults: results,
		BestParams: map[string]interface{}{"a": 2},
		BestValue:  100,
	}

	// Test sorted results
	sorted := r.SortedResults()
	if sorted[0].Value != 100 {
		t.Errorf("First sorted value = %f, want 100", sorted[0].Value)
	}
	if sorted[2].Value != 50 {
		t.Errorf("Last sorted value = %f, want 50", sorted[2].Value)
	}

	// Test TopN
	top2 := r.TopN(2)
	if len(top2) != 2 {
		t.Errorf("TopN(2) returned %d results, want 2", len(top2))
	}
	if top2[0].Value != 100 {
		t.Errorf("Top value = %f, want 100", top2[0].Value)
	}
}

// TestSmaCrossStrategy is a simple strategy for testing optimization.
type TestSmaCrossStrategy struct {
	backtesting.StrategyBase
	FastPeriod int
	SlowPeriod int
	fastSMA    *data.Indicator
	slowSMA    *data.Indicator
}

func (s *TestSmaCrossStrategy) Init() {
	s.fastSMA = s.I("SMA_fast", data.SMA(s.Close(), s.FastPeriod))
	s.slowSMA = s.I("SMA_slow", data.SMA(s.Close(), s.SlowPeriod))
}

func (s *TestSmaCrossStrategy) Next() {
	idx := s.BarIndex()
	if idx < 1 {
		return
	}

	fastCurr := s.IndicatorAt(s.fastSMA, idx)
	fastPrev := s.IndicatorAt(s.fastSMA, idx-1)
	slowCurr := s.IndicatorAt(s.slowSMA, idx)
	slowPrev := s.IndicatorAt(s.slowSMA, idx-1)

	if fastPrev <= slowPrev && fastCurr > slowCurr {
		if !s.Position().IsLong() {
			s.Buy()
		}
	}

	if fastPrev >= slowPrev && fastCurr < slowCurr {
		if s.Position().IsLong() {
			s.Position().Close(1.0)
		}
	}
}

func TestGridSearch(t *testing.T) {
	// Load test data
	ohlcv, err := data.LoadCSV("../testdata/csv/GOOG.csv")
	if err != nil {
		t.Skipf("Test data not available: %v", err)
		return
	}

	// Define strategy factory
	factory := func(params map[string]interface{}) backtesting.Strategy {
		fast := params["FastPeriod"].(int)
		slow := params["SlowPeriod"].(int)
		return &TestSmaCrossStrategy{
			FastPeriod: fast,
			SlowPeriod: slow,
		}
	}

	// Run grid search
	result, err := GridSearch(
		ohlcv,
		factory,
		backtesting.BacktestConfig{Cash: 10000},
		GridConfig{
			Params: map[string][]interface{}{
				"FastPeriod": {5, 10, 15},
				"SlowPeriod": {20, 30, 40},
			},
			Constraint: ParamLessThan("FastPeriod", "SlowPeriod"),
			Maximize:   "ReturnPct",
			Workers:    2,
		},
	)

	if err != nil {
		t.Fatalf("GridSearch failed: %v", err)
	}

	// Check results
	if result.Len() != 9 { // 3 * 3 combinations, all satisfy constraint
		t.Errorf("Expected 9 results, got %d", result.Len())
	}

	if result.BestValue <= 0 {
		t.Errorf("Best value should be positive, got %f", result.BestValue)
	}

	t.Logf("Best params: %v", result.BestParams)
	t.Logf("Best return: %.2f%%", result.BestValue)

	// Check that all results have valid stats
	for _, r := range result.AllResults {
		if r.Error != nil {
			t.Errorf("Result has error: %v", r.Error)
		}
		if math.IsNaN(r.Value) {
			t.Error("Result has NaN value")
		}
	}
}
