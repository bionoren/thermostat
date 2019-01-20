package db

import (
	"encoding/json"
	"github.com/spf13/viper"
	"go.etcd.io/bbolt"
	"math"
)

type mode struct {
	ID         int64 `json:"ID"`
	Name       string `json:"name"`
	MinTempv    float64 `json:"minTemp"`
	MaxTempv    float64 `json:"maxTemp"`
	Correctionv float64 `json:"correction"`
}

func (m mode) MinTemp() float64 {
	return m.MinTempv
}

func (m mode) MaxTemp() float64 {
	return m.MaxTempv
}

func (m mode) Correction() float64 {
	return m.Correctionv
}

type Setting interface {
	MinTemp() float64
	MaxTemp() float64
	Correction() float64
}

func GetSetting(id int64) (Setting, error) {
	var m mode
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		data := b.Get(itob(id))
		return json.Unmarshal(data, &m)
	})

	m.MinTempv = math.Max(m.MinTempv, viper.GetFloat64("minTemp"))
	m.MaxTempv = math.Min(m.MaxTempv, viper.GetFloat64("maxTemp"))
	if m.MaxTempv - m.MinTempv < 2 {
		m.MinTempv = viper.GetFloat64("minTemp")
		m.MaxTempv = viper.GetFloat64("maxTemp")
	}
	if m.Correctionv < 0.5 {
		m.Correctionv = 0.5
	}
	if m.Correctionv*2 > m.MaxTempv - m.MinTempv {
		m.Correctionv = (m.MaxTempv - m.MinTempv) / 2
	}
	return m, err
}

func Modes() ([]json.RawMessage, error) {
	modes := make([]json.RawMessage, 0, 4)

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		return b.ForEach(func(k, v []byte) error {
			modes = append(modes, v)
			return nil
		})
	})

	return modes, err
}

func AddMode(name string, min, max, correction float64) (int64, error) {
	m := mode{
		Name:       name,
		MinTempv:    min,
		MaxTempv:    max,
		Correctionv: correction,
	}

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		id, _ := b.NextSequence()
		m.ID = int64(id)
		buf, err := json.Marshal(m)
		if err != nil {
			return err
		}
		return b.Put(itob(m.ID), buf)
	})

	return m.ID, err
}

func EditMode(id int64, name string, min, max, correction float64) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		data := b.Get(itob(id))
		var m mode
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		m.Name = name
		m.MinTempv = min
		m.MaxTempv = max
		m.Correctionv = correction

		buf, err := json.Marshal(m)
		if err != nil {
			return err
		}

		return b.Put(itob(m.ID), buf)
	})
}

func DeleteMode(id int64) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		return b.Delete(itob(id))
	})
}

func CustomMode() (modeID int64, err error) {
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		return b.ForEach(func(k, v []byte) error {
			var m mode
			if err := json.Unmarshal(v, &m); err != nil {
				return err
			}
			if m.Name == "custom" {
				modeID = m.ID
			}
			return nil
		})
	})

	return
}
