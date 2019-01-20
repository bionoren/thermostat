package system

import (
	"encoding/hex"
	"github.com/spf13/viper"
	"thermostat/db"
	"time"
)

func Startup(update <-chan []db.Schedule) {
	h := NewHVAC(viper.GetInt("fanPin"), viper.GetInt("acPin"), viper.GetInt("heatPin"))
	sys = h

	addr, err := hex.DecodeString(viper.GetString("tempSensor")[2:])
	if err != nil {
		panic(err)
	}
	sens = NewHIH6020(uint16(addr[0]), 0, 0)

	go monitor(sens, h, update)
}

var sys System
var sens Sensor
var schedule db.Schedule

type System interface {
	Fan() bool
	Heat() bool
	AC() bool
}

func Systems() []System {
	return []System{sys}
}

func Sensors() []Sensor {
	return []Sensor{sens}
}

func Configuration() db.Schedule {
	return schedule
}

func NextConfigChange() time.Time {
	panic("todo")
	return schedule.EndDay()
}
