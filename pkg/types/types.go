package types

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/thoas/go-funk"
)

type BaseBackend interface {
	Execute(Endpoint) (*bytes.Buffer, *bytes.Buffer, error)
}

type Config struct {
	BaseURL    string `yaml:"baseURL"`
	Host       string
	Port       string
	HealthPath string     `yaml:"healthcheckPath"`
	Backend    string     `yaml:"backend"`
	Endpoints  []Endpoint `yaml:"endpoints"`
}
type Endpoint struct {
	Path        string      `yaml:"path"`
	Method      string      `yaml:"method"`
	Cache       Cache       `yaml:"cache"`
	Async       bool        `yaml:"async"`
	Form        bool        `yaml:"form"`
	TemplateURL string      `yaml:"template_url"`
	Response    string      `yaml:"response"`
	Params      []Param     `yaml:"params"`
	SecretName  string      `yaml:"secretName"`
	Env         []string    `yaml:"env"`
	Containers  []Container `yaml:"containers"`
}

type Container struct {
	Name    string   `yaml:"name"`
	Image   string   `yaml:"image"`
	Command []string `yaml:"command"` // contains go template

}

type Param struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Value       string   `yaml:"default"`
	Choices     []Choice `yaml:"choice"`
	Optional    bool     `yaml:"optional"`
}

type Choice struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Cache struct {
	Enabled bool          `yaml:"enabled"`
	Expire  time.Duration `yaml:"expire"`
	Store   *cache.Cache
}

func (c Cache) GetStdout(key string) (*bytes.Buffer, bool) {
	fmt.Println("key: " + key)
	val, found := c.Store.Get(key)
	if !found {
		return nil, found
	}
	return val.(*bytes.Buffer), found
}
func (c Cache) SetStdout(key string, val *bytes.Buffer) {
	fmt.Println("key: " + key)
	fmt.Println(fmt.Sprintf("val: %v", val))
	c.Store.Set(key, val, c.Expire*time.Minute)
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
func (e *Endpoint) BuildCommand(c Container) []string {
	result := []string{}
	for _, cmd := range c.Command {
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
