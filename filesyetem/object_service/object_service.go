package object_service

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

	//internal queues and channels
	encryptChan  chan encryptChanItem
	uploadChan   chan uploadChanItem
	encryptQueue *EncryptQueue
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
	service := Service{
		downloader:   dn,
		cache:        cache,
		objectRepo:   objectRepo,
		keystoreRepo: keystoreRepo,
		clientId:     clientId,
		encryptChan:  make(chan encryptChanItem, 10),
		uploadChan:   make(chan uploadChanItem, 10),
	}

	service.encryptQueue = service.NewEncryptQueue()

	go service.StartEncryptRoutine()
	go service.StartUploadRoutine()

	return service
}

func (o *Service) Read(id string, buff []byte, ofst int64) (n int, err error) {
	err = o.availableInCache(id)
	if err != nil {
		return 0, err
	}

	return o.cache.Read(id, buff, ofst)
}

func (o *Service) Write(id string, buff []byte, ofst int64) (n int, err error) {
	n, err = o.cache.Write(id, buff, ofst)
	o.encryptQueue.Enqueue(id)
	return n, err
}

func (o *Service) Create(id string) (err error) {
	err = o.cache.Create(id)
	if err != nil {
		return err
	}
	o.encryptQueue.Enqueue(id)
	return nil
}

func (o *Service) Move(oldId string, newId string) (err error) {
	return o.cache.Move(oldId, newId)
}

func (o *Service) Truncate(id string, size int64) (err error) {
	return o.cache.Truncate(id, size)
}

func (o *Service) IsInQueue(id string) bool {
	return o.encryptQueue.IsInQueue(id)
}

func (o *Service) availableInCache(id string) error {
	if o.cache.IsInCache(id) {
		return nil
	}
	if o.objectRepo.IsInRepo(id) == false {
		err := o.downloadToObject(id)
		if err != nil {
			return err
		}
	}
	err := o.decryptToCache(id)
	if err != nil {
		return err
	}
	return nil
}

func (o *Service) decryptToCache(id string) error {
	openObject, _ := o.objectRepo.OpenObject(id)
	defer openObject.Close()
	decryptedReader, _ := o.decryptReader(openObject, id)
	writer, err := o.cache.CacheObjectWriter(id)
	defer writer.Close()
	_, err = io.Copy(writer, decryptedReader)
	return err
}

func (o *Service) downloadToObject(id string) error {
	file, _ := o.objectRepo.CreateFile(id)
	err := o.downloader.Download(id, file)
	defer file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (o *Service) encryptWriter(writer io.Writer, fileId string) (write io.WriteCloser, err error) {
	recoveryItems, err := o.keystoreRepo.GetRecoveryItems()
	if err != nil {
		return nil, err
	}
	pair, err := recovery.GenerateKeyPair(recoveryItems)
	if err != nil {
		return nil, err
	}
	err = o.keystoreRepo.Insert(fileId, pair.Key)
	if err != nil {
		return nil, err
	}
	return file_crypto.NewWriter(writer, pair.Key, o.clientId, fileId, pair.RecoveryBlobs)
}

func (o *Service) decryptReader(reader io.Reader, fileId string) (read io.Reader, err error) {
	key, err := o.keystoreRepo.Get(fileId)
	if err != nil {
		return nil, err
	}
	read, err = file_crypto.NewReader(key, reader)
	return read, err
}
