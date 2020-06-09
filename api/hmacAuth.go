package api

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"hash"
	"net/http"
	"net/url"
	"strconv"
	"thermostat/api/request"
	"time"
)

type hmacAuth struct {
	hash.Hash
	key string
}

const MAXDRIFT = 10

func newHmacAuth(secret string) hmacAuth {
	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		logrus.WithError(err).Panic("Unable to load the api secret key")
	}

	return hmacAuth{Hash: hmac.New(func() hash.Hash { return crypto.SHA3_512.New() }, key), key: secret}
}

func (h hmacAuth) authorize(url *url.URL, body []byte) (json.RawMessage, request.ApiResponse) {
	var data struct {
		Sig     string
		Time    int64 // unix timestamp in UTC
		Payload json.RawMessage
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, request.ApiResponse{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
		}
	}

	if delta := time.Now().UTC().Unix() - data.Time; delta > MAXDRIFT || delta < -MAXDRIFT {
		return nil, request.ApiResponse{
			Code: http.StatusBadRequest,
			Msg:  "invalid timestamp",
		}
	}

	sig, err := hex.DecodeString(data.Sig)
	if err != nil {
		return nil, request.ApiResponse{
			Code: http.StatusUnauthorized,
			Msg:  "invalid signature encoding",
		}
	}

	var buf bytes.Buffer
	buf.WriteString(url.Path)
	buf.WriteString(strconv.FormatInt(data.Time, 10))
	buf.Write(data.Payload)

	h.Reset()
	_, _ = h.Write(buf.Bytes())
	check := h.Sum(nil)

	if !hmac.Equal(sig, check) {
		return nil, request.ApiResponse{
			Code: http.StatusUnauthorized,
			Msg:  "denied",
		}
	}

	return data.Payload, request.ApiResponse{}
}
