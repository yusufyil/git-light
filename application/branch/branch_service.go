package branch

import (
	"fmt"
	"git-light/application/repository"
	"git-light/util"
	"log"
	"path/filepath"
)

type BranchService interface {
	CreateBranch(branchName string)
	DeleteBranch(branchName string)
	ListAllBranches() []string
}

type branchService struct {
	repo repository.Repository
}

func NewBranchService(repo repository.Repository) BranchService {
	return branchService{
		repo: repo,
	}
}

func (bs branchService) CreateBranch(branchName string) {
	lines, err := bs.repo.GetFileLines(filepath.Join(util.BaseFilePath, util.Head))
	if err != nil {
		log.Fatal("couldn't get HEAD err: " + err.Error())
	}

	commitHash, err := bs.repo.GetFileLines(filepath.Join(util.BaseFilePath, util.BranchFolder, lines[0]))
	if err != nil {
		commitHash[0] = lines[0]
	}

	err = bs.repo.WriteToFile(filepath.Join(util.BaseFilePath, util.BranchFolder, branchName), commitHash)
	if err != nil {
		log.Fatal("couldn't create new branch. err: " + err.Error())
	}
}

func (bs branchService) DeleteBranch(branchName string) {
	lines, err := bs.repo.GetFileLines(filepath.Join(util.BaseFilePath, util.Head))
	if err != nil {
		log.Fatal("couldn't get HEAD err: " + err.Error())
	}

	if lines[0] == branchName {
		log.Fatal("couldn't delete branch. you should checkout different branch before deleting it.")
	}

	err = bs.repo.DeleteFiles(filepath.Join(util.BaseFilePath, util.BranchFolder, branchName))
	if err != nil {
		log.Fatal("couldn't delete branch. err: " + err.Error())
	}
}

func (bs branchService) ListAllBranches() []string {
	allBranches, err := bs.repo.ListAllFiles(filepath.Join(util.BaseFilePath, util.BranchFolder))
	if err != nil {
		log.Fatal("couldn't get list of branches. err: " + err.Error())
	}

	for _, branch := range allBranches {
		fmt.Println(filepath.Base(branch))
	}
	return allBranches
}
