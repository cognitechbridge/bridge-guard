package filesyetem

// Persist defines an interface for a file system
type Persist interface {
	SavePath(path string, key string, isDir bool) error
	GetPath(path string) (string, error)
	RemovePath(path string) error
	PathExist(path string) (bool, error)
}

// FileSystem implements the FileSystem interface
type FileSystem struct {
	persist Persist
}

// NewPersistFileSystem creates a new instance of PersistFileSystem
func NewPersistFileSystem(persist Persist) *FileSystem {
	return &FileSystem{
		persist: persist,
	}
}

// SavePath saves a path with the given id
func (p *FileSystem) SavePath(id string, path string, isDir bool) error {
	return p.persist.SavePath(path, id, isDir)
}

// GetPath retrieves a path
func (p *FileSystem) GetPath(path string) (string, error) {
	return p.persist.GetPath(path)
}

func (p *FileSystem) PathExist(path string) bool {
	_, err := p.persist.PathExist(path)
	return err == nil
}

func (p *FileSystem) RemovePath(path string) error {
	err := p.persist.RemovePath(path)
	return err
}
