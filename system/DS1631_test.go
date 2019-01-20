package system

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDS1631_celciusFromRaw(t *testing.T) {
	t.Parallel()

	sensor := DS1631{}
	tests := []struct{
		name string
		input []byte
		celcius float64
	}{
		{"40C from datasheet", []byte{0x28, 0x00}, 40},
		{"10C from datasheet", []byte{0x0A, 0x00}, 10},
		{"25.0625C from datasheet", []byte{0x19, 0x10}, 25.0625},
		{"10.125C from datasheet", []byte{0x0A, 0x20}, 10.125},
		{"0.5C from datasheet", []byte{0x00, 0x80}, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp := sensor.celciusFromRaw(tt.input)
			assert.Equal(t, tt.celcius, temp)
		})
	}
}
