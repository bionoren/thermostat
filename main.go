package main

import (
	"context"
	"encoding/hex"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"thermostat/api"
	"thermostat/config"
	"thermostat/sensor"
	"thermostat/system"
)

func main() {
	config.Ready()
	ctx := context.Background()

	var test bool
	var stop bool
	flag.BoolVar(&test, "post", false, "Perform a power on self test of systems and sensors")
	flag.BoolVar(&stop, "stop", false, "Turn off all relays")
	flag.Parse()
	if test {
		selfTest()
		return
	}
	if stop {
		turnOff()
		return
	}

	sys := system.NewHVAC(viper.GetInt("fanPin"), viper.GetInt("acPin"), viper.GetInt("heatPin"))

	addr, err := hex.DecodeString(viper.GetString("tempSensor")[2:])
	if err != nil {
		panic(err)
	}
	sens := sensor.NewHIH6020(uint16(addr[0]), viper.GetFloat64("tempCorrection"), viper.GetFloat64("temperatureRangeDivider"), viper.GetFloat64("humCorrection"))
	zone, err := system.NewZone(ctx, "default", sys, sens)
	if err != nil {
		panic(err)
	}

	zone.Startup()

	cert, key := loadApiCert()
	api.StartApi(cert, key)
}

func loadApiCert() (cert, key []byte) {
	cert = []byte(viper.GetString("apiCert"))
	key = []byte(viper.GetString("apiKey"))
	return
}

func selfTest() {
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.Info("running Power On Self Test...")

	addr, err := hex.DecodeString(viper.GetString("tempSensor")[2:])
	if err != nil {
		panic(err)
	}
	sens := sensor.NewHIH6020(uint16(addr[0]), 0, 1, 0)
	logrus.WithField("temp", sens.Temperature()).Info("current temp")

	h := system.NewHVAC(viper.GetInt("fanPin"), viper.GetInt("acPin"), viper.GetInt("heatPin"))
	h.Test()

	logrus.WithField("temp", sens.Temperature()).Info("current temp")
}

func turnOff() {
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.Info("Powering off...")

	h := system.NewHVAC(viper.GetInt("fanPin"), viper.GetInt("acPin"), viper.GetInt("heatPin"))
	h.Reset()
}
