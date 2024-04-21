package cmd

import (
	"git-light/application/checkout"
	"git-light/application/myersdiff"
	"git-light/application/repository"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initializes empty repository to current path",
	Long:  `this command initializes empty git-light object store, if it's already there program exists.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		repo := repository.NewRepository()
		myersDiff := myersdiff.NewMyersDiffCalculator()
		commitService := checkout.NewCommitService(repo, myersDiff)
		commitService.Initialize()
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
