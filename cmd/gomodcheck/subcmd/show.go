package subcmd

import (
	"fmt"
	"log"
	"os"

	"mod-check/config"
	"mod-check/internal/module"

	"golang.org/x/mod/modfile"
)

const goModFile = "go.mod"

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

	num := 0
	// Loop through required mods and exclude indirect ones
	for _, req := range file.Require {
		if req.Indirect {
			continue
		}

		// Create a Module
		mod, err := module.NewModule(req.Mod.Path, req.Mod.Version)
		if err != nil {
			log.Fatalf("%s", err)
		}
		if mod == nil {
			continue
		}

		if cfg.ShowOld && mod.IsCurrent {
			continue
		}

		// List dependencies
		fmt.Println(mod.Path)
		num++
	}
	fmt.Printf("Total entries: %d\n", num)
}
