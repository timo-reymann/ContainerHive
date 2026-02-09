package model

type VariantConfig struct {
	Name      string    `yaml:"name" json:"name" jsonschema:"Name of the variant"`
	TagSuffix string    `yaml:"tag_suffix" json:"tag_suffix" jsonschema:"Suffix to append to the tag name for this variant"`
	Versions  Versions  `yaml:"versions" json:"versions,omitempty" jsonschema:"Versions to use for this variant"`
	BuildArgs BuildArgs `yaml:"build_args" json:"build_args,omitempty" jsonschema:"Build args to add for this variant"`
}

type ImageDefinitionConfig struct {
	Tags      []*Tag          `yaml:"tags" json:"tags" jsonschema:"Tags to create for this image"`
	Variants  []VariantConfig `yaml:"variants" json:"variants,omitempty" jsonschema:"Variants to create for this image"`
	Versions  Versions        `yaml:"versions" json:"versions,omitempty" jsonschema:"Versions to use for this image"`
	BuildArgs BuildArgs       `yaml:"build_args" json:"build_args,omitempty" jsonschema:"Build args to add for this image"`
	DependsOn []string        `yaml:"depends_on" json:"depends_on,omitempty" jsonschema:"Names of other images in this project that must be built before this image"`
}

type HiveProjectConfig struct {
}
