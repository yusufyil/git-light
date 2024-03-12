package main

import (
	"git-light/application/checkout"
	"git-light/application/myersdiff"
	"git-light/application/repository"
	"path/filepath"
)

func main() {
	calculator := myersdiff.NewMyersDiffCalculator()
	repo := repository.NewRepository()
	//t1, _ := repo.GetFileLines("test1.txt")
	//t2, _ := repo.GetFileLines("test2.txt")
	//calculator.GenerateDiff(t1, t2)
	//diffScript := calculator.GenerateDiffScript(t1, t2)
	//fmt.Println(diffScript)
	//_ = repo.WriteToFile("diff.txt", diffScript)
	//err := repo.CompressAndSaveStrings("test", diffScript)
	//if err != nil {
	//	log.Fatal(err)
	//}

	commitService := checkout.NewCommitService(repo, calculator)
	//lines1, _ := repo.GetFileLines("1.txt")
	//lines2, _ := repo.GetFileLines("2.txt")
	//calculator.GenerateDiff(lines1, lines2)
	//commitService.Initialize()
	commitService.AddToStage("deneme.txt")
	commitService.CommitChanges("second commit message", "yusuf.yildirim@trendyol.com")
	commitService.Log()

	//var commit checkout.Commit
	//var diff myersdiff.Diff
	//_ = repo.DecompressFromFileAndConvert(".git-light/objects/a36fa2f1c4713ef75b7a334bc8d34dfeffb2ed77", &commit)
	//_ = repo.DecompressFromFileAndConvert(".git-light/objects/023bcc862d142ae728649d30cc1ab7c68410e5f1", &diff)
	//
	//fmt.Println(commit)
	//fmt.Println(diff)
	//fmt.Println(commitService)

	//TODO apply changes to source in order to create destination
	// applyDelta (source []string, editScript []string)

	/*strings, err := loadAndDecompressStrings("test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings)
	*/
	/*last := applyDelta(t1, diffScript)
	fmt.Println(last)
	dir, err := FilePathWalkDir("./")
	fmt.Println(dir)
	*/
}

func FilterDirsGlob(dir, suffix string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, suffix))
}
