package docker

import (
	"errors"

	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/factory"
)

const muzeumFactory = "muzeum"

var (
	errConfiguration = errors.New("Storgae driven configuration error")
)

func init() {
	factory.Register(muzeumFactory, storageFactory{})
}

type storageFactory struct{}

func (f storageFactory) Create(parameters map[string]interface{}) (driver.StorageDriver, error) {
	if d, ok := parameters["driver"].(driver.StorageDriver); ok {
		return d, nil
	}
	return nil, errConfiguration
}
