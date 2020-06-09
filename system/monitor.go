package system

import (
	"thermostat/db/setting"
	"time"
)

func currentSetting(schedules []setting.Setting) setting.Setting {
	now := time.Now().Round(time.Second) // rounding this lets me prevent instantaneous schedule overlap, while also preventing brief gaps in scheduling
	var current setting.Setting
	for _, s := range schedules {
		if s.Priority < current.Priority {
			continue
		}

		if s.DayOfWeek&setting.WeekdayMask(now.Weekday()) == 0 {
			continue
		}
		if s.StartDay.After(now) || s.EndDay.Before(now) {
			continue
		}

		sec := daySeconds(now)
		if sec < s.StartTime || sec > s.EndTime {
			continue
		}

		current = s
	}

	return current
}

func daySeconds(t time.Time) int {
	hour, min, sec := t.Clock()
	return hour*60*60 + min*60 + sec
}
