package cmd

import (
	"fmt"
	"git-light/application/checkout"
	"git-light/application/myersdiff"
	"git-light/application/repository"

	"github.com/spf13/cobra"
)

var commitMessage string
var committerEmail string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "commits given file",
	Long:  `long exp..`,
	//Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := repository.NewRepository()
		myersDiff := myersdiff.NewMyersDiffCalculator()
		commitService := checkout.NewCommitService(repo, myersDiff)
		commitService.CommitChanges(commitMessage, committerEmail)
		fmt.Println("commit called")
	},
}

func init() {
	RootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message")
	commitCmd.Flags().StringVarP(&committerEmail, "committer", "c", "default committer", "Committer's email")

}
