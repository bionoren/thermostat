package db

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"time"
)

func init1(tx *bbolt.Tx, migrations *bbolt.Bucket) error {
	if data := migrations.Get(itob(1)); data == nil {
		modes, err := tx.CreateBucket([]byte("modes"))
		if err != nil {
			return err
		}
		settings, err := tx.CreateBucket([]byte("settings"))
		if err != nil {
			return err
		}

		id, _ := modes.NextSequence()
		m := mode{
			ID:         int64(id),
			Name:       "default",
			MinTempv:    60,
			MaxTempv:    85,
			Correctionv: 2,
		}
		buf, err := json.Marshal(m)
		if err != nil {
			return err
		}
		if err = modes.Put(itob(m.ID), buf); err != nil {
			return err
		}

		id, _ = settings.NextSequence()
		setting := setting{
			IDv:         int64(id),
			ModeIDv:    m.ID,
			Priorityv:  DEFAULT,
			DayOfWeekv: 255,
			StartDayv:  time.Unix(0, 0),
			EndDayv:    time.Date(2200, 1, 1, 1, 0, 0, 0, time.Local),
			StartTimev: 0,
			EndTimev:   86400,
		}
		buf, err = json.Marshal(setting)
		if err != nil {
			return err
		}
		if err = settings.Put(itob(setting.ID()), buf); err != nil {
			return err
		}

		id, _ = modes.NextSequence()
		m = mode{
			ID:         int64(id),
			Name:       "custom",
			MinTempv:    60,
			MaxTempv:    85,
			Correctionv: 1,
		}
		buf, err = json.Marshal(m)
		if err != nil {
			return err
		}
		if err = modes.Put(itob(m.ID), buf); err != nil {
			return err
		}

		if err := migrations.Put(itob(1), []byte{}); err != nil {
			return err
		}
	}

	return nil
}
