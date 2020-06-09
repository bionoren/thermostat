package web

import (
	"context"
	"encoding/json"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"thermostat/api/request"
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
