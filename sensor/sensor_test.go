package sensor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFarenheitFromCelcius(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		celcius   float64
		farenheit float64
	}{
		{"40c", 40, 104},
		{"10c", 10, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FarenheitFromCelcius(tt.celcius)
			assert.Equal(t, tt.farenheit, f)
		})
	}
}

func TestHeatIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		temp     float64
		hum      float64
		expected float64
	}{
		{60, 0, 55.7},
		{60, 25, 56.875},
		{60, 50, 58.05},
		{60, 75, 59.225},
		{60, 100, 60.4},
		{75, 0, 72.2},
		{75, 25, 73.375},
		{75, 50, 74.55},
		{75, 75, 75.725},
		{75, 100, 76.9},
		{85, 0, 80.2984837},
		{85, 25, 82.3624083},
		{85, 50, 86.4593188},
		{85, 75, 94.6747043},
		{85, 100, 107.60856480},
	}

	for i, tt := range tests {
		temp := HeatIndex(tt.temp, tt.hum)
		assert.InDelta(t, tt.expected, temp, 0.0001, "Test %d: %.2f℉, %.2f% hum", i, tt.temp, tt.hum)
	}
}
