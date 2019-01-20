package system

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"thermostat/db"
	"thermostat/log"
	"time"
)

type Sensor interface {
	Temperature() float64
	Humidity() float64
}

type controller interface {
	SetFan(on bool) bool
	SetAC(on bool) bool
	SetHeat(on bool) bool
}

func monitor(s Sensor, c controller, update <-chan []db.Schedule) {
	interval := 30 * time.Second
	schedules := <-update

	tick := time.NewTicker(interval)

	c.SetFan(false)
	ac := c.SetAC(false)
	heat := c.SetHeat(false)

	if viper.IsSet("log.report") {
		if err := log.Setup(viper.GetString("log.report")); err != nil {
			panic(err)
		}
	}

	temp := s.Temperature()
	for {
		setting := currentSetting(schedules)

		if temp >= setting.MaxTemp()-setting.Correction() {
			heat = c.SetHeat(false)
			ac = c.SetAC(true)
		} else if temp <= setting.MinTemp()+setting.Correction() {
			ac = c.SetAC(false)
			heat = c.SetHeat(true)
		} else if heat || ac {
			heat = c.SetHeat(false)
			ac = c.SetAC(false)
		}

		log.Log(sys.Fan(), sys.AC(), sys.Heat(), temp, s.Humidity())

		select {
		case schedules = <-update:
		case <-tick.C:
			temp = s.Temperature()
		}
	}
}

func currentSetting(schedules []db.Schedule) db.Setting {
	now := time.Now()
	var sched db.Schedule
	for _, s := range schedules {
		if sched != nil && s.Priority() < sched.Priority() {
			continue
		}

		if s.DayOfWeek()&weekdayMask(now.Weekday()) == 0 {
			continue
		}
		if s.StartDay().After(now) || s.EndDay().Before(now) {
			continue
		}

		sec := daySeconds(now)
		if sec < s.StartTime() || sec > s.EndTime() {
			continue
		}

		sched = s
	}
	schedule = sched

	setting, err := db.GetSetting(sched.SettingID())
	if err != nil {
		logrus.WithError(err).WithField("settingID", sched.SettingID()).Panic("Could not get schedule")
	}

	return setting
}

func weekdayMask(d time.Weekday) int {
	return 2 << uint(d)
}

func daySeconds(t time.Time) int {
	hour, min, sec := t.Clock()
	return hour*60*60 + min*60 + sec
}
