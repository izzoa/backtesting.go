package lib

import (
	"time"

	"github.com/izzoa/backtesting.go/data"
)

// ResampleRule defines how to group time periods.
type ResampleRule string

const (
	RuleDaily   ResampleRule = "D"
	RuleWeekly  ResampleRule = "W"
	RuleMonthly ResampleRule = "M"
	RuleHourly  ResampleRule = "H"
	Rule4Hour   ResampleRule = "4H"
)

// AggFunc is an aggregation function type.
type AggFunc func([]float64) float64

// Common aggregation functions
var (
	First = func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		return values[0]
	}

	Last = func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		return values[len(values)-1]
	}

	Max = func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	}

	Min = func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	}

	Sum = func(values []float64) float64 {
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	}

	Mean = func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		return Sum(values) / float64(len(values))
	}
)

// Resample resamples OHLCV data to a higher timeframe.
func Resample(d *data.OHLCV, rule ResampleRule) *data.OHLCV {
	if d == nil || d.Len() == 0 {
		return data.NewOHLCV(0)
	}

	// Group bars by period
	groups := groupByPeriod(d.Time, rule)
	if len(groups) == 0 {
		return data.NewOHLCV(0)
	}

	result := data.NewOHLCV(len(groups))

	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		// Extract values for this group
		opens := make([]float64, len(group))
		highs := make([]float64, len(group))
		lows := make([]float64, len(group))
		closes := make([]float64, len(group))
		volumes := make([]float64, len(group))

		for i, idx := range group {
			opens[i] = d.Open[idx]
			highs[i] = d.High[idx]
			lows[i] = d.Low[idx]
			closes[i] = d.Close[idx]
			if len(d.Volume) > idx {
				volumes[i] = d.Volume[idx]
			}
		}

		// Aggregate
		bar := data.Bar{
			Time:   d.Time[group[0]],
			Open:   First(opens),
			High:   Max(highs),
			Low:    Min(lows),
			Close:  Last(closes),
			Volume: Sum(volumes),
		}

		result.Append(bar)
	}

	return result
}

// ResampleApply applies an aggregation function to a series with resampling.
func ResampleApply(series []float64, times []time.Time, rule ResampleRule, fn AggFunc) []float64 {
	if len(series) == 0 || len(times) == 0 || len(series) != len(times) {
		return nil
	}

	groups := groupByPeriod(times, rule)
	result := make([]float64, len(groups))

	for i, group := range groups {
		values := make([]float64, len(group))
		for j, idx := range group {
			values[j] = series[idx]
		}
		result[i] = fn(values)
	}

	return result
}

// ResampleApplyAligned applies resampling and aligns back to original timeframe.
// Each bar gets the value from its resampled period.
func ResampleApplyAligned(series []float64, times []time.Time, rule ResampleRule, fn AggFunc) []float64 {
	if len(series) == 0 || len(times) == 0 || len(series) != len(times) {
		return nil
	}

	groups := groupByPeriod(times, rule)
	result := make([]float64, len(series))

	for _, group := range groups {
		values := make([]float64, len(group))
		for j, idx := range group {
			values[j] = series[idx]
		}
		aggValue := fn(values)

		// Assign to all bars in this group
		for _, idx := range group {
			result[idx] = aggValue
		}
	}

	return result
}

// groupByPeriod groups indices by time period.
func groupByPeriod(times []time.Time, rule ResampleRule) [][]int {
	if len(times) == 0 {
		return nil
	}

	var groups [][]int
	var currentGroup []int
	var currentPeriod string

	for i, t := range times {
		period := getPeriodKey(t, rule)

		if period != currentPeriod {
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
			}
			currentGroup = []int{i}
			currentPeriod = period
		} else {
			currentGroup = append(currentGroup, i)
		}
	}

	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

// getPeriodKey returns a unique key for the time period.
func getPeriodKey(t time.Time, rule ResampleRule) string {
	switch rule {
	case RuleHourly:
		return t.Format("2006-01-02-15")
	case Rule4Hour:
		hour := (t.Hour() / 4) * 4
		return t.Format("2006-01-02") + "-" + string(rune('0'+hour/10)) + string(rune('0'+hour%10))
	case RuleDaily:
		return t.Format("2006-01-02")
	case RuleWeekly:
		year, week := t.ISOWeek()
		return t.Format("2006") + "-W" + string(rune('0'+year/1000%10)) + string(rune('0'+week/10)) + string(rune('0'+week%10))
	case RuleMonthly:
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02")
	}
}
