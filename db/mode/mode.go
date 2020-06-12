package mode

import (
	"context"
	"errors"
	"github.com/spf13/viper"
	"math"
	"thermostat/db"
)

type Mode struct {
	ID         int64
	ZoneID     int64   `json:"zoneID"`
	Name       string  `json:"name"`
	MinTemp    float64 `json:"minTemp"`
	MaxTemp    float64 `json:"maxTemp"`
	Correction float64 `json:"correction"`
}

func (m Mode) Validate() error {
	if len(m.Name) < 2 {
		return errors.New("name must be at least 2 characters long")
	}
	if m.MaxTemp-m.MinTemp < 2 {
		return errors.New("max temperature must be at least 2 degrees higher than the min temperature")
	}
	if m.Correction < 0.5 {
		return errors.New("correction must be >= 0.5")
	}
	if m.Correction*2 > m.MaxTemp-m.MinTemp {
		return errors.New("correction cannot be more than half the difference in temperature range")
	}

	return nil
}

func Get(ctx context.Context, id int64) (Mode, error) {
	row := db.DB.QueryRowContext(ctx, "select id, zoneID, name, minTemp, maxTemp, correction from mode where id=?", id)
	var m Mode
	if err := row.Scan(&m.ID, &m.ZoneID, &m.Name, &m.MinTemp, &m.MaxTemp, &m.Correction); err != nil {
		return Mode{}, err
	}

	m.MinTemp = math.Max(m.MinTemp, viper.GetFloat64("minTemp"))
	m.MaxTemp = math.Min(m.MaxTemp, viper.GetFloat64("maxTemp"))
	if m.MaxTemp-m.MinTemp < 2 {
		m.MinTemp = viper.GetFloat64("minTemp")
		m.MaxTemp = viper.GetFloat64("maxTemp")
	}
	if m.Correction < 0.5 {
		m.Correction = 0.5
	}
	if m.Correction*2 > m.MaxTemp-m.MinTemp {
		m.Correction = (m.MaxTemp - m.MinTemp) / 2
	}
	return m, nil
}

func All(ctx context.Context, zone int64) ([]Mode, error) {
	rows, err := db.DB.QueryContext(ctx, "select id, name, minTemp, maxTemp, correction from mode where zoneID=?", zone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	modes := make([]Mode, 0, 4)
	for rows.Next() {
		m := Mode{
			ZoneID: zone,
		}
		if err := rows.Scan(&m.ID, &m.Name, &m.MinTemp, &m.MaxTemp, &m.Correction); err != nil {
			return nil, err
		}
		modes = append(modes, m)
	}

	return modes, err
}

func New(ctx context.Context, zoneID int64, name string, minTemp, maxTemp, correction float64) (Mode, error) {
	m := Mode{
		ZoneID:     zoneID,
		Name:       name,
		MinTemp:    minTemp,
		MaxTemp:    maxTemp,
		Correction: correction,
	}

	result, err := db.DB.ExecContext(ctx, "insert into mode (zoneID, name, minTemp, maxTemp, correction) values (?, ?, ?, ?, ?)", zoneID, name, minTemp, maxTemp, correction)
	if err != nil {
		return Mode{}, err
	}

	m.ID, err = result.LastInsertId()

	return m, err
}

func (m Mode) Update(ctx context.Context) error {
	_, err := db.DB.ExecContext(ctx, "UPDATE mode SET name=?, minTemp=?, maxTemp=?, correction=? where id=?", m.Name, m.MinTemp, m.MaxTemp, m.Correction, m.ID)
	return err
}

func (m Mode) Delete(ctx context.Context) error {
	_, err := db.DB.ExecContext(ctx, "delete from mode where id=?", m.ID)
	return err
}
