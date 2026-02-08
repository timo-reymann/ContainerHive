package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/timo-reymann/ContainerHive/internal/buildkit"
	"github.com/timo-reymann/ContainerHive/internal/buildkit/build_context"
	"github.com/timo-reymann/ContainerHive/internal/buildkit/cache"
	containerStructureTest "github.com/timo-reymann/ContainerHive/internal/container_structure_test"
	"github.com/timo-reymann/ContainerHive/internal/docker"
	"github.com/timo-reymann/ContainerHive/internal/syft"
	"github.com/timo-reymann/ContainerHive/pkg/discovery"
	"github.com/timo-reymann/ContainerHive/pkg/rendering"
)

const (
	// Matches hack/docker-compose.yml buildkitd service
	buildkitAddr = "tcp://127.0.0.1:8502"

	// Matches hack/garage/init.sh S3 cache configuration
	// Note: Use docker-compose service name 'garage' since buildkitd runs in container
	s3Endpoint  = "http://garage:3900"
	s3Bucket    = "buildkit-cache"
	s3Region    = "garage"
	s3AccessKey = "GK31337cafe000000000000000"
	s3SecretKey = "1337cafe0000000000000000000000000000000000000000000000000000dead"

	imageName  = "ch-smoke-test:latest"
	sbomFormat = "syft-json"
)

var platform = "linux/" + runtime.GOARCH

const dockerfile = `FROM alpine:latest
RUN echo hello > /hello.txt
`

const testConfig = `schemaVersion: 2.0.0

fileExistenceTests:
  - name: 'hello.txt exists'
    path: '/hello.txt'
    shouldExist: true
`

func main() {
	ctx := context.TODO()
	project, err := discovery.DiscoverProject(ctx, "example")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Discovered project: %v", project)

	if err := rendering.RenderProject(ctx, project, "example/dist"); err != nil {
		log.Fatal(err)
	}

	tmpDir, err := os.MkdirTemp("", "ch-smoke-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write dummy Dockerfile
	buildCtxDir := filepath.Join(tmpDir, "build")
	if err := os.MkdirAll(buildCtxDir, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(buildCtxDir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		log.Fatal(err)
	}

	tarFile := filepath.Join(tmpDir, "image.tar")
	testDefPath := filepath.Join(tmpDir, "test.yml")
	reportFile := filepath.Join(tmpDir, "report.xml")
	sbomFile := filepath.Join(tmpDir, "sbom.json")

	// Write container structure test config
	if err := os.WriteFile(testDefPath, []byte(testConfig), 0644); err != nil {
		log.Fatal(err)
	}

	// Step 1: Build image with BuildKit
	log.Println("Step 1: Building image with BuildKit (with S3 cache)...")
	bkClient, err := buildkit.NewClient(ctx, buildkitAddr)
	if err != nil {
		log.Fatalf("Failed to connect to BuildKit at %s: %v", buildkitAddr, err)
	}
	defer bkClient.Close()

	version, err := bkClient.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get BuildKit version: %v", err)
	}
	log.Printf("BuildKit version: %s", version)

	// Configure S3 cache (matches hack/docker-compose.yml garage service)
	s3Cache := &cache.S3BuildKitCache{
		EndpointUrl:     s3Endpoint,
		Bucket:          s3Bucket,
		Region:          s3Region,
		AccessKeyId:     s3AccessKey,
		SecretAccessKey: s3SecretKey,
		UsePathStyle:    true,
		CacheKey:        imageName,
	}
	log.Printf("S3 cache configured: endpoint=%s, bucket=%s", s3Endpoint, s3Bucket)

	err = bkClient.Build(ctx, &buildkit.BuildOpts{
		ImageName: imageName,
		Platform:  platform,
		TarFile:   tarFile,
		Cache:     s3Cache,
		BuildContext: &build_context.DockerfileBuildContext{
			Root: buildCtxDir,
		},
	}, func(ch chan *client.SolveStatus) error {
		d, err := progressui.NewDisplay(os.Stdout, progressui.TtyMode)
		if err != nil {
			// If an error occurs while attempting to create the tty display,
			// fallback to using plain mode on stdout (in contrast to stderr).
			d, _ = progressui.NewDisplay(os.Stdout, progressui.PlainMode)
		}
		// not using shared context to not disrupt display but let is finish reporting errors
		_, err = d.UpdateFrom(context.TODO(), ch)
		return err
	})
	if err != nil {
		log.Fatalf("Build failed: %v", err)
	}
	log.Printf("Image built: %s -> %s", imageName, tarFile)

	// Step 2: Run container structure tests
	log.Println("Step 2: Running container structure tests...")
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	runner := &containerStructureTest.TestRunner{
		TestDefinitionPaths: []string{testDefPath},
		Image:               tarFile,
		Platform:            platform,
		ReportFile:          reportFile,
		DockerClient:        dockerClient,
	}
	if err := runner.Run(); err != nil {
		log.Fatalf("Container structure tests failed: %v", err)
	}
	log.Printf("Test report written to %s", reportFile)

	// Step 3: Generate SBOM
	log.Println("Step 3: Generating SBOM...")
	sbomTool, err := syft.NewSBOMImageTool()
	if err != nil {
		log.Fatalf("Failed to create SBOM tool: %v", err)
	}

	generatedSBOM, err := sbomTool.GenerateSBOM(ctx, tarFile)
	if err != nil {
		log.Fatalf("SBOM generation failed: %v", err)
	}

	sbomBytes, err := sbomTool.SerializeSBOM(generatedSBOM, sbomFormat)
	if err != nil {
		log.Fatalf("SBOM serialization failed: %v", err)
	}

	if err := os.WriteFile(sbomFile, sbomBytes, 0644); err != nil {
		log.Fatalf("Failed to write SBOM: %v", err)
	}
	log.Printf("SBOM written to %s (%s)", sbomFile, sbomFormat)

	fmt.Println("All steps completed successfully.")
}
