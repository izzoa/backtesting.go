package data

import (
	"testing"
	"time"
)

func TestNewOHLCV(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
	}{
		{"zero capacity", 0},
		{"small capacity", 10},
		{"large capacity", 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ohlcv := NewOHLCV(tt.capacity)
			if ohlcv.Len() != 0 {
				t.Errorf("Len() = %d, want 0", ohlcv.Len())
			}
		})
	}
}

func TestOHLCV_Append(t *testing.T) {
	ohlcv := NewOHLCV(10)

	bar := Bar{
		Time:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Open:   100,
		High:   105,
		Low:    95,
		Close:  102,
		Volume: 1000000,
	}

	ohlcv.Append(bar)

	if ohlcv.Len() != 1 {
		t.Errorf("Len() = %d, want 1", ohlcv.Len())
	}

	got := ohlcv.At(0)
	if got != bar {
		t.Errorf("At(0) = %+v, want %+v", got, bar)
	}
}

func TestOHLCV_At(t *testing.T) {
	ohlcv := NewOHLCV(3)
	for i := 0; i < 3; i++ {
		ohlcv.Append(Bar{
			Time:   time.Date(2020, 1, i+1, 0, 0, 0, 0, time.UTC),
			Open:   float64(100 + i),
			High:   float64(105 + i),
			Low:    float64(95 + i),
			Close:  float64(102 + i),
			Volume: float64(1000000 + i*100000),
		})
	}

	t.Run("valid indices", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			bar := ohlcv.At(i)
			if bar.Open != float64(100+i) {
				t.Errorf("At(%d).Open = %f, want %f", i, bar.Open, float64(100+i))
			}
		}
	})

	t.Run("negative index panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for negative index")
			}
		}()
		ohlcv.At(-1)
	})

	t.Run("out of bounds panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for out of bounds index")
			}
		}()
		ohlcv.At(10)
	})
}

func TestOHLCV_Slice(t *testing.T) {
	ohlcv := makeTestOHLCV(100)

	tests := []struct {
		name      string
		start     int
		end       int
		wantLen   int
		wantPanic bool
	}{
		{"full slice", 0, 100, 100, false},
		{"partial start", 10, 100, 90, false},
		{"partial end", 0, 50, 50, false},
		{"middle slice", 25, 75, 50, false},
		{"single element", 50, 51, 1, false},
		{"empty slice", 50, 50, 0, false},
		{"negative start", -1, 50, 0, true},
		{"end > length", 0, 150, 0, true},
		{"start > end", 50, 25, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic")
					}
				}()
			}
			slice := ohlcv.Slice(tt.start, tt.end)
			if !tt.wantPanic && slice.Len() != tt.wantLen {
				t.Errorf("Slice(%d, %d).Len() = %d, want %d",
					tt.start, tt.end, slice.Len(), tt.wantLen)
			}
		})
	}
}

func TestOHLCV_LastN(t *testing.T) {
	ohlcv := makeTestOHLCV(100)

	tests := []struct {
		name    string
		n       int
		wantLen int
	}{
		{"last 10", 10, 10},
		{"last 50", 50, 50},
		{"last 100", 100, 100},
		{"more than length", 150, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			last := ohlcv.LastN(tt.n)
			if last.Len() != tt.wantLen {
				t.Errorf("LastN(%d).Len() = %d, want %d", tt.n, last.Len(), tt.wantLen)
			}
		})
	}
}

func TestOHLCV_Copy(t *testing.T) {
	original := makeTestOHLCV(50)
	copy := original.Copy()

	// Verify same content
	if copy.Len() != original.Len() {
		t.Errorf("copy.Len() = %d, want %d", copy.Len(), original.Len())
	}

	for i := 0; i < original.Len(); i++ {
		if copy.At(i) != original.At(i) {
			t.Errorf("copy.At(%d) differs from original", i)
		}
	}

	// Verify independence (modifying copy doesn't affect original)
	origClose0 := original.Close[0]
	copy.Close[0] = 999999
	if original.Close[0] != origClose0 {
		t.Error("modifying copy affected original")
	}
}

func TestOHLCV_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ohlcv   *OHLCV
		wantErr bool
	}{
		{
			name:    "valid data",
			ohlcv:   makeTestOHLCV(10),
			wantErr: false,
		},
		{
			name: "high < low",
			ohlcv: &OHLCV{
				Time:   []time.Time{time.Now()},
				Open:   []float64{100},
				High:   []float64{95}, // Invalid: high < low
				Low:    []float64{105},
				Close:  []float64{100},
				Volume: []float64{1000},
			},
			wantErr: true,
		},
		{
			name: "high < open",
			ohlcv: &OHLCV{
				Time:   []time.Time{time.Now()},
				Open:   []float64{110},
				High:   []float64{105}, // Invalid: high < open
				Low:    []float64{95},
				Close:  []float64{100},
				Volume: []float64{1000},
			},
			wantErr: true,
		},
		{
			name: "low > close",
			ohlcv: &OHLCV{
				Time:   []time.Time{time.Now()},
				Open:   []float64{100},
				High:   []float64{110},
				Low:    []float64{105}, // Invalid: low > close
				Close:  []float64{100},
				Volume: []float64{1000},
			},
			wantErr: true,
		},
		{
			name: "negative volume",
			ohlcv: &OHLCV{
				Time:   []time.Time{time.Now()},
				Open:   []float64{100},
				High:   []float64{105},
				Low:    []float64{95},
				Close:  []float64{100},
				Volume: []float64{-1000}, // Invalid: negative volume
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ohlcv.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOHLCV_String(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		ohlcv := NewOHLCV(0)
		s := ohlcv.String()
		if s != "OHLCV(empty)" {
			t.Errorf("String() = %q, want %q", s, "OHLCV(empty)")
		}
	})

	t.Run("with data", func(t *testing.T) {
		ohlcv := makeTestOHLCV(100)
		s := ohlcv.String()
		if s == "" || s == "OHLCV(empty)" {
			t.Errorf("String() = %q, want non-empty string", s)
		}
	})
}

// makeTestOHLCV creates test OHLCV data without depending on testutil.
func makeTestOHLCV(length int) *OHLCV {
	ohlcv := NewOHLCV(length)
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	price := 100.0

	for i := 0; i < length; i++ {
		open := price
		close := price * 1.001
		high := close * 1.002
		low := open * 0.998

		ohlcv.Append(Bar{
			Time:   baseTime.AddDate(0, 0, i),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: float64(1000000 + i*1000),
		})

		price = close
	}

	return ohlcv
}
