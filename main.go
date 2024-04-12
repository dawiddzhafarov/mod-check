package main

import (
	"mod-check/cmd/gomodcheck"
	"mod-check/config"
)

func main() {
	gomodcheck.Run(config.New())
}
