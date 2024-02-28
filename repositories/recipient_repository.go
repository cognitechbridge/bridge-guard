package repositories

import (
	"ctb-cli/core"
	"encoding/json"
	"errors"
	"github.com/samber/lo"
	"io"
	"os"
	"path/filepath"
)

type RecipientRepository interface {
	GetRecipientByEmail(email string) (core.Recipient, error)
	InsertRecipient(recipient core.Recipient) error
}

var (
	ErrorWritingToRepository = errors.New("error writing to recipient repo")
)

// RecipientRepositoryFile File implementation of RecipientRepository
// @Todo: This implementation is not safe
type RecipientRepositoryFile struct {
	rootPath string
}

func NewRecipientRepositoryFile(rootPath string) *RecipientRepositoryFile {
	return &RecipientRepositoryFile{
		rootPath: rootPath,
	}
}

func (r *RecipientRepositoryFile) InsertRecipient(recipient core.Recipient) error {
	list, _ := r.openRepository()
	list = append(list, recipient)
	res, _ := json.Marshal(list)

	path := filepath.Join(r.rootPath, "recipients.txt")
	file, _ := os.OpenFile(path, os.O_RDWR, 0666)
	defer file.Close()
	_, err := file.WriteString(string(res))
	if err != nil {
		return ErrorWritingToRepository
	}
	return nil
}

func (r *RecipientRepositoryFile) GetRecipientByEmail(email string) (core.Recipient, error) {
	list, _ := r.openRepository()
	rec, _ := lo.Find(list, func(r core.Recipient) bool {
		return r.Email == email
	})
	return rec, nil
}

func (r *RecipientRepositoryFile) openRepository() ([]core.Recipient, error) {
	path := filepath.Join(r.rootPath, "recipients.txt")
	file, _ := os.Open(path)
	defer file.Close()
	data, _ := io.ReadAll(file)
	var list []core.Recipient
	_ = json.Unmarshal(data, &list)
	return list, nil
}
