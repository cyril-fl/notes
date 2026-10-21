package config

import (
	"fmt"
	"os"
	"strings"

	plumcliconfig "github.com/cyril-fl/plum/cli/config"
	plummillload "github.com/cyril-fl/plum/services/factory/millload"
	"github.com/joho/godotenv"
)

var Current config

type config struct {
	Name     string   `yaml:"name"`
	Commands commands `yaml:"commands"`
}

type commands struct {
	Create plumcliconfig.Cmdparams `yaml:"create"`
	Read   plumcliconfig.Cmdparams `yaml:"read"`
	Update plumcliconfig.Cmdparams `yaml:"update"`
	Delete plumcliconfig.Cmdparams `yaml:"delete"`
	Move   plumcliconfig.Cmdparams `yaml:"move"`
	Finder plumcliconfig.Cmdparams `yaml:"finder"`
	Graph  plumcliconfig.Cmdparams `yaml:"graph"`
	Root   plumcliconfig.Cmdparams `yaml:"root"`
}

func load() (*config, error) {
	file := strings.TrimSpace(os.Getenv("CONFIG_FILE"))

	loader := plummillload.New(file)
	current := loader.LoadFromFile[config]()

	if loader.HasErr() {
		return nil, fmt.Errorf("%w", loader.GetErr())
	}

	return current, nil
}

func init() {
	godotenv.Load()
	loaded, err := load()
	if err != nil {
		panic(err)
	}
	Current = *loaded
}
