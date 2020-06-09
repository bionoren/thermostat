package log

import (
	"encoding/csv"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"thermostat/config"
	"time"
)

var logger *csv.Writer

func init() {
	config.Ready()

	if viper.IsSet("log.report") {
		file := viper.GetString("log.report")
		f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 660)
		if err != nil {
			panic(err)
		}
		logger = csv.NewWriter(f)
	}
}

func Log(fan, ac, heat bool, temp, hum float64) {
	if logger == nil {
		return
	}

	data := []string{
		time.Now().Format(time.Stamp),
		strconv.FormatBool(fan),
		strconv.FormatBool(ac),
		strconv.FormatBool(heat),
		strconv.FormatFloat(temp, 'f', 6, 64),
		strconv.FormatFloat(hum, 'f', 6, 64),
	}

	_ = logger.Write(data)
	logger.Flush()
}
