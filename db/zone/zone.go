package zone

import (
	"context"
	"encoding/json"
	"thermostat/db"
)

type Zone struct {
	ID   int64
	Name string `json:"name"`
}

func Get(ctx context.Context, name string) (Zone, error) {
	row := db.DB.QueryRowContext(ctx, "select id from zone where name=?", name)
	z := Zone{
		Name: name,
	}
	err := row.Scan(&z.ID)

	return z, err
}

func All(ctx context.Context) ([]json.RawMessage, error) {
	rows, err := db.DB.QueryContext(ctx, "select id, name from zone")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	zones := make([]json.RawMessage, 0, 4)
	for rows.Next() {
		var z Zone
		if err := rows.Scan(&z.ID, &z.Name); err != nil {
			return nil, err
		}
		data, err := json.Marshal(z)
		if err != nil {
			return nil, err
		}
		zones = append(zones, data)
	}

	return zones, nil
}

func New(ctx context.Context, name string) (Zone, error) {
	result, err := db.DB.ExecContext(ctx, "insert into zone (name) values (?)", name)
	if err != nil {
		return Zone{}, err
	}

	z := Zone{
		Name: name,
	}
	z.ID, err = result.LastInsertId()
	return z, err
}

func (z *Zone) Update(ctx context.Context) error {
	_, err := db.DB.ExecContext(ctx, "update zone set name=? where id=?", z.Name, z.ID)
	return err
}

func (z Zone) Delete(ctx context.Context) error {
	_, err := db.DB.ExecContext(ctx, "delete from zone where id=?", z.ID)
	return err
}
