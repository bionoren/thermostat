package web

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"thermostat/api/request"
	"time"
)

func Page(name string) func(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	tmplDir := viper.GetString("templateDir")
	data, err := ioutil.ReadFile(tmplDir + "/" + name)
	if err != nil {
		panic("failed to load file " + tmplDir + "/" + name)
	}
	fileData := string(data)

	return func(ctx context.Context, msg json.RawMessage) request.ApiResponse {
		return request.ApiResponse{
			Code: http.StatusOK,
			Msg:  fileData,
		}
	}
}

func MimePage(name, mimeType string) func(w http.ResponseWriter, r *http.Request) {
	page := Page(name)
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logrus.WithFields(logrus.Fields{
					"recovered": rec,
					"stack":     string(debug.Stack()),
					"path":      r.URL.Path,
				}).Error("Recovered from panic in API")
			}
		}()

		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("Method not allowed"))
			return
		}

		_ = r.Body.Close()

		log := logrus.WithFields(logrus.Fields{
			"path":      r.URL.Path,
			"ip":        r.RemoteAddr,
			"useragent": r.UserAgent(),
		})

		ctx, cancel := context.WithTimeout(context.WithValue(r.Context(), "log", log), 5*time.Second)
		defer cancel()

		resp := page(ctx, nil)

		w.Header().Set("Content-Type", mimeType)
		w.WriteHeader(resp.Code)
		_, _ = w.Write([]byte(resp.Msg))
	}
}
