package data

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func getTestDataPath(name string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine test file path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "testdata", "csv", name)
}

func TestLoadCSV_GOOG(t *testing.T) {
	path := getTestDataPath("GOOG.csv")
	ohlcv, err := LoadCSV(path)
	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}

	if ohlcv.Len() == 0 {
		t.Error("loaded OHLCV is empty")
	}

	// Validate data
	if err := ohlcv.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	// Check first and last values are reasonable
	first := ohlcv.At(0)
	if first.Open <= 0 || first.Close <= 0 {
		t.Error("first bar has invalid prices")
	}

	last := ohlcv.At(ohlcv.Len() - 1)
	if last.Open <= 0 || last.Close <= 0 {
		t.Error("last bar has invalid prices")
	}

	t.Logf("Loaded GOOG: %d bars, %s to %s",
		ohlcv.Len(),
		first.Time.Format("2006-01-02"),
		last.Time.Format("2006-01-02"))
}

func TestLoadCSV_EURUSD(t *testing.T) {
	path := getTestDataPath("EURUSD.csv")
	ohlcv, err := LoadCSV(path)
	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}

	if ohlcv.Len() == 0 {
		t.Error("loaded OHLCV is empty")
	}

	if err := ohlcv.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	t.Logf("Loaded EURUSD: %d bars", ohlcv.Len())
}

func TestLoadCSV_BTCUSD(t *testing.T) {
	path := getTestDataPath("BTCUSD.csv")
	ohlcv, err := LoadCSV(path)
	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}

	if ohlcv.Len() == 0 {
		t.Error("loaded OHLCV is empty")
	}

	if err := ohlcv.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	t.Logf("Loaded BTCUSD: %d bars", ohlcv.Len())
}

func TestLoadCSV_NonexistentFile(t *testing.T) {
	_, err := LoadCSV("nonexistent_file.csv")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadCSV_EmptyFile(t *testing.T) {
	// Create a temporary empty file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.csv")
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := LoadCSV(tmpFile)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestLoadCSV_HeaderOnly(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "header_only.csv")
	content := "Date,Open,High,Low,Close,Volume\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ohlcv, err := LoadCSV(tmpFile)
	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}

	if ohlcv.Len() != 0 {
		t.Errorf("expected 0 rows, got %d", ohlcv.Len())
	}
}

func TestLoadCSV_CustomColumns(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "custom.csv")
	content := `date_col,open_price,high_price,low_price,close_price,vol
2020-01-01,100,105,95,102,1000000
2020-01-02,102,108,100,106,1100000
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ohlcv, err := LoadCSV(tmpFile,
		WithTimeColumn("date_col"),
		WithOpenColumn("open_price"),
		WithHighColumn("high_price"),
		WithLowColumn("low_price"),
		WithCloseColumn("close_price"),
		WithVolumeColumn("vol"),
	)

	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}

	if ohlcv.Len() != 2 {
		t.Errorf("expected 2 rows, got %d", ohlcv.Len())
	}

	if ohlcv.Open[0] != 100 {
		t.Errorf("Open[0] = %f, want 100", ohlcv.Open[0])
	}
}

func TestLoadCSV_DifferentDateFormats(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "ISO date",
			content: `Date,Open,High,Low,Close,Volume
2020-01-15,100,105,95,102,1000000
`,
		},
		{
			name: "US date",
			content: `Date,Open,High,Low,Close,Volume
01/15/2020,100,105,95,102,1000000
`,
		},
		{
			name: "datetime",
			content: `Date,Open,High,Low,Close,Volume
2020-01-15 09:30:00,100,105,95,102,1000000
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "data.csv")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			ohlcv, err := LoadCSV(tmpFile)
			if err != nil {
				t.Fatalf("LoadCSV() error = %v", err)
			}

			if ohlcv.Len() != 1 {
				t.Errorf("expected 1 row, got %d", ohlcv.Len())
			}

			// Verify the date was parsed to January 15, 2020
			if ohlcv.Time[0].Year() != 2020 || ohlcv.Time[0].Month() != 1 || ohlcv.Time[0].Day() != 15 {
				t.Errorf("date parsed incorrectly: %v", ohlcv.Time[0])
			}
		})
	}
}

func TestLoadCSV_MissingColumns(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "missing.csv")
	content := `Date,Open,High
2020-01-01,100,105
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err := LoadCSV(tmpFile)
	if err == nil {
		t.Error("expected error for missing columns")
	}
}

func TestLoadCSV_NoVolumeColumn(t *testing.T) {
	// Volume should be optional
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "no_volume.csv")
	content := `Date,Open,High,Low,Close
2020-01-01,100,105,95,102
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ohlcv, err := LoadCSV(tmpFile)
	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}

	if ohlcv.Len() != 1 {
		t.Errorf("expected 1 row, got %d", ohlcv.Len())
	}

	// Volume should be 0 when not provided
	if ohlcv.Volume[0] != 0 {
		t.Errorf("Volume[0] = %f, want 0", ohlcv.Volume[0])
	}
}
