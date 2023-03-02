package main

//go:generate go run ./generate.go

import (
	"log"

	"github.com/spf13/cobra/doc"
	"gitlab.com/fynbos/cli/cmd"
)

func main() {
	rootCmd := cmd.NewCmdRoot(nil)
	err := doc.GenMarkdownTree(rootCmd, "./")
	if err != nil {
		log.Fatal(err)
	}
}
