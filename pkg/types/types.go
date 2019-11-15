package types

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/thoas/go-funk"
)

func NewConfig() Config {
	return Config{
		Endpoints: []Endpoint{},
	}
}

type Config struct {
	Endpoints []Endpoint `yaml:"endpoints"`
}
type Endpoint struct {
	Path      string    `yaml:"path"`
	Method    string    `yaml:"method"`
	Async     bool      `yaml:"async"`
	Params    []Param   `yaml:"params"`
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

// BuildCommand build command with params
func (e *Endpoint) BuildCommand() []string {
	result := make([]string, len(e.Container.Command))
	for _, cmd := range e.Container.Command {
		params := funk.Filter(e.Params, func(p Param) bool {
			return strings.Contains(cmd, p.Name)
		}).([]Param)
		if len(params) != 0 {
			for _, param := range params {
				cmd = strings.ReplaceAll(cmd, param.Name, ".Value")
				tmpl, _ := template.New(param.Name).Parse(cmd)
				var doc bytes.Buffer
				tmpl.Execute(&doc, param)
				cmd = doc.String()
			}
		}
		result = append(result, cmd)
	}
	fmt.Println(result)
	return result
}

func (e *Endpoint) Execute() string {
	return strings.Join(e.BuildCommand(), ",")
}
