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

const switchInterval = 120

type hvac struct {
	fan  bool
	ac   bool
	heat bool

	last  time.Time
	mutex sync.Mutex

	fanPin  gpio.PinIO
	heatPin gpio.PinIO
	acPin   gpio.PinIO
}

// NewHVAC returns a new HVAC controller using the given fan, ac, and heat GPIO pins
func NewHVAC(fan, ac, heat int) *hvac {
	cont := &hvac{}

	pin := fmt.Sprintf("GPIO%d", fan)
	cont.fanPin = gpioreg.ByName(pin)
	if cont.fanPin == nil {
		logrus.WithField("pin", pin).Panic("Failed to find pin")
	}

	pin = fmt.Sprintf("GPIO%d", ac)
	cont.acPin = gpioreg.ByName(pin)
	if cont.acPin == nil {
		logrus.WithField("pin", pin).Panic("Failed to find pin")
	}

	pin = fmt.Sprintf("GPIO%d", heat)
	cont.heatPin = gpioreg.ByName(pin)
	if cont.heatPin == nil {
		logrus.WithField("pin", pin).Panic("Failed to find pin")
	}

	cont.Reset()

	return cont
}

func (c *hvac) Reset() {
	if err := c.fanPin.Out(gpio.Low); err != nil {
		logrus.WithField("pin", c.fanPin.String()).Fatal(err)
	}
	if err := c.acPin.Out(gpio.Low); err != nil {
		logrus.WithField("pin", c.acPin.String()).Fatal(err)
	}
	if err := c.heatPin.Out(gpio.Low); err != nil {
		logrus.WithField("pin", c.heatPin.String()).Fatal(err)
	}
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

	now := time.Now()
	if !c.canSwitch(now, on) {
		return c.ac
	}

	if on {
		time.AfterFunc(time.Second*15, func() {
			c.SetFan(true)
		})
	} else {
		// this needs to be long enough to dehumidify the air duct
		time.AfterFunc(time.Second*30, func() {
			c.SetFan(false)
		})
	}

	logrus.WithField("on", on).Info("toggling AC")

	level := gpio.Low
	if on {
		level = gpio.High
	}
	if err := c.acPin.Out(level); err != nil {
		log.Fatal(err)
	}

	c.ac = on
	c.last = now
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

	now := time.Now()
	if !c.canSwitch(now, on) {
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

	if on {
		// this is probably too short, but I don't want the heater to overheat
		time.AfterFunc(time.Second*15, func() {
			c.SetFan(true)
		})
	} else {
		time.AfterFunc(time.Second*30, func() {
			c.SetFan(false)
		})
	}

	c.heat = on
	c.last = now
	return c.Heat()
}

func (c hvac) canSwitch(now time.Time, on bool) bool {
	delta := time.Since(c.last).Round(time.Second)
	if on && delta < time.Second*switchInterval {
		logrus.WithFields(logrus.Fields{
			"since": delta.String(),
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
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetFan(true)
	c.SetHeat(true)
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetHeat(false)
	time.Sleep(time.Second * (switchInterval + 1))
	c.SetFan(false)
}
