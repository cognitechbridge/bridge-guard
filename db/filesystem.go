package db

import (
	"ctb-cli/config"
	"ctb-cli/filesyetem"
	"io"
	"os"
	"path/filepath"
)

var _ filesyetem.Persist = (*SqlLiteConnection)(nil)

// SavePath saves a serialized key in the database
func (conn *SqlLiteConnection) SavePath(path string, key string, isDir bool) error {
	absPath, err := getAbsPath(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(absPath)
	defer file.Close()
	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}
	return nil
}

// GetPath retrieves a path id by path
func (conn *SqlLiteConnection) GetPath(path string) (string, error) {
	absPath, err := getAbsPath(path)
	if err != nil {
		return "", err
	}
	file, err := os.Open(absPath)
	defer file.Close()
	id, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(id), nil
}

func (conn *SqlLiteConnection) RemovePath(path string) error {
	absPath, err := getAbsPath(path)
	if err != nil {
		return err
	}
	err = os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

func (conn *SqlLiteConnection) PathExist(path string) (bool, error) {
	absPath, err := getAbsPath(path)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func getAbsPath(path string) (string, error) {
	basePath, err := getFilesysPath()
	if err != nil {
		return "", err
	}
	absPath := filepath.Join(basePath, path)
	return absPath, nil
}

func getFilesysPath() (string, error) {
	rootPath, err := config.GetRepoCtbRoot()
	if err != nil {
		return "", err
	}
	path := filepath.Join(rootPath, "filesystem")
	return path, nil
}
