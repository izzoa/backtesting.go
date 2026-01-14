package optimize

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
)

// Result contains the results of an optimization run.
type Result struct {
	// BestParams are the parameters that produced the best result.
	BestParams map[string]interface{}

	// BestValue is the value of the optimized metric for the best parameters.
	BestValue float64

	// AllResults contains results for all parameter combinations tested.
	AllResults []ParamResult

	// Metric is the name of the metric that was optimized.
	Metric string

	// Maximize indicates whether the metric was maximized (true) or minimized (false).
	Maximize bool
}

// ParamResult contains the result for a single parameter combination.
type ParamResult struct {
	// Params are the parameter values used.
	Params map[string]interface{}

	// Stats contains all computed statistics for this run.
	Stats map[string]float64

	// Value is the value of the optimized metric.
	Value float64

	// Error is set if the backtest failed for this parameter combination.
	Error error
}

// Len returns the number of results.
func (r *Result) Len() int {
	return len(r.AllResults)
}

// SortedResults returns results sorted by value (descending if maximize, ascending otherwise).
func (r *Result) SortedResults() []ParamResult {
	sorted := make([]ParamResult, len(r.AllResults))
	copy(sorted, r.AllResults)

	sort.Slice(sorted, func(i, j int) bool {
		if r.Maximize {
			return sorted[i].Value > sorted[j].Value
		}
		return sorted[i].Value < sorted[j].Value
	})

	return sorted
}

// TopN returns the top N results.
func (r *Result) TopN(n int) []ParamResult {
	sorted := r.SortedResults()
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// Heatmap generates a 2D heatmap for two parameters.
func (r *Result) Heatmap(param1, param2 string) *Heatmap {
	return GenerateHeatmap(r.AllResults, param1, param2, r.Metric)
}

// ToCSV exports results to a CSV file.
func (r *Result) ToCSV(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Collect all parameter names and stat names
	paramNames := make(map[string]bool)
	statNames := make(map[string]bool)

	for _, res := range r.AllResults {
		for k := range res.Params {
			paramNames[k] = true
		}
		for k := range res.Stats {
			statNames[k] = true
		}
	}

	// Create sorted slices for consistent column order
	var paramCols []string
	for k := range paramNames {
		paramCols = append(paramCols, k)
	}
	sort.Strings(paramCols)

	var statCols []string
	for k := range statNames {
		statCols = append(statCols, k)
	}
	sort.Strings(statCols)

	// Write header
	header := append(paramCols, statCols...)
	header = append(header, "OptimizedValue")
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, res := range r.AllResults {
		row := make([]string, 0, len(header))

		// Parameters
		for _, col := range paramCols {
			if v, ok := res.Params[col]; ok {
				row = append(row, fmt.Sprintf("%v", v))
			} else {
				row = append(row, "")
			}
		}

		// Stats
		for _, col := range statCols {
			if v, ok := res.Stats[col]; ok {
				row = append(row, strconv.FormatFloat(v, 'f', 6, 64))
			} else {
				row = append(row, "")
			}
		}

		// Optimized value
		row = append(row, strconv.FormatFloat(res.Value, 'f', 6, 64))

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// String returns a summary of the optimization results.
func (r *Result) String() string {
	s := fmt.Sprintf("Optimization Results\n")
	s += fmt.Sprintf("====================\n")
	s += fmt.Sprintf("Metric: %s (maximize=%v)\n", r.Metric, r.Maximize)
	s += fmt.Sprintf("Total combinations tested: %d\n", len(r.AllResults))
	s += fmt.Sprintf("\nBest Parameters:\n")

	for k, v := range r.BestParams {
		s += fmt.Sprintf("  %s: %v\n", k, v)
	}
	s += fmt.Sprintf("\nBest %s: %.4f\n", r.Metric, r.BestValue)

	return s
}
