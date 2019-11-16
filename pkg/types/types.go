package types

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/thoas/go-funk"
)

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
		if cmd != "" {
			result = append(result, cmd)
		}
	}
	return result
}

func (e *Endpoint) Execute() string {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	name := funk.RandomString(10)
	cli.ImagePull(ctx, e.Container.Image, types.ImagePullOptions{})
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: e.Container.Image,
		Cmd:   e.BuildCommand(),
	}, nil, nil, name)
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	if _, err = cli.ContainerWait(ctx, resp.ID); err != nil {
		panic(err)
	}
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	go cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	fmt.Println(buf.String())
	return buf.String()
}
