package optimizer_test

import (
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
)

// BenchmarkShouldFilterMove tests performance of filtering logic
func BenchmarkShouldFilterMove(b *testing.B) {
	// Setup test data
	metadata := &gcode.Metadata{
		Is4Axis:    false,
		ZReference: gcode.ZRefMetadata,
		MinZ:       -10.0,
		MaxZ:       5.0,
	}

	testCases := []struct {
		name      string
		cmd       gcode.Command
		allowance float64
		strategy  optimizer.FilterStrategy
	}{
		{
			name: "Shallow G1 move (should filter)",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1.0,
				Params: map[string]float64{
					"X": 10.0,
					"Y": 20.0,
					"Z": -0.5,
					"F": 1500.0,
				},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
		},
		{
			name: "Deep G1 move (should keep)",
			cmd: gcode.Command{
				Letter: "G",
				Value:  1.0,
				Params: map[string]float64{
					"X": 10.0,
					"Y": 20.0,
					"Z": -2.5,
					"F": 1500.0,
				},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
		},
		{
			name: "G0 rapid move (should keep)",
			cmd: gcode.Command{
				Letter: "G",
				Value:  0.0,
				Params: map[string]float64{
					"X": 10.0,
					"Y": 20.0,
					"Z": -0.5,
				},
			},
			allowance: 1.0,
			strategy:  optimizer.StrategySafe,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = optimizer.ShouldFilterMove(tc.cmd, tc.allowance, metadata, tc.strategy)
			}
		})
	}
}

// BenchmarkBatchFiltering simulates processing 100k lines
func BenchmarkBatchFiltering(b *testing.B) {
	metadata := &gcode.Metadata{
		Is4Axis:    false,
		ZReference: gcode.ZRefMetadata,
		MinZ:       -10.0,
		MaxZ:       5.0,
	}

	// Create realistic mix of commands
	commands := make([]gcode.Command, 1000)
	for i := 0; i < len(commands); i++ {
		var depth float64
		if i%3 == 0 {
			depth = -0.5 // Shallow (will be filtered)
		} else {
			depth = -2.0 // Deep (will be kept)
		}

		commands[i] = gcode.Command{
			Letter: "G",
			Value:  1.0,
			Params: map[string]float64{
				"X": float64(i % 100),
				"Y": float64(i % 50),
				"Z": depth,
				"F": 1500.0,
			},
		}
	}

	allowance := 1.0
	strategy := optimizer.StrategySafe

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filtered := 0
		for _, cmd := range commands {
			if optimizer.ShouldFilterMove(cmd, allowance, metadata, strategy) {
				filtered++
			}
		}
	}
}

// BenchmarkTimeSavingsCalculation tests performance of time calculations
func BenchmarkTimeSavingsCalculation(b *testing.B) {
	positions := [][6]float64{
		{0, 0, 0, 10, 10, -1},
		{10, 10, -1, 20, 20, -2},
		{20, 20, -2, 30, 30, -3},
		{30, 30, -3, 40, 40, -4},
		{40, 40, -4, 50, 50, -5},
	}

	feedRate := 1500.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pos := range positions {
			_ = optimizer.CalculateTimeSaved(
				pos[0], pos[1], pos[2],
				pos[3], pos[4], pos[5],
				feedRate,
			)
		}
	}
}

// BenchmarkStrategyComparison compares performance of different strategies
func BenchmarkStrategyComparison(b *testing.B) {
	metadata := &gcode.Metadata{
		Is4Axis:    false,
		ZReference: gcode.ZRefMetadata,
		MinZ:       -10.0,
		MaxZ:       5.0,
	}

	// Multi-axis move
	cmd := gcode.Command{
		Letter: "G",
		Value:  1.0,
		Params: map[string]float64{
			"X": 10.0,
			"Y": 20.0,
			"Z": -0.5,
			"F": 1500.0,
		},
	}

	allowance := 1.0

	strategies := []struct {
		name     string
		strategy optimizer.FilterStrategy
	}{
		{"Safe", optimizer.StrategySafe},
		{"AllAxes", optimizer.StrategyAllAxes},
		{"Aggressive", optimizer.StrategyAggressive},
	}

	for _, s := range strategies {
		b.Run(s.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = optimizer.ShouldFilterMove(cmd, allowance, metadata, s.strategy)
			}
		})
	}
}
