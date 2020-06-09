package db

import (
	"context"
	"github.com/spf13/viper"
)

func init() {
	ctx := context.Background()
	viper.SetDefault("db.file", "/tmp/thermostat-db.db")

	if _, err := DB.ExecContext(ctx, "delete from setting"); err != nil {
		panic(err)
	}
	if _, err := DB.ExecContext(ctx, "delete from mode"); err != nil {
		panic(err)
	}
	if _, err := DB.ExecContext(ctx, "delete from zone"); err != nil {
		panic(err)
	}
}
