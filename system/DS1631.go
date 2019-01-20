package system

import (
	"github.com/sirupsen/logrus"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"time"
)

type DS1631 struct {
	conn conn.Conn

	calibration float64
	temp float64
	lastUpdate time.Time
}

func NewDS1631(addr uint16, calibration float64) *DS1631 {
	var err error
	var i2cbus i2c.BusCloser
	if i2cbus, err = i2creg.Open(""); err != nil {
		logrus.WithError(err).Panic("Could not open i2c bus")
	}

	// Address the device with address addr on the I²C bus:
	dev := i2c.Dev{
		Bus:  i2cbus,
		Addr: addr,
	}

	// 10001101 - oneshot, 12 bit resolution, max conversion time 750ms
	// 0x8D
	// 10001001 - oneshot, 11 bit resolution, max conversion time 375ms
	// 0x89
	// 10000101 - oneshot, 10 bit resolution, max conversion time 187.5ms
	// 0x85
	if err := dev.Tx([]byte{0xAC, 0x8D}, nil); err != nil {
		logrus.WithError(err).Panic("Unable to configure temp DS1631")
	}
	logrus.Debug("DS1631 configured")

	return &DS1631{
		// This is now a point-to-point connection and implements conn.Conn:
		conn:       &dev,
		calibration: calibration,
	}
}

func (s DS1631) Temperature() float64 {
	// keep things like the /status API from pounding on the sensor
	if time.Now().Sub(s.lastUpdate).Seconds() < 20 {
		return s.temp
	}

	// start conversion
	logrus.Debug("starting temperature conversion")
	if err := s.conn.Tx([]byte{0x51}, nil); err != nil {
		logrus.WithError(err).Error("Unable to get temperature data")
		return s.temp
	}

	// poll for conversion to finish
	for {
		time.Sleep(125 * time.Millisecond)

		resp := make([]byte, 1)
		if err := s.conn.Tx([]byte{0xAC}, resp); err != nil {
			logrus.WithError(err).Error("Unable to get temperature data")
			continue
		}
		if resp[0] & 0x80 == 0x80 {
			break
		}
	}

	// read temperature data
	logrus.Debug("parsing temperature data")
	resp := make([]byte, 2)
	if err := s.conn.Tx([]byte{0xAA}, resp); err != nil {
		logrus.WithError(err).Error("Unable to get temperature data")
		return s.temp
	}

	// convert data to farenheit
	celcius := s.celciusFromRaw(resp)
	s.temp = FarenheitFromCelcius(celcius) + s.calibration
	s.lastUpdate = time.Now()

	logrus.WithField("temp", s.temp).Debug("read temperature")
	return s.temp
}

func (s DS1631) celciusFromRaw(data []byte) float64 {
	celcius := float64(data[0])
	for i := uint(7); i > 0; i-- {
		if data[1] & byte(2 << i) > 0 {
			divisor := 1 << (7-i)
			celcius += float64(1)/float64(divisor)
		}
	}

	return celcius
}

func (s DS1631) Humidity() float64 {
	return 0
}
