package main

import (
	"fmt"
	"os"

	"github.com/stackable-specs/agent-checkers/src/cli/commands"
)

func main() {
	if err := commands.NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
