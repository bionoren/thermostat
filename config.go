package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	viper.SetDefault("minTemp", 60)
	viper.SetDefault("maxTemp", 85)
	viper.SetDefault("apiPort", 443)
	viper.SetDefault("templateDir", "/usr/share/thermostat")

	viper.SetConfigName("thermostat")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.WithField("eventName", e.Name).Debug("Config file changed")
	})

	switch viper.GetString("log.type") {
	case "stderr":
		logrus.SetOutput(os.Stderr)
	case "file":
		logrus.SetFormatter(&logrus.JSONFormatter{})
		l := &lumberjack.Logger{
			Filename:  viper.GetString("log.file"),
			MaxAge:    365,
			LocalTime: true,
			Compress:  true,
		}
		logrus.SetOutput(l)

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP)

		go func() {
			for {
				<-c
				logrus.Debug("Rotating log file...")
				if err := l.Rotate(); err != nil {
					logrus.WithError(err).Error("Failed to rotate log")
				}
			}
		}()
	}

	switch viper.GetString("log.level") {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	}
}
