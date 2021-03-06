package system

import (
	"context"
	"thermostat/db"
)

func init() {
	ctx := context.Background()
	if _, err := db.DB.ExecContext(ctx, "delete from setting"); err != nil {
		panic(err)
	}
	if _, err := db.DB.ExecContext(ctx, "delete from mode"); err != nil {
		panic(err)
	}
	if _, err := db.DB.ExecContext(ctx, "delete from zone"); err != nil {
		panic(err)
	}
}
