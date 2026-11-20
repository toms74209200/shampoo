package main

import (
	"os"

	"github.com/toms74209200/shampoo/internal/run"
)

func main() {
	os.Exit(run.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
