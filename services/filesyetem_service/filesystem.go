package filesyetem_service

import (
	"ctb-cli/core"
	"ctb-cli/repositories"
	"ctb-cli/services/object_service"
	"fmt"
	"io/fs"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	objectService object_service.Service
	linkRepo      *repositories.LinkRepository
	keyService    core.KeyService

	openToWrite map[string]openToWrite
}

// openToWrite is a map of files open for writing
type openToWrite struct {
	id string
}

// Make sure FileSystem implements the FileSystemService interface
var _ core.FileSystemService = &FileSystem{}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(keyService core.KeyService, objectSerivce object_service.Service, linkRepository *repositories.LinkRepository) *FileSystem {
	fileSys := FileSystem{
		objectService: objectSerivce,
		linkRepo:      linkRepository,
		keyService:    keyService,
		openToWrite:   make(map[string]openToWrite),
	}

	return &fileSys
}

// CreateDir creates a directory at the specified path.
// It first creates a vault in the specified path and then creates the directory in the link repository.
// If any error occurs during the process, it returns the error.
func (f *FileSystem) CreateDir(path string) error {
	//Create vault in the specified path
	err := f.CreateVaultInPath(path)
	if err != nil {
		return err
	}
	//Create directory in link repo
	return f.linkRepo.CreateDir(path)
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
		vaultLink, err := f.linkRepo.GetVaultLinkByPath(parentPath)
		if err != nil {
			return err
		}
		parentVaultId = vaultLink.VaultId
	}
	//Create vault in the parent vault using the key service
	vault, err := f.keyService.CreateVault(parentVaultId, path)
	if err != nil {
		return err
	}
	//Create vault link
	link := core.NewVaultLink(vault.Id, vault.KeyId)
	err = f.linkRepo.InsertVaultLink(path, link)
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
		//Ignore .vault and .object folders
		if subFile.Name() == ".vault" || subFile.Name() == ".object" {
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
				size:  link.Size,
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
	//Remove vault link
	err := f.linkRepo.RemoveVaultLink(path)
	if err != nil {
		return err
	}
	//Remove directory in link repo
	return f.linkRepo.RemoveDir(path)
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
	_ = f.linkRepo.Create(path, core.Link{
		ObjectId: id,
		Size:     0,
	})
	//Create file in object service
	err = f.objectService.Create(id)
	if err != nil {
		return err
	}
	//Add file to open to write
	f.openToWrite[path] = openToWrite{id: id}
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
	id := link.ObjectId
	//Write file using object service
	n, err = f.objectService.Write(id, buff, ofst)
	//Update file size in link repo
	if link, _ := f.linkRepo.GetByPath(path); link.Size < ofst+int64(len(buff)) {
		link.Size = ofst + int64(len(buff))
		err = f.linkRepo.Update(path, link)
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
	oldId := link.ObjectId
	newId, _ = core.NewUid()
	link.ObjectId = newId
	err = f.linkRepo.Update(path, link)
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
	dir := filepath.Dir(path)
	//Get file link
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return 0, err
	}
	//Get file vault link
	vaultLink, vaultPath, err := f.linkRepo.GetFileVaultLink(path)
	if err != nil {
		return 0, err
	}
	//Get file key id
	keyId, err := f.objectService.GetKeyIdByObjectId(link.ObjectId, dir)
	if err != nil {
		return 0, err
	}
	//Get file key
	key, err := f.keyService.Get(keyId, vaultLink.VaultId, vaultPath)
	if err != nil {
		return 0, err
	}
	//Read file
	return f.objectService.Read(link.ObjectId, dir, buff, ofst, key)
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
	link.Size = size
	err = f.linkRepo.Update(path, link)
	if err != nil {
		return err
	}
	//Resize file in object service
	err = f.objectService.Truncate(link.ObjectId, size)
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
	oldVault, err := f.linkRepo.GetVaultLinkByPath(oldVaultPath)
	if err != nil {
		return err
	}
	newVaultPath := filepath.Dir(newPath)
	newVault, err := f.linkRepo.GetVaultLinkByPath(newVaultPath)
	if err != nil {
		return err
	}
	if isDir {
		//If the path is a directory, move the vault to the new parent vault
		vault, err := f.linkRepo.GetVaultLinkByPath(oldPath)
		if err != nil {
			return err
		}
		err = f.keyService.MoveVault(vault.VaultId, oldPath, newPath, oldVault.VaultId, oldVaultPath, newVault.VaultId, newVaultPath)
		if err != nil {
			return err
		}
	} else {
		//If the path is a file, move the file key to the new vault
		obj, err := f.linkRepo.GetByPath(oldPath)
		if err != nil {
			return err
		}
		oldDir := filepath.Dir(oldPath)
		keyId, err := f.objectService.GetKeyIdByObjectId(obj.ObjectId, oldDir)
		if err != nil {
			return err
		}
		//Move the file key to the new vault
		err = f.keyService.MoveKey(keyId, oldVault.VaultId, oldVaultPath, newVault.VaultId, newVaultPath)
		if err != nil {
			return err
		}
		//If the file moved to a different directory, change the directory of the file in the object service
		newDir := filepath.Dir(newPath)
		if newDir != oldDir {
			//Change the directory of the file in the object service
			err = f.objectService.ChangeDir(obj.ObjectId, oldDir, newDir)
			if err != nil {
				return err
			}
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
func (f *FileSystem) Commit(path string) error {
	_, ex := f.openToWrite[path]
	if ex {
		//Remove file from open to write
		delete(f.openToWrite, path)
		//Get vault link
		link, _ := f.linkRepo.GetByPath(path)
		vaultLink, vaultPath, err := f.linkRepo.GetFileVaultLink(path)
		if err != nil {
			return err
		}
		//Generate key in vault
		keyInfo, err := f.keyService.GenerateKeyInVault(vaultLink.VaultId, vaultPath)
		if err != nil {
			return err
		}
		//Commit changes
		dir := filepath.Dir(path)
		return f.objectService.Commit(link, dir, keyInfo)
	}
	return nil
}

// OpenInWrite opens the file at the specified path for writing.
// If the file is not already open for writing, it assigns a new ID to the file and adds it to the list of files open for writing.
// Returns an error if there was an issue changing the file ID or if the file is already open for writing.
func (f *FileSystem) OpenInWrite(path string) error {
	_, ex := f.openToWrite[path]
	if !ex {
		newId, err := f.changeFileId(path)
		if err != nil {
			return err
		}
		f.openToWrite[path] = openToWrite{id: newId}
	}
	return nil
}

// GetUserFileAccess returns the file mode for a given path and whether it is a directory.
// It checks the user's access to the file or directory and returns the corresponding file mode.
// If the user has access, it returns 0777, otherwise it returns 0000.
// If the path is a directory, it checks if there are any sub-files that the user has access to.
// If there are, it returns 0555, otherwise it returns 0000.
func (f *FileSystem) GetUserFileAccess(path string, isDir bool) fs.FileMode {
	//Get vault link
	vaultLink, vaultPath, err := f.linkRepo.GetFileVaultLink(path)
	if err != nil {
		return 0000
	}
	//If user has access to vault, he has access to the file
	if _, err := f.keyService.Get(vaultLink.KeyId, vaultLink.VaultId, vaultPath); err == nil {
		return 0777
	}
	//If the path is a file
	if !isDir {
		//Get file key id
		link, err := f.linkRepo.GetByPath(path)
		if err != nil {
			return 0000
		}
		dir := filepath.Dir(path)
		keyId, err := f.objectService.GetKeyIdByObjectId(link.ObjectId, dir)
		if err != nil {
			return 0000
		}
		//If user has access to file key, he has access to the file
		if _, err := f.keyService.Get(keyId, vaultLink.VaultId, vaultPath); err == nil {
			return 0777
		}
		//If user does not have access to file key, he does not have access to the file
		return 0000
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

// getDiskUsage returns the total and free bytes available in the directory's disk partition
func (f *FileSystem) GetDiskUsage() (totalBytes, freeBytes uint64, err error) {
	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(f.linkRepo.GetRootPath()),
		&freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return 0, 0, err
	}
	return totalNumberOfBytes, freeBytesAvailable, nil
}
