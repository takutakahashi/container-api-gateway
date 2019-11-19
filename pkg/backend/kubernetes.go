package backend

import (
	"bytes"

	"github.com/takutakahashi/container-api-gateway/pkg/types"
)

type KubernetesBackend struct{}

func (b KubernetesBackend) Execute(e types.Endpoint) (*bytes.Buffer, *bytes.Buffer, error) {
	return nil, nil, nil
}
