package discovery

import (
	"errors"
	"os"
	"path/filepath"
)

var hiveConfigFileNames = []string{
	"hive.yaml",
	"hive.yml",
	"container-hive.yaml",
	"container-hive.yml",
}

func getContainerHiveConfigFile(root string) (string, error) {
	for _, name := range hiveConfigFileNames {
		path := filepath.Join(root, name)
		_, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			return "", errors.Join(errors.New("failed to state ContainerHive config file path "+path), err)
		}

		if err == nil {
			return path, nil
		}
	}

	return "", errors.New("no ContainerHive config file found")
}
