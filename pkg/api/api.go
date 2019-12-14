package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/takutakahashi/container-api-gateway/pkg/handler"
	"k8s.io/client-go/rest"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	b "github.com/takutakahashi/container-api-gateway/pkg/backend"
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
	e.GET(s.config.HealthPath, func(c echo.Context) error {
		return c.String(http.StatusOK, "healty")

	})
	var backend types.BaseBackend
	switch s.config.Backend {
	case "k8s":
		backend = b.KubernetesBackend{}
		fmt.Println("use kubernetes")
	default:
		backend = b.DockerBackend{}
		fmt.Println("use docker")
	}
	if _, err := rest.InClusterConfig(); err == nil {
		backend = b.KubernetesBackend{}
		fmt.Println("use kubernetes")
	}
	for _, endpoint := range s.config.Endpoints {
		e.Add(endpoint.Method, endpoint.Path, handler.GetHandler(endpoint, backend))
		if endpoint.Form {
			e.Add("GET", endpoint.Path, handler.GetFormHandler(endpoint))
		}
	}
	e.Start(s.config.GenServerURI())
}
