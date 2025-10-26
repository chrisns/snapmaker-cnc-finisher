package cli_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/cli"
)

func TestProgressTrackerUpdate(t *testing.T) {
	tests := []struct {
		name               string
		totalLines         int
		currentLine        int
		linesRemoved       int
		elapsedTime        time.Duration
		expectedPercent    float64
		expectUpdate       bool
		lastUpdateLine     int
		lastUpdateTime     time.Time
		timeSinceLastCheck time.Duration
	}{
		{
			name:               "Update at 10k boundary",
			totalLines:         100000,
			currentLine:        10000,
			linesRemoved:       2500,
			elapsedTime:        2 * time.Second,
			expectedPercent:    10.0,
			expectUpdate:       true,
			lastUpdateLine:     0,
			lastUpdateTime:     time.Now().Add(-2 * time.Second),
			timeSinceLastCheck: 2 * time.Second,
		},
		{
			name:               "Update at 2 second interval",
			totalLines:         100000,
			currentLine:        5000,
			linesRemoved:       1250,
			elapsedTime:        2100 * time.Millisecond,
			expectedPercent:    5.0,
			expectUpdate:       true,
			lastUpdateLine:     0,
			lastUpdateTime:     time.Now().Add(-2100 * time.Millisecond),
			timeSinceLastCheck: 2100 * time.Millisecond,
		},
		{
			name:               "No update before 10k lines or 2 seconds",
			totalLines:         100000,
			currentLine:        5000,
			linesRemoved:       1000,
			elapsedTime:        1 * time.Second,
			expectedPercent:    5.0,
			expectUpdate:       false,
			lastUpdateLine:     0,
			lastUpdateTime:     time.Now().Add(-1 * time.Second),
			timeSinceLastCheck: 1 * time.Second,
		},
		{
			name:               "Update at 50% completion",
			totalLines:         200000,
			currentLine:        100000,
			linesRemoved:       25000,
			elapsedTime:        10 * time.Second,
			expectedPercent:    50.0,
			expectUpdate:       true,
			lastUpdateLine:     90000,
			lastUpdateTime:     time.Now().Add(-3 * time.Second),
			timeSinceLastCheck: 3 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := cli.NewProgressTracker(tt.totalLines)
			tracker.Update(tt.currentLine, tt.linesRemoved)

			shouldUpdate := tracker.ShouldUpdate(tt.lastUpdateLine, tt.timeSinceLastCheck)
			if shouldUpdate != tt.expectUpdate {
				t.Errorf("ShouldUpdate() = %v, want %v", shouldUpdate, tt.expectUpdate)
			}

			percent := tracker.PercentComplete()
			if percent != tt.expectedPercent {
				t.Errorf("PercentComplete() = %.2f, want %.2f", percent, tt.expectedPercent)
			}
		})
	}
}

func TestProgressDisplayFormat(t *testing.T) {
	tests := []struct {
		name         string
		totalLines   int
		currentLine  int
		linesRemoved int
		elapsedTime  time.Duration
		wantContains []string
	}{
		{
			name:         "Standard progress output",
			totalLines:   100000,
			currentLine:  45230,
			linesRemoved: 12450,
			elapsedTime:  3200 * time.Millisecond,
			wantContains: []string{
				"45,230",
				"100,000",
				"45.2%",
				"12,450",
				"3.2s",
				"lines/s", // Throughput
				"Speed:",  // Throughput label
			},
		},
		{
			name:         "High progress near completion",
			totalLines:   50000,
			currentLine:  48000,
			linesRemoved: 15000,
			elapsedTime:  5500 * time.Millisecond,
			wantContains: []string{
				"48,000",
				"50,000",
				"96.0%",
				"15,000",
				"5.5s",
				"lines/s", // Throughput
			},
		},
		{
			name:         "Early progress",
			totalLines:   1000000,
			currentLine:  10000,
			linesRemoved: 3000,
			elapsedTime:  1000 * time.Millisecond,
			wantContains: []string{
				"10,000",
				"1,000,000",
				"1.0%",
				"3,000",
				"1.0s",
				"lines/s", // Throughput
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tracker := cli.NewProgressTracker(tt.totalLines)
			tracker.Update(tt.currentLine, tt.linesRemoved)

			// Test display without newline (uses \r for overwrite)
			tracker.Display(&buf, tt.elapsedTime)

			output := buf.String()
			for _, want := range tt.wantContains {
				if !bytes.Contains([]byte(output), []byte(want)) {
					t.Errorf("Display() output missing %q\nGot: %s", want, output)
				}
			}

			// Verify uses \r for single-line overwrite
			if !bytes.HasPrefix([]byte(output), []byte("\r")) {
				t.Error("Display() should start with \\r for line overwrite")
			}
		})
	}
}

func TestEstimatedTimeRemaining(t *testing.T) {
	tests := []struct {
		name        string
		totalLines  int
		currentLine int
		elapsed     time.Duration
		wantETA     time.Duration
		tolerance   time.Duration
	}{
		{
			name:        "Halfway through, linear projection",
			totalLines:  100000,
			currentLine: 50000,
			elapsed:     5 * time.Second,
			wantETA:     5 * time.Second,
			tolerance:   100 * time.Millisecond,
		},
		{
			name:        "25% complete",
			totalLines:  200000,
			currentLine: 50000,
			elapsed:     2 * time.Second,
			wantETA:     6 * time.Second,
			tolerance:   100 * time.Millisecond,
		},
		{
			name:        "95% complete, nearly done",
			totalLines:  100000,
			currentLine: 95000,
			elapsed:     10 * time.Second,
			wantETA:     526 * time.Millisecond, // ~0.5s
			tolerance:   100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := cli.NewProgressTracker(tt.totalLines)
			tracker.Update(tt.currentLine, 0)

			eta := tracker.EstimatedTimeRemaining(tt.elapsed)
			diff := eta - tt.wantETA
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("EstimatedTimeRemaining() = %v, want %v (±%v)", eta, tt.wantETA, tt.tolerance)
			}
		})
	}
}
