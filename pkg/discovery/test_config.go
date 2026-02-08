package discovery

import (
	"github.com/timo-reymann/ContainerHive/internal/file_resolver"
)

var testConfigFileNames = file_resolver.GetFileCandidates("test", "yml", "yaml")

func getTestConfigFilePath(root string) (string, error) {
	return file_resolver.ResolveFirstExistingFile(root, testConfigFileNames...)
}
