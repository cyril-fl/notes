package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	plumcliconfig "github.com/cyril-fl/plum/cli/config"
	plummillload "github.com/cyril-fl/plum/services/factory/millload"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// embedded is the config built into the binary, so it runs from any folder; CONFIG_FILE overrides it.
//
//go:embed config.yaml
var embedded []byte

var Current config

type config struct {
	Name     string   `yaml:"name"`
	Commands commands `yaml:"commands"`
	Trash    trash    `yaml:"trash"`
}

type trash struct {
	RetentionDays int `yaml:"retention-days"` // <= 0: deleted items are never purged
}

type commands struct {
	Create        plumcliconfig.Cmdparams `yaml:"create"`
	CreateNote    plumcliconfig.Cmdparams `yaml:"create-note"`
	CreateFolder  plumcliconfig.Cmdparams `yaml:"create-folder"`
	CreateTag     plumcliconfig.Cmdparams `yaml:"create-tag"`
	Read          plumcliconfig.Cmdparams `yaml:"read"`
	ReadNote      plumcliconfig.Cmdparams `yaml:"read-note"`
	ReadFolder    plumcliconfig.Cmdparams `yaml:"read-folder"`
	ReadTag       plumcliconfig.Cmdparams `yaml:"read-tag"`
	Edit          plumcliconfig.Cmdparams `yaml:"edit"`
	EditNote      plumcliconfig.Cmdparams `yaml:"edit-note"`
	EditFolder    plumcliconfig.Cmdparams `yaml:"edit-folder"`
	EditTag       plumcliconfig.Cmdparams `yaml:"edit-tag"`
	Delete        plumcliconfig.Cmdparams `yaml:"delete"`
	DeleteNote    plumcliconfig.Cmdparams `yaml:"delete-note"`
	DeleteFolder  plumcliconfig.Cmdparams `yaml:"delete-folder"`
	DeleteTag     plumcliconfig.Cmdparams `yaml:"delete-tag"`
	Move          plumcliconfig.Cmdparams `yaml:"move"`
	Show          plumcliconfig.Cmdparams `yaml:"show"`
	ShowFinder    plumcliconfig.Cmdparams `yaml:"show-finder"`
	ShowGraph     plumcliconfig.Cmdparams `yaml:"show-graph"`
	Trash         plumcliconfig.Cmdparams `yaml:"trash"`
	Restore       plumcliconfig.Cmdparams `yaml:"restore"`
	Search        plumcliconfig.Cmdparams `yaml:"search"`
	Recent        plumcliconfig.Cmdparams `yaml:"recent"`
	SyncTags      plumcliconfig.Cmdparams `yaml:"sync-tags"`
	SyncTitles    plumcliconfig.Cmdparams `yaml:"sync-titles"`
	SyncLinks     plumcliconfig.Cmdparams `yaml:"sync-links"`
	SyncEntities  plumcliconfig.Cmdparams `yaml:"sync-entities"`
	Export        plumcliconfig.Cmdparams `yaml:"export"`
	Import        plumcliconfig.Cmdparams `yaml:"import"`
	Serve         plumcliconfig.Cmdparams `yaml:"serve"`
	Login         plumcliconfig.Cmdparams `yaml:"login"`
	Logout        plumcliconfig.Cmdparams `yaml:"logout"`
	User          plumcliconfig.Cmdparams `yaml:"user"`
	UserAdd       plumcliconfig.Cmdparams `yaml:"user-add"`
	ResetPassword plumcliconfig.Cmdparams `yaml:"reset-password"`
	Root          plumcliconfig.Cmdparams `yaml:"root"`
}

func load() (*config, error) {
	file := strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	if file == "" {
		var current config
		if err := yaml.Unmarshal(embedded, &current); err != nil {
			return nil, fmt.Errorf("embedded config: %w", err)
		}
		return &current, nil
	}

	loader := plummillload.New(file)
	current := loader.LoadFromFile[config]()

	if loader.HasErr() {
		return nil, fmt.Errorf("%w", loader.GetErr())
	}

	return current, nil
}

// loadEnv reads the user file first: godotenv never overwrites a value, so the .env of whatever
// project the CLI runs from cannot point it at another database.
func loadEnv() {
	if home, err := os.UserHomeDir(); err == nil {
		godotenv.Load(filepath.Join(home, ".config", "notes", ".env"))
	}
	godotenv.Load()
}

func init() {
	loadEnv()
	loaded, err := load()
	if err != nil {
		panic(err)
	}
	Current = *loaded
}
