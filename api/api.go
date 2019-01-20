package api

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	_ "golang.org/x/crypto/sha3"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"
	"thermostat/api/request"
	"time"
)

type authorizer interface {
	authorize(url *url.URL, body []byte) (json.RawMessage, request.ApiResponse)
}
type handler func(ctx context.Context, msg json.RawMessage) request.ApiResponse

func handlerWrapper(f handler, auth authorizer, allowGet, logRequest bool) func(w http.ResponseWriter, r *http.Request) {
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

		if r.Method != "POST" && (!allowGet || r.Method != "GET") {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("Method not allowed"))
			return
		}

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024*2))
		if err != nil {
			logrus.WithError(err).Error("Error reading request body")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
			return
		}

		payload, resp := auth.authorize(r.URL, body)
		if resp.Code != 0 {
			w.WriteHeader(resp.Code)
			_, _ = w.Write([]byte(resp.Msg))
			return
		}

		log := logrus.WithFields(logrus.Fields{
			"path":      r.URL.Path,
			"ip":        r.RemoteAddr,
			"useragent": r.UserAgent(),
		})
		if logRequest {
			log.Info("api request")
			log.Debug(string(body))
		}

		ctx, cancel := context.WithTimeout(context.WithValue(r.Context(), "log", log), 5*time.Second)
		defer cancel()

		resp = f(ctx, payload)

		if logRequest {
			log = log.WithField("code", resp.Code)
			log.Info("api response")
			log.Debug(resp.Msg)
		}

		w.WriteHeader(resp.Code)
		_, _ = w.Write([]byte(resp.Msg))
	}
}
