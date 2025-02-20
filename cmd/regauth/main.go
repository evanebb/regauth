package main

import (
	"github.com/evanebb/regauth/cli"
	"os"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
