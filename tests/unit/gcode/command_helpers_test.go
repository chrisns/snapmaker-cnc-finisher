package gcode_test

import (
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

func TestIsRapidMove(t *testing.T) {
	tests := []struct {
		name string
		cmd  gcode.Command
		want bool
	}{
		{
			name: "G0 is rapid move",
			cmd:  gcode.Command{Letter: "G", Value: 0},
			want: true,
		},
		{
			name: "G1 is not rapid move",
			cmd:  gcode.Command{Letter: "G", Value: 1},
			want: false,
		},
		{
			name: "M3 is not rapid move",
			cmd:  gcode.Command{Letter: "M", Value: 3},
			want: false,
		},
		{
			name: "Empty command is not rapid move",
			cmd:  gcode.Command{},
			want: false,
		},
		{
			name: "G0 with parameters is still rapid move",
			cmd: gcode.Command{
				Letter: "G",
				Value:  0,
				Params: map[string]float64{"X": 10.0, "Y": 20.0},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.IsRapidMove(); got != tt.want {
				t.Errorf("IsRapidMove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCuttingMove(t *testing.T) {
	tests := []struct {
		name string
		cmd  gcode.Command
		want bool
	}{
		{
			name: "G1 with feed rate is cutting move",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "F": 1000.0},
			},
			want: true,
		},
		{
			name: "G1 with parameters but no feed rate is cutting move",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"Z": -1.0},
			},
			want: true,
		},
		{
			name: "G1 with no parameters is not cutting move",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{},
			},
			want: false,
		},
		{
			name: "G0 is not cutting move",
			cmd: gcode.Command{
				Letter: "G",
				Value:  0,
				Params: map[string]float64{"X": 10.0, "F": 1000.0},
			},
			want: false,
		},
		{
			name: "M3 is not cutting move",
			cmd:  gcode.Command{Letter: "M", Value: 3},
			want: false,
		},
		{
			name: "Empty command is not cutting move",
			cmd:  gcode.Command{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.IsCuttingMove(); got != tt.want {
				t.Errorf("IsCuttingMove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMachineCode(t *testing.T) {
	tests := []struct {
		name string
		cmd  gcode.Command
		want bool
	}{
		{
			name: "M3 is machine code",
			cmd:  gcode.Command{Letter: "M", Value: 3},
			want: true,
		},
		{
			name: "M5 is machine code",
			cmd:  gcode.Command{Letter: "M", Value: 5},
			want: true,
		},
		{
			name: "G0 is not machine code",
			cmd:  gcode.Command{Letter: "G", Value: 0},
			want: false,
		},
		{
			name: "G1 is not machine code",
			cmd:  gcode.Command{Letter: "G", Value: 1},
			want: false,
		},
		{
			name: "Empty command is not machine code",
			cmd:  gcode.Command{},
			want: false,
		},
		{
			name: "M-code with parameters is machine code",
			cmd: gcode.Command{
				Letter: "M",
				Value:  3,
				Params: map[string]float64{"S": 1000.0},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.IsMachineCode(); got != tt.want {
				t.Errorf("IsMachineCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsComment(t *testing.T) {
	tests := []struct {
		name string
		cmd  gcode.Command
		want bool
	}{
		{
			name: "Pure comment line is comment",
			cmd:  gcode.Command{Comment: "; This is a comment"},
			want: true,
		},
		{
			name: "Empty comment string is not comment",
			cmd:  gcode.Command{Comment: ""},
			want: false,
		},
		{
			name: "Command with inline comment is not pure comment",
			cmd: gcode.Command{
				Letter:  "G",
				Value:   1,
				Comment: "; Move",
			},
			want: false,
		},
		{
			name: "Empty command is not comment",
			cmd:  gcode.Command{},
			want: false,
		},
		{
			name: "G-code without comment is not comment",
			cmd:  gcode.Command{Letter: "G", Value: 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.IsComment(); got != tt.want {
				t.Errorf("IsComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasParam(t *testing.T) {
	tests := []struct {
		name  string
		cmd   gcode.Command
		param string
		want  bool
	}{
		{
			name: "Command has X parameter",
			cmd: gcode.Command{
				Params: map[string]float64{"X": 10.0, "Y": 20.0},
			},
			param: "X",
			want:  true,
		},
		{
			name: "Command has Y parameter",
			cmd: gcode.Command{
				Params: map[string]float64{"X": 10.0, "Y": 20.0},
			},
			param: "Y",
			want:  true,
		},
		{
			name: "Command does not have Z parameter",
			cmd: gcode.Command{
				Params: map[string]float64{"X": 10.0, "Y": 20.0},
			},
			param: "Z",
			want:  false,
		},
		{
			name:  "Empty params map",
			cmd:   gcode.Command{Params: map[string]float64{}},
			param: "X",
			want:  false,
		},
		{
			name: "Parameter with zero value exists",
			cmd: gcode.Command{
				Params: map[string]float64{"Z": 0.0},
			},
			param: "Z",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.HasParam(tt.param); got != tt.want {
				t.Errorf("HasParam(%q) = %v, want %v", tt.param, got, tt.want)
			}
		})
	}
}

func TestGetParam(t *testing.T) {
	tests := []struct {
		name  string
		cmd   gcode.Command
		param string
		want  float64
	}{
		{
			name: "Get existing X parameter",
			cmd: gcode.Command{
				Params: map[string]float64{"X": 10.5, "Y": 20.3},
			},
			param: "X",
			want:  10.5,
		},
		{
			name: "Get existing Y parameter",
			cmd: gcode.Command{
				Params: map[string]float64{"X": 10.5, "Y": 20.3},
			},
			param: "Y",
			want:  20.3,
		},
		{
			name: "Get non-existing parameter returns zero",
			cmd: gcode.Command{
				Params: map[string]float64{"X": 10.5},
			},
			param: "Z",
			want:  0.0,
		},
		{
			name:  "Get from empty params returns zero",
			cmd:   gcode.Command{Params: map[string]float64{}},
			param: "X",
			want:  0.0,
		},
		{
			name: "Get parameter with zero value",
			cmd: gcode.Command{
				Params: map[string]float64{"Z": 0.0},
			},
			param: "Z",
			want:  0.0,
		},
		{
			name: "Get negative parameter value",
			cmd: gcode.Command{
				Params: map[string]float64{"Z": -1.5},
			},
			param: "Z",
			want:  -1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.GetParam(tt.param); got != tt.want {
				t.Errorf("GetParam(%q) = %v, want %v", tt.param, got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		cmd  gcode.Command
		want string
	}{
		{
			name: "G0 rapid move",
			cmd: gcode.Command{
				Letter: "G",
				Value:  0,
				Params: map[string]float64{"X": 10.0, "Y": 20.0},
			},
			want: "G0 X10 Y20",
		},
		{
			name: "G1 with feed rate",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.5, "Y": 20.3, "Z": -1.2, "F": 1500.0},
			},
			want: "G1 X10.5 Y20.3 Z-1.2 F1500",
		},
		{
			name: "M-code with spindle speed",
			cmd: gcode.Command{
				Letter: "M",
				Value:  3,
				Params: map[string]float64{"S": 1000.0},
			},
			want: "M3 S1000",
		},
		{
			name: "Comment line",
			cmd: gcode.Command{
				Comment: "; This is a comment",
			},
			want: "; This is a comment",
		},
		{
			name: "Empty command",
			cmd:  gcode.Command{},
			want: "",
		},
		{
			name: "Command with inline comment",
			cmd: gcode.Command{
				Letter:  "G",
				Value:   0,
				Params:  map[string]float64{"X": 0.0},
				Comment: "; Move to origin",
			},
			want: "G0 X0 ; Move to origin",
		},
		{
			name: "4-axis command with B-axis",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1,
				Params: map[string]float64{"X": 10.0, "Y": 20.0, "B": 45.0, "F": 1000.0},
			},
			want: "G1 X10 Y20 B45 F1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
