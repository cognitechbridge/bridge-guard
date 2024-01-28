package manager

import (
	"os"
	"path/filepath"
)

type FileWalker struct {
	rootPath string
	name     string
	mn       *Manager
	force    bool
}

type WalkResult struct {
	list []File
}

type File struct {
	path string
}

func (mn *Manager) NewFileWalker(rootPath string, name string, force bool) *FileWalker {
	return &FileWalker{
		rootPath: rootPath,
		name:     name,
		force:    force,
		mn:       mn,
	}
}

func (f *FileWalker) Read() (*WalkResult, error) {
	files, err := walkDir(f.rootPath)
	if err != nil {
		return nil, err
	}
	res := WalkResult{
		files,
	}
	return &res, nil
}

func walkDir(rootPath string) ([]File, error) {
	var files []File

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == rootPath {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Compute the relative path
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		f := File{
			path: relativePath,
		}
		files = append(files, f)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (f *FileWalker) Upload() error {
	res, err := f.Read()
	if err != nil {
		return err
	}
	for _, file := range res.list {
		name := filepath.Join(f.name, file.path)
		path := filepath.Join(f.rootPath, file.path)
		u := f.mn.NewUploader(path, name, f.force)
		_, err := u.Upload()
		if err != nil {
			return err
		}
	}
	return nil
}
