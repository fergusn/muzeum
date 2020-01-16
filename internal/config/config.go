package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"

	registry "github.com/docker/distribution/configuration"
)

// Configuration root
type Configuration struct {
	Repositories []Repository
	Storage      registry.Storage
	Certificate  Certificate `json:"certificate"`
}

// Repository configuration
type Repository struct {
	Name   string
	Path   string
	Host   string
	Plugin map[string]map[string]interface{} `yaml:",inline"`
}

// Certificate configuration
type Certificate struct {
	Crt string `json:"crt"`
	Key string `json:"key"`
}

// Parse reads a yaml configuration file
func Parse(path string) *Configuration {
	config := &Configuration{
		Repositories: []Repository{},
	}
	file, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	decoder := yaml.NewDecoder(file)
	decoder.Decode(config)
	return config
}
