package repositories

import (
	"ctb-cli/core"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrUserNotJoined = errors.New("user not joined")
)

// KeyRepository KeyStorePersist is an interface for persisting keys
type KeyRepository interface {
	SaveDataKey(keyId, key, recipient string, path string) error
	GetDataKey(keyID string, userId string, path string) (string, error)
	DataKeyExist(keyId string, userId string, path string) bool
	IsUserJoined(userId string) bool
	ListUsers() ([]string, error)
	DeleteDataKey(keyID string, userId string, path string) error
}

type KeyRepositoryFile struct {
	rootPath string
}

var _ KeyRepository = &KeyRepositoryFile{}

func NewKeyRepositoryFile(rootPath string) *KeyRepositoryFile {
	return &KeyRepositoryFile{
		rootPath: rootPath,
	}
}

func (k *KeyRepositoryFile) SaveDataKey(keyId, key, recipient string, path string) error {
	datapath := k.getDataPath(recipient, path)
	err := os.MkdirAll(datapath, os.ModePerm)
	if err != nil {
		return err
	}
	p := filepath.Join(datapath, keyId)
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyRepositoryFile) GetDataKey(keyID string, userId string, path string) (string, error) {
	datapath := k.getDataPath(userId, path)
	if _, err := os.Stat(datapath); err != nil {
		return "", err
	}
	p := filepath.Join(datapath, keyID)
	file, err := os.Open(p)
	if os.IsNotExist(err) {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), err
}

// DataKeyExist checks if a data key with the given key ID exists for the specified user.
// It returns true if the data key exists, and false otherwise.
func (k *KeyRepositoryFile) DataKeyExist(keyId string, userId string, path string) bool {
	datapath := k.getDataPath(userId, path)
	if _, err := os.Stat(datapath); err != nil {
		return false
	}
	p := filepath.Join(datapath, keyId)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (k *KeyRepositoryFile) IsUserJoined(userId string) bool {
	users, err := k.GetJoinedUsers()
	if err != nil {
		return false
	}
	for _, user := range users {
		if user.Recipient == userId {
			return true
		}
	}
	return false
}

// ListUsers returns a list of users stored in the key repository.
func (k *KeyRepositoryFile) ListUsers() ([]string, error) {
	joinedUser, err := k.GetJoinedUsers()
	if err != nil {
		return nil, err
	}
	users := make([]string, 0)
	for _, user := range joinedUser {
		users = append(users, user.Recipient)
	}
	return users, nil
}

// DeleteDataKey deletes the data key associated with the given keyID and userId.
// It removes the file corresponding to the keyID from the user's data path.
// If an error occurs during the deletion process, it is returned.
func (k *KeyRepositoryFile) DeleteDataKey(keyID string, userId string, path string) error {
	datapath := k.getDataPath(userId, path)
	if _, err := os.Stat(datapath); err != nil {
		return err
	}
	p := filepath.Join(datapath, keyID)
	err := os.Remove(p)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyRepositoryFile) GetJoinedUsers() ([]core.JoinedUser, error) {
	return k.getJoinedUsersInPath("")
}

func (k *KeyRepositoryFile) getJoinedUsersInPath(path string) ([]core.JoinedUser, error) {
	list := make([]core.JoinedUser, 0)

	entries, err := os.ReadDir(k.getKeysPath(path))
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return list, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			list = append(list, core.JoinedUser{
				Recipient: entry.Name(),
				Path:      path,
			})
		}
	}

	subs, err := k.getSubFolders(path)
	if err != nil {
		return list, err
	}
	for _, sub := range subs {
		users, err := k.getJoinedUsersInPath(sub)
		if err != nil {
			return list, err
		}
		list = append(list, users...)
	}

	return list, nil
}

func (k *KeyRepositoryFile) getSubFolders(path string) ([]string, error) {
	list := make([]string, 0)

	entries, err := os.ReadDir(filepath.Join(k.rootPath, path))
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return list, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() != ".meta" {
				list = append(list, filepath.Join(path, entry.Name()))
			}
		}
	}

	return list, nil
}

func (k *KeyRepositoryFile) getDataPath(recipient string, path string) string {
	p := filepath.Join(k.getKeysPath(path), recipient)
	return p
}

func (k *KeyRepositoryFile) getKeysPath(path string) string {
	dir := filepath.Join(k.rootPath, path, ".meta", ".key-share")
	return dir
}
