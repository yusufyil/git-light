package checkout

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"time"
)

type Commit struct {
	Committer      string
	Date           time.Time
	PreviousCommit string
	Message        string
	Files          []File
}

type File struct {
	Path string
	Hash string
}

func (c Commit) CalculateHashForCommit() string {
	hasher := sha1.New()
	for _, file := range c.Files {
		_, err := hasher.Write([]byte(file.Hash))
		if err != nil {
			log.Fatal("got error while hashing commit, err: ", err.Error())
		}
	}

	hashSum := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	return hashString
}

func (c Commit) GetAllFilePaths() map[string]string {
	filePaths := make(map[string]string)
	for _, file := range c.Files {
		filePaths[file.Path] = file.Hash
	}

	return filePaths
}

func (c Commit) GetFilePathList() []string {
	filePaths := make([]string, len(c.Files))
	for _, file := range c.Files {
		filePaths = append(filePaths, file.Path)
	}

	return filePaths
}
