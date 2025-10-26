package optimizer_test

import (
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
)

func TestShouldFilterByDepth(t *testing.T) {
	tests := []struct {
		name      string
		z         float64
		allowance float64
		meta      *gcode.Metadata
		want      bool
	}{
		{
			name:      "Shallow cut within allowance - should filter",
			z:         -0.5,
			allowance: 1.0,
			meta: &gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: true,
		},
		{
			name:      "Deep cut beyond allowance - should keep",
			z:         -1.5,
			allowance: 1.0,
			meta: &gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: false,
		},
		{
			name:      "Exactly at threshold - should keep (not greater than)",
			z:         -1.0,
			allowance: 1.0,
			meta: &gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: false,
		},
		{
			name:      "Very shallow - should filter",
			z:         -0.1,
			allowance: 1.0,
			meta: &gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: true,
		},
		{
			name:      "Zero allowance - only positive Z filtered",
			z:         -0.1,
			allowance: 0.0,
			meta: &gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: false,
		},
		{
			name:      "Positive MaxZ reference",
			z:         1.5,
			allowance: 1.0,
			meta: &gcode.Metadata{
				MinZ:       -3.0,
				MaxZ:       2.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: true, // 1.5 > (2.0 - 1.0) = 1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optimizer.ShouldFilterByDepth(tt.z, tt.allowance, tt.meta)
			if got != tt.want {
				t.Errorf("ShouldFilterByDepth(%v, %v, meta) = %v, want %v",
					tt.z, tt.allowance, got, tt.want)
			}
		})
	}
}

func TestPreserveRapidAndMachineCodes(t *testing.T) {
	tests := []struct {
		name string
		cmd  gcode.Command
		want bool // true if command should be preserved
	}{
		{
			name: "G0 rapid move - always preserve",
			cmd: gcode.Command{
				Letter: "G",
				Value:  0,
				Params: map[string]float64{"X": 10.0, "Y": 20.0, "Z": -0.5},
			},
			want: true,
		},
		{
			name: "G0 shallow rapid - preserve even if shallow",
			cmd: gcode.Command{
				Letter: "G",
				Value:  0,
				Params: map[string]float64{"Z": -0.2},
			},
			want: true,
		},
		{
			name: "M3 spindle on - always preserve",
			cmd: gcode.Command{
				Letter: "M",
				Value:  3,
				Params: map[string]float64{"S": 1000.0},
			},
			want: true,
		},
		{
			name: "M5 spindle off - always preserve",
			cmd: gcode.Command{
				Letter: "M",
				Value:  5,
			},
			want: true,
		},
		{
			name: "M2 end program - always preserve",
			cmd: gcode.Command{
				Letter: "M",
				Value:  2,
			},
			want: true,
		},
		{
			name: "G1 cutting move - depends on depth (not always preserved)",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Z": -0.5, "F": 1000.0},
			},
			want: false,
		},
		{
			name: "Comment line - should preserve",
			cmd: gcode.Command{
				Comment: "; This is a comment",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optimizer.ShouldAlwaysPreserve(tt.cmd)
			if got != tt.want {
				t.Errorf("ShouldAlwaysPreserve(%+v) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestFeedRatePreservation(t *testing.T) {
	tests := []struct {
		name     string
		cmd      gcode.Command
		wantFeed float64
	}{
		{
			name: "G1 cutting move with F1000",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Y": 20.0, "Z": -2.0, "F": 1000.0},
			},
			wantFeed: 1000.0,
		},
		{
			name: "G1 cutting move with F1500",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 5.0, "Z": -1.5, "F": 1500.0},
			},
			wantFeed: 1500.0,
		},
		{
			name: "G1 cutting move with F500",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"Y": 30.0, "Z": -3.0, "F": 500.0},
			},
			wantFeed: 500.0,
		},
		{
			name: "G1 cutting move with high speed F3000",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 100.0, "F": 3000.0},
			},
			wantFeed: 3000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the command preserves its feed rate parameter
			if feed, ok := tt.cmd.Params["F"]; !ok {
				t.Error("Command should have feed rate parameter 'F'")
			} else if feed != tt.wantFeed {
				t.Errorf("Feed rate = %v, want %v", feed, tt.wantFeed)
			}

			// Verify that feed rate is part of the command params and not lost during filtering decisions
			// This ensures that when a move is kept (not filtered), it retains its feed rate
			meta := &gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			}

			// These moves should NOT be filtered (deep cuts), so feed rate should be preserved
			shouldFilter := optimizer.ShouldFilterMove(tt.cmd, 1.0, meta, optimizer.StrategySafe)
			if shouldFilter {
				t.Error("Deep cut should not be filtered - feed rate preservation test invalid")
			}

			// Verify the params map still contains the feed rate after filtering decision
			if feed, ok := tt.cmd.Params["F"]; !ok || feed != tt.wantFeed {
				t.Errorf("Feed rate not preserved after filtering decision: got %v, want %v", feed, tt.wantFeed)
			}
		})
	}
}

func TestShouldFilterMultiAxisMove(t *testing.T) {
	meta := &gcode.Metadata{
		MinZ:       -5.0,
		MaxZ:       0.0,
		ZReference: gcode.ZRefMetadata,
	}

	tests := []struct {
		name      string
		cmd       gcode.Command
		allowance float64
		strategy  optimizer.FilterStrategy
		want      bool
	}{
		{
			name: "Safe strategy: shallow Z with XY motion - keep (multi-axis)",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Y": 20.0, "Z": -0.5, "F": 1000.0},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
			want:      false, // Keep because it's multi-axis
		},
		{
			name: "Safe strategy: shallow Z-only move - filter",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"Z": -0.5, "F": 1000.0},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
			want:      true, // Filter because Z-only and shallow
		},
		{
			name: "Safe strategy: deep Z with XY - keep (deep)",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Z": -2.0, "F": 1000.0},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
			want:      false, // Keep because deep
		},
		{
			name: "AllAxes strategy: shallow Z with XY - filter",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Y": 20.0, "Z": -0.5, "F": 1000.0},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategyAllAxes,
			want:      true, // Filter based on Z depth regardless of other axes
		},
		{
			name: "AllAxes strategy: deep Z with XY - keep",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Z": -2.0, "F": 1000.0},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategyAllAxes,
			want:      false, // Keep because deep
		},
		{
			name: "Aggressive strategy: shallow multi-axis - filter",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Y": 20.0, "B": 45.0, "Z": -0.3, "F": 1000.0},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategyAggressive,
			want:      true, // Filter aggressively
		},
		{
			name: "Safe strategy: 4-axis shallow - keep",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "B": 30.0, "Z": -0.4, "F": 1000.0},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
			want:      false, // Keep multi-axis move in safe mode
		},
		{
			name: "Not a cutting move - don't filter",
			cmd: gcode.Command{
				Letter: "G",
				Value:  0,
				Params: map[string]float64{"X": 10.0, "Z": -0.5},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
			want:      false, // G0 never filtered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optimizer.ShouldFilterMove(tt.cmd, tt.allowance, meta, tt.strategy)
			if got != tt.want {
				t.Errorf("ShouldFilterMove() = %v, want %v", got, tt.want)
			}
		})
	}
}
