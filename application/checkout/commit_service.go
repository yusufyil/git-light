package checkout

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"git-light/application/myersdiff"
	"git-light/application/repository"
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

const (
	baseFilePath      = ".git-light"
	defaultBranchName = "main"
)

//todo use filepath.join with all paths

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
		// called method already has log.fatal remove this part.
	}

	err := os.Mkdir(baseFilePath, 0700)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.Mkdir(baseFilePath+"/branches", 0700)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.Mkdir(baseFilePath+"/stage", 0700)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.Mkdir(baseFilePath+"/objects", 0700)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.Mkdir(baseFilePath+"/temp", 0700)
	if err != nil {
		log.Fatal(err)
	}

	err = cs.repo.WriteToFile(baseFilePath+"/HEAD", []string{defaultBranchName})
	if err != nil {
		log.Fatal(err)
		return
	}

	err = cs.repo.WriteToFile(baseFilePath+"/branches/"+defaultBranchName, []string{"nil"})
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (cs commitService) CommitChanges(commitMessage string, committer string) {
	var commit Commit
	err := cs.repo.DecompressFromFileAndConvert(baseFilePath+"/stage/"+"commit", &commit)
	if err != nil {
		log.Println("nothing found in staging area, you should first add your changes")
	}

	commit.Message = commitMessage
	commit.Committer = committer
	commit.Date = time.Now()
	err = cs.repo.CompressAndSaveToFile(commit, baseFilePath+"/stage/"+"commit")
	if err != nil {
		log.Println("failed to write commit object from staging area")
	}

	commitHash := commit.CalculateHashForCommit()
	err = os.Rename(baseFilePath+"/stage/"+"commit", baseFilePath+"/stage/"+commitHash)
	if err != nil {
		log.Println("failed to rename commit object from staging area")
	}

	cs.UpdateCurrentBranch(commitHash)

	err = cs.repo.MoveFiles(baseFilePath+"/stage", baseFilePath+"/objects")
	if err != nil {
		log.Println("failed to move staging area to permanent object store")
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
			log.Println(err)
			return
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
				log.Println(err)
			}
			deltaFound = true
			sha1Hash := cs.CalculateSHA1Hash(lines)
			stageCommit.Files = append(stageCommit.Files, File{Path: filePath, Hash: sha1Hash})
			err = cs.repo.CompressAndSaveToFile(myersdiff.Diff{PreviousBlobHash: "nil", Commands: "nil", Data: lines}, baseFilePath+"/stage/"+sha1Hash)
			if err != nil {
				return
			}
		}
	} else {
		stageCommit.PreviousCommit = lastCommit.CalculateHashForCommit()
		for _, filePath := range filePaths {

			currentFile, err := cs.repo.GetFileLines(filePath)
			if err != nil {
				log.Println(err)
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
						err := cs.repo.CompressAndSaveToFile(diff, baseFilePath+"/stage/"+currentFileHash)
						if err != nil {
							log.Println(err)
							return
						}
					}
				}
			}
		}
	}

	if deltaFound {
		err = cs.repo.CompressAndSaveToFile(stageCommit, baseFilePath+"/stage/"+"commit")
		if err != nil {
			log.Println(err)
		}
	}
}

func (cs commitService) Checkout(commitHash string) {
	var commit Commit
	//lines, err := cs.repo.GetFileLines(baseFilePath + "/branches/" + commitHash)
	//if err != nil {
	//	commitHash = lines[0]
	//	err := cs.repo.WriteToFile(baseFilePath+"/HEAD", []string{commitHash})
	//	if err != nil {
	//		fmt.Println("an error occurred when updating head")
	//	}
	//}

	err := cs.repo.DecompressFromFileAndConvert(baseFilePath+"/objects/"+commitHash, &commit)
	if err != nil {
		fmt.Println("given commit hash / branch name is not valid")
		return
	}

	var checkoutFailure = false
	for _, file := range commit.Files {
		content := cs.ExtractFileFromObjectStore(file.Hash)

		err = cs.repo.WriteToFile(baseFilePath+"/temp/"+file.Path, content)
		if err != nil {
			fmt.Println("failed to write file to object store with name: " + file.Path)
			checkoutFailure = true
			break
		}
	}

	if checkoutFailure {
		err = os.Remove(baseFilePath + "/temp")
		if err != nil {
			fmt.Println("failed to remove temporary files")
		}
		return
	}

	err = cs.repo.MoveFiles(baseFilePath+"/temp", "./")
	if err != nil {
		fmt.Println("failed to move extracted files")
	}
}

func (cs commitService) checkObjectStore() bool {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir := currentDir + "/" + baseFilePath
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (cs commitService) GetCurrentBranch() string {
	lines, err := cs.repo.GetFileLines(baseFilePath + "/HEAD")
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return lines[0]
}

func (cs commitService) UpdateCurrentBranch(commitSha string) {
	currentBranch := cs.GetCurrentBranch()
	err := cs.repo.WriteToFile(baseFilePath+"/branches/"+currentBranch, []string{commitSha})
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (cs commitService) GetLastCommitOnCurrentBranch() (Commit, error) {
	currentBranch := cs.GetCurrentBranch()
	lines, err := cs.repo.GetFileLines(baseFilePath + "/branches/" + currentBranch)
	if err != nil {
		log.Println(err)
		return Commit{}, err
	}

	if lines[0] == "nil" {
		log.Println("branch has no commit")
		return Commit{}, errors.New("empty branch")
	}

	var commit Commit
	err = cs.repo.DecompressFromFileAndConvert(baseFilePath+"/objects/"+lines[0], &commit)
	if err != nil {
		log.Println("can't find any object")
		return Commit{}, err
	}

	return commit, nil
}

func (cs commitService) ExtractFileFromObjectStore(hash string) []string {
	var diff myersdiff.Diff
	err := cs.repo.DecompressFromFileAndConvert(baseFilePath+"/objects/"+hash, &diff)
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
				log.Fatal(err)
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
				log.Fatal(err)
			}
			insertionSourceIndex, err := strconv.Atoi(insert[1])
			if err != nil {
				log.Fatal(err)
			}
			source = slices.Insert(source, insertionDestIndex, diff.Data[insertionSourceIndex])
		}
	}

	return source
}

func (cs commitService) CalculateSHA1Hash(strings []string) string {
	hasher := sha1.New()

	for _, str := range strings {
		_, err := hasher.Write([]byte(str))
		if err != nil {
			fmt.Println("Error while writing to hasher:", err)
			return ""
		}
	}

	hashSum := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	return hashString
}

func (cs commitService) Log() {
	commit, err := cs.GetLastCommitOnCurrentBranch()

	for err == nil {
		fmt.Println("****")
		fmt.Println(commit.Message)
		fmt.Println(commit.Committer)
		fmt.Println(commit.CalculateHashForCommit())
		fmt.Println(commit.Files)

		err = cs.repo.DecompressFromFileAndConvert(baseFilePath+"/objects/"+commit.PreviousCommit, &commit)
		if err != nil {
			break
		}
	}
}
