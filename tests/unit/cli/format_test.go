package cli_test

import (
	"testing"
	"time"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/cli"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"Under 1000", 123, "123"},
		{"Exactly 1000", 1000, "1,000"},
		{"Thousands", 12450, "12,450"},
		{"Hundreds of thousands", 123456, "123,456"},
		{"Millions", 1234567, "1,234,567"},
		{"Zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.FormatNumber(tt.input)
			if got != tt.want {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{"Small bytes", 500, "500"},
		{"Kilobytes", 1500, "1,500"},
		{"Large bytes", 1234567, "1,234,567"},
		{"Zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.FormatBytes(tt.input)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"Milliseconds", 500 * time.Millisecond, "0.5s"},
		{"Under 1 second", 800 * time.Millisecond, "0.8s"},
		{"Exactly 1 second", 1 * time.Second, "1.0s"},
		{"Several seconds", 5 * time.Second, "5.0s"},
		{"Under 1 minute", 45 * time.Second, "45.0s"},
		{"Exactly 1 minute", 60 * time.Second, "1m 0s"},
		{"Minutes and seconds", 5*time.Minute + 30*time.Second, "5m 30s"},
		{"Under 1 hour", 45*time.Minute + 15*time.Second, "45m 15s"},
		{"Hours and minutes", 2*time.Hour + 15*time.Minute, "2h 15m"},
		{"Multiple hours", 5*time.Hour + 45*time.Minute, "5h 45m"},
		{"Zero", 0, "0.0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.FormatDuration(tt.input)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatThroughput(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"Low throughput", 45.67, "46"},
		{"Under 1000", 999.9, "1000"},
		{"Thousands", 12450.5, "12,450"},
		{"High throughput", 123456.78, "123,456"},
		{"Zero", 0.0, "0"},
		{"Very low", 0.5, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.FormatThroughput(tt.input)
			if got != tt.want {
				t.Errorf("FormatThroughput(%f) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
