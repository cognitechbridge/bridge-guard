package repositories

import (
	"ctb-cli/types"
	"encoding/json"
	"github.com/samber/lo"
	"io"
	"os"
	"path/filepath"
)

type RecipientRepository interface {
	GetRecipientByEmail(email string) (types.Recipient, error)
}

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

func (r *RecipientRepositoryFile) GetRecipientByEmail(email string) (types.Recipient, error) {
	path := filepath.Join(r.rootPath, "recipients.txt")
	file, _ := os.Open(path)
	data, _ := io.ReadAll(file)
	var list []types.Recipient
	_ = json.Unmarshal(data, &list)
	rec, _ := lo.Find(list, func(r types.Recipient) bool {
		return r.Email == email
	})
	return rec, nil
}
