package main

import (
	"github.com/evanebb/regauth/cli/server"
	"os"
)

func main() {
	if err := server.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
