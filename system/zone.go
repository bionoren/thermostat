package system

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"thermostat/db/setting"
	"thermostat/db/zone"
	"thermostat/log"
	"time"
)

type Sensor interface {
	Temperature() float64
	Humidity() float64
}

type Controller interface {
	Fan() bool
	SetFan(on bool) bool
	AC() bool
	SetAC(on bool) bool
	Heat() bool
	SetHeat(on bool) bool
	Reset()
}

type Zone struct {
	zoneID     int64
	controller Controller
	sensor     Sensor
	update     chan []setting.Setting

	setting setting.Setting
}

func (z Zone) ID() int64 {
	return z.zoneID
}

func (z Zone) Controller() Controller {
	return z.controller
}

func (z Zone) Sensor() Sensor {
	return z.sensor
}

func (z Zone) Setting() setting.Setting {
	return z.setting
}

func (z Zone) Update(settings []setting.Setting) {
	z.update <- settings
}

func (z *Zone) monitor(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("recovered", r).WithField("zone", z.zoneID).Error("monitoring panicked")
			z.controller.Reset()
		}
	}()

	interval := 60 * time.Second
	schedules, err := setting.All(ctx, z.ID())
	if err != nil {
		logrus.WithField("zoneID", z.ID()).WithError(err).Panic("unable to load settings to initialize zone monitoring")
	}
	if len(schedules) == 0 { // tests may delete all existing schedules. In that case, we need to wait for schedules to be provided
		schedules = <-z.update
	}

	tick := time.NewTicker(interval)

	var ac, heat bool
	temp := z.sensor.Temperature()
	for {
		z.setting = currentSetting(schedules)
		mode := z.setting.Mode(ctx)

		switch {
		case ac:
			if temp <= mode.MaxTemp-mode.Correction {
				ac = z.controller.SetAC(false)
			}
		case heat:
			if temp >= mode.MinTemp+mode.Correction {
				heat = z.controller.SetHeat(false)
			}
		default:
			if temp > mode.MaxTemp {
				ac = z.controller.SetAC(true)
			} else if temp < mode.MinTemp {
				heat = z.controller.SetHeat(true)
			}
		}

		log.Log(z.controller.Fan(), z.controller.AC(), z.controller.Heat(), temp, z.sensor.Humidity())

		select {
		case schedules = <-z.update:
		case <-tick.C:
			temp = z.sensor.Temperature()
		}
	}
}

var zones = make(map[int64]*Zone)

func NewZone(ctx context.Context, name string, controller Controller, sensor Sensor) (Zone, error) {
	z, err := zone.Get(ctx, name)
	if err != nil {
		return Zone{}, err
	}

	return Zone{
		zoneID:     z.ID,
		controller: controller,
		sensor:     sensor,
		update:     make(chan []setting.Setting),
	}, nil
}

func GetZone(id int64) (*Zone, error) {
	if z, ok := zones[id]; ok {
		return z, nil
	}
	return nil, errors.New("no zone found")
}

func (z *Zone) Startup() {
	zones[z.zoneID] = z

	go func() {
		lastPanic := time.Now()
		for {
			z.controller.Reset()
			z.monitor(context.Background())
			now := time.Now()
			if now.Sub(lastPanic) < time.Hour {
				defer z.controller.Reset() // worth a try. Deferred in case it also causes a panic
				logrus.WithField("timeSincePanic", now.Sub(lastPanic).String()).Fatal("exiting because panics are in close proximity. Not sure what state the system is in anymore...")
				break
			} else {
				lastPanic = now
			}
		}
	}()
}
