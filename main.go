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
	//commitService.Initialize()
	commitService.AddToStage("*")
	commitService.CommitChanges("first commit message", "yusuf.yildirim@trendyol.com")

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
