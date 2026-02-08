package buildconfig_resolver

import "github.com/timo-reymann/ContainerHive/pkg/model"

type ResolvedBuildValues struct {
	BuildArgs model.BuildArgs
	Versions  model.Versions
}

func ForTag(image *model.Image, tag *model.Tag) *ResolvedBuildValues {
	resolved := &ResolvedBuildValues{
		BuildArgs: tag.BuildArgs,
		Versions:  image.Versions,
	}

	if resolved.Versions == nil {
		resolved.Versions = make(model.Versions)
	}

	if resolved.BuildArgs == nil {
		resolved.BuildArgs = make(model.BuildArgs)
	}

	for k, v := range tag.Versions {
		resolved.Versions[k] = v
	}

	for k, v := range image.BuildArgs {
		resolved.BuildArgs[k] = v
	}

	return resolved
}

func ForTagVariant(image *model.Image, variant *model.ImageVariant, tag *model.Tag) *ResolvedBuildValues {
	resolved := ForTag(image, tag)

	for k, v := range variant.Versions {
		resolved.Versions[k] = v
	}

	for k, v := range variant.BuildArgs {
		resolved.BuildArgs[k] = v
	}

	return resolved
}
