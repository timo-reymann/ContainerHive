package buildkit

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/moby/buildkit/client"
	"github.com/timo-reymann/ContainerHive/internal/buildkit/build_context"
	"github.com/timo-reymann/ContainerHive/internal/buildkit/cache"
	"github.com/timo-reymann/ContainerHive/internal/utils"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	buildkit *client.Client
}

type BuildOpts struct {
	ImageName    string
	Platform     string
	TarFile      string
	BuildArgs    map[string]string
	Labels       map[string]string
	Cache        cache.BuildkitCache
	BuildContext build_context.BuildContext
}

func NewClient(ctx context.Context, endpoint string) (*Client, error) {
	buildkit, err := client.New(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{buildkit}, nil
}

func (c *Client) Close() error {
	return c.buildkit.Close()
}

func (c *Client) Build(ctx context.Context, opts *BuildOpts, statusUpdateHandler func(chan *client.SolveStatus) error) error {
	var buildCache []client.CacheOptionsEntry
	if opts.Cache != nil {
		cacheOpt := opts.Cache
		buildCache = []client.CacheOptionsEntry{
			{
				Type:  cacheOpt.Name(),
				Attrs: cacheOpt.ToAttributes(),
			},
		}
	}

	localMounts, err := opts.BuildContext.ToLocalMounts()
	if err != nil {
		return errors.Join(errors.New("failed to mount build context"), err)
	}

	frontendAttrs := map[string]string{
		"filename":                    opts.BuildContext.FileName(),
		"build-arg:SOURCE_DATE_EPOCH": "1770336000",
		"platform":                    opts.Platform,
		// this will be done using syft explicitly
		// as this should not rely on a upstream image
		// "attest:sbom":                 "",
	}

	utils.MergeMapWithPrefix("label:", frontendAttrs, opts.Labels)
	utils.MergeMapWithPrefix("build-arg:", frontendAttrs, opts.BuildArgs)

	solveOpts := client.SolveOpt{
		CacheExports: buildCache,
		CacheImports: buildCache,
		Exports: []client.ExportEntry{
			{
				Type: "docker",
				Attrs: map[string]string{
					"name":              opts.ImageName,
					"rewrite-timestamp": "true",
				},
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return os.Create(opts.TarFile)
				},
			},
		},
		LocalMounts:   localMounts,
		Frontend:      opts.BuildContext.FrontendType(),
		FrontendAttrs: frontendAttrs,
	}

	statusUpdates := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return statusUpdateHandler(statusUpdates)
	})

	eg.Go(func() error {
		_, err = c.buildkit.Build(ctx, solveOpts, "ContainerHive", opts.BuildContext.RunBuild, statusUpdates)
		return err
	})

	return eg.Wait()
}
