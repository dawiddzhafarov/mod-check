package main

import (
	"os"

	"mod-check/cmd/gomodcheck"
	"mod-check/cmd/gomodcheck/subcmd"
	"mod-check/config"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "show" {
		subcmd.Run(config.New())
	} else {
		gomodcheck.Run(config.New())
	}
}
