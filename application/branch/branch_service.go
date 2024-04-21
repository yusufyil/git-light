package branch

import "fmt"

type BranchService interface {
	CreateBranch(branchName string) error
	DeleteBranch(branchName string) error
	ListAllBranches() []string
}

type branchService struct {
}

func NewBranchService() BranchService {
	return branchService{}
}

func (bs branchService) CreateBranch(branchName string) error {
	fmt.Println("create branch called with arguments ->", branchName)
	return nil
}

func (bs branchService) DeleteBranch(branchName string) error {
	fmt.Println("delete branch called with arguments ->", branchName)
	return nil
}

func (bs branchService) ListAllBranches() []string {
	fmt.Println("list branches called")
	return []string{}
}
