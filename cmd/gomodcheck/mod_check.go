package gomodcheck

import (
	"fmt"
	"log"
	"os"
	"strings"

	"mod-check/config"
	"mod-check/internal/module"
	"mod-check/internal/version"

	"github.com/fatih/color"
	ver "github.com/hashicorp/go-version"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/mod/modfile"
)

const (
	// goModFile is the name of go mod file
	goModFile = "go.mod"
)

var (
	blue   = color.New(color.FgBlue).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgGreen).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

func Run(cfg *config.Config) {

	// Read in go.mod file
	data, err := os.ReadFile("./" + goModFile)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Parse go.mod file
	file, err := modfile.Parse(goModFile, data, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var modules module.Modules

	// Loop through required mods and exclude indirect ones
	for _, req := range file.Require {
		if req.Indirect {
			continue
		}

		// Create a Module
		mod, err := module.NewModule(req.Mod.Path, req.Mod.Version)
		if err != nil {
			log.Fatalf("%s", err) // it was continue, no checking the error
		}
		if mod == nil { // there was no such check previously
			continue
		}

		// Do nothing if mod is current
		if mod.IsCurrent {
			continue
		}

		modules = append(modules, *mod)
	}

	// If dependency is provided, only process it
	if cfg.Dependency != "" {
		for _, dep := range modules {
			if dep.Path == cfg.Dependency {
				modules = module.Modules{dep}
			}
		}
		if len(modules) > 1 {
			log.Printf("dependency '%s' not found. Listing all dependencies\n", cfg.Dependency)
		}
	}

	if cfg.Pretty {
		processPretty(modules, *cfg, cfg.VersionCharSpace, cfg.VersionsToPrint)
	} else {
		processDefault(modules, *cfg)
	}
}

func processDefault(modules module.Modules, cfg config.Config) {
	for _, mod := range modules {
		var newVersions []string
		for _, availableVer := range mod.AvailableVersions {
			currentVersion, err := ver.NewVersion(mod.CurrentVersion.Original)
			if err != nil {
				panic(err)
			}
			availableVersion, err := ver.NewVersion(availableVer.Original)
			if availableVersion.GreaterThan(currentVersion) {
				if v := filterVersions(availableVer, cfg.IsIncompatible, cfg.Filter); v != "" {
					newVersions = append(newVersions, v)
				}
			}
		}
		if len(newVersions) != 0 {
			fmt.Println(cyan(mod.Path), fmt.Sprintf("current: %s; available: %s", blue(mod.CurrentVersion.Original), strings.Join(newVersions, ", ")))
		}
	}
}

func processPretty(modules module.Modules, cfg config.Config, versionSpace, versionNumber int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Dependency", "Current Version", "Available Versions"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:     "#",
			WidthMax: 3,
			Align:    text.AlignCenter,
		},
		{
			Name:     "Dependency",
			WidthMax: 30,
			Align:    text.AlignCenter,
		},
		{
			Name:     "Current Version",
			WidthMax: 17,
			Align:    text.AlignCenter,
		},
		{
			Name:     "Available Versions",
			WidthMax: versionSpace,
		},
	})

	for i, mod := range modules {
		var newVersions []string
		for _, availableVer := range mod.AvailableVersions {
			currentVersion, err := ver.NewVersion(mod.CurrentVersion.Original)
			if err != nil {
				panic(err)
			}
			availableVersion, err := ver.NewVersion(availableVer.Original)
			if availableVersion.GreaterThan(currentVersion) {
				if v := filterVersions(availableVer, cfg.IsIncompatible, cfg.Filter); v != "" {
					newVersions = append(newVersions, v)
				}
			}
		}
		if len(newVersions) != 0 {
			if len(newVersions) <= cfg.MaxVersions {
				unwrapRows(i, cyan(mod.Path), blue(mod.CurrentVersion.Original), versionNumber, newVersions, t)
			} else {
				unwrapRows(i, cyan(mod.Path), blue(mod.CurrentVersion.Original), versionNumber, newVersions[:cfg.MaxVersions], t)
			}
		}
	}
	if t.Length() != 0 {
		t.AppendSeparator()
		t.Render()
	} else {
		fmt.Println("There are no newer versions that fulfill provided requirements.")
		fmt.Printf("Filter: %s\n", cfg.Filter)
		fmt.Printf("Max-versions: %d\n", cfg.MaxVersions)
		fmt.Printf("Show incompatible: %v\n", cfg.IsIncompatible)
	}
}

func unwrapRows(i int, path string, curVersion string, num int, versions []string, t table.Writer) {
	length := len(versions)
	if length > num {
		t.AppendRow([]interface{}{i + 1, blue(path), curVersion, strings.Join(versions[:num], ", ")})
		for length > num {
			length -= num
			versions = versions[num:]
			if len(versions) < num {
				t.AppendRow([]interface{}{"", "", "", strings.Join(versions, ", ")})
				t.AppendSeparator()
			} else {
				t.AppendRow([]interface{}{"", "", "", strings.Join(versions[:num], ", ")})
			}
		}
	} else {
		t.AppendRow([]interface{}{i + 1, blue(path), curVersion, strings.Join(versions, ", ")})
		t.AppendSeparator()
	}
}

func filterVersions(ver version.Version, showIncompatible bool, filter string) string {
	if !showIncompatible && ver.Incompatible {
		return ""
	}

	if filter == "" {
		return colorVersion(ver.Status, ver.Original)
	}

	filters := strings.Split(filter, ",")
	for _, fil := range filters {
		if strings.Contains(ver.Status, fil) {
			return colorVersion(ver.Status, ver.Original)
		}
	}

	return ""
}

func colorVersion(status string, version string) string {
	switch status {
	case "minor":
		return yellow(version)
	case "major":
		return red(version)
	case "patch":
		return green(version)
	default:
		return blue(version)
	}
}
