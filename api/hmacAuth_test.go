package api

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hash"
	"net/url"
	"testing"
	"time"
)

func TestHmacAuth(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	secret := "hqwII3HznSQ="
	h := newHmacAuth(secret)
	reqUrl := &url.URL{Path: "/v1/test"}

	data := struct {
		Sig     string
		Time    int64
		Payload interface{}
	}{
		Time: now.Unix(),
		Payload: struct {
			Foo string
		}{
			"testing123",
		},
	}
	data.Sig = hmacTest(t, secret, reqUrl.Path, data.Time, data.Payload)

	body, err := json.Marshal(data)
	require.NoError(t, err)

	msg, resp := h.authorize(reqUrl, body)
	assert.Equal(t, 0, resp.Code, resp.Msg)

	var payload struct {
		Foo string
	}
	err = json.Unmarshal(msg, &payload)
	assert.NoError(t, err)
	assert.Equal(t, data.Payload, payload)
}

func hmacTest(t testing.TB, secret, path string, time int64, payload interface{}) string {
	key, err := base64.StdEncoding.DecodeString(secret)
	require.NoError(t, err)
	h := hmac.New(func() hash.Hash { return crypto.SHA3_512.New() }, key)

	payloadData, err := json.Marshal(payload)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.WriteString(path)
	buf.WriteString(fmt.Sprintf("%d", time))
	buf.Write(payloadData)

	h.Write(buf.Bytes())
	return hex.EncodeToString(h.Sum(nil))
}
