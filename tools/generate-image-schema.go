//go:build ignore

package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/timo-reymann/ContainerHive/pkg/model"
)

func main() {
	log.Println("Generating image schema...")
	schema, err := jsonschema.For[model.ImageDefinitionConfig](&jsonschema.ForOptions{})
	if err != nil {
		log.Fatal("failed to generate schema", err)
	}

	schema.ID = "https://container-hive.timo-reymann.de/schemas/image.schema.json"
	schema.Title = "Image definition"
	schema.Description = "Image definition configuration schema for ContainerHive."

	log.Println("Writing schema to file...")
	indented, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal("failed to marshal indented schema", err)
	}

	err = os.WriteFile("schemas/image.schema.json", indented, 0644)
	if err != nil {
		log.Fatal("failed to write schema file", err)
	}
}
