package zone

import (
	"context"
	"github.com/spf13/viper"
	"thermostat/db"
)

func init() {
	ctx := context.Background()
	viper.SetDefault("db.file", "/tmp/thermostat-db-zone.db")

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
