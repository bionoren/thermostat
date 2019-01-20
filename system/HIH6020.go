package system

import (
	"encoding/hex"
	"github.com/sirupsen/logrus"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"time"
)


type HIH6020 struct {
	conn conn.Conn
	addr uint16

	conversionTime time.Duration
	tempCalibration float64
	humCalibration float64
	temp float64
	hum float64
	lastUpdate time.Time
}

func NewHIH6020(addr uint16, tempCalibration, humCalibration float64) *HIH6020 {
	var err error
	var i2cbus i2c.BusCloser
	if i2cbus, err = i2creg.Open(""); err != nil {
		logrus.WithError(err).Panic("Could not open i2c bus")
	}

	// Address the device with address addr on the I²C bus:
	dev := i2c.Dev{
		Bus:  i2cbus,
		Addr: addr, // default 0x27
	}

	return &HIH6020{
		// This is now a point-to-point connection and implements conn.Conn:
		conn:       &dev,
		addr:       addr,
		conversionTime: 41 * time.Millisecond, // The measurement cycle duration is typically 36.65 ms for temperature and humidity readings.
		tempCalibration: tempCalibration,
		humCalibration: humCalibration,
	}
}

func (s *HIH6020) Temperature() float64 {
	// keep things like the /status API from pounding on the sensor
	if time.Now().Sub(s.lastUpdate).Seconds() < 20 {
		return s.temp
	}

	s.update()
	return s.temp
}

func (s *HIH6020) Humidity() float64 {
	// keep things like the /status API from pounding on the sensor
	if time.Now().Sub(s.lastUpdate).Seconds() < 20 {
		return s.hum
	}

	s.update()
	return s.hum
}

func (s *HIH6020) update() {
	logrus.Debug("starting conversion")

	// I think this can be literally any command
	convertCmd := s.addr << 1
	if err := s.conn.Tx([]byte{byte(convertCmd)}, nil); err != nil {
		logrus.WithError(err).Error("Unable to get data")
		return
	}
	time.Sleep(s.conversionTime)

	resp := make([]byte, 4)
	if err := s.conn.Tx([]byte{}, resp); err != nil {
		logrus.WithError(err).Error("Unable to get data")
		return
	}

	logrus.WithField("data", hex.EncodeToString(resp)).Debug("sensor data")

	status := resp[0] >> 6
	switch status {
	case 0:
	case 1:
		logrus.WithField("wait", s.conversionTime).Warn("returning stale sensor data")
		s.conversionTime += 5 * time.Millisecond
	case 2:
		logrus.Panic("device is unexpectedly in command mode!")
	case 3:
		logrus.Panic("unexpected device status")
	}

	s.temp = FarenheitFromCelcius(s.celciusFromRaw(resp[2:])) + s.tempCalibration
	s.hum = s.humidityFromRaw(resp[0:2]) + s.humCalibration

	s.lastUpdate = time.Now()
}

func (s HIH6020) celciusFromRaw(data []byte) float64 {
	num := (uint64(data[0]) << 6) | (uint64(data[1]) >> 2)
	const denom float64 = 16382 // math.Exp2(14) - 2
	return float64(num) / denom * 165 - 40
}

func (s HIH6020) humidityFromRaw(data []byte) float64 {
	num := (uint64(data[0] & 0x3F) << 2) | uint64(data[1])
	const denom float64 = 16382 // math.Exp2(14) - 2
	return float64(num) / denom * 100
}
