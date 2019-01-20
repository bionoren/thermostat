package db

import (
	"github.com/spf13/viper"
	"os"
)

func init() {
	viper.Set("db.file", "/tmp/thermostat.db")

	_ = os.Remove(viper.GetString("db.file"))

	if err := Connect(); err != nil {
		panic(err)
	}
	if err := Migrate(); err != nil {
		panic(err)
	}
}
