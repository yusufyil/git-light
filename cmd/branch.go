package cmd

import (
	"git-light/application/branch"
	"git-light/application/repository"
	"github.com/spf13/cobra"
)

var (
	deleteBranch    string
	listAllBranches bool
)

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Manage branches",
	Long:  `The branch command allows you to create, delete, and list branches.`,
	Run: func(cmd *cobra.Command, args []string) {
		if listAllBranches {
			repo := repository.NewRepository()
			branchService := branch.NewBranchService(repo)
			_ = branchService.ListAllBranches()
			return
		}
		if deleteBranch != "" {
			repo := repository.NewRepository()
			branchService := branch.NewBranchService(repo)
			branchService.DeleteBranch(deleteBranch)
			return
		}
		if len(args) > 0 {
			repo := repository.NewRepository()
			branchService := branch.NewBranchService(repo)
			branchService.CreateBranch(args[0])
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(branchCmd)

	branchCmd.Flags().StringVarP(&deleteBranch, "delete", "d", "", "Delete a branch")
	branchCmd.Flags().BoolVarP(&listAllBranches, "all", "a", false, "List all branches")
}
