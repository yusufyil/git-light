package repository

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Repository interface {
	GetFileLines(p string) ([]string, error)
	WriteToFile(p string, content []string) error
	CompressAndSaveStrings(p string, input []string) error
	LoadAndDecompressStrings(p string) ([]string, error)
	CompressAndSaveToFile(data interface{}, filename string) error
	DecompressFromFileAndConvert(filename string, data interface{}) error
	ListAllFiles(root string) ([]string, error)
	MoveFiles(sourceDir, destinationDir string) error
}

type repository struct {
}

func NewRepository() Repository {
	return repository{}
}

func (r repository) GetFileLines(p string) ([]string, error) {
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
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func (r repository) WriteToFile(p string, content []string) error {
	file, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, line := range content {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (r repository) CompressAndSaveStrings(p string, input []string) error {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	for _, str := range input {
		_, err := gzipWriter.Write([]byte(str + "\n"))
		if err != nil {
			return err
		}
	}
	gzipWriter.Close()

	err := ioutil.WriteFile(p, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (r repository) LoadAndDecompressStrings(p string) ([]string, error) {
	compressedData, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(compressedData)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	decompressedData, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(decompressedData), "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

func (r repository) CompressAndSaveToFile(data interface{}, filename string) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	compressedData := new(bytes.Buffer)
	compressor := gzip.NewWriter(compressedData)
	_, err = compressor.Write(buf.Bytes())
	if err != nil {
		return err
	}
	compressor.Close()

	err = os.WriteFile(filename, compressedData.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r repository) DecompressFromFileAndConvert(filename string, data interface{}) error {
	compressedData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	decompressor, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return err
	}
	defer decompressor.Close()

	decoder := gob.NewDecoder(decompressor)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func (r repository) ListAllFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && !strings.HasPrefix(path, ".git") && !strings.HasPrefix(path, ".idea") && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (r repository) MoveFiles(sourceDir, destinationDir string) error {
	source, err := os.Open(sourceDir)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(destinationDir)
	if os.IsNotExist(err) {
		return errors.New("destination folder does not exist")
	}

	fileInfos, err := source.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		sourcePath := filepath.Join(sourceDir, fileInfo.Name())
		destinationPath := filepath.Join(destinationDir, fileInfo.Name())

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destinationFile, err := os.Create(destinationPath)
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		_, err = io.Copy(destinationFile, sourceFile)
		if err != nil {
			return err
		}

		err = os.Remove(sourcePath)
		if err != nil {
			return err
		}
	}

	return nil
}
