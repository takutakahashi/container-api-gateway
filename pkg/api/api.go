package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/takutakahashi/container-api-gateway/pkg/handler"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/takutakahashi/container-api-gateway/pkg/types"
	"gopkg.in/yaml.v2"
)

// Server contains server config and func
type Server struct {
	config types.Config
}

// LoadConfig loads config
func (s *Server) LoadConfig(configPath string) error {
	str, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	s.config = types.Config{}
	err = yaml.Unmarshal(str, &s.config)
	if err != nil {
		return err
	}
	return nil
}

// Start api server
func (s *Server) Start() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HideBanner = true
	for _, endpoint := range s.config.Endpoints {
		e.Add(endpoint.Method, endpoint.Path, handler.GetHandler(endpoint))
	}
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	e.Start(":8080")
}
