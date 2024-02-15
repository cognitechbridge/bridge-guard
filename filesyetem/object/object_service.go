package object

import (
	"ctb-cli/crypto/file_crypto"
	"ctb-cli/crypto/recovery"
	"ctb-cli/filesyetem/object_cache"
	"ctb-cli/filesyetem/object_repository"
	"ctb-cli/types"
	"io"
)

type Service struct {
	cache        *object_cache.ObjectCache
	objectRepo   *object_repository.ObjectRepository
	downloader   Downloader
	keystoreRepo keystoreRepo
	clientId     string
}

type keystoreRepo interface {
	Get(keyID string) (*types.Key, error)
	Insert(keyID string, key types.Key) error
	GetRecoveryItems() ([]types.RecoveryItem, error)
}

type Downloader interface {
	Download(id string, writeAt io.WriterAt) error
	Upload(reader io.Reader, fileId string) error
}

func NewService(keystoreRepo keystoreRepo, clientId string, cache *object_cache.ObjectCache, objectRepo *object_repository.ObjectRepository, dn Downloader) Service {
	return Service{
		downloader:   dn,
		cache:        cache,
		objectRepo:   objectRepo,
		keystoreRepo: keystoreRepo,
		clientId:     clientId,
	}
}

func (f *Service) Read(id string, buff []byte, ofst int64) (n int, err error) {
	if f.cache.IsInCache(id) {
		return f.cache.Read(id, buff, ofst)
	}
	if f.objectRepo.IsInRepo(id) == false {
		err := f.downloadToObject(id, err)
		if err != nil {
			return 0, err
		}
	}
	err = f.decryptToCache(id, err)
	if err != nil {
		return 0, err
	}

	return f.cache.Read(id, buff, ofst)
}

func (f *Service) Write(id string, buff []byte, ofst int64) (n int, err error) {
	return f.cache.Write(id, buff, ofst)
}

func (f *Service) Create(id string) (err error) {
	return f.cache.Create(id)
}

func (f *Service) Move(oldId string, newId string) (err error) {
	return f.cache.Move(oldId, newId)
}

func (f *Service) Truncate(id string, size int64) (err error) {
	return f.cache.Truncate(id, size)
}

func (f *Service) decryptToCache(id string, err error) error {
	openObject, _ := f.objectRepo.OpenObject(id)
	defer openObject.Close()
	decryptedReader, _ := f.Decrypt(openObject, id)
	writer, err := f.cache.CacheObjectWriter(id)
	defer writer.Close()
	_, err = io.Copy(writer, decryptedReader)
	return err
}

func (f *Service) downloadToObject(id string, err error) error {
	file, _ := f.objectRepo.CreateFile(id)
	err = f.downloader.Download(id, file)
	defer file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (f *Service) Encrypt(writer io.Writer, fileId string) (write io.WriteCloser, err error) {
	recoveryItems, err := f.keystoreRepo.GetRecoveryItems()
	if err != nil {
		return nil, err
	}
	pair, err := recovery.GenerateKeyPair(recoveryItems)
	if err != nil {
		return nil, err
	}
	err = f.keystoreRepo.Insert(fileId, pair.Key)
	if err != nil {
		return nil, err
	}
	return file_crypto.NewWriter(writer, pair.Key, f.clientId, fileId, pair.RecoveryBlobs)
}

func (f *Service) Decrypt(reader io.Reader, fileId string) (read io.Reader, err error) {
	key, err := f.keystoreRepo.Get(fileId)
	if err != nil {
		return nil, err
	}
	read, err = file_crypto.NewReader(key, reader)
	return read, err
}
