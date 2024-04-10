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
	ErrVaultLinkNotFount     = errors.New("vault link not found")
	ErrReadingLinkFile       = errors.New("error reading link file")
	ErrRemovingVaultLinkFile = errors.New("error removing vault link file")
	ErrPathIsNotDir          = errors.New("path is not a valid directory")
)

type LinkRepository struct {
	rootPath string
}

func NewLinkRepository(rootPath string) *LinkRepository {
	return &LinkRepository{
		rootPath: rootPath,
	}
}

// Create creates a new file at the specified path and writes the JSON representation of the given link to it.
// If the file or any necessary directories do not exist, they will be created.
// The path parameter specifies the relative path to the file, and the link parameter contains the data to be written.
// Returns an error if any error occurs during the creation or writing process.
func (c *LinkRepository) Create(path string, link core.Link) error {
	absPath := filepath.Join(c.rootPath, path)
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	js, _ := json.Marshal(link)
	_, _ = file.Write(js)
	return nil
}

// InsertVaultLink inserts a VaultLink into the specified path.
// It creates the necessary directories and writes the link data to a file.
// The link data is serialized as JSON before writing to the file.
// If any error occurs during the process, it is returned.
func (c *LinkRepository) InsertVaultLink(path string, link core.VaultLink) error {
	absPath := filepath.Join(c.rootPath, path, ".vault")
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	js, _ := json.Marshal(link)
	_, _ = file.Write(js)
	return nil
}

// GetVaultLinkByPath retrieves the vault link by the given path.
// The path should be a directory path.
// It reads the vault link file located at the specified path and returns the corresponding VaultLink object.
// If an error occurs during file reading or unmarshaling, it returns an empty VaultLink object and the error.
func (c *LinkRepository) GetVaultLinkByPath(path string) (core.VaultLink, error) {
	// Make sure the path is a directory
	if !c.IsDir(path) {
		return core.VaultLink{}, ErrPathIsNotDir
	}
	// Read the vault link file
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

// Update updates the link file at the specified path with the provided link data.
// It returns an error if there was a problem updating the file.
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

// GetByPath retrieves a link from the repository based on the given path.
// It returns the retrieved link and an error, if any.
func (c *LinkRepository) GetByPath(path string) (core.Link, error) {
	p := filepath.Join(c.rootPath, path)
	js, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return core.Link{}, ErrVaultLinkNotFount
	}
	if err != nil {
		return core.Link{}, ErrReadingLinkFile
	}
	var link core.Link
	err = json.Unmarshal(js, &link)
	if err != nil {
		return core.Link{}, fmt.Errorf("error unmarshalink link file: %v", err)
	}
	return link, nil
}

// Remove deletes the file at the specified path.
// It takes the relative path of the file as input and returns an error if any.
func (c *LinkRepository) Remove(path string) error {
	absPath := filepath.Join(c.rootPath, path)
	err := os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

// RemoveDir removes the directory at the specified path.
// It takes the path of the directory to be removed as a parameter.
// Returns an error if the directory removal fails.
func (c *LinkRepository) RemoveDir(path string) error {
	p := filepath.Join(c.rootPath, path)
	err := os.Remove(p)
	return err
}

// Rename renames a file or directory from the old path to the new path.
// It takes the old path and the new path as parameters and returns an error if any.
func (c *LinkRepository) Rename(oldPath string, newPath string) error {
	o := filepath.Join(c.rootPath, oldPath)
	n := filepath.Join(c.rootPath, newPath)
	err := os.Rename(o, n)
	if err != nil {
		return err
	}
	return nil
}

// CreateDir creates a directory at the specified path.
// If the directory already exists, it does nothing.
// It returns an error if there was a problem creating the directory.
func (c *LinkRepository) CreateDir(path string) (err error) {
	absPath := filepath.Join(c.rootPath, path)
	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// GetSubFiles returns a list of sub-files in the specified directory path.
// It takes a path string as input and returns a slice of os.FileInfo and an error.
func (c *LinkRepository) GetSubFiles(path string) ([]os.FileInfo, error) {
	// Make sure the path is a directory
	if c.IsDir(path) {
		return nil, ErrPathIsNotDir
	}
	// Read the sub-files
	p := filepath.Join(c.rootPath, path)
	file, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("error opening dir to Read sub files: %v", err)
	}
	defer file.Close()
	subFiles, _ := file.Readdir(0)
	return subFiles, nil
}

// IsDir checks if the given path is a valid directory.
// It returns true if the path is a directory.
// It returns false if the path is not a directory or if there was an issue accessing the file system.
func (c *LinkRepository) IsDir(path string) bool {
	p := filepath.Join(c.rootPath, path)
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// IsFile checks if the given path corresponds to a valid file.
// It returns true if the path is a file.
// It returns false if the path is not a file or if there was an issue accessing the file system.
func (c *LinkRepository) IsFile(path string) bool {
	p := filepath.Join(c.rootPath, path)
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

// RemoveVaultLink removes the vault link file for the specified path.
// It takes the path of the link file as input and returns an error if any.
func (c *LinkRepository) RemoveVaultLink(path string) error {
	absPath := filepath.Join(c.rootPath, path, ".vault")
	err := os.Remove(absPath)
	if err != nil {
		return ErrRemovingVaultLinkFile
	}
	return nil
}

// getVaultLink retrieves the vault link for the file located at the specified path.
// It takes a path string as input and returns a core.VaultLink and an error.
func (c *LinkRepository) GetFileVaultLink(path string) (core.VaultLink, error) {
	dir := filepath.Dir(path)
	vaultLink, err := c.GetVaultLinkByPath(dir)
	return vaultLink, err
}

// IsValidPath checks if the given path is a valid path.
func (c *LinkRepository) IsValidPath(path string) bool {
	absPath := filepath.Join(c.rootPath, path)
	_, err := os.Stat(absPath)
	return err == nil
}
