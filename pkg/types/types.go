package types

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/thoas/go-funk"
)

type BaseBackend interface {
	Execute(Endpoint) (*bytes.Buffer, *bytes.Buffer, error)
}

type Config struct {
	Host       string
	Port       string
	HealthPath string     `yaml:"healthcheckPath"`
	Backend    string     `yaml:"backend"`
	Endpoints  []Endpoint `yaml:"endpoints"`
}
type Endpoint struct {
	Path      string    `yaml:"path"`
	Method    string    `yaml:"method"`
	Async     bool      `yaml:"async"`
	Params    []Param   `yaml:"params"`
	Env       []string  `yaml:"env"`
	Container Container `yaml:"container"`
}

type Container struct {
	Image   string   `yaml:"image"`
	Command []string `yaml:"command"` // contains go template

}

type Param struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Value    string `yaml:"default"`
	Optional bool   `yaml:"optional"`
}

func (c *Config) GenServerURI() string {
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if c.Port == "" {
		c.Port = "8080"
	}
	return c.Host + ":" + c.Port
}

// BuildCommand build command with params
func (e *Endpoint) BuildCommand() []string {
	result := []string{}
	for _, cmd := range e.Container.Command {
		params := funk.Filter(e.Params, func(p Param) bool {
			return strings.Contains(cmd, p.Name)
		}).([]Param)
		if len(params) != 0 {
			for _, param := range params {
				target := "." + param.Name
				cmd = strings.ReplaceAll(cmd, target, ".Value")
				tmpl, _ := template.New(param.Name).Parse(cmd)
				var doc bytes.Buffer
				tmpl.Execute(&doc, param)
				cmd = doc.String()
			}
		}
		if cmd != "" {
			result = append(result, cmd)
		}
	}
	return result
}

func (e *Endpoint) BuildEnv() []string {
	return funk.Map(e.Env, func(key string) string {
		return fmt.Sprintf("%s=%s", key, os.Getenv(key))
	}).([]string)
}
