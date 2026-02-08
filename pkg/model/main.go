package model

type Versions map[string]string

type BuildArgs map[string]string

type Tag struct {
	Name     string
	Versions Versions
	BuildArgs
}

type Image struct {
	Identifier         string
	Name               string
	RootDir            string
	RootFSDir          string
	TestConfigFilePath string
	DefinitionFilePath string
	Versions           Versions
	BuildArgs          BuildArgs `yaml:"build_args"`
	Tags               map[string]*Tag
	Variants           map[string]*ImageVariant
}

type ImageVariant struct {
	Name               string
	RootDir            string
	RootFSDir          string
	TagSuffix          string `yaml:"tag_suffix"`
	TestConfigFilePath string
	Versions           Versions
	BuildArgs          BuildArgs `yaml:"build_args"`
}

type ContainerHiveProject struct {
	RootDir            string
	ConfigFilePath     string
	ImagesByIdentifier map[string]*Image
	ImagesByName       map[string][]*Image
}
