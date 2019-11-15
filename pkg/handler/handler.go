package handler

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/takutakahashi/container-api-gateway/pkg/types"
)

func GetHandler(endpoint types.Endpoint) echo.HandlerFunc {
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
		return c.String(http.StatusOK, endpoint.Execute())
	}
}
