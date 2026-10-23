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
	Create       plumcliconfig.Cmdparams `yaml:"create"`
	CreateNote   plumcliconfig.Cmdparams `yaml:"create-note"`
	CreateFolder plumcliconfig.Cmdparams `yaml:"create-folder"`
	CreateTag    plumcliconfig.Cmdparams `yaml:"create-tag"`
	Read         plumcliconfig.Cmdparams `yaml:"read"`
	ReadNote     plumcliconfig.Cmdparams `yaml:"read-note"`
	ReadFolder   plumcliconfig.Cmdparams `yaml:"read-folder"`
	ReadTag      plumcliconfig.Cmdparams `yaml:"read-tag"`
	Edit         plumcliconfig.Cmdparams `yaml:"edit"`
	EditNote     plumcliconfig.Cmdparams `yaml:"edit-note"`
	EditFolder   plumcliconfig.Cmdparams `yaml:"edit-folder"`
	EditTag      plumcliconfig.Cmdparams `yaml:"edit-tag"`
	Delete       plumcliconfig.Cmdparams `yaml:"delete"`
	DeleteNote   plumcliconfig.Cmdparams `yaml:"delete-note"`
	DeleteFolder plumcliconfig.Cmdparams `yaml:"delete-folder"`
	DeleteTag    plumcliconfig.Cmdparams `yaml:"delete-tag"`
	Move         plumcliconfig.Cmdparams `yaml:"move"`
	Show         plumcliconfig.Cmdparams `yaml:"show"`
	ShowFinder   plumcliconfig.Cmdparams `yaml:"show-finder"`
	ShowGraph    plumcliconfig.Cmdparams `yaml:"show-graph"`
	Root         plumcliconfig.Cmdparams `yaml:"root"`
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
