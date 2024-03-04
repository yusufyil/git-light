package checkout

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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
			fmt.Println("Error while writing to hasher:", err)
			return ""
		}
	}

	hashSum := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	return hashString
}
