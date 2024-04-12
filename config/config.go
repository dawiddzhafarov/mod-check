package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"golang.org/x/crypto/ssh/terminal"
)

// Config represents configuration of invoked command
type Config struct {
	Width          int
	Filter         string
	IsIncompatible bool
	TerminalWidth  int
	MaxVersions    int
}

func New() *Config {
	cfg := new(Config)
	flag.IntVar(&cfg.MaxVersions, "max-versions", 15, "Specify maximum number of versions to display for each dependency.")
	flag.StringVar(&cfg.Filter, "filter", "major,minor,patch", "Filter out the version types for display, available values: `major`, `minor`, `patch`. In order to use two filters, separate them with a comma (,). By default, all version types are included.")
	flag.BoolVar(&cfg.IsIncompatible, "show-incompatible", false, "Show incompatible versions, disabled by default.")
	flag.Parse()
	if err := cfg.validateFlags(); err != nil {
		log.Fatalf("%v", err)
	}

	fd := int(os.Stdin.Fd())
	width, _, err := terminal.GetSize(fd)
	if err != nil {
		log.Fatalf("%v", err)
	}

	cfg.Width = width

	return cfg
}

func (c *Config) validateFlags() error {
	if c.MaxVersions <= 0 || c.MaxVersions > 1000 {
		return fmt.Errorf("`max-version` can be between 1 and 1000")
	}
	if len(c.Filter) != 0 {
		flags := strings.Split(c.Filter, ",")
		allowed := []string{"major", "minor", "patch"}
		for _, f := range flags {
			if !contains(allowed, f) {
				return fmt.Errorf("`filter` flag can be made up only from `major`, `minor` and `patch` values")
			}
		}
	}

	return nil
}

func contains(slice []string, sub string) bool {
	for _, el := range slice {
		if el == sub {
			return true
		}
	}
	return false
}
