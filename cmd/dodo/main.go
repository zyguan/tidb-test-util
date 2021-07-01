package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zyguan/tidb-test-util/cmd/dodo/command"
)

var (
	Version   = "latest"
	BuildTime = "unknown"
)

func main() {
	cmd := command.FileServer()
	cmd.Use = "dodo"
	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run:   func(*cobra.Command, []string) { fmt.Printf("%s@%s (%s)\n", cmd.Use, Version, BuildTime) },
	})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[0;31mError: %+v\x1b[0m\n", err)
		os.Exit(1)
	}
}
