package cmd

import (
	"git-light/application/checkout"
	"git-light/application/myersdiff"
	"git-light/application/repository"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "prints commit history",
	Long:  `this command prints log history in descending order`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		repo := repository.NewRepository()
		myersDiff := myersdiff.NewMyersDiffCalculator()
		commitService := checkout.NewCommitService(repo, myersDiff)
		commitService.Log()
	},
}

func init() {
	RootCmd.AddCommand(logCmd)
}
