package main

import (
	"bufio"
	"fmt"
	"git-light/application/myersdiff"
	"os"
)

func main() {
	calculator := myersdiff.NewMyersDiffCalculator()
	t1, _ := getFileLines("test1.txt")
	t2, _ := getFileLines("test2.txt")
	//calculator.GenerateDiff(t1, t2)
	diffScript := calculator.GenerateDiffScript(t1, t2)
	fmt.Println(diffScript)
}

func getFileLines(p string) ([]string, error) {
	f, err := os.Open(p)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
