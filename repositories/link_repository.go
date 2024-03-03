package repositories

import (
	"ctb-cli/core"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrorVaultLinkNotFount     = errors.New("vault link not found")
	ErrorReadingLinkFile       = errors.New("error reading link file")
	ErrorRemovingVaultLinkFile = errors.New("error removing vault link file")
)

type LinkRepository struct {
	rootPath string
}

func NewLinkRepository(rootPath string) *LinkRepository {
	return &LinkRepository{
		rootPath: rootPath,
	}
}

// Create link file
func (c *LinkRepository) Create(path string, link core.Link) error {
	absPath := filepath.Join(c.rootPath, path)
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	js, _ := json.Marshal(link)
	_, _ = file.Write(js)
	return nil
}

// InsertVaultLink create vault link file
func (c *LinkRepository) InsertVaultLink(path string, link core.VaultLink) error {
	absPath := filepath.Join(c.rootPath, path, ".vault")
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	js, _ := json.Marshal(link)
	_, _ = file.Write(js)
	return nil
}

func (c *LinkRepository) GetVaultLinkByPath(path string) (core.VaultLink, error) {
	p := filepath.Join(c.rootPath, path, ".vault")
	js, err := os.ReadFile(p)
	if err != nil {
		return core.VaultLink{}, fmt.Errorf("error reading vault link file: %v", err)
	}
	var link core.VaultLink
	err = json.Unmarshal(js, &link)
	if err != nil {
		return core.VaultLink{}, fmt.Errorf("error unmarshalink vault link file: %v", err)
	}
	return link, nil
}

func (c *LinkRepository) Update(path string, link core.Link) error {
	absPath := filepath.Join(c.rootPath, path)
	file, err := os.OpenFile(absPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error updating link file: %v", err)
	}
	defer file.Close()
	js, _ := json.Marshal(link)
	_, err = file.Write(js)
	return err
}

func (c *LinkRepository) GetByPath(path string) (core.Link, error) {
	p := filepath.Join(c.rootPath, path)
	js, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return core.Link{}, ErrorVaultLinkNotFount
	}
	if err != nil {
		return core.Link{}, ErrorReadingLinkFile
	}
	var link core.Link
	err = json.Unmarshal(js, &link)
	if err != nil {
		return core.Link{}, fmt.Errorf("error unmarshalink link file: %v", err)
	}
	return link, nil
}

func (c *LinkRepository) open(path string) (*os.File, error) {
	p := filepath.Join(c.rootPath, path)
	file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	return file, err
}

func (c *LinkRepository) ListIdsByRegex(regex string) ([]string, error) {
	var matchedIds []string
	list, err := c.listFilesByRegex(regex)
	if err != nil {
		return nil, err
	}
	for _, file := range list {
		link, err := c.GetByPath(file)
		if err != nil {
			return nil, err
		}
		matchedIds = append(matchedIds, link.ObjectId)
	}
	return matchedIds, nil
}

func (c *LinkRepository) Remove(path string) error {
	absPath := filepath.Join(c.rootPath, path)
	err := os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

func (c *LinkRepository) RemoveDir(path string) error {
	p := filepath.Join(c.rootPath, path)
	err := os.Remove(p)
	return err
}

func (c *LinkRepository) Rename(oldPath string, newPath string) error {
	o := filepath.Join(c.rootPath, oldPath)
	n := filepath.Join(c.rootPath, newPath)
	err := os.Rename(o, n)
	if err != nil {
		return err
	}
	return nil
}

func (c *LinkRepository) CreateDir(path string) (err error) {
	absPath := filepath.Join(c.rootPath, path)
	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (c *LinkRepository) GetSubFiles(path string) ([]os.FileInfo, error) {
	p := filepath.Join(c.rootPath, path)
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error opening dir to Read sub files: %v", err)
	}
	subFiles, _ := file.Readdir(0)
	return subFiles, nil
}

// listFilesByRegex lists all files in the specified directory that match the given regex pattern.
// Returns a slice of strings containing the names of matching files.
func (c *LinkRepository) listFilesByRegex(pattern string) ([]string, error) {
	var matchedFiles []string

	dirPath := c.rootPath

	// Walk the directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Return the error to stop the walk
		}
		if info.IsDir() {
			return nil // Skip directories
		}

		// Generate the relative path
		relativePath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err // Return the error to stop the walk
		}

		// Check if the file matches the pattern
		match, err := filepath.Match(pattern, relativePath)
		if err != nil {
			return err // Return the error to stop the walk
		}
		if match {
			matchedFiles = append(matchedFiles, relativePath) // Add matching file to the slice
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %w", dirPath, err)
	}

	return matchedFiles, nil // Return the slice of matched file paths
}

func (c *LinkRepository) IsDir(path string) (bool, error) {
	p := filepath.Join(c.rootPath, path)
	fi, err := os.Stat(p)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func (c *LinkRepository) RemoveVaultLink(path string) error {
	absPath := filepath.Join(c.rootPath, path, ".vault")
	err := os.Remove(absPath)
	if err != nil {
		return ErrorRemovingVaultLinkFile
	}
	return nil
}
