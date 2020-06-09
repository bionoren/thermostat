package api

import (
	"encoding/json"
	"net/url"
	"thermostat/api/request"
)

type nullAuth struct{}

func (auth nullAuth) authorize(_ *url.URL, body []byte) (json.RawMessage, request.ApiResponse) {
	return body, request.ApiResponse{}
}
