package system

import (
	"github.com/spf13/viper"
	"thermostat/config"
)

func init() {
	config.Ready()

	viper.Set("maxTemp", 100)
}
