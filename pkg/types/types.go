package types

import (
	"bytes"
	"context"
	"html/template"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/thoas/go-funk"
)

type Config struct {
	Host      string
	Port      string
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
				cmd = strings.ReplaceAll(cmd, param.Name, ".Value")
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

func (e *Endpoint) Execute() (*bytes.Buffer, *bytes.Buffer, error) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, nil, err
	}
	name := funk.RandomString(10)
	cli.ImagePull(ctx, e.Container.Image, types.ImagePullOptions{})
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: e.Container.Image,
		Cmd:   e.BuildCommand(),
	}, nil, nil, name)
	if err != nil {
		return nil, nil, err
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, nil, err
	}
	if _, err = cli.ContainerWait(ctx, resp.ID); err != nil {
		return nil, nil, err
	}
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, nil, err
	}
	go cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	stdcopy.StdCopy(stdout, stderr, out)
	if err != nil {
		return nil, nil, err
	}
	return stdout, stderr, nil
}
