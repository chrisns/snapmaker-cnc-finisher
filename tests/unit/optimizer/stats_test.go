package optimizer_test

import (
	"testing"
	"time"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
)

func TestStatisticsTracking(t *testing.T) {
	tests := []struct {
		name            string
		totalLines      int
		removedLines    int
		bytesIn         int64
		bytesOut        int64
		wantReduction   float64
		wantSizeReduced float64
	}{
		{
			name:            "25% line reduction",
			totalLines:      100,
			removedLines:    25,
			bytesIn:         10000,
			bytesOut:        7500,
			wantReduction:   25.0,
			wantSizeReduced: 25.0,
		},
		{
			name:            "50% line reduction",
			totalLines:      200,
			removedLines:    100,
			bytesIn:         20000,
			bytesOut:        10000,
			wantReduction:   50.0,
			wantSizeReduced: 50.0,
		},
		{
			name:            "No lines removed",
			totalLines:      100,
			removedLines:    0,
			bytesIn:         10000,
			bytesOut:        10000,
			wantReduction:   0.0,
			wantSizeReduced: 0.0,
		},
		{
			name:            "All lines removed (edge case)",
			totalLines:      100,
			removedLines:    100,
			bytesIn:         10000,
			bytesOut:        0,
			wantReduction:   100.0,
			wantSizeReduced: 100.0,
		},
		{
			name:            "Low reduction (< 10%)",
			totalLines:      1000,
			removedLines:    50,
			bytesIn:         100000,
			bytesOut:        95000,
			wantReduction:   5.0,
			wantSizeReduced: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := optimizer.NewStatistics()
			stats.TotalLines = tt.totalLines
			stats.RemovedLines = tt.removedLines
			stats.BytesIn = tt.bytesIn
			stats.BytesOut = tt.bytesOut

			if got := stats.LineReductionPercent(); got != tt.wantReduction {
				t.Errorf("LineReductionPercent() = %v, want %v", got, tt.wantReduction)
			}

			if got := stats.FileSizeReductionPercent(); got != tt.wantSizeReduced {
				t.Errorf("FileSizeReductionPercent() = %v, want %v", got, tt.wantSizeReduced)
			}
		})
	}
}

func TestStatisticsKeptLines(t *testing.T) {
	stats := optimizer.NewStatistics()
	stats.TotalLines = 100
	stats.RemovedLines = 25

	want := 75
	if got := stats.KeptLines(); got != want {
		t.Errorf("KeptLines() = %v, want %v", got, want)
	}
}

func TestCalculateTimeSaved(t *testing.T) {
	tests := []struct {
		name       string
		x1, y1, z1 float64 // Start position
		x2, y2, z2 float64 // End position
		feedRate   float64 // Feed rate in mm/min
		want       time.Duration
	}{
		{
			name: "Horizontal move 100mm at 1000mm/min",
			x1:   0, y1: 0, z1: 0,
			x2: 100, y2: 0, z2: 0,
			feedRate: 1000,
			want:     6 * time.Second, // 100mm / 1000mm/min = 0.1 min = 6s
		},
		{
			name: "Vertical move 50mm at 500mm/min",
			x1:   0, y1: 0, z1: 0,
			x2: 0, y2: 0, z2: -50,
			feedRate: 500,
			want:     6 * time.Second, // 50mm / 500mm/min = 0.1 min = 6s
		},
		{
			name: "Diagonal move at 1500mm/min",
			x1:   0, y1: 0, z1: 0,
			x2: 30, y2: 40, z2: 0, // 3-4-5 triangle = 50mm
			feedRate: 1500,
			want:     2 * time.Second, // 50mm / 1500mm/min = 0.0333 min = 2s
		},
		{
			name: "3D diagonal move",
			x1:   0, y1: 0, z1: 0,
			x2: 10, y2: 10, z2: 10, // sqrt(300) ≈ 17.32mm
			feedRate: 1000,
			want:     1040 * time.Millisecond, // ~17.32mm / 1000mm/min ≈ 1.04s
		},
		{
			name: "Zero distance move",
			x1:   10, y1: 20, z1: -5,
			x2: 10, y2: 20, z2: -5,
			feedRate: 1000,
			want:     0,
		},
		{
			name: "Default feed rate (1000mm/min) when zero provided",
			x1:   0, y1: 0, z1: 0,
			x2: 100, y2: 0, z2: 0,
			feedRate: 0, // Should use default 1000mm/min
			want:     6 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optimizer.CalculateTimeSaved(
				tt.x1, tt.y1, tt.z1,
				tt.x2, tt.y2, tt.z2,
				tt.feedRate,
			)

			// Allow 100ms tolerance for floating point calculations
			tolerance := 100 * time.Millisecond
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}

			if diff > tolerance {
				t.Errorf("CalculateTimeSaved() = %v, want %v (tolerance: %v, diff: %v)",
					got, tt.want, tolerance, diff)
			}
		})
	}
}
