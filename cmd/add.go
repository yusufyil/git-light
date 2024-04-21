package cmd

import (
	"git-light/application/checkout"
	"git-light/application/myersdiff"
	"git-light/application/repository"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "adds given file to stage",
	Long: `this command calculates diffs according to given files and saves them into staging area if any difference exist
		   between working directory and previous commit.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := repository.NewRepository()
		myersDiff := myersdiff.NewMyersDiffCalculator()
		commitService := checkout.NewCommitService(repo, myersDiff)
		commitService.AddToStage(args[0])
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}
