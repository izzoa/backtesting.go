package optimize

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sync"

	backtesting "github.com/quickfixgo/backtesting"
	"github.com/quickfixgo/backtesting/data"
)

// GridConfig configures grid search optimization.
type GridConfig struct {
	// Params maps parameter names to their possible values.
	Params map[string][]interface{}

	// Constraint is an optional function to filter invalid parameter combinations.
	Constraint ConstraintFunc

	// Maximize is the metric to maximize (e.g., "ReturnPct", "SharpeRatio").
	Maximize string

	// MaxTries limits the number of combinations to test. 0 = test all.
	MaxTries int

	// RandomState is the seed for random sampling when MaxTries < total combinations.
	RandomState int64

	// Workers is the number of parallel workers. 0 = use all CPUs.
	Workers int
}

// StrategyFactory creates a new strategy instance with the given parameters.
type StrategyFactory func(params map[string]interface{}) backtesting.Strategy

// GridSearch performs grid search optimization over parameter combinations.
func GridSearch(
	ohlcv *data.OHLCV,
	factory StrategyFactory,
	cfg backtesting.BacktestConfig,
	gridCfg GridConfig,
) (*Result, error) {
	// Generate all parameter combinations
	combinations := generateCombinations(gridCfg.Params)

	// Apply constraint filter
	if gridCfg.Constraint != nil {
		filtered := make([]map[string]interface{}, 0, len(combinations))
		for _, combo := range combinations {
			if gridCfg.Constraint(combo) {
				filtered = append(filtered, combo)
			}
		}
		combinations = filtered
	}

	if len(combinations) == 0 {
		return nil, fmt.Errorf("no valid parameter combinations after applying constraints")
	}

	// Sample if MaxTries is set
	if gridCfg.MaxTries > 0 && gridCfg.MaxTries < len(combinations) {
		rng := rand.New(rand.NewSource(gridCfg.RandomState))
		rng.Shuffle(len(combinations), func(i, j int) {
			combinations[i], combinations[j] = combinations[j], combinations[i]
		})
		combinations = combinations[:gridCfg.MaxTries]
	}

	// Run backtests
	results := runParallel(ohlcv, factory, cfg, combinations, gridCfg.Workers)

	// Find best result
	result := &Result{
		Metric:     gridCfg.Maximize,
		Maximize:   true,
		AllResults: results,
	}

	bestValue := math.Inf(-1)
	for _, r := range results {
		if r.Error == nil && r.Value > bestValue {
			bestValue = r.Value
			result.BestValue = r.Value
			result.BestParams = r.Params
		}
	}

	return result, nil
}

// generateCombinations generates all combinations of parameter values.
func generateCombinations(params map[string][]interface{}) []map[string]interface{} {
	if len(params) == 0 {
		return nil
	}

	// Get parameter names in consistent order
	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}

	// Calculate total combinations
	total := 1
	for _, values := range params {
		total *= len(values)
	}

	combinations := make([]map[string]interface{}, 0, total)

	// Generate combinations using indices
	indices := make([]int, len(names))

	for {
		// Create combination from current indices
		combo := make(map[string]interface{})
		for i, name := range names {
			combo[name] = params[name][indices[i]]
		}
		combinations = append(combinations, combo)

		// Increment indices (like counting)
		carry := true
		for i := len(indices) - 1; i >= 0 && carry; i-- {
			indices[i]++
			if indices[i] >= len(params[names[i]]) {
				indices[i] = 0
			} else {
				carry = false
			}
		}

		if carry {
			break // All combinations generated
		}
	}

	return combinations
}

// runParallel runs backtests in parallel.
func runParallel(
	ohlcv *data.OHLCV,
	factory StrategyFactory,
	cfg backtesting.BacktestConfig,
	combinations []map[string]interface{},
	workers int,
) []ParamResult {
	if workers <= 0 {
		workers = 4 // Default to 4 workers
	}

	results := make([]ParamResult, len(combinations))
	jobs := make(chan int, len(combinations))
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				params := combinations[idx]
				results[idx] = runBacktest(ohlcv, factory, cfg, params)
			}
		}()
	}

	// Send jobs
	for i := range combinations {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	return results
}

// runBacktest runs a single backtest with the given parameters.
func runBacktest(
	ohlcv *data.OHLCV,
	factory StrategyFactory,
	cfg backtesting.BacktestConfig,
	params map[string]interface{},
) ParamResult {
	result := ParamResult{
		Params: params,
		Stats:  make(map[string]float64),
	}

	defer func() {
		if r := recover(); r != nil {
			result.Error = fmt.Errorf("panic: %v", r)
		}
	}()

	// Create strategy with parameters
	strategy := factory(params)

	// Run backtest
	bt := backtesting.NewBacktest(ohlcv, strategy, cfg)
	btResult, err := bt.Run()
	if err != nil {
		result.Error = err
		return result
	}

	// Extract statistics
	result.Stats["ReturnPct"] = btResult.ReturnPct
	result.Stats["NumTrades"] = float64(btResult.NumTrades)
	result.Stats["WinRate"] = btResult.WinRate
	result.Stats["ProfitFactor"] = btResult.ProfitFactor
	result.Stats["MaxDrawdownPct"] = btResult.MaxDrawdownPct
	result.Stats["AvgTradePct"] = btResult.AvgTradePct
	result.Stats["FinalEquity"] = btResult.FinalEquity

	// Set the optimization value (default to ReturnPct)
	result.Value = btResult.ReturnPct

	return result
}

// SetOptimizeMetric sets which metric to use for the optimization value.
func SetOptimizeMetric(results []ParamResult, metric string) {
	for i := range results {
		if v, ok := results[i].Stats[metric]; ok {
			results[i].Value = v
		}
	}
}

// InjectParams injects parameter values into a strategy using reflection.
func InjectParams(strategy interface{}, params map[string]interface{}) error {
	v := reflect.ValueOf(strategy)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("strategy must be a struct or pointer to struct")
	}

	t := v.Type()
	for name, value := range params {
		// Try to find field by name
		field := v.FieldByName(name)
		if !field.IsValid() {
			// Try case-insensitive match
			for i := 0; i < t.NumField(); i++ {
				if t.Field(i).Name == name {
					field = v.Field(i)
					break
				}
			}
		}

		if !field.IsValid() || !field.CanSet() {
			continue // Skip fields that don't exist or can't be set
		}

		// Set the value
		val := reflect.ValueOf(value)
		if val.Type().ConvertibleTo(field.Type()) {
			field.Set(val.Convert(field.Type()))
		}
	}

	return nil
}

// Range generates a slice of values from start to end (inclusive) with step.
func Range(start, end, step int) []interface{} {
	if step == 0 {
		step = 1
	}
	if (step > 0 && start > end) || (step < 0 && start < end) {
		return nil
	}

	var values []interface{}
	for v := start; (step > 0 && v <= end) || (step < 0 && v >= end); v += step {
		values = append(values, v)
	}
	return values
}

// RangeFloat generates a slice of float values from start to end with step.
func RangeFloat(start, end, step float64) []interface{} {
	if step == 0 {
		step = 1
	}
	if (step > 0 && start > end) || (step < 0 && start < end) {
		return nil
	}

	var values []interface{}
	for v := start; (step > 0 && v <= end+step/2) || (step < 0 && v >= end-step/2); v += step {
		values = append(values, v)
	}
	return values
}

// Values creates a slice of interface{} from the given values.
func Values(vals ...interface{}) []interface{} {
	return vals
}
