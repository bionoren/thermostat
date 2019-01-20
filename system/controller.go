package system

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"runtime/debug"
	"sync"
	"time"
)

const switchInterval = 60

type hvac struct {
	fan  bool
	ac   bool
	heat bool

	last  time.Time
	mutex sync.Mutex

	fanPin gpio.PinIO
	heatPin gpio.PinIO
	acPin gpio.PinIO
}

// NewHVAC returns a new HVAC controller using the given fan, ac, and heat GPIO pins
func NewHVAC(fan, ac, heat int) *hvac {
	cont := &hvac{}

	pin := fmt.Sprintf("GPIO%d", fan)
	cont.fanPin = gpioreg.ByName(pin)
	if cont.fanPin == nil {
		logrus.WithField("pin", pin).Panic("Failed to find pin")
	}
	if err := cont.fanPin.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}

	pin = fmt.Sprintf("GPIO%d", ac)
	cont.acPin = gpioreg.ByName(pin)
	if cont.acPin == nil {
		logrus.WithField("pin", pin).Panic("Failed to find pin")
	}
	if err := cont.acPin.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}

	pin = fmt.Sprintf("GPIO%d", heat)
	cont.heatPin = gpioreg.ByName(pin)
	if cont.heatPin == nil {
		logrus.WithField("pin", pin).Panic("Failed to find pin")
	}
	if err := cont.heatPin.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}

	return cont
}

func (c hvac) Fan() bool {
	return c.fan
}

func (c hvac) AC() bool {
	return c.ac
}

func (c hvac) Heat() bool {
	return c.heat
}

func (c *hvac) SetFan(on bool) bool {
	if on == c.fan {
		return on
	}

	if !on && (c.ac || c.heat) {
		on = true
	}

	logrus.WithField("on", on).Info("toggling fan")

	level := gpio.Low
	if on {
		level = gpio.High
	}
	if err := c.fanPin.Out(level); err != nil {
		log.Fatal(err)
	}

	c.fan = on
	return c.Fan()
}

func (c *hvac) SetAC(on bool) bool {
	if on == c.ac {
		return on
	}
	if c.heat {
		logrus.Error("Illegal attempt to engage AC while heat is on")
		return c.AC()
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.canSwitch(on) {
		return c.ac
	}

	c.SetFan(on)

	logrus.WithField("on", on).Info("toggling AC")

	level := gpio.Low
	if on {
		level = gpio.High
	}
	if err := c.acPin.Out(level); err != nil {
		log.Fatal(err)
	}

	c.ac = on
	c.last = time.Now()
	return c.AC()
}

func (c *hvac) SetHeat(on bool) bool {
	if on == c.heat {
		return on
	}
	if c.ac {
		logrus.Error("Illegal attempt to engage heat while AC is on")
		return c.Heat()
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.canSwitch(on) {
		return c.heat
	}

	logrus.WithField("on", on).Info("toggling heat")

	level := gpio.Low
	if on {
		level = gpio.High
	}
	if err := c.heatPin.Out(level); err != nil {
		log.Fatal(err)
	}

	c.SetFan(on)

	c.heat = on
	c.last = time.Now()
	return c.Heat()
}

func (c hvac) canSwitch(on bool) bool {
	if on && time.Since(c.last) < time.Second*switchInterval {
		logrus.WithFields(logrus.Fields{
			"since": time.Since(c.last),
			"stack": string(debug.Stack()),
		}).Warn("Attempted to engage a system too soon")
		return false
	}

	return true
}

func (c hvac) Test() {
	c.SetFan(true)
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetAC(true)
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetAC(false)
	c.SetFan(true)
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetHeat(true)
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetHeat(false)
	c.SetFan(true)
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetFan(false)
}
