package backend

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	ctypes "github.com/takutakahashi/container-api-gateway/pkg/types"
	"github.com/thoas/go-funk"
)

type DockerBackend struct{}

func (b DockerBackend) Execute(e ctypes.Endpoint) (*bytes.Buffer, *bytes.Buffer, error) {
	if e.Async {
		go execute(e)
		return nil, nil, nil
	}
	return execute(e)
}

func execute(e ctypes.Endpoint) (*bytes.Buffer, *bytes.Buffer, error) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, nil, err
	}
	name := funk.RandomString(10)
	progress, err := cli.ImagePull(ctx, e.Container.Image, types.ImagePullOptions{})
	if err != nil {
		return nil, nil, err
	}
	io.Copy(os.Stdout, progress)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: e.Container.Image,
		Cmd:   e.BuildCommand(),
		Env:   e.BuildEnv(),
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
