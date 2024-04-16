package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"golang.org/x/crypto/ssh/terminal"
)

// minEntryLength is a length of the dependency number, name and current version chain of characters used in pretty print
const minEntryLength = 60

// Config represents configuration of invoked command
type Config struct {
	Dependency       string
	Filter           string
	IsIncompatible   bool
	MaxVersions      int
	Pretty           bool
	ShowOld          bool
	Width            int
	VersionsToPrint  int
	VersionCharSpace int
}

func New() *Config {
	cfg := new(Config)

	// Parse flags and subcommands
	parseFlags(cfg)
	cfg.ShowOld = parseShowSubCmd()
	cfg.Width = getTerminalWidth()

	// validate flags
	if err := cfg.validateFlags(); err != nil {
		log.Fatalf("%v", err)
	}

	cfg.VersionCharSpace = cfg.Width - minEntryLength
	cfg.VersionsToPrint = cfg.VersionCharSpace / 10

	return cfg
}

func parseFlags(cfg *Config) {
	flag.IntVar(&cfg.MaxVersions, "max-versions", 15, "Specify maximum number of versions to display for each dependency.")
	flag.StringVar(&cfg.Filter, "filter", "major,minor,patch", "Filter out the version types for display, available values: `major`, `minor`, `patch`. In order to use two filters, separate them with a comma (,). By default, all version types are included.")
	flag.StringVar(&cfg.Dependency, "dependency", "", "Check only provided dependency.")
	flag.BoolVar(&cfg.IsIncompatible, "show-incompatible", false, "Show incompatible versions, disabled by default.")
	flag.BoolVar(&cfg.Pretty, "pretty", false, "Print in pretty way, disabled by default.")
	flag.Parse()
}

func parseShowSubCmd() bool {
	showCmd := flag.NewFlagSet("show", flag.ExitOnError)
	showOld := showCmd.Bool("old", false, "old")
	if len(os.Args) > 1 && os.Args[1] == "show" {
		showCmd.Parse(os.Args[2:])
		return *showOld
	}

	return false
}

func getTerminalWidth() int {
	fd := int(os.Stdin.Fd())
	width, _, err := terminal.GetSize(fd)
	if err != nil {
		log.Fatalf("%v", err)
	}

	return width
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
