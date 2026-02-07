---
id: 002
status: accepted
date: 2025-06-02
---

# SBOM generation

## Context and Problem Statement

ContainerHive needs to generate Software Bills of Materials (SBOMs) for built container images. BuildKit has built-in
SBOM attestation support, but it relies on running a Docker container (by default pulled from Docker Hub) to perform the
scan. In enterprise environments with restricted network access - a primary target audience for ContainerHive - this
creates friction. How should ContainerHive integrate SBOM generation in a way that works seamlessly across environments
without external runtime dependencies?

## Decision Drivers

* Must work in network-restricted enterprise environments potentially without access to Docker Hub
* Should not require Docker-in-Docker or access to a container runtime for SBOM generation
* Seamless, built-in experience - SBOM generation should feel like a native part of the tool
* Support for local image tar files to enable offline workflows without loading into a registry first
* Well-documented and maintainable integration

## Considered Options

* Option 1: Use BuildKit's built-in SBOM attestation
* Option 2: Use Syft as a Go library
* Option 3: Use Syft as a Docker container

## Decision Outcome

Chosen option: "Option 2 - Use Syft as a Go library", because it provides a well-documented Go API that integrates
directly into ContainerHive without any external runtime dependencies. It operates on local OCI tar files produced by
BuildKit, avoiding the need for a container runtime or registry access during SBOM generation. This keeps the tool
self-contained and works reliably in restricted environments.

## Pros and Cons of the Options

### Option 1: Use BuildKit's built-in SBOM attestation

BuildKit supports SBOM generation via the `attest:sbom` frontend option. It runs a scanner image (default:
`docker/buildkit-syft-scanner` from Docker Hub) as a container during the build process.

* Good, because it is built into BuildKit with no additional integration effort
* Good, because the SBOM is attached as a build attestation, following emerging OCI standards
* Good, because the scanner image could be customized to point to an internal registry
* Bad, because it pulls a Docker Hub image by default, which fails in network-restricted enterprise environments
* Bad, because even with a custom scanner image, it effectively requires a Docker-in-Docker setup
* Bad, because Docker-in-Docker is difficult to configure in Kubernetes-based CI environments (security policies,
  privileged containers)
* Bad, because the complexity is hidden - failures in the attestation step are harder to diagnose than explicit SBOM
  generation
* Bad, because ties SBOM format and tooling choices to what BuildKit's attestation pipeline supports

### Option 2: Use Syft as a Go library

Import `github.com/anchore/syft` as a Go dependency and call its API directly to scan OCI tar files and produce SBOMs in
multiple formats.

* Good, because Syft has a well-documented Go API
* Good, because it integrates as a library - no subprocess management, no container runtime needed
* Good, because it operates on local `image.tar` files, enabling fully offline SBOM generation
* Good, because it supports multiple output formats (syft-json, CycloneDX, SPDX, etc.) with a simple encoder API
* Good, because no network access required at SBOM generation time - works in air-gapped environments
* Good, because ContainerHive controls when and how the SBOM is generated, independent of the build step
* Bad, because it adds a significant dependency tree (Syft pulls in SQLite for rpmdb, multiple archive libraries, etc.)
* Bad, because SBOM results may differ slightly from BuildKit's attestation scanner since they are separate tools

### Option 3: Use Syft as a Docker container

Run the official Syft Docker container image to scan images, with the user providing the container.

* Good, because uses the official Syft distribution without embedding anything
* Good, because version management is delegated to the user
* Bad, because it requires a container runtime available during SBOM generation
* Bad, because it reintroduces the same network-restriction problems as Option 1 (pulling images)
* Bad, because it contradicts the goal of a seamless, all-in-one solution
* Bad, because subprocess/container orchestration adds complexity for little benefit over the library approach

## Links

* [Syft GitHub repository](https://github.com/anchore/syft)
* [Syft Go library documentation](https://pkg.go.dev/github.com/anchore/syft/syft)
* [BuildKit SBOM attestation](https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md)
* Relates to [ADR-001: BuildKit Integration](001-buildkit-integration.md) - SBOM generation is explicitly separated from
  the BuildKit build step

<!-- markdownlint-disable-file MD013 -->