package filesystem_service

import (
	"ctb-cli/core"
	"ctb-cli/repositories"
	"ctb-cli/services/config_service"
	"ctb-cli/services/object_service"
	"fmt"
	"io/fs"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	objectService object_service.Service
	linkRepo      *repositories.LinkRepository
	vaultRepo     repositories.VaultRepository
	keyService    core.KeyService
	configService config_service.ConfigService
}

// Make sure FileSystem implements the FileSystemService interface
var _ core.FileSystemService = &FileSystem{}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(
	keyService core.KeyService,
	objectSerivce object_service.Service,
	linkRepository *repositories.LinkRepository,
	vaultRepo repositories.VaultRepository,
	configService config_service.ConfigService,
) *FileSystem {
	fileSys := FileSystem{
		objectService: objectSerivce,
		linkRepo:      linkRepository,
		vaultRepo:     vaultRepo,
		keyService:    keyService,
		configService: configService,
	}

	return &fileSys
}

// CreateDir creates a directory at the specified path.
// It creates the directory in the link repository and creates a vault in the specified path.
// Returns an error if any operation fails.
func (f *FileSystem) CreateDir(path string) error {
	err := f.linkRepo.CreateDir(path)
	if err != nil {
		return err
	}
	//Create vault in the specified path
	err = f.CreateVaultInPath(path)
	if err != nil {
		return err
	}
	err = f.configService.InitConfig(path)
	if err != nil {
		return err
	}
	return nil
}

// CreateVaultInPath creates a new vault in the specified path.
// If the path is not the root directory, it gets the parent vault ID and creates the new vault in the parent vault.
// It uses the key service to create the vault and inserts the vault link into the link repository.
// Returns an error if any operation fails.
func (f *FileSystem) CreateVaultInPath(path string) error {
	parentVaultId := ""
	//If the path is not the root directory, get the parent vault id
	if filepath.Clean(path) != string(filepath.Separator) {
		parentPath := filepath.Dir(path)
		vault, err := f.vaultRepo.GetVaultByPath(parentPath)
		if err != nil {
			return err
		}
		parentVaultId = vault.Id
	}
	//Create vault in the parent vault using the key service
	_, err := f.keyService.CreateVault(parentVaultId, path)
	if err != nil {
		return err
	}
	return nil
}

// RemovePath removes the file or directory at the specified path.
func (f *FileSystem) RemovePath(path string) (err error) {
	return f.linkRepo.Remove(path)
}

// GetSubFiles returns a list of sub files in the specified path.
// It ignores ".vault" files and creates file or directory info based on the sub file type.
// For directories, it checks the user's access to the directory and sets the mode accordingly.
// For files, it retrieves the file link and sets the size and user access mode.
// The function returns the list of file info and any error encountered during the process.
func (f *FileSystem) GetSubFiles(path string) (res []fs.FileInfo, err error) {
	//Get sub files in link repo
	subFiles, err := f.linkRepo.GetSubFiles(path)
	if err != nil {
		return nil, err
	}
	//Create list of file info
	var infos []fs.FileInfo
	//Iterate through sub files
	for _, subFile := range subFiles {
		//Ignore .vault, .key-share, and .object folders
		if subFile.Name() == ".meta" {
			continue
		}
		if subFile.IsDir() {
			//If sub file is a directory, create directory info
			var info fs.FileInfo = FileInfo{
				isDir: true,
				name:  subFile.Name(),
				//Check user access to directory (Read only if user has access to at least one file in the directory)
				mode: f.GetUserFileAccess(filepath.Join(path, subFile.Name()), true),
			}
			//Add directory info to list
			infos = append(infos, info)
			continue
		} else {
			//If sub file is a file, create file info
			p := filepath.Join(path, subFile.Name())
			//Get file link
			link, err := f.linkRepo.GetByPath(p)
			if err != nil {
				return nil, fmt.Errorf("error reading file size: %v", err)
			}
			var info fs.FileInfo = FileInfo{
				isDir: false,
				name:  subFile.Name(),
				size:  link.Data.Size,
				//Check user access to file
				mode: f.GetUserFileAccess(filepath.Join(path, subFile.Name()), false),
			}
			//Add file info to list
			infos = append(infos, info)
		}

	}
	return infos, nil
}

// RemoveDir removes a directory at the specified path.
// It first removes the vault link associated with the directory,
// and then removes the directory itself from the link repository.
// If any error occurs during the removal process, it is returned.
func (f *FileSystem) RemoveDir(path string) error {
	// Remove vault
	err := f.vaultRepo.RemoveVault(path)
	if err != nil {
		return err
	}
	//Remove Share folder
	err = f.linkRepo.RemoveDir(path)
	if err != nil {
		return err
	}
	return nil

}

// CreateFile creates a new file at the specified path.
// It generates a new file ID, creates a file link, and creates the file in the object service.
// The file is then added to the list of files open for writing.
func (f *FileSystem) CreateFile(path string) (err error) {
	//Create new file id
	id, err := core.NewUid()
	if err != nil {
		return err
	}
	//Create file link
	_ = f.linkRepo.Create(core.Link{
		Data: core.LinkData{
			ObjectId: id,
			Size:     0,
		},
		Path: path,
	})
	//Create file in object service
	err = f.objectService.Create(id)
	if err != nil {
		return err
	}
	return
}

// Write writes the given byte slice to the file at the specified path, starting at the specified offset.
// It returns the number of bytes written and any error encountered.
func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	//Open file in write
	if err := f.OpenInWrite(path); err != nil {
		return 0, err
	}
	//Get file link
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return 0, err
	}
	//Write file using object service
	n, err = f.objectService.Write(link.Id(), buff, ofst)
	//Update file size in link repo
	if link, _ := f.linkRepo.GetByPath(path); link.Data.Size < ofst+int64(len(buff)) {
		link.Data.Size = ofst + int64(len(buff))
		err = f.linkRepo.Update(link)
		if err != nil {
			return 0, err
		}
	}
	return
}

// changeFileId changes the ID of a file identified by the given path.
// It retrieves the file link from the link repository, updates the ID in the link repository,
// and moves the file in the object service to the new ID.
// If any error occurs during the process, it is returned along with an empty string for the new ID.
// Otherwise, the new ID is returned along with a nil error.
func (f *FileSystem) changeFileId(path string) (newId string, err error) {
	//Get file link
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return "", err
	}
	//Change file id in link repo
	oldId := link.Id()
	newId, _ = core.NewUid()
	link.Data.ObjectId = newId
	err = f.linkRepo.Update(link)
	if err != nil {
		return "", err
	}
	//Move file in object service (Move the file in object cache to the new id)
	err = f.objectService.Move(oldId, newId)
	if err != nil {
		return "", err
	}
	return newId, nil
}

// Read reads data from a file at the specified path into the provided buffer starting from the given offset.
// It returns the number of bytes read and any error encountered.
func (f *FileSystem) Read(path string, buff []byte, ofst int64) (n int, err error) {
	//Get file link
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return 0, err
	}
	//Get file key
	key, err := f.getKeyByPath(path)
	if err != nil {
		return 0, err
	}
	//Read file
	return f.objectService.Read(link, buff, ofst, key)
}

// Resize resizes a file to the specified size.
// It opens the file in write mode, updates the file size in the link repository,
// and truncates the file in the object service to the specified size.
// If any error occurs during the process, it returns the error.
func (f *FileSystem) Resize(path string, size int64) (err error) {
	//Open file in write
	if err := f.OpenInWrite(path); err != nil {
		return err
	}
	//Get file link
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return err
	}
	//Resize file in link repo
	link.Data.Size = size
	err = f.linkRepo.Update(link)
	if err != nil {
		return err
	}
	//Resize file in object service
	err = f.objectService.Truncate(link.Id(), size)
	if err != nil {
		return err
	}
	return nil
}

// Rename renames a file or directory from the oldPath to the newPath.
// If the oldPath and newPath are in different directories, the file or directory is moved to the new location.
// If the path is a directory, the vault is moved to the new parent vault.
// If the path is a file, the file key is moved to the new vault.
// Returns an error if any operation fails.
func (f *FileSystem) Rename(oldPath string, newPath string) (err error) {
	//Check if the path is a directory
	isDir := f.linkRepo.IsDir(oldPath)
	//Get the vault links for the oldPath and newPath
	oldVaultPath := filepath.Dir(oldPath)
	oldVault, err := f.vaultRepo.GetVaultByPath(oldVaultPath)
	if err != nil {
		return err
	}
	newVaultPath := filepath.Dir(newPath)
	newVault, err := f.vaultRepo.GetVaultByPath(newVaultPath)
	if err != nil {
		return err
	}
	if isDir {
		//If the path is a directory, move the vault to the new parent vault
		vault, err := f.vaultRepo.GetVaultByPath(oldPath)
		if err != nil {
			return err
		}
		err = f.keyService.MoveVault(vault.Id, oldPath, newPath, oldVault.Id, oldVaultPath, newVault.Id, newVaultPath)
		if err != nil {
			return err
		}
	} else {
		//If the path is a file, move the file key to the new vault
		link, err := f.linkRepo.GetByPath(oldPath)
		if err != nil {
			return err
		}
		keyId, err := f.objectService.GetKeyIdByObjectId(link)
		if err != nil {
			return err
		}
		//Move the file key to the new vault
		err = f.keyService.MoveKey(keyId, oldVault.Id, oldVaultPath, newVault.Id, newVaultPath)
		if err != nil {
			return err
		}
		//Change the path of the file in the object service
		err = f.objectService.ChangePath(link, newPath)
		if err != nil {
			return err
		}
	}
	return f.linkRepo.Rename(oldPath, newPath)
}

// Commit commits changes made to a file at the specified path.
// If the file is open for writing, it removes it from the list of open files,
// retrieves the link associated with the path, generates a key in the vault,
// and commits the changes using the object service.
// Returns an error if there was an issue retrieving the vault link or generating the key.
// Returns nil if the file is not open for writing.
// If the file is not open for writing, it removes the file from the object cache.
func (f *FileSystem) Commit(path string) error {
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return err
	}
	ex := f.objectService.IsOpenForWrite(link)
	// If the file is open for writing
	if ex {
		//Get vault
		link, _ := f.linkRepo.GetByPath(path)
		vault, vaultPath, err := f.vaultRepo.GetFileVault(path)
		if err != nil {
			return err
		}
		//Generate key in vault
		keyInfo, err := f.keyService.GenerateKeyInVault(vault.Id, vaultPath)
		if err != nil {
			return err
		}
		//Commit changes
		return f.objectService.Commit(link, keyInfo)
	} else {
		//Remove file from object cache if it is not open for writing
		link, err = f.linkRepo.GetByPath(path)
		if err != nil {
			return err
		}
		err = f.objectService.RemoveFromCache(link.Id())
		if err != nil {
			return err
		}
	}
	return nil
}

// OpenInWrite opens the file at the specified path for writing.
// If the file is not already open for writing, it assigns a new ID to the file and adds it to the list of files open for writing.
// Returns an error if there was an issue changing the file ID or if the file is already open for writing.
func (f *FileSystem) OpenInWrite(path string) error {
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return err
	}
	ex := f.objectService.IsOpenForWrite(link)
	if !ex {
		// Make sure the file is available in the cache
		link, err := f.linkRepo.GetByPath(path)
		if err != nil {
			return err
		}
		key, err := f.getKeyByPath(path)
		if err != nil {
			return err
		}
		err = f.objectService.AvailableInCache(link, key)
		if err != nil {
			return err
		}
		// Change the file ID
		_, err = f.changeFileId(path)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetUserFileAccess returns the file mode for a given path and whether it is a directory.
// It checks the user's access to the file or directory and returns the corresponding file mode.
// If the user has access, it returns 0777, otherwise it returns 0000.
// If the path is a directory, it checks if there are any sub-files that the user has access to.
// If there are, it returns 0555, otherwise it returns 0000.
// @TODO: Improve this function
func (f *FileSystem) GetUserFileAccess(path string, isDir bool) fs.FileMode {
	//If the path is a file, find the access based on it's
	if !isDir {
		dir := filepath.Dir(path)
		perm := f.GetUserFileAccess(dir, true)
		if perm == 0777 {
			return 0777
		} else {
			return 0000
		}
	}
	vault, err := f.vaultRepo.GetVaultByPath(path)
	if err != nil {
		return 0000
	}
	parentVaultPath, parentVault, err := f.vaultRepo.GetVaultParent(path)
	if err != nil {
		return 0000
	}
	//If user has access to vault, he has access to the file
	if _, err := f.keyService.Get(vault.KeyId, parentVault.Id, parentVaultPath); err == nil {
		return 0777
	}
	//If the path is a directory get all sub files
	subFiles, err := f.linkRepo.GetSubFiles(path)
	if err != nil {
		return 0000
	}
	//Check if we have a file in the directory that the user has access to
	for _, subFile := range subFiles {
		if subFile.Mode()&0777 != 0 {
			return 0555
		}
	}
	return 0000
}

// keyFileKey returns the key for a given file.
func (f *FileSystem) getKeyByPath(path string) (*core.KeyInfo, error) {
	//Get file linkc
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return nil, err
	}
	//Get file vault
	vault, vaultPath, err := f.vaultRepo.GetFileVault(path)
	if err != nil {
		return nil, err
	}
	//Get file key id
	keyId, err := f.objectService.GetKeyIdByObjectId(link)
	if err != nil {
		return nil, err
	}
	//Get file key
	key, err := f.keyService.Get(keyId, vault.Id, vaultPath)
	if err != nil {
		return nil, err
	}
	return key, nil
}
