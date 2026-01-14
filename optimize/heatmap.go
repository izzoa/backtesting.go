package optimize

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

// Heatmap represents a 2D heatmap of optimization results.
type Heatmap struct {
	// XParam is the name of the x-axis parameter.
	XParam string
	// YParam is the name of the y-axis parameter.
	YParam string
	// Metric is the name of the metric displayed.
	Metric string

	// XLabels are the unique values of the x-axis parameter.
	XLabels []string
	// YLabels are the unique values of the y-axis parameter.
	YLabels []string

	// Values is the 2D array of metric values [y][x].
	Values [][]float64
}

// GenerateHeatmap creates a heatmap from optimization results.
func GenerateHeatmap(results []ParamResult, xParam, yParam, metric string) *Heatmap {
	if len(results) == 0 {
		return nil
	}

	// Collect unique values for each parameter
	xValues := make(map[string]bool)
	yValues := make(map[string]bool)

	for _, r := range results {
		if xv, ok := r.Params[xParam]; ok {
			xValues[fmt.Sprintf("%v", xv)] = true
		}
		if yv, ok := r.Params[yParam]; ok {
			yValues[fmt.Sprintf("%v", yv)] = true
		}
	}

	// Sort labels
	xLabels := make([]string, 0, len(xValues))
	for k := range xValues {
		xLabels = append(xLabels, k)
	}
	sortNumericStrings(xLabels)

	yLabels := make([]string, 0, len(yValues))
	for k := range yValues {
		yLabels = append(yLabels, k)
	}
	sortNumericStrings(yLabels)

	// Create index maps
	xIndex := make(map[string]int)
	for i, l := range xLabels {
		xIndex[l] = i
	}
	yIndex := make(map[string]int)
	for i, l := range yLabels {
		yIndex[l] = i
	}

	// Initialize values with NaN
	values := make([][]float64, len(yLabels))
	for i := range values {
		values[i] = make([]float64, len(xLabels))
		for j := range values[i] {
			values[i][j] = math.NaN()
		}
	}

	// Fill in values
	for _, r := range results {
		xv := fmt.Sprintf("%v", r.Params[xParam])
		yv := fmt.Sprintf("%v", r.Params[yParam])

		xi, xok := xIndex[xv]
		yi, yok := yIndex[yv]

		if xok && yok {
			// Get metric value
			if v, ok := r.Stats[metric]; ok {
				values[yi][xi] = v
			} else {
				values[yi][xi] = r.Value
			}
		}
	}

	return &Heatmap{
		XParam:  xParam,
		YParam:  yParam,
		Metric:  metric,
		XLabels: xLabels,
		YLabels: yLabels,
		Values:  values,
	}
}

// ToCSV exports the heatmap to a CSV file.
func (h *Heatmap) ToCSV(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header row (x labels)
	header := append([]string{h.YParam + "/" + h.XParam}, h.XLabels...)
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for i, yLabel := range h.YLabels {
		row := make([]string, len(h.XLabels)+1)
		row[0] = yLabel
		for j, val := range h.Values[i] {
			if math.IsNaN(val) {
				row[j+1] = ""
			} else {
				row[j+1] = strconv.FormatFloat(val, 'f', 4, 64)
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// Get returns the value at the specified x and y labels.
func (h *Heatmap) Get(xLabel, yLabel string) (float64, bool) {
	xi := -1
	for i, l := range h.XLabels {
		if l == xLabel {
			xi = i
			break
		}
	}

	yi := -1
	for i, l := range h.YLabels {
		if l == yLabel {
			yi = i
			break
		}
	}

	if xi < 0 || yi < 0 {
		return 0, false
	}

	return h.Values[yi][xi], !math.IsNaN(h.Values[yi][xi])
}

// Max returns the maximum value and its coordinates.
func (h *Heatmap) Max() (value float64, xLabel, yLabel string) {
	value = math.Inf(-1)
	for yi, row := range h.Values {
		for xi, v := range row {
			if !math.IsNaN(v) && v > value {
				value = v
				xLabel = h.XLabels[xi]
				yLabel = h.YLabels[yi]
			}
		}
	}
	return
}

// Min returns the minimum value and its coordinates.
func (h *Heatmap) Min() (value float64, xLabel, yLabel string) {
	value = math.Inf(1)
	for yi, row := range h.Values {
		for xi, v := range row {
			if !math.IsNaN(v) && v < value {
				value = v
				xLabel = h.XLabels[xi]
				yLabel = h.YLabels[yi]
			}
		}
	}
	return
}

// sortNumericStrings sorts strings as numbers if possible, otherwise alphabetically.
func sortNumericStrings(s []string) {
	sort.Slice(s, func(i, j int) bool {
		ni, erri := strconv.ParseFloat(s[i], 64)
		nj, errj := strconv.ParseFloat(s[j], 64)

		if erri == nil && errj == nil {
			return ni < nj
		}
		return s[i] < s[j]
	})
}
