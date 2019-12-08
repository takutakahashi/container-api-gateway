package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/takutakahashi/container-api-gateway/pkg/types"
)

type Response struct {
	Stdout string `json:"stdout" xml:"stdout"`
	Stderr string `json:"stderr" xml:"stderr"`
}

// GetHandler generate handler from endpoint
func GetHandler(endpoint types.Endpoint, b types.BaseBackend) echo.HandlerFunc {
	return func(c echo.Context) error {
		params := make([]types.Param, len(endpoint.Params))
		for i, param := range endpoint.Params {
			p := c.FormValue(param.Name)
			if p == "" && !param.Optional {
				return c.String(http.StatusBadRequest, "required param "+param.Name+" was not found.")
			}
			// if p is null, use default
			if p != "" {
				param.Value = p
			}
			params[i] = param
		}
		endpoint.Params = params
		stdout, _, err := b.Execute(endpoint)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		} else {
			var o interface{}
			in := bytes.Replace(stdout.Bytes(), []byte("'"), []byte("\""), -1)
			fmt.Println(string(in))
			err = json.Unmarshal(in, &o)
			fmt.Println(o)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			return c.JSON(http.StatusOK, o)
		}
	}
}
