package cmd

import (
	"git-light/application/branch"
	"github.com/spf13/cobra"
)

var branchName string

var createBranchCmd = &cobra.Command{
	Use:   "branch",
	Short: "creates branch",
	Long:  `this command creates branch on top of existing commit hash`,
	Run: func(cmd *cobra.Command, args []string) {
		branchService := branch.NewBranchService()
		_ = branchService.CreateBranch(branchName)
	},
}

var deleteBranchCmd = &cobra.Command{
	Use:   "branch",
	Short: "deletes branch",
	Long:  `this command deletes`,
	Run: func(cmd *cobra.Command, args []string) {
		branchService := branch.NewBranchService()
		_ = branchService.DeleteBranch(branchName)
	},
}

var listBranchCmd = &cobra.Command{
	Use:   "branch",
	Short: "list all branches",
	Long:  `this command deletes`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		branchService := branch.NewBranchService()
		_ = branchService.ListAllBranches()
	},
}

func init() {
	//RootCmd.AddCommand(createBranchCmd)
	//RootCmd.AddCommand(deleteBranchCmd)
	//RootCmd.AddCommand(listBranchCmd)

	//createBranchCmd.Flags().StringVarP(&branchName, "branchName", "b", "", "branch name")
	//deleteBranchCmd.Flags().StringVarP(&branchName, "branchName", "d", "", "branch name")
	//listBranchCmd.Flags().StringVarP(&branchName, "branchName", "a", "", "branch name")

}
