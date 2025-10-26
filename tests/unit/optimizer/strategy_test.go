package optimizer_test

import (
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
)

func TestParseFilterStrategy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    optimizer.FilterStrategy
		wantErr bool
	}{
		{
			name:    "Parse 'safe' strategy",
			input:   "safe",
			want:    optimizer.StrategySafe,
			wantErr: false,
		},
		{
			name:    "Parse 'all-axes' strategy",
			input:   "all-axes",
			want:    optimizer.StrategyAllAxes,
			wantErr: false,
		},
		{
			name:    "Parse 'split' strategy",
			input:   "split",
			want:    optimizer.StrategySplit,
			wantErr: false,
		},
		{
			name:    "Parse 'aggressive' strategy",
			input:   "aggressive",
			want:    optimizer.StrategyAggressive,
			wantErr: false,
		},
		{
			name:    "Parse with uppercase",
			input:   "SAFE",
			want:    optimizer.StrategySafe,
			wantErr: false,
		},
		{
			name:    "Parse with mixed case",
			input:   "All-Axes",
			want:    optimizer.StrategyAllAxes,
			wantErr: false,
		},
		{
			name:    "Invalid strategy returns error",
			input:   "invalid",
			want:    optimizer.StrategySafe, // Default
			wantErr: true,
		},
		{
			name:    "Empty string returns error",
			input:   "",
			want:    optimizer.StrategySafe, // Default
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := optimizer.ParseFilterStrategy(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilterStrategy(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFilterStrategy(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterStrategyString(t *testing.T) {
	tests := []struct {
		name     string
		strategy optimizer.FilterStrategy
		want     string
	}{
		{
			name:     "Safe strategy to string",
			strategy: optimizer.StrategySafe,
			want:     "safe",
		},
		{
			name:     "AllAxes strategy to string",
			strategy: optimizer.StrategyAllAxes,
			want:     "all-axes",
		},
		{
			name:     "Split strategy to string",
			strategy: optimizer.StrategySplit,
			want:     "split",
		},
		{
			name:     "Aggressive strategy to string",
			strategy: optimizer.StrategyAggressive,
			want:     "aggressive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.want {
				t.Errorf("FilterStrategy.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
