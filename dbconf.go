package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/lib/pq"
	"os"
	"path/filepath"
)

// global options. available to any subcommands.
var dbPath = flag.String("path", "db", "folder containing db info")
var dbEnv = flag.String("env", "development", "which DB environment to use")

// DBDriver encapsulates the info needed to work with
// a specific database driver
type DBDriver struct {
	Name    string
	OpenStr string
	Import  string
}

type DBConf struct {
	MigrationsDir string
	Env           string
	Driver        DBDriver
}

// default helper - makes a DBConf from the dbPath and dbEnv flags
func MakeDBConf() (*DBConf, error) {
	return makeDBConfDetails(*dbPath, *dbEnv)
}

// extract configuration details from the given file
func makeDBConfDetails(p, env string) (*DBConf, error) {

	cfgFile := filepath.Join(p, "dbconf.yml")

	f, err := yaml.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	drv, err := f.Get(fmt.Sprintf("%s.driver", env))
	if err != nil {
		return nil, err
	}

	open, err := f.Get(fmt.Sprintf("%s.open", env))
	if err != nil {
		return nil, err
	}
	open = os.ExpandEnv(open)

	// Automatically parse postgres urls
	if drv == "postgres" {

		// Assumption: If we can parse the URL, we should
		if parsedURL, err := pq.ParseURL(open); err == nil && parsedURL != "" {
			open = parsedURL
		}
	}

	d := NewDBDriver(drv, open)

	// XXX: allow an import entry to override DBDriver.Import

	if !d.IsValid() {
		return nil, errors.New(fmt.Sprintf("Invalid DBConf: %v", d))
	}

	return &DBConf{
		MigrationsDir: filepath.Join(p, "migrations"),
		Env:           env,
		Driver:        d,
	}, nil
}

// Create a new DBDriver and populate driver specific
// fields for drivers that we know about.
// Further customization may be done in NewDBConf
func NewDBDriver(name, open string) DBDriver {

	d := DBDriver{
		Name:    name,
		OpenStr: open,
	}

	switch name {
	case "postgres":
		d.Import = "github.com/lib/pq"

	case "mymysql":
		d.Import = "github.com/ziutek/mymysql/godrv"
	}

	return d
}

// ensure we have enough info about this driver
func (drv *DBDriver) IsValid() bool {
	if len(drv.Import) == 0 {
		return false
	}

	return true
}
