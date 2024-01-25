package filesyetem

// Persist defines an interface for a file system
type Persist interface {
	SavePath(path string, key string) error
	GetPath(path string) (string, error)
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
func (p *FileSystem) SavePath(id string, path string) error {
	return p.persist.SavePath(path, id)
}

// GetPath retrieves a path
func (p *FileSystem) GetPath(path string) (string, error) {
	return p.persist.GetPath(path)
}
