package docker

import (
	"context"
	"errors"
	"os"

	dockerClient "github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

type Client struct {
	docker *dockerClient.Client
}

func (c *Client) Close() error {
	return c.docker.Close()
}

func NewClient() (*Client, error) {
	docker, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Client{
		docker,
	}, nil
}

func (c *Client) LoadImageFromTar(_ context.Context, tarPath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "oci-layout-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	if err := extractTar(tarPath, tmpDir); err != nil {
		return "", errors.Join(errors.New("failed to extract OCI tar"), err)
	}

	layoutPath, err := layout.FromPath(tmpDir)
	if err != nil {
		return "", errors.Join(errors.New("failed to read OCI layout"), err)
	}

	idx, err := layoutPath.ImageIndex()
	if err != nil {
		return "", err
	}

	idxManifest, err := idx.IndexManifest()
	if err != nil {
		return "", err
	}

	if len(idxManifest.Manifests) == 0 {
		return "", errors.New("no manifests in OCI layout")
	}

	imageName, ok := idxManifest.Manifests[0].Annotations["io.containerd.image.name"]
	if !ok || imageName == "" {
		return "", errors.New("no image name annotation in OCI index")
	}

	img, err := layoutPath.Image(idxManifest.Manifests[0].Digest)
	if err != nil {
		return "", errors.Join(errors.New("failed to read image from layout"), err)
	}

	tag, err := name.NewTag(imageName)
	if err != nil {
		return "", errors.Join(errors.New("invalid image name"), err)
	}

	if _, err := daemon.Write(tag, img); err != nil {
		return "", errors.Join(errors.New("failed to load image into Docker"), err)
	}

	return imageName, nil
}
