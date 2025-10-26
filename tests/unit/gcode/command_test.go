package gcode_test

import (
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    gcode.Command
		wantErr bool
	}{
		{
			name:  "G1 command with coordinates and feed rate",
			input: "G1 X10.5 Y20.3 Z-1.2 F1500",
			want: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{
					"X": 10.5,
					"Y": 20.3,
					"Z": -1.2,
					"F": 1500,
				},
			},
			wantErr: false,
		},
		{
			name:  "G0 rapid move",
			input: "G0 X5.0 Y10.0",
			want: gcode.Command{
				Letter: "G",
				Value:  0,
				Params: map[string]float64{
					"X": 5.0,
					"Y": 10.0,
				},
			},
			wantErr: false,
		},
		{
			name:  "M-code",
			input: "M3 S1000",
			want: gcode.Command{
				Letter: "M",
				Value:  3,
				Params: map[string]float64{
					"S": 1000,
				},
			},
			wantErr: false,
		},
		{
			name:  "Comment line",
			input: "; This is a comment",
			want: gcode.Command{
				Comment: "; This is a comment",
			},
			wantErr: false,
		},
		{
			name:    "Empty line",
			input:   "",
			want:    gcode.Command{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gcode.ParseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !commandsEqual(got, tt.want) {
				t.Errorf("ParseCommand() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
