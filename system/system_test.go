package system

import (
	"github.com/spf13/viper"
	"thermostat/db"
)

func init() {
	viper.Set("db.file", "/tmp/thermostat-sys.db")

	if err := db.Connect(); err != nil {
		panic(err)
	}
	if err := db.Migrate(); err != nil {
		panic(err)
	}

	viper.Set("maxTemp", 100)
}
