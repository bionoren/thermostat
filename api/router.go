package api

import (
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	golog "log"
	"net/http"
	"thermostat/api/web"
	"time"
)

type logWriter struct {
	src string
}

func (log logWriter) Write(data []byte) (int, error) {
	logrus.WithField("src", log.src).Warn(string(data))
	return len(data), nil
}

func StartApi(cert, key []byte) {
	auth := newHmacAuth(viper.GetString("apiSecret"))

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/zones", handlerWrapper(zones, auth, false, true))
	mux.HandleFunc("/v1/status", handlerWrapper(status, auth, false, true))
	mux.HandleFunc("/v1/schedule", handlerWrapper(schedules, auth, false, true))
	mux.HandleFunc("/v1/schedule/add", handlerWrapper(addSchedule, auth, false, true))
	mux.HandleFunc("/v1/schedule/delete", handlerWrapper(deleteSchedule, auth, false, true))
	mux.HandleFunc("/v1/mode", handlerWrapper(modes, auth, false, true))
	mux.HandleFunc("/v1/mode/add", handlerWrapper(addMode, auth, false, true))
	mux.HandleFunc("/v1/mode/edit", handlerWrapper(editMode, auth, false, true))
	// TODO need to be able to validate that the mode isn't used by any schedules
	// mux.HandleFunc("/v1/mode/delete", handlerWrapper(deleteMode(update), auth, false, true))
	mux.HandleFunc("/v1/edit", handlerWrapper(editHandler, auth, false, true))

	mux.HandleFunc("/", handlerWrapper(web.Page("index.html"), nullAuth{}, true, false))
	mux.HandleFunc("/sha3.js", handlerWrapper(web.Page("sha3.js"), nullAuth{}, true, false))

	mux.HandleFunc("/v1/main.html", handlerWrapper(web.Page("main.html"), auth, false, false))
	mux.HandleFunc("/js.js", handlerWrapper(web.Page("thermostat.js"), auth, false, false))
	mux.HandleFunc("/v1/thermostat.html", handlerWrapper(web.Page("thermostat.html"), auth, false, false))
	mux.HandleFunc("/v1/thermostat.css", handlerWrapper(web.Page("thermostat.css"), auth, false, false))
	mux.HandleFunc("/v1/modes.html", handlerWrapper(web.Page("modes.html"), auth, false, false))
	mux.HandleFunc("/v1/addModes.html", handlerWrapper(web.Page("addModes.html"), auth, false, false))
	mux.HandleFunc("/v1/editModes.html", handlerWrapper(web.Page("editModes.html"), auth, false, false))
	mux.HandleFunc("/v1/schedules.html", handlerWrapper(web.Page("schedules.html"), auth, false, false))
	mux.HandleFunc("/v1/addSchedule.html", handlerWrapper(web.Page("addSchedule.html"), auth, false, false))
	mux.HandleFunc("/v1/editSchedule.html", handlerWrapper(web.Page("editSchedule.html"), auth, false, false))

	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}

	portString := fmt.Sprintf(":%d", viper.GetInt("apiPort"))
	srv := &http.Server{
		Addr:              portString,
		Handler:           mux,
		ReadTimeout:       time.Duration(15) * time.Second,
		ReadHeaderTimeout: time.Duration(15) * time.Second,
		WriteTimeout:      time.Duration(15) * time.Second,
		IdleTimeout:       time.Duration(120) * time.Second,
		ErrorLog:          golog.New(logWriter{"REMOTE-API"}, "", 0),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS13,
		},
	}

	logrus.WithField("port", portString).Info("starting API server")
	if err := srv.ListenAndServeTLS("", ""); err != nil {
		logrus.WithError(err).Panic("failed to start server")
	}
}
