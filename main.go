package main

import (
	"encoding/hex"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"thermostat/api"
	"thermostat/db"
	"thermostat/system"
)

func main() {
	var test bool
	flag.BoolVar(&test, "post", false, "Perform a power on self test of systems and sensors")
	flag.Parse()
	if test {
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{})
		logrus.Info("running Power On Self Test...")

		addr, err := hex.DecodeString(viper.GetString("tempSensor")[2:])
		if err != nil {
			panic(err)
		}
		sens := system.NewDS1631(uint16(addr[0]), -2)
		logrus.WithField("temp", sens.Temperature()).Info("current temp")

		h := system.NewHVAC(viper.GetInt("fanPin"), viper.GetInt("acPin"), viper.GetInt("heatPin"))
		h.Test()

		logrus.WithField("temp", sens.Temperature()).Info("current temp")
		return
	}

	if err := db.Connect(); err != nil {
		panic(err)
	}
	if err := db.Migrate(); err != nil {
		panic(err)
	}

	update := make(chan []db.Schedule, 1)
	settings, err := db.Settings()
	if err != nil {
		panic(err)
	}
	update <- settings
	system.Startup(update)

	cert, key, err := loadApiCert()
	if err != nil {
		panic(err)
	}
	api.StartApi(update, cert, key)
}

func loadApiCert() (cert, key []byte, err error) {
	cert = []byte(viper.GetString("apiCert"))
	key = []byte(viper.GetString("apiKey"))

	return
}
