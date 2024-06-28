package entitiesmap

import (
	"os"

	"gopkg.in/yaml.v3"
)

type entityMap struct {
	Sensors []Entity `yaml:"sensors"`
}

type Entity struct {
	Name     string `yaml:"name"`
	EntityId string `yaml:"entity_id"`
}

func Read(file string) ([]Entity, error) {
	var c entityMap
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return c.Sensors, err
	}
	err = yaml.Unmarshal(yamlFile, &c)

	return c.Sensors, err
}
