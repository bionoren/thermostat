package db

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"time"
)

type Priority int

const (
	DEFAULT Priority = iota + 1
	SCHEDULED
	OVERRIDE
)

type setting struct {
	IDv        int64 `json:"ID"`
	ModeIDv    int64 `json:"modeID"`
	Priorityv  Priority `json:"priority"`
	DayOfWeekv int `json:"dayOfWeek"`
	StartDayv  time.Time `json:"startDay"`
	EndDayv    time.Time `json:"endDay"`
	StartTimev int `json:"startTime"`
	EndTimev   int `json:"endTime"`
}

func (s setting) ID() int64 {
	return s.IDv
}

func (s setting) Priority() Priority {
	return s.Priorityv
}

func (s setting) DayOfWeek() int {
	return s.DayOfWeekv
}

func (s setting) StartDay() time.Time {
	return s.StartDayv
}

func (s setting) EndDay() time.Time {
	return s.EndDayv
}

func (s setting) StartTime() int {
	return s.StartTimev
}

func (s setting) EndTime() int {
	return s.EndTimev
}

func (s setting) SettingID() int64 {
	return s.ModeIDv
}

type Schedule interface {
	ID() int64
	Priority() Priority
	DayOfWeek() int
	StartDay() time.Time
	EndDay() time.Time
	StartTime() int
	EndTime() int
	SettingID() int64
}

func Settings() ([]Schedule, error) {
	settings := make([]Schedule, 0, 8)
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))

		return b.ForEach(func(k, v []byte) error {
			var s setting
			if err := json.Unmarshal(v, &s); err != nil {
				return err
			}

			settings = append(settings, s)
			return nil
		})
	})

	return settings, err
}

func AddSetting(modeID int64, priority Priority, dayOfWeek int, start, end time.Time, startTime, endTime int) (Schedule, error) {
	setting := setting{
		ModeIDv:     modeID,
		Priorityv:  priority,
		DayOfWeekv: dayOfWeek,
		StartDayv:  start,
		EndDayv:    end,
		StartTimev: startTime,
		EndTimev:   endTime,
	}

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))

		id, _ := b.NextSequence()
		setting.IDv = int64(id)
		buf, err := json.Marshal(setting)
		if err != nil {
			return err
		}
		return b.Put(itob(setting.IDv), buf)
	})

	return setting, err
}

func DeleteSetting(id int64) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))

		return b.Delete(itob(id))
	})
}
