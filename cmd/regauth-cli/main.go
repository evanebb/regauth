package main

import (
	"fmt"
	"github.com/evanebb/regauth/cli/client"
	"os"
)

func main() {
	cmd, err := client.NewRootCmd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
