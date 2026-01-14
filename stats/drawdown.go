package stats

import (
	"time"
)

// Drawdown represents a single drawdown period.
type Drawdown struct {
	Peak        float64   // Peak equity value
	Trough      float64   // Trough equity value
	DrawdownPct float64   // Drawdown percentage
	StartBar    int       // Bar when drawdown started
	TroughBar   int       // Bar when trough was reached
	EndBar      int       // Bar when recovered (-1 if not recovered)
	Duration    int       // Duration in bars
	StartTime   time.Time // Time when drawdown started
	TroughTime  time.Time // Time when trough was reached
	EndTime     time.Time // Time when recovered (zero if not recovered)
}

// CalcDrawdowns calculates all drawdown periods from equity curve.
func CalcDrawdowns(equity []float64, times []time.Time) []Drawdown {
	if len(equity) == 0 {
		return nil
	}

	var drawdowns []Drawdown
	peak := equity[0]
	peakBar := 0
	var peakTime time.Time
	if len(times) > 0 {
		peakTime = times[0]
	}

	inDrawdown := false
	var currentDD Drawdown

	for i, eq := range equity {
		if eq >= peak {
			// New peak or recovery
			if inDrawdown {
				// End current drawdown
				currentDD.EndBar = i
				if len(times) > i {
					currentDD.EndTime = times[i]
				}
				currentDD.Duration = i - currentDD.StartBar
				drawdowns = append(drawdowns, currentDD)
				inDrawdown = false
			}
			peak = eq
			peakBar = i
			if len(times) > i {
				peakTime = times[i]
			}
		} else {
			// In drawdown
			ddPct := (peak - eq) / peak * 100

			if !inDrawdown {
				// Start new drawdown
				inDrawdown = true
				currentDD = Drawdown{
					Peak:        peak,
					Trough:      eq,
					DrawdownPct: ddPct,
					StartBar:    peakBar,
					TroughBar:   i,
					EndBar:      -1, // Not recovered yet
					StartTime:   peakTime,
				}
				if len(times) > i {
					currentDD.TroughTime = times[i]
				}
			} else {
				// Update trough if deeper
				if eq < currentDD.Trough {
					currentDD.Trough = eq
					currentDD.DrawdownPct = ddPct
					currentDD.TroughBar = i
					if len(times) > i {
						currentDD.TroughTime = times[i]
					}
				}
			}
		}
	}

	// Handle ongoing drawdown at end of data
	if inDrawdown {
		currentDD.Duration = len(equity) - 1 - currentDD.StartBar
		drawdowns = append(drawdowns, currentDD)
	}

	return drawdowns
}

// CalcMaxDrawdown returns the maximum drawdown from equity curve.
func CalcMaxDrawdown(equity []float64) (float64, float64) {
	if len(equity) == 0 {
		return 0, 0
	}

	peak := equity[0]
	maxDD := 0.0
	maxDDValue := 0.0

	for _, eq := range equity {
		if eq > peak {
			peak = eq
		}

		dd := (peak - eq) / peak * 100
		ddValue := peak - eq

		if dd > maxDD {
			maxDD = dd
			maxDDValue = ddValue
		}
	}

	return maxDD, maxDDValue
}

// CalcAvgDrawdown calculates the average drawdown percentage.
func CalcAvgDrawdown(drawdowns []Drawdown) float64 {
	if len(drawdowns) == 0 {
		return 0
	}

	sum := 0.0
	for _, dd := range drawdowns {
		sum += dd.DrawdownPct
	}
	return sum / float64(len(drawdowns))
}

// CalcDrawdownDurations returns max and average drawdown durations.
func CalcDrawdownDurations(drawdowns []Drawdown, times []time.Time) (max, avg time.Duration) {
	if len(drawdowns) == 0 {
		return 0, 0
	}

	var maxDuration time.Duration
	var totalDuration time.Duration

	for _, dd := range drawdowns {
		var duration time.Duration

		if dd.EndBar >= 0 && len(times) > dd.EndBar {
			// Recovered drawdown
			duration = times[dd.EndBar].Sub(dd.StartTime)
		} else if dd.TroughBar >= 0 && len(times) > dd.TroughBar {
			// Ongoing drawdown - use trough time as minimum
			duration = times[dd.TroughBar].Sub(dd.StartTime)
		}

		totalDuration += duration
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	return maxDuration, totalDuration / time.Duration(len(drawdowns))
}

// CalcDrawdownSeries calculates drawdown percentage at each bar.
func CalcDrawdownSeries(equity []float64) []float64 {
	if len(equity) == 0 {
		return nil
	}

	drawdowns := make([]float64, len(equity))
	peak := equity[0]

	for i, eq := range equity {
		if eq > peak {
			peak = eq
		}
		if peak > 0 {
			drawdowns[i] = (peak - eq) / peak * 100
		}
	}

	return drawdowns
}

// CalcDrawdownDurationSeries calculates bars in current drawdown at each bar.
func CalcDrawdownDurationSeries(equity []float64) []int {
	if len(equity) == 0 {
		return nil
	}

	durations := make([]int, len(equity))
	peak := equity[0]
	peakBar := 0

	for i, eq := range equity {
		if eq >= peak {
			peak = eq
			peakBar = i
			durations[i] = 0
		} else {
			durations[i] = i - peakBar
		}
	}

	return durations
}
