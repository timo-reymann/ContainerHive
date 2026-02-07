---
id: 003
status: accepted
date: 2025-06-02
---

# Accept Docker daemon dependency for full container structure test support

## Context and Problem Statement

ContainerHive uses Google's container-structure-test to validate built images. However, container images can be
represented in multiple formats and at their core are just a series of tarball layers. Container-structure-test supports
two modes of operation with different capabilities:

- **Docker driver**: Requires a Docker daemon. Supports all test types including command tests, which execute commands
  inside a running container.
- **Tar/filesystem driver**: No daemon needed. Only supports file existence and file content tests by inspecting the
  image layers directly.

BuildKit produces OCI-format tar archives, not Docker-format images. This means that to use the Docker driver,
ContainerHive must convert the OCI tar to a Docker-compatible format and load it into a daemon. The question is whether
this conversion and Docker dependency is worth the full test coverage, or whether the limited tar-only mode is
sufficient.

## Decision Drivers

* Command tests (running commands inside the container) are essential for verifying runtime behavior, not just file
  presence
* BuildKit outputs OCI tar format, which is not directly compatible with Docker's image store
* Users should be able to test images locally without pushing to a registry first
* The Docker daemon is already commonly available in development environments and most CI can be configured with a DinD service
* Consistency with ADR-001: ContainerHive follows a "bring your own" philosophy for infrastructure
* Test behavior should be consistent regardless of environment to avoid "works on my machine" issues

## Considered Options

* Option 1: Use Docker driver with OCI-to-Docker conversion for full test support
* Option 2: Use tar/filesystem driver only, accept no command test support
* Option 3: Dynamically choose driver based on test configuration content

## Decision Outcome

Chosen option: "Option 1 - Use Docker driver with OCI-to-Docker conversion", because command tests are a critical
validation capability that verifies runtime behavior beyond static file checks. The conversion from OCI tar to Docker is
handled transparently using `go-containerregistry`, and requiring a Docker daemon is consistent with the "bring your own
infrastructure" philosophy established in ADR-001.

The implementation:

1. Detects whether the image reference is a tar file or a Docker image name
2. For tar files: extracts the OCI tar, reads the OCI layout, resolves the image name from manifest annotations, and
   loads into Docker via `daemon.Write`
3. Runs container-structure-test with the `docker` driver against the loaded image
4. Produces JUnit XML reports for CI integration

## Pros and Cons of the Options

### Option 1: Use Docker driver with OCI-to-Docker conversion

Convert OCI tar archives to Docker-compatible images using `go-containerregistry` (OCI layout parsing + `daemon.Write`),
then run container-structure-test with the Docker driver.

* Good, because all test types are supported: command tests, file existence, file content, metadata, and license tests
* Good, because command tests verify actual runtime behavior (installed packages work, entrypoints execute correctly,
  etc.)
* Good, because the OCI-to-Docker conversion is handled transparently - users pass a tar file and it just works
* Good, because `go-containerregistry` provides a clean, library-based conversion path without shelling out
* Good, because images are tested locally from the OCI tar without requiring a registry push
* Good, because it aligns with the existing "bring your own" pattern - the user provides a Docker daemon just like they
  provide a BuildKit daemon
* Good, because test behavior is identical across all environments
* Bad, because it requires a Docker daemon available at test time
* Bad, because the OCI-to-Docker load adds time and temporary disk usage
* Bad, because it introduces a dependency on `go-containerregistry` for the format conversion
* Bad, because in Kubernetes-based CI, a Docker daemon may require a DinD sidecar

### Option 2: Use tar/filesystem driver only

Run container-structure-test in tar mode, inspecting image layers directly without a container runtime.

* Good, because no Docker daemon required - works anywhere
* Good, because simpler implementation with no format conversion needed
* Good, because faster execution since no image loading or container creation
* Bad, because no command test support - cannot verify that installed software actually works
* Bad, because cannot test entrypoints, environment variable resolution, or runtime behavior
* Bad, because file existence and content tests alone provide limited confidence in image correctness
* Bad, because users would need a separate workflow to run command-based validation

### Option 3: Dynamically choose driver based on test configuration

Inspect the test definition file and select the Docker driver when command tests are present, fall back to the tar
driver when only file tests are defined.

* Good, because no Docker daemon needed for simple file-only test suites
* Good, because command tests still work when Docker is available
* Bad, because the same test suite runs through different code paths depending on its content, leading to inconsistent
  behavior across environments
* Bad, because subtle differences between drivers (e.g. file resolution, layer merging) may cause tests to pass with one
  driver but fail with the other
* Bad, because it creates a "works on my machine" problem - adding a command test to a config silently changes the
  driver and may break in environments without Docker
* Bad, because increased complexity in driver selection logic for marginal benefit

## Links

* [container-structure-test GitHub repository](https://github.com/GoogleContainerTools/container-structure-test)
* [container-structure-test driver documentation](https://github.com/GoogleContainerTools/container-structure-test#drivers)
* [go-containerregistry](https://github.com/google/go-containerregistry) - used for OCI layout parsing and Docker daemon
  loading
* Relates to [ADR-001: BuildKit Integration](001-buildkit-integration.md) - follows the same "bring your own"
  infrastructure pattern

<!-- markdownlint-disable-file MD013 -->