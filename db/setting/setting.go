package setting

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"thermostat/db"
	"thermostat/db/mode"
	"time"
)

type Priority int

const (
	DEFAULT   Priority = iota + 1 // if the user does nothing, this schedule be used
	SCHEDULED                     // regularly scheduled things
	OVERRIDE                      // unusual things (vacation, etc)
	CUSTOM                        // manual override
)

type Setting struct {
	ID        int64
	ZoneID    int64     `json:"zoneID"`
	ModeID    int64     `json:"modeID"`
	Priority  Priority  `json:"priority"`
	DayOfWeek int       `json:"dayOfWeek"`
	StartDay  time.Time `json:"startDay"`
	EndDay    time.Time `json:"endDay"`
	StartTime int       `json:"startTime"`
	EndTime   int       `json:"endTime"`
}

func (s Setting) Mode(ctx context.Context) mode.Mode {
	m, err := mode.Get(ctx, s.ModeID)
	if err != nil {
		logrus.WithField("setting", s.ID).Panic("failed to load mode: " + err.Error())
	}
	return m
}

// Runtime returns the next time this schedule would run if it was the highest priority schedule at that time
func (s Setting) Runtime(now time.Time) time.Time {
	if now.After(s.EndDay) {
		return time.Time{}
	}

	if now.Before(s.StartDay) {
		return s.StartDay
	}

	start := now
	if s.DayOfWeek&WeekdayMask(now.Weekday()) > 0 {
		nowSeconds := now.Hour()*60*60 + now.Minute()*60 + now.Second()
		if nowSeconds < s.EndTime {
			if nowSeconds < s.StartTime {
				start = start.Add(time.Second * time.Duration(s.StartTime-nowSeconds))
			}
			return start
		}
	}

	_, offset := now.Zone()
	start = start.Truncate(time.Hour * 24).Add(time.Duration(-offset) * time.Second)
	for i, d := 1, now.Weekday()+1; i <= 7; i, d = i+1, (d+1)%(time.Saturday+1) {
		if s.DayOfWeek&WeekdayMask(d) > 0 {
			start = start.Add(time.Hour * 24 * time.Duration(i)).Add(time.Second * time.Duration(s.StartTime))
			break
		}
	}
	return start
}

func Overlaps(a, b Setting) bool {
	if a.ZoneID != b.ZoneID {
		return false
	}
	if a.Priority != b.Priority {
		return false
	}
	if a.EndDay.Before(b.StartDay) || a.StartDay.After(b.EndDay) {
		return false
	}
	if a.DayOfWeek&b.DayOfWeek == 0 {
		return false
	}
	if a.EndTime < b.StartTime || a.StartTime > b.EndTime {
		return false
	}

	return true
}

func WeekdayMask(d time.Weekday) int {
	return 2 << uint(d)
}

func All(ctx context.Context, zoneID int64) ([]Setting, error) {
	rows, err := db.DB.QueryContext(ctx, "select id, modeID, priority, dayOfWeek, startDay, endDay, startTime, endTime from setting where zoneID=?", zoneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make([]Setting, 0, 16)
	for rows.Next() {
		s := Setting{
			ZoneID: zoneID,
		}
		if err := rows.Scan(&s.ID, &s.ModeID, &s.Priority, &s.DayOfWeek, &s.StartDay, &s.EndDay, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}

	return settings, err
}

func allPriority(ctx context.Context, zoneID int64, priority Priority) ([]Setting, error) {
	rows, err := db.DB.QueryContext(ctx, "select id, modeID, dayOfWeek, startDay, endDay, startTime, endTime from setting where zoneID=? and priority=?", zoneID, priority)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make([]Setting, 0, 16)
	for rows.Next() {
		s := Setting{
			ZoneID:   zoneID,
			Priority: priority,
		}
		if err := rows.Scan(&s.ID, &s.ModeID, &s.DayOfWeek, &s.StartDay, &s.EndDay, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}

	return settings, err
}

func Validate(ctx context.Context, setting Setting) error {
	if setting.ZoneID == 0 {
		return errors.New("setting must be in a zone")
	}
	if setting.Priority == 0 {
		return errors.New("setting must have a priority")
	}
	if !setting.StartDay.Before(setting.EndDay) {
		return errors.New("setting start must be before setting end")
	}
	if setting.EndTime <= setting.StartTime {
		return errors.New("setting end time must be after start time")
	}
	if setting.DayOfWeek == 0 {
		return errors.New("setting must be active on at least one day of the week")
	}

	settings, err := allPriority(ctx, setting.ZoneID, setting.Priority)
	if err != nil {
		return err
	}

	for _, s := range settings {
		if Overlaps(setting, s) {
			return errors.New(fmt.Sprintf("new setting overlaps with setting %d", s.ID))
		}
	}

	return nil
}

func New(ctx context.Context, zoneID, modeID int64, priority Priority, dayOfWeek int, startDay, endDay time.Time, startTime, endTime int) (Setting, error) {
	s := Setting{
		ZoneID:    zoneID,
		ModeID:    modeID,
		Priority:  priority,
		DayOfWeek: dayOfWeek,
		StartDay:  startDay,
		EndDay:    endDay,
		StartTime: startTime,
		EndTime:   endTime,
	}
	if err := Validate(ctx, s); err != nil {
		return Setting{}, err
	}

	result, err := db.DB.ExecContext(ctx, "insert into setting (zoneID, modeID, priority, dayOfWeek, startDay, endDay, startTime, endTime) values (?, ?, ?, ?, ?, ?, ?, ?)", zoneID, modeID, priority, dayOfWeek, startDay, endDay, startTime, endTime)
	if err != nil {
		return Setting{}, err
	}

	s.ID, err = result.LastInsertId()

	return s, err
}

func (s Setting) Delete(ctx context.Context) error {
	if s.Priority == DEFAULT {
		return errors.New("cannot delete the default schedule")
	}
	_, err := db.DB.ExecContext(ctx, "delete from setting where id=?", s.ID)
	return err
}
