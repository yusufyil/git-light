package cmd

import (
	"git-light/application/checkout"
	"git-light/application/myersdiff"
	"git-light/application/repository"

	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "checkouts for given commit hash or branch name",
	Long:  `this command first looks for branches and then checks for commits to retrieve files from object store.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := repository.NewRepository()
		myersDiff := myersdiff.NewMyersDiffCalculator()
		commitService := checkout.NewCommitService(repo, myersDiff)
		commitService.Checkout(args[0])
	},
}

func init() {
	RootCmd.AddCommand(checkoutCmd)
}
