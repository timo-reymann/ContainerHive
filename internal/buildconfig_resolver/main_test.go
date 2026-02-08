package buildconfig_resolver

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timo-reymann/ContainerHive/pkg/model"
)

func TestForTag(t *testing.T) {
	tests := map[string]struct {
		image    *model.Image
		tag      *model.Tag
		expected *ResolvedBuildValues
	}{
		"empty image and tag": {
			image: &model.Image{},
			tag:   &model.Tag{},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{},
				Versions:  model.Versions{},
			},
		},
		"image with versions only": {
			image: &model.Image{
				Versions: model.Versions{
					"python": "3.11",
					"pip":    "23.0",
				},
			},
			tag: &model.Tag{},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{},
				Versions: model.Versions{
					"python": "3.11",
					"pip":    "23.0",
				},
			},
		},
		"image with build args only": {
			image: &model.Image{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:latest",
					"WORKDIR":    "/app",
				},
			},
			tag: &model.Tag{},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:latest",
					"WORKDIR":    "/app",
				},
				Versions: model.Versions{},
			},
		},
		"tag with versions only": {
			image: &model.Image{},
			tag: &model.Tag{
				Versions: model.Versions{
					"node": "20.0.0",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{},
				Versions: model.Versions{
					"node": "20.0.0",
				},
			},
		},
		"tag with build args only": {
			image: &model.Image{},
			tag: &model.Tag{
				BuildArgs: model.BuildArgs{
					"BUILD_TYPE": "release",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{
					"BUILD_TYPE": "release",
				},
				Versions: model.Versions{},
			},
		},
		"tag versions override image versions": {
			image: &model.Image{
				Versions: model.Versions{
					"python": "3.10",
					"pip":    "22.0",
				},
			},
			tag: &model.Tag{
				Versions: model.Versions{
					"python": "3.11",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{},
				Versions: model.Versions{
					"python": "3.11", // overridden by tag
					"pip":    "22.0", // from image
				},
			},
		},
		"image build args override tag build args": {
			image: &model.Image{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:3.18",
					"EXTRA_ARG":  "value",
				},
			},
			tag: &model.Tag{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:latest",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:3.18", // image overrides tag
					"EXTRA_ARG":  "value",       // from image
				},
				Versions: model.Versions{},
			},
		},
		"complex merge scenario": {
			image: &model.Image{
				Versions: model.Versions{
					"python": "3.10",
					"poetry": "1.5.0",
				},
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:3.18",
					"WORKDIR":    "/app",
				},
			},
			tag: &model.Tag{
				Versions: model.Versions{
					"python": "3.11",
					"pip":    "23.0",
				},
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:latest",
					"BUILD_TYPE": "release",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:3.18", // image overrides tag
					"WORKDIR":    "/app",        // from image
					"BUILD_TYPE": "release",     // from tag
				},
				Versions: model.Versions{
					"python": "3.11",  // tag overrides image
					"poetry": "1.5.0", // from image
					"pip":    "23.0",  // from tag
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := ForTag(tc.image, tc.tag)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("ForTag() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestForTagVariant(t *testing.T) {
	tests := map[string]struct {
		image    *model.Image
		variant  *model.ImageVariant
		tag      *model.Tag
		expected *ResolvedBuildValues
	}{
		"empty image, variant, and tag": {
			image:   &model.Image{},
			variant: &model.ImageVariant{},
			tag:     &model.Tag{},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{},
				Versions:  model.Versions{},
			},
		},
		"variant with versions only": {
			image: &model.Image{},
			variant: &model.ImageVariant{
				Versions: model.Versions{
					"nodejs": "20.0.0",
				},
			},
			tag: &model.Tag{},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{},
				Versions: model.Versions{
					"nodejs": "20.0.0",
				},
			},
		},
		"variant with build args only": {
			image: &model.Image{},
			variant: &model.ImageVariant{
				BuildArgs: model.BuildArgs{
					"VARIANT_ARG": "value",
				},
			},
			tag: &model.Tag{},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{
					"VARIANT_ARG": "value",
				},
				Versions: model.Versions{},
			},
		},
		"variant versions override tag and image versions": {
			image: &model.Image{
				Versions: model.Versions{
					"python": "3.10",
					"poetry": "1.5.0",
				},
			},
			variant: &model.ImageVariant{
				Versions: model.Versions{
					"python": "3.11",
					"nodejs": "20.0.0",
				},
			},
			tag: &model.Tag{
				Versions: model.Versions{
					"python": "3.9",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{},
				Versions: model.Versions{
					"python": "3.11",   // variant overrides all
					"poetry": "1.5.0",  // from image
					"nodejs": "20.0.0", // from variant
				},
			},
		},
		"variant build args override image build args": {
			image: &model.Image{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:3.18",
					"WORKDIR":    "/app",
				},
			},
			variant: &model.ImageVariant{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE":  "ubuntu:22.04",
					"VARIANT_ARG": "value",
				},
			},
			tag: &model.Tag{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:latest",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE":  "ubuntu:22.04", // variant overrides all
					"WORKDIR":     "/app",         // from image
					"VARIANT_ARG": "value",        // from variant
				},
				Versions: model.Versions{},
			},
		},
		"complex three-way merge": {
			image: &model.Image{
				Versions: model.Versions{
					"python": "3.10",
					"poetry": "1.5.0",
					"pip":    "22.0",
				},
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:3.18",
					"WORKDIR":    "/app",
					"USER":       "appuser",
				},
			},
			variant: &model.ImageVariant{
				Versions: model.Versions{
					"python": "3.11",
					"nodejs": "20.0.0",
				},
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE":  "ubuntu:22.04",
					"VARIANT_ARG": "variant_value",
				},
			},
			tag: &model.Tag{
				Versions: model.Versions{
					"python": "3.9",
					"go":     "1.21",
				},
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE": "alpine:latest",
					"BUILD_TYPE": "release",
				},
			},
			expected: &ResolvedBuildValues{
				BuildArgs: model.BuildArgs{
					"BASE_IMAGE":  "ubuntu:22.04",  // variant overrides all
					"WORKDIR":     "/app",          // from image
					"USER":        "appuser",       // from image
					"BUILD_TYPE":  "release",       // from tag
					"VARIANT_ARG": "variant_value", // from variant
				},
				Versions: model.Versions{
					"python": "3.11",   // variant overrides tag and image
					"poetry": "1.5.0",  // from image
					"pip":    "22.0",   // from image
					"go":     "1.21",   // from tag
					"nodejs": "20.0.0", // from variant
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := ForTagVariant(tc.image, tc.variant, tc.tag)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("ForTagVariant() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}
