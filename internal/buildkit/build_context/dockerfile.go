package build_context

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/builder"
	gatewayClient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/tonistiigi/fsutil"
)

var defaultDockerfile = "Dockerfile"

const hivePrefix = "__hive__/"

// RewriteHiveRefs replaces all __hive__/ prefixes in a Dockerfile with the actual registry address.
func RewriteHiveRefs(src, target string, registryAddress string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return errors.Join(errors.New("failed to read Dockerfile for rewriting"), err)
	}

	replaced := strings.ReplaceAll(string(content), hivePrefix, registryAddress+"/")
	return os.WriteFile(target, []byte(replaced), 0644)
}

type DockerfileBuildContext struct {
	Root       string
	Dockerfile string
}

func (d DockerfileBuildContext) RunBuild(ctx context.Context, client gatewayClient.Client) (*gatewayClient.Result, error) {
	return builder.Build(ctx, client)
}

func (d DockerfileBuildContext) FileName() string {
	if d.Dockerfile == "" {
		return defaultDockerfile
	}
	return d.Dockerfile
}

func (d DockerfileBuildContext) FrontendType() string {
	return "dockerfile.v0"
}

func (d DockerfileBuildContext) ToLocalMounts() (map[string]fsutil.FS, error) {
	var dockerFilePath string
	if d.Dockerfile == "" {
		dockerFilePath = filepath.Join(d.Root, defaultDockerfile)
	} else {
		dockerFilePath = filepath.Join(d.Root, d.Dockerfile)
	}

	ctxFs, err := fsutil.NewFS(d.Root)
	if err != nil {
		return nil, errors.Join(errors.New("invalid Dockerfile path"), err)
	}

	dockerfileFs, err := fsutil.NewFS(filepath.Dir(dockerFilePath))
	if err != nil {
		return nil, errors.Join(errors.New("invalid Dockerfile path"), err)
	}

	return map[string]fsutil.FS{
		"context":    ctxFs,
		"dockerfile": dockerfileFs,
	}, nil
}
