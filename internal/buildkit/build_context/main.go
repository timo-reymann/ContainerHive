package build_context

import (
	"context"

	gatewayClient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/tonistiigi/fsutil"
)

type BuildContext interface {
	FrontendType() string
	ToLocalMounts() (map[string]fsutil.FS, error)
	FileName() string
	RunBuild(context.Context, gatewayClient.Client) (*gatewayClient.Result, error)
}
