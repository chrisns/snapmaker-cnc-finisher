package gcode_test

import (
	"strings"
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

func TestExtractMetadata(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMinZ  float64
		wantMaxZ  float64
		want4Axis bool
		wantRef   gcode.ZReference
	}{
		{
			name: "Valid Snapmaker header with min/max Z",
			input: `;Header Start
;MIN_Z: -5.0
;MAX_Z: 0.0
;Header End
G0 X0 Y0`,
			wantMinZ:  -5.0,
			wantMaxZ:  0.0,
			want4Axis: false,
			wantRef:   gcode.ZRefMetadata,
		},
		{
			name: "4-axis file with B-axis commands",
			input: `;MIN_Z: -3.0
;MAX_Z: 0.5
G0 X0 Y0
G1 X10 Y10 B45.0 F1000`,
			wantMinZ:  -3.0,
			wantMaxZ:  0.5,
			want4Axis: true,
			wantRef:   gcode.ZRefMetadata,
		},
		{
			name: "Incomplete Z metadata - fallback to machine origin",
			input: `;Header Start
;MIN_Z: -2.0
;TOOL: 3mm End Mill
;Header End
G0 X0 Y0`,
			wantMinZ:  -2.0,
			wantMaxZ:  0,
			want4Axis: false,
			wantRef:   gcode.ZRefMachineOrigin,
		},
		{
			name: "No header - fallback to surface convention",
			input: `G0 X0 Y0
G1 Z-1.0 F1000`,
			wantMinZ:  0,
			wantMaxZ:  0,
			want4Axis: false,
			wantRef:   gcode.ZRefSurface,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			meta, err := gcode.ExtractMetadata(r)
			if err != nil {
				t.Fatalf("ExtractMetadata() error = %v", err)
			}

			if meta.MinZ != tt.wantMinZ {
				t.Errorf("MinZ = %v, want %v", meta.MinZ, tt.wantMinZ)
			}
			if meta.MaxZ != tt.wantMaxZ {
				t.Errorf("MaxZ = %v, want %v", meta.MaxZ, tt.wantMaxZ)
			}
			if meta.Is4Axis != tt.want4Axis {
				t.Errorf("Is4Axis = %v, want %v", meta.Is4Axis, tt.want4Axis)
			}
			if meta.ZReference != tt.wantRef {
				t.Errorf("ZReference = %v, want %v", meta.ZReference, tt.wantRef)
			}
		})
	}
}

func TestZReferenceFallbackChain(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRef     gcode.ZReference
		wantMessage string
	}{
		{
			name: "Metadata present - use metadata",
			input: `;MIN_Z: -2.0
;MAX_Z: 0.0
G0 X0`,
			wantRef:     gcode.ZRefMetadata,
			wantMessage: "Using Z-axis reference from GCode header metadata",
		},
		{
			name: "Missing min_z - fallback to machine origin",
			input: `;MAX_Z: 0.0
G0 X0`,
			wantRef:     gcode.ZRefMachineOrigin,
			wantMessage: "Z-axis reference: falling back to machine work origin (metadata incomplete)",
		},
		{
			name:        "No metadata - fallback to surface convention",
			input:       "G0 X0 Y0",
			wantRef:     gcode.ZRefSurface,
			wantMessage: "Z-axis reference: using material surface convention (Z=0 = top surface)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			meta, err := gcode.ExtractMetadata(r)
			if err != nil {
				t.Fatalf("ExtractMetadata() error = %v", err)
			}

			if meta.ZReference != tt.wantRef {
				t.Errorf("ZReference = %v, want %v", meta.ZReference, tt.wantRef)
			}

			if !strings.Contains(meta.ZRefMessage, tt.wantMessage) {
				t.Errorf("ZRefMessage = %q, want to contain %q", meta.ZRefMessage, tt.wantMessage)
			}
		})
	}
}
