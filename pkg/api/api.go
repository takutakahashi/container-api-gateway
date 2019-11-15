package api

import (
	"io/ioutil"

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

// Start starts api server
func (s *Server) Start() {
	s.config.Endpoints[2].BuildCommand()
}
