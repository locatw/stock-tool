package main

import (
	"fmt"
	"os"
	"stock-tool/cmd/cli/cmd"
)

func main() {
	command := cmd.NewRootCmd()

	if err := command.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
