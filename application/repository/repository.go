package repository

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Repository interface {
	GetFileLines(p string) ([]string, error)
	WriteToFile(p string, content []string) error
	CompressAndSaveToFile(data interface{}, filename string) error
	DecompressFromFileAndConvert(filename string, data interface{}) error
	ListAllFiles(root string) ([]string, error)
	MoveFiles(sourceDir, destinationDir string) error
	DeleteFiles(path string) error
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
		if err != nil {
			return err
		}

		if info.IsDir() {
			if strings.HasPrefix(path, filepath.Join(root, ".git")) ||
				strings.HasPrefix(path, filepath.Join(root, ".idea")) {
				return filepath.SkipDir
			}
		} else {
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

func (r repository) DeleteFiles(path string) error {
	return os.Remove(path)
}
