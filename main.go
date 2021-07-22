package main

import (
	"fmt"
	"os"

	"github.com/cuducos/docs-cpi-pandemia/cli"
)

func main() {
	if err := cli.CLI().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
