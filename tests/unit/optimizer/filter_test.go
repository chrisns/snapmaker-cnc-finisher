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
