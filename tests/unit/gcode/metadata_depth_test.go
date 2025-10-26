package gcode_test

import (
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

func TestGetZReference(t *testing.T) {
	tests := []struct {
		name string
		meta gcode.Metadata
		want float64
	}{
		{
			name: "ZRefMetadata uses MaxZ as reference",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: 0.0,
		},
		{
			name: "ZRefMetadata with positive MaxZ",
			meta: gcode.Metadata{
				MinZ:       -2.0,
				MaxZ:       1.5,
				ZReference: gcode.ZRefMetadata,
			},
			want: 1.5,
		},
		{
			name: "ZRefMetadata with negative MaxZ",
			meta: gcode.Metadata{
				MinZ:       -10.0,
				MaxZ:       -2.0,
				ZReference: gcode.ZRefMetadata,
			},
			want: -2.0,
		},
		{
			name: "ZRefMachineOrigin uses zero",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMachineOrigin,
			},
			want: 0.0,
		},
		{
			name: "ZRefSurface uses zero",
			meta: gcode.Metadata{
				ZReference: gcode.ZRefSurface,
			},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.meta.GetZReference(); got != tt.want {
				t.Errorf("GetZReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsShallowDepth(t *testing.T) {
	tests := []struct {
		name      string
		meta      gcode.Metadata
		z         float64
		allowance float64
		want      bool
	}{
		{
			name: "Z above threshold is shallow (ZRefMetadata)",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         -0.5, // Above threshold of -0.8 (0 - 0.8)
			allowance: 0.8,
			want:      true,
		},
		{
			name: "Z below threshold is not shallow (ZRefMetadata)",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         -1.5, // Below threshold of -0.8 (0 - 0.8)
			allowance: 0.8,
			want:      false,
		},
		{
			name: "Z exactly at threshold is not shallow",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         -0.8, // Exactly at threshold
			allowance: 0.8,
			want:      false, // Not greater than threshold
		},
		{
			name: "Shallow depth with machine origin reference",
			meta: gcode.Metadata{
				MinZ:       -3.0,
				ZReference: gcode.ZRefMachineOrigin,
			},
			z:         -0.3, // Above threshold of -0.5 (0 - 0.5)
			allowance: 0.5,
			want:      true,
		},
		{
			name: "Deep cut with machine origin reference",
			meta: gcode.Metadata{
				MinZ:       -3.0,
				ZReference: gcode.ZRefMachineOrigin,
			},
			z:         -2.0, // Below threshold of -0.5 (0 - 0.5)
			allowance: 0.5,
			want:      false,
		},
		{
			name: "Shallow depth with surface convention",
			meta: gcode.Metadata{
				ZReference: gcode.ZRefSurface,
			},
			z:         -0.2, // Above threshold of -1.0 (0 - 1.0)
			allowance: 1.0,
			want:      true,
		},
		{
			name: "Deep cut with surface convention",
			meta: gcode.Metadata{
				ZReference: gcode.ZRefSurface,
			},
			z:         -2.5, // Below threshold of -1.0 (0 - 1.0)
			allowance: 1.0,
			want:      false,
		},
		{
			name: "Zero allowance - only positive Z is shallow",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         0.1, // Above threshold of 0 (0 - 0)
			allowance: 0.0,
			want:      true,
		},
		{
			name: "Zero allowance - zero Z is not shallow",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         0.0, // Exactly at threshold
			allowance: 0.0,
			want:      false,
		},
		{
			name: "Positive MaxZ reference with allowance",
			meta: gcode.Metadata{
				MinZ:       -3.0,
				MaxZ:       2.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         1.5, // Above threshold of 1.0 (2.0 - 1.0)
			allowance: 1.0,
			want:      true,
		},
		{
			name: "Positive MaxZ reference below allowance",
			meta: gcode.Metadata{
				MinZ:       -3.0,
				MaxZ:       2.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         0.5, // Below threshold of 1.0 (2.0 - 1.0)
			allowance: 1.0,
			want:      false,
		},
		{
			name: "Boundary case - very small allowance",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         -0.001, // Above threshold of -0.01 (0 - 0.01)
			allowance: 0.01,
			want:      true,
		},
		{
			name: "Boundary case - large allowance covers entire range",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         -3.0, // Above threshold of -10.0 (0 - 10.0)
			allowance: 10.0,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.meta.IsShallowDepth(tt.z, tt.allowance); got != tt.want {
				threshold := tt.meta.GetZReference() - tt.allowance
				t.Errorf("IsShallowDepth(%v, %v) = %v, want %v (threshold: %v, ref: %v)",
					tt.z, tt.allowance, got, tt.want, threshold, tt.meta.GetZReference())
			}
		})
	}
}

func TestIsShallowDepthEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		meta      gcode.Metadata
		z         float64
		allowance float64
		want      bool
	}{
		{
			name: "Negative allowance (edge case)",
			meta: gcode.Metadata{
				MinZ:       -5.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         0.5, // Above threshold of 1.0 (0 - (-1.0))
			allowance: -1.0,
			want:      false, // 0.5 is not > 1.0
		},
		{
			name: "Very large Z value",
			meta: gcode.Metadata{
				MinZ:       -100.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         50.0, // Well above threshold
			allowance: 1.0,
			want:      true,
		},
		{
			name: "Very small Z value",
			meta: gcode.Metadata{
				MinZ:       -100.0,
				MaxZ:       0.0,
				ZReference: gcode.ZRefMetadata,
			},
			z:         -99.0, // Well below threshold of -1.0 (0 - 1.0)
			allowance: 1.0,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.meta.IsShallowDepth(tt.z, tt.allowance); got != tt.want {
				t.Errorf("IsShallowDepth(%v, %v) = %v, want %v",
					tt.z, tt.allowance, got, tt.want)
			}
		})
	}
}
