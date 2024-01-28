package manager

import (
	"os"
	"path/filepath"
)

type DirWalker struct {
	rootPath string
	name     string
	mn       *Manager
	force    bool
}

type walkResult struct {
	list []file
}

type file struct {
	path string
}

func (mn *Manager) NewDirWalker(rootPath string, name string, force bool) *DirWalker {
	return &DirWalker{
		rootPath: rootPath,
		name:     name,
		force:    force,
		mn:       mn,
	}
}

func (f *DirWalker) read() (*walkResult, error) {
	files, err := walkDir(f.rootPath)
	if err != nil {
		return nil, err
	}
	res := walkResult{
		files,
	}
	return &res, nil
}

func walkDir(rootPath string) ([]file, error) {
	var files []file

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

		f := file{
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

func (f *DirWalker) Upload() error {
	res, err := f.read()
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
