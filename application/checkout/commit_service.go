package checkout

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"git-light/application/myersdiff"
	"git-light/application/repository"
	"git-light/util"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type CommitService interface {
	Initialize()
	AddToStage(path string)
	CommitChanges(commitMessage string, committer string)
	Checkout(commitHash string)
	Log()
}

type commitService struct {
	repo  repository.Repository
	myers myersdiff.Myers
}

func NewCommitService(repo repository.Repository, myers myersdiff.Myers) CommitService {
	return commitService{repo: repo, myers: myers}
}

func (cs commitService) Initialize() {
	if cs.checkObjectStore() {
		log.Fatal("this directory has already initialized with git-light")
	}

	err := os.Mkdir(util.BaseFilePath, 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(util.BaseFilePath, util.BranchFolder), 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(util.BaseFilePath, util.StageFolder), 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(util.BaseFilePath, util.ObjectFolder), 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(util.BaseFilePath, util.TempFolder), 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = cs.repo.WriteToFile(filepath.Join(util.BaseFilePath, util.Head), []string{util.DefaultBranchName})
	if err != nil {
		log.Fatal(err)
	}

	err = cs.repo.WriteToFile(filepath.Join(util.BaseFilePath, util.BranchFolder, util.DefaultBranchName), []string{"nil"})
	if err != nil {
		log.Fatal(err)
	}
}

func (cs commitService) CommitChanges(commitMessage string, committer string) {
	var commit Commit
	err := cs.repo.DecompressFromFileAndConvert(filepath.Join(util.BaseFilePath, util.StageFolder, "commit"), &commit)
	if err != nil {
		log.Fatal("nothing found in staging area, you should first add your changes")
	}

	commit.Message = commitMessage
	commit.Committer = committer
	commit.Date = time.Now()
	err = cs.repo.CompressAndSaveToFile(commit, filepath.Join(util.BaseFilePath, util.StageFolder, "commit"))
	if err != nil {
		log.Fatal("failed to write commit object from staging area")
	}

	commitHash := commit.CalculateHashForCommit()
	err = os.Rename(filepath.Join(util.BaseFilePath, util.StageFolder, "commit"), filepath.Join(util.BaseFilePath, util.StageFolder, commitHash))
	if err != nil {
		log.Fatal("failed to rename commit object from staging area")
	}

	cs.UpdateCurrentBranch(commitHash)

	err = cs.repo.MoveFiles(filepath.Join(util.BaseFilePath, util.StageFolder), filepath.Join(util.BaseFilePath, util.ObjectFolder))
	if err != nil {
		log.Fatal("failed to move staging area to permanent object store")
	}
}

func (cs commitService) AddToStage(path string) {
	var deltaFound = false
	stageCommit := Commit{
		Committer:      "",
		Date:           time.Time{},
		PreviousCommit: "",
		Message:        "",
		Files:          make([]File, 0),
	}

	var filePaths []string
	if path == "*" || path == "." {
		allFiles, err := cs.repo.ListAllFiles("./")
		if err != nil {
			log.Fatal("failed to get all file paths for root repository.")
		}
		filePaths = append(filePaths, allFiles...)
	} else {
		filePaths = append(filePaths, path)
	}

	lastCommit, err := cs.GetLastCommitOnCurrentBranch()
	if err != nil {
		stageCommit.PreviousCommit = "nil"
		for _, filePath := range filePaths {
			lines, err := cs.repo.GetFileLines(filePath)
			if err != nil {
				log.Fatal("failed to read files. file path: " + filePath)
			}
			deltaFound = true
			sha1Hash := cs.CalculateSHA1Hash(lines)
			stageCommit.Files = append(stageCommit.Files, File{Path: filePath, Hash: sha1Hash})
			err = cs.repo.CompressAndSaveToFile(myersdiff.Diff{PreviousBlobHash: "nil", Commands: "nil", Data: lines}, filepath.Join(util.BaseFilePath, util.BranchFolder, sha1Hash))
			if err != nil {
				log.Fatal("failed to save given file to stage: " + filePath)
			}
		}
	} else {
		stageCommit.PreviousCommit = lastCommit.CalculateHashForCommit()
		for _, filePath := range filePaths {

			currentFile, err := cs.repo.GetFileLines(filePath)
			if err != nil {
				log.Fatal("failed to read files. file path: " + filePath)
			}

			for _, file := range lastCommit.Files {
				if path == file.Path {
					previousFile := cs.ExtractFileFromObjectStore(file.Hash)
					currentFileHash := cs.CalculateSHA1Hash(currentFile)
					previousFileHash := cs.CalculateSHA1Hash(previousFile)
					if currentFileHash != previousFileHash {
						deltaFound = true
						diff := cs.myers.GenerateDiffScript(previousFile, currentFile)
						diff.PreviousBlobHash = previousFileHash
						stageCommit.Files = append(stageCommit.Files, File{Path: path, Hash: currentFileHash})
						err := cs.repo.CompressAndSaveToFile(diff, filepath.Join(util.BaseFilePath, util.StageFolder, currentFileHash))
						if err != nil {
							log.Fatal("failed to save given file to stage: " + filePath)
						}
					}
					//todo may be else can be added here.
				}
			}
			//todo should also commit newly added files to new commit
		}
	}

	if deltaFound {
		err = cs.repo.CompressAndSaveToFile(stageCommit, filepath.Join(util.BaseFilePath, util.StageFolder, "commit"))
		if err != nil {
			log.Fatal("failed to save commit to stage")
		}
	}
}

func (cs commitService) Checkout(commitHashOrBranch string) {
	var commit Commit
	lines, err := cs.repo.GetFileLines(filepath.Join(util.BaseFilePath, util.BranchFolder, commitHashOrBranch))
	if err == nil {
		err := cs.repo.WriteToFile(filepath.Join(util.BaseFilePath, util.Head), []string{commitHashOrBranch})
		if err != nil {
			log.Fatal("an error occurred when updating head")
		}
		commitHashOrBranch = lines[0]
	} else {
		err := cs.repo.WriteToFile(filepath.Join(util.BaseFilePath, util.Head), []string{commitHashOrBranch})
		if err != nil {
			log.Fatal("an error occurred when updating head")
		}
	}

	err = cs.repo.DecompressFromFileAndConvert(filepath.Join(util.BaseFilePath, util.ObjectFolder, commitHashOrBranch), &commit)
	if err != nil {
		log.Fatal("no such commit hash / branch found")
	}

	var checkoutFailure = false
	for _, file := range commit.Files {
		content := cs.ExtractFileFromObjectStore(file.Hash)

		err = cs.repo.WriteToFile(filepath.Join(util.BaseFilePath, util.TempFolder, file.Path), content)
		if err != nil {
			log.Println("failed to write file to object store with name: " + file.Path)
			checkoutFailure = true
			break
		}
	}

	if checkoutFailure {
		err = os.Remove(filepath.Join(util.BaseFilePath, util.TempFolder))
		if err != nil {
			log.Fatal("failed to remove temporary files")
		}
	}

	err = cs.repo.MoveFiles(filepath.Join(util.BaseFilePath, util.TempFolder), "./")
	if err != nil {
		log.Fatal("failed to move extracted files")
	}
}

func (cs commitService) checkObjectStore() bool {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(currentDir, util.BaseFilePath)); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (cs commitService) GetCurrentBranch() string {
	lines, err := cs.repo.GetFileLines(filepath.Join(util.BaseFilePath, util.Head))
	if err != nil {
		log.Fatal("failed to get current branch")
	}
	return lines[0]
}

func (cs commitService) UpdateCurrentBranch(commitSha string) {
	currentBranch := cs.GetCurrentBranch()
	err := cs.repo.WriteToFile(filepath.Join(util.BaseFilePath, util.BranchFolder, currentBranch), []string{commitSha})
	if err != nil {
		log.Fatal("failed to update current branch")
	}
}

func (cs commitService) GetLastCommitOnCurrentBranch() (Commit, error) {
	currentBranch := cs.GetCurrentBranch()
	lines, err := cs.repo.GetFileLines(filepath.Join(util.BaseFilePath, util.BranchFolder, currentBranch))
	if err != nil {
		log.Println(err)
		return Commit{}, err
	}

	if lines[0] == "nil" {
		log.Println("branch has no commit")
		return Commit{}, errors.New("empty branch")
	}

	var commit Commit
	err = cs.repo.DecompressFromFileAndConvert(filepath.Join(util.BaseFilePath, util.ObjectFolder, lines[0]), &commit)
	if err != nil {
		log.Println("failed to get last commit from repository")
		return Commit{}, err
	}

	return commit, nil
}

func (cs commitService) ExtractFileFromObjectStore(hash string) []string {
	var diff myersdiff.Diff
	err := cs.repo.DecompressFromFileAndConvert(filepath.Join(util.BaseFilePath, util.ObjectFolder, hash), &diff)

	if err != nil {
		log.Fatal("failed to decompress delta err: " + err.Error())
	}

	if diff.PreviousBlobHash == "nil" {
		return diff.Data
	}

	source := cs.ExtractFileFromObjectStore(diff.PreviousBlobHash)
	return cs.applyDelta(source, diff)
}

func (cs commitService) applyDelta(source []string, diff myersdiff.Diff) []string {
	editScript := strings.Split(diff.Commands, "$")
	deletedRowCount := 0
	for _, command := range editScript {
		if strings.Contains(command, "d") {
			deletedRow, err := strconv.Atoi(strings.Replace(command, "d", "", 1))
			if err != nil {
				log.Fatal("delta calculation error, failed to decompose instructions: " + err.Error())
			}
			source = append(source[:deletedRow-deletedRowCount], source[deletedRow-deletedRowCount+1:]...)
			deletedRowCount++
		}
	}

	for _, command := range editScript {
		if strings.Contains(command, "i") {
			insert := strings.Split(strings.Replace(command, "i", "", 1), "-")
			insertionDestIndex, err := strconv.Atoi(insert[0])
			if err != nil {
				log.Fatal("delta calculation error, failed to decompose instructions: " + err.Error())
			}
			insertionSourceIndex, err := strconv.Atoi(insert[1])
			if err != nil {
				log.Fatal("delta calculation error, failed to decompose instructions: " + err.Error())
			}
			source = slices.Insert(source, insertionDestIndex, diff.Data[insertionSourceIndex])
		}
	}

	return source
}

func (cs commitService) CalculateSHA1Hash(lines []string) string {
	hasher := sha1.New()

	for _, str := range lines {
		_, err := hasher.Write([]byte(str))
		if err != nil {
			log.Fatal("an error occurred during hash process", err.Error())
		}
	}

	hashSum := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	return hashString
}

func (cs commitService) Log() {
	commit, err := cs.GetLastCommitOnCurrentBranch()

	for err == nil {
		fmt.Printf("\033[32m  commit %s\n", commit.CalculateHashForCommit())
		fmt.Printf("\033[34m Author: %s\n", commit.Committer)
		fmt.Printf("\033[36m Date:   %s\n\n", commit.Date.Format("Mon Jan 2 15:04:05 2006 -0700"))
		fmt.Printf("\033[31m    %s\n\n", commit.Message)

		err = cs.repo.DecompressFromFileAndConvert(filepath.Join(util.BaseFilePath, util.ObjectFolder, commit.PreviousCommit), &commit)
		if err != nil {
			break
		}
	}
}
