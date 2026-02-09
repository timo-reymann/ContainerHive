package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/timo-reymann/ContainerHive/internal/utils"
	"zotregistry.dev/zot/v2/pkg/api"
	"zotregistry.dev/zot/v2/pkg/api/config"
)

// ZotRegistry is an embedded OCI registry for local development builds.
// It runs zot in-process on a random port.
type ZotRegistry struct {
	ctlr    *api.Controller
	dataDir string
	port    int
}

// NewZotRegistry creates a new ZotRegistry instance.
func NewZotRegistry() *ZotRegistry {
	return &ZotRegistry{}
}

func (z *ZotRegistry) Start(ctx context.Context) error {
	dataDir, err := os.MkdirTemp("", "containerhive-zot-*")
	if err != nil {
		return errors.Join(errors.New("failed to create zot data directory"), err)
	}
	z.dataDir = dataDir

	conf := config.New()
	conf.HTTP.Address = "127.0.0.1"
	conf.HTTP.Port = "0"
	conf.Storage.RootDirectory = dataDir
	conf.Storage.GC = false
	conf.Storage.Dedupe = false
	conf.Log = &config.LogConfig{
		Level:  "error",
		Output: "",
	}

	z.ctlr = api.NewController(conf)

	if err := z.ctlr.Init(); err != nil {
		os.RemoveAll(dataDir)
		return errors.Join(errors.New("failed to initialize zot"), err)
	}

	go func() {
		if err := z.ctlr.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Run returned unexpectedly; nothing to do since Stop will handle cleanup
		}
	}()

	if err := z.waitForReady(ctx); err != nil {
		z.ctlr.Shutdown()
		os.RemoveAll(dataDir)
		return errors.Join(errors.New("zot failed to become ready"), err)
	}

	return nil
}

func (z *ZotRegistry) waitForReady(ctx context.Context) error {
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return errors.New("timeout waiting for zot to start")
		case <-ticker.C:
			port := z.ctlr.GetPort()
			if port <= 0 {
				continue
			}
			url := fmt.Sprintf("http://127.0.0.1:%d/v2/", port)
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				z.port = port
				return nil
			}
		}
	}
}

func (z *ZotRegistry) Stop(_ context.Context) error {
	if z.ctlr != nil {
		z.ctlr.Shutdown()
	}
	if z.dataDir != "" {
		os.RemoveAll(z.dataDir)
	}
	return nil
}

func (z *ZotRegistry) Address() string {
	return fmt.Sprintf("127.0.0.1:%d", z.port)
}

func (z *ZotRegistry) IsLocal() bool {
	return true
}

func (z *ZotRegistry) Push(_ context.Context, imageName, tag, ociTarPath string) error {
	tmpDir, err := os.MkdirTemp("", "oci-push-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := utils.ExtractTar(ociTarPath, tmpDir); err != nil {
		return errors.Join(errors.New("failed to extract OCI tar for push"), err)
	}

	layoutPath, err := layout.FromPath(tmpDir)
	if err != nil {
		return errors.Join(errors.New("failed to read OCI layout"), err)
	}

	idx, err := layoutPath.ImageIndex()
	if err != nil {
		return err
	}

	idxManifest, err := idx.IndexManifest()
	if err != nil {
		return err
	}

	if len(idxManifest.Manifests) == 0 {
		return errors.New("no manifests in OCI layout")
	}

	img, err := layoutPath.Image(idxManifest.Manifests[0].Digest)
	if err != nil {
		return errors.Join(errors.New("failed to read image from layout"), err)
	}

	ref, err := name.NewTag(fmt.Sprintf("%s/%s:%s", z.Address(), imageName, tag), name.Insecure)
	if err != nil {
		return errors.Join(errors.New("invalid image reference"), err)
	}

	if err := remote.Write(ref, img); err != nil {
		return errors.Join(errors.New("failed to push image to zot"), err)
	}

	return nil
}
