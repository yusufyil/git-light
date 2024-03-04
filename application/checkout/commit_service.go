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
	"strconv"
	"strings"
	"time"
)

type CommitService interface {
	Initialize()
	AddToStage(path string)
	CommitChanges(commitMessage string, committer string)
}

const baseFilePath = ".git-light"
const defaultBranchName = "main"

// TODO
// staging area isimli alanda maximum 1 commit tutulmali git-light add ile bir islem yapilirsa buraya eklennekli
// branches
// HEAD
// objects

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
	//todo update date message and committer of commit object in staging area and move them to objects folder
	var commit Commit
	err := cs.repo.DecompressFromFileAndConvert(baseFilePath+"/stage/"+"commit", &commit)
	if err != nil {
		log.Println("failed to read commit object from staging area")
	}

	commit.Message = commitMessage
	commit.Committer = committer
	commit.Date = time.Now()
	err = cs.repo.CompressAndSaveToFile(commit, baseFilePath+"/stage/"+"commit")
	if err != nil {
		log.Println("failed to write commit object from staging area")
	}

	err = cs.repo.MoveFiles(baseFilePath+"/stage", baseFilePath+"/objects")
	if err != nil {
		log.Println("failed to move staging area to permanent object store")
	}
}

func (cs commitService) AddToStage(path string) {
	// todo you should also create a commit object at the beginning and update it whenever add command called
	stageCommit := Commit{
		Committer:      "",
		Date:           time.Time{},
		PreviousCommit: "",
		Message:        "",
		Files:          make([]File, 0),
	}

	var filePaths []string
	if path == "*" || path == "." {
		// add all the contents
		allFiles, err := cs.repo.ListAllFiles("./")
		if err != nil {
			log.Println(err)
			return
		}
		filePaths = append(filePaths, allFiles...)
	} else {
		// todo (you should check if given file existing later)
		filePaths = append(filePaths, path)
	}

	lastCommit, err := cs.GetLastCommitOnCurrentBranch()
	if err != nil {
		stageCommit.PreviousCommit = "nil"
		// todo branch initially empty, so create everything from scratch
		for _, filePath := range filePaths {
			lines, err := cs.repo.GetFileLines(filePath)
			if err != nil {
				log.Fatal(err)
			}
			sha1Hash := cs.CalculateSHA1Hash(lines)
			stageCommit.Files = append(stageCommit.Files, File{Path: filePath, Hash: sha1Hash})
			err = cs.repo.WriteToFile(baseFilePath+"/stage/"+sha1Hash, append([]string{"nil"}, lines...))
			if err != nil {
				return
			}
		}

	} else {
		// todo get filePaths from commit, if new file created, create that file accordingly if a file modified create  another object on top of existing one
		stageCommit.PreviousCommit = lastCommit.CalculateHashForCommit()
		for _, path := range filePaths {

			currentFile, err := cs.repo.GetFileLines(path)
			if err != nil {
				log.Println(err)
			}

			for _, file := range lastCommit.Files {
				if path == file.Path {
					previousFile := cs.ExtractFileFromObjectStore(baseFilePath + "/objects/" + file.Hash)
					currentFileHash := cs.CalculateSHA1Hash(currentFile)
					previousFileHash := cs.CalculateSHA1Hash(previousFile)
					if cs.CalculateSHA1Hash(currentFile) != previousFileHash {
						diffScript := cs.myers.GenerateDiffScript(previousFile, currentFile)
						stageCommit.Files = append(stageCommit.Files, File{Path: path, Hash: currentFileHash})
						err := cs.repo.WriteToFile(baseFilePath+"/stage/"+currentFileHash, append([]string{currentFileHash}, diffScript...))
						if err != nil {
							log.Println(err)
							return
						}
					}
				}
			}
		}
	}

	// todo also save commit object to staging area
	err = cs.repo.CompressAndSaveToFile(stageCommit, baseFilePath+"/stage/"+"commit")
	if err != nil {
		log.Println(err)
	}

}

func (cs commitService) checkObjectStore() bool {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir := currentDir + baseFilePath
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

	parsedTime, err := time.Parse(time.ANSIC, lines[1])
	if err != nil {
		log.Fatal(err)
		return Commit{}, err
	}

	files := make([]File, len(lines)-4)

	for _, file := range lines[3:] {
		s := strings.Split(file, " ")
		files = append(files, File{Path: s[0], Hash: s[1]})
	}

	return Commit{
		Committer:      lines[0],
		Date:           parsedTime,
		PreviousCommit: lines[2],
		Message:        lines[3],
		Files:          files,
	}, nil
}

func (cs commitService) ExtractFileFromObjectStore(hash string) []string {
	delta, err := cs.repo.LoadAndDecompressStrings(baseFilePath + "/objects/" + hash)
	if err != nil {
		log.Fatal("failed to decompress delta err: " + err.Error())
	}
	if delta[0] == "nil" {
		return delta[1:]
	}
	previousDeltaHash := delta[0]
	source := cs.ExtractFileFromObjectStore(previousDeltaHash)
	return cs.applyDelta(source[1:], delta)
}

func (cs commitService) applyDelta(source []string, delta []string) []string {
	editScript := strings.Split(delta[0], "$")
	deletedRowCount := 0
	for _, deleteCommand := range editScript {
		if strings.Contains(deleteCommand, "d") {
			deletedRow, err := strconv.Atoi(strings.Replace(deleteCommand, "d", "", 1))
			if err != nil {
				log.Fatal(err)
			}
			source = append(source[:deletedRow-deletedRowCount], source[deletedRow-deletedRowCount+1:]...)
			deletedRowCount++
		}
	}

	for _, insertCommand := range editScript {
		if strings.Contains(insertCommand, "i") {
			insert := strings.Split(strings.Replace(insertCommand, "i", "", 1), "-")
			insertionDestIndex, err := strconv.Atoi(insert[0])
			if err != nil {
				log.Fatal(err)
			}
			insertionSourceIndex, err := strconv.Atoi(insert[1])
			if err != nil {
				log.Fatal(err)
			}
			source = append(source[:insertionDestIndex+1], source[insertionDestIndex:]...)
			source[insertionDestIndex] = delta[insertionSourceIndex]
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