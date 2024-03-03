package object_service

import (
	"ctb-cli/core"
	"ctb-cli/crypto/file_crypto"
	"ctb-cli/repositories"
	"io"
)

type Service struct {
	objectCacheRepo *repositories.ObjectCacheRepository
	objectRepo      *repositories.ObjectRepository
	downloader      core.CloudStorage
	keystore        core.KeyService
	userId          string

	//internal queues and channels
	encryptChan chan encryptChanItem
	uploadChan  chan uploadChanItem
}

func NewService(keystoreRepo core.KeyService, userId string, cache *repositories.ObjectCacheRepository, objectRepo *repositories.ObjectRepository, dn core.CloudStorage) Service {
	service := Service{
		downloader:      dn,
		objectCacheRepo: cache,
		objectRepo:      objectRepo,
		keystore:        keystoreRepo,
		userId:          userId,
		encryptChan:     make(chan encryptChanItem, 10),
		uploadChan:      make(chan uploadChanItem, 10),
	}

	go service.StartEncryptRoutine()
	go service.StartUploadRoutine()

	return service
}

func (o *Service) Read(id string, buff []byte, ofst int64, key *core.KeyInfo) (n int, err error) {
	err = o.availableInCache(id, key)
	if err != nil {
		return 0, err
	}

	return o.objectCacheRepo.Read(id, buff, ofst)
}

func (o *Service) Write(id string, buff []byte, ofst int64) (n int, err error) {
	n, err = o.objectCacheRepo.Write(id, buff, ofst)
	return n, err
}

func (o *Service) Create(id string) (err error) {
	err = o.objectCacheRepo.Create(id)
	if err != nil {
		return err
	}
	return nil
}

func (o *Service) Move(oldId string, newId string) (err error) {
	return o.objectCacheRepo.Move(oldId, newId)
}

func (o *Service) Truncate(id string, size int64) (err error) {
	return o.objectCacheRepo.Truncate(id, size)
}

func (o *Service) availableInCache(id string, key *core.KeyInfo) error {
	if o.objectCacheRepo.IsInCache(id) {
		return nil
	}
	if o.objectRepo.IsInRepo(id) == false {
		err := o.downloadToObject(id)
		if err != nil {
			return err
		}
	}
	err := o.decryptToCache(id, key)
	if err != nil {
		return err
	}
	return nil
}

func (o *Service) decryptToCache(id string, key *core.KeyInfo) error {
	openObject, _ := o.objectRepo.OpenObject(id)
	defer openObject.Close()
	decryptedReader, _ := o.decryptReader(openObject, key)
	writer, err := o.objectCacheRepo.CacheObjectWriter(id)
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

func (o *Service) encryptWriter(writer io.Writer, fileId string, vaultId string) (write io.WriteCloser, err error) {
	keyInfo, err := o.keystore.GenerateKeyInVault(vaultId)
	if err != nil {
		return nil, err
	}
	return file_crypto.NewWriter(writer, keyInfo, o.userId, fileId)
}

func (o *Service) decryptReader(reader io.Reader, key *core.KeyInfo) (read io.Reader, err error) {
	_, enc, err := file_crypto.Parse(reader)
	if err != nil {
		return nil, err
	}
	read, err = enc.Decrypt(key)
	return read, err
}

func (o *Service) GetKeyIdByObjectId(id string) (string, error) {
	reader, err := o.objectRepo.OpenObject(id)
	defer reader.Close()
	header, _, err := file_crypto.Parse(reader)
	if err != nil {
		return "", err
	}
	return header.KeyId, nil
}

func (o *Service) Commit(link core.Link, vaultId string) error {
	o.encryptChan <- encryptChanItem{id: link.ObjectId, vaultId: vaultId}
	return nil
}
