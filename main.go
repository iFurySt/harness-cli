package main

import (
	"context"
	"fmt"
	"os"

	"github.com/iFurySt/harness-cli/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(context.Background(), version, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
