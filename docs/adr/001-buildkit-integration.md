---
id: 001
status: accepted
date: 2025-06-02
---

# BuildKit integration for building the container images

## Context and Problem Statement

ContainerHive needs to build container images. BuildKit is the modern build engine behind Docker and provides advanced
features like efficient caching, concurrent build steps, and OCI-compliant output. The question is how ContainerHive
should integrate with BuildKit: should it embed and manage the daemon itself, shell out to `buildctl`, or expect the
user to provide a running BuildKit instance?

## Decision Drivers

* Must work for both local development and CI environments
* Should not force a specific daemon lifecycle on users (rootless, rootful, Docker-managed, remote, etc.)
* Avoid packaging and management complexity
* Leverage BuildKit's Go client library for type safety and direct gRPC integration
* Keep ContainerHive focused on orchestration, not daemon management

## Considered Options

* Option 1: Embed BuildKit daemon and run directly on the host
* Option 2: Embed daemon for Linux (CI), fall back to Docker for local
* Option 3: "Bring your own" - connect to a user-provided BuildKit daemon via endpoint (e.g. `BUILDKIT_HOST`)
* Option 4: Embed `buildctl` and run as subprocess
* Option 5: Require users to install `buildctl` locally

## Decision Outcome

Chosen option: "Option 3 - Bring your own BuildKit daemon", because it cleanly separates daemon lifecycle from
ContainerHive's responsibility while leveraging the BuildKit Go client API for direct, type-safe integration over gRPC.
The user controls how BuildKit runs (rootless, Docker socket, remote TCP, etc.) and ContainerHive connects to it via an
endpoint string, following the same convention as `buildctl` with `BUILDKIT_HOST`.

For local development, ContainerHive will attempt to connect to standard paths/URLs. CI configuration will specify the
buildkitd address explicitly. This also leaves the door open for ContainerHive to optionally spin up a buildkitd
container using the image version specified in the project-wide configuration, though this is subject to further
evaluation.

## Pros and Cons of the Options

### Option 1: Embed BuildKit daemon and run directly on the host

Run buildkitd in-process on the host machine.

* Good, because no external daemon dependency
* Good, because simplest user experience on Linux
* Bad, because only works on Linux - buildkitd cannot run natively on macOS or Windows
* Bad, because requires root or complex rootless setup managed by ContainerHive
* Bad, because tight coupling to daemon lifecycle creates failure modes unrelated to building

### Option 2: Embed daemon for Linux, use Docker for local

Hybrid approach: embedded buildkitd on Linux (CI), Docker's built-in BuildKit on developer machines.

* Good, because works on all platforms via Docker
* Good, because optimized for CI where Docker overhead is unnecessary
* Bad, because two different code paths to maintain and test
* Bad, because subtle behavioral differences between embedded and Docker-managed BuildKit
* Bad, because Docker's BuildKit integration may lag behind standalone BuildKit features

### Option 3: "Bring your own" BuildKit daemon

Connect to a user-provided BuildKit instance via endpoint. Use the BuildKit Go client library (
`github.com/moby/buildkit/client`) for direct gRPC communication.

* Good, because users choose their own daemon setup (rootless, Docker, remote, Kubernetes, etc.)
* Good, because BuildKit has a Go API that provides type-safe, direct gRPC integration without subprocess overhead
* Good, because works identically across platforms - the daemon runs wherever the user wants
* Good, because ContainerHive stays focused on orchestration rather than daemon management
* Bad, because requires users to have a running BuildKit daemon before using ContainerHive
* Bad, because the BuildKit Go API is largely undocumented, increasing maintenance burden

### Option 4: Embed `buildctl` and run as subprocess

Bundle the `buildctl` binary and invoke it via `os/exec`.

* Good, because uses the official CLI interface which is stable and documented
* Bad, because requires packaging platform-specific binaries, increasing distribution complexity
* Bad, because subprocess management adds error handling complexity (exit codes, stderr parsing)
* Bad, because loses type safety - build options become string arguments
* Bad, because harder to handle streaming build status updates

### Option 5: Require users to install `buildctl` locally

Expect `buildctl` to be on `$PATH` and shell out to it.

* Good, because zero packaging burden
* Good, because users get the exact version they want
* Bad, because adds an external dependency users must manage
* Bad, because same subprocess management downsides as Option 4
* Bad, because version mismatches between `buildctl` and the BuildKit daemon can cause subtle failures
* Bad, because poor developer experience - another tool to install and keep updated

## Links

* [BuildKit GitHub repository](https://github.com/moby/buildkit)
* [BuildKit Go client package](https://pkg.go.dev/github.com/moby/buildkit/client)

<!-- markdownlint-disable-file MD013 -->