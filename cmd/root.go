package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "git-light",
	Short:   "git-light is a basic version control system.",
	Long:    `this application is a light weight version of git. It basically uses myers-diff algorithm to calculate differences between text files and saves them as deltas to a git like object store.`,
	Version: "0.1",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
