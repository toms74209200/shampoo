package main

import (
	"os"

	"github.com/toms74209200/shampoo/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
