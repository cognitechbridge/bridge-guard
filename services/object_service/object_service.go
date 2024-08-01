package object_service

import (
	"bytes"
	"crypto/md5"
	"ctb-cli/core"
	"ctb-cli/crypto/file_crypto"
	"ctb-cli/repositories"
	"errors"
	"io"

	log "github.com/sirupsen/logrus"
)

// Service represents the object service.
type Service struct {
	objectCacheRepo *repositories.ObjectCacheRepository
	objectRepo      *repositories.ObjectRepository
	downloader      core.CloudStorage

	uploadChan chan uploadChanItem
}

// Make sure Service implements the core.ObjectService interface
var _ core.ObjectService = (*Service)(nil)

// NewService creates a new instance of the object service.
// It takes in a cache repository, an object repository, and a cloud storage instance.
// It initializes the service with the provided repositories and channels for encryption and upload routines.
// It starts the encryption and upload routines in separate goroutines.
// It returns the initialized service.
func NewService(cache *repositories.ObjectCacheRepository, objectRepo *repositories.ObjectRepository, dn core.CloudStorage) Service {
	service := Service{
		downloader:      dn,
		objectCacheRepo: cache,
		objectRepo:      objectRepo,
		uploadChan:      make(chan uploadChanItem, 10),
	}

	//start the encryption and upload routines in separate goroutines
	go service.StartUploadRoutine()

	return service
}

// Read reads the object with the specified ID from the object service.
// It populates the provided buffer with the object data starting from the specified offset.
// Returns the number of bytes read and any error encountered.
func (o *Service) Read(link core.Link, buff []byte, ofst int64, key *core.KeyInfo) (n int, err error) {
	err = o.AvailableInCache(link, key)
	if err != nil {
		return 0, err
	}
	return o.objectCacheRepo.Read(link.Id(), buff, ofst)
}

// Write writes the given byte slice to the object cache repository at the specified offset.
// It returns the number of bytes written and any error encountered.
func (o *Service) Write(id string, buff []byte, ofst int64) (n int, err error) {
	n, err = o.objectCacheRepo.Write(id, buff, ofst)
	return n, err
}

// Create creates a new object with the specified ID.
// It returns an error if the object creation fails.
func (o *Service) Create(id string) (err error) {
	err = o.objectCacheRepo.Create(id)
	if err != nil {
		return err
	}
	return nil
}

// Move moves an object from the oldId to the newId.
// It returns an error if the move operation fails.
func (o *Service) Move(oldId string, newId string) (err error) {
	return o.objectCacheRepo.MoveToWrite(oldId, newId)
}

func (o *Service) ChangePath(link core.Link, newPath string) (err error) {
	log.Debugf("ChangePath: %s to %s", link.Id(), newPath)
	return o.objectRepo.ChangePath(link, newPath)
}

// Truncate truncates the object with the specified ID to the given size.
// It returns an error if the truncation operation fails.
func (o *Service) Truncate(id string, size int64) (err error) {
	return o.objectCacheRepo.Truncate(id, size)
}

// availableInCache makes sure that the object is available in the cache. It does the following:
// If the object is already in the cache, it immediately returns.
// If the object is not in the cache, it checks if the object is in the repository.
// If the object is not in the repository, it downloads the object and stores it in the repository.
// Finally, it decrypts the object and stores it in the cache.
// It returns an error if any error occurs during the process.
func (o *Service) AvailableInCache(link core.Link, key *core.KeyInfo) error {
	//check if object is already in cache, if yes, return
	if o.objectCacheRepo.IsInCache(link.Id()) {
		return nil
	}
	//if not, check if object is in repo, if not, download it
	if !o.objectRepo.IsInRepo(link) {
		//download object
		err := o.downloadToObject(link)
		if err != nil {
			return err
		}
	}
	//decrypt object to cache
	err := o.decryptToCache(link, key)
	if err != nil {
		return err
	}
	return nil
}

// decryptToCache decrypts an object with the given ID using the provided key and writes the decrypted object to the cache.
// It opens the object from the repository, creates an unencrypted reader from the encrypted file and the key,
// and writes the decrypted object to the cache using the created writer and reader.
// The decrypted object is written to the cache using the object's ID as the cache key.
// If any error occurs during the decryption or writing process, it is returned.
func (o *Service) decryptToCache(link core.Link, key *core.KeyInfo) error {
	//open object from repo
	openObject, _ := o.objectRepo.OpenObject(link)
	defer openObject.Close()
	//Create an unencrypted reader from encrypted file (reader interface) and the key
	decryptedReader, _ := o.decryptReader(openObject, key)
	//Create a writer to write the decrypted object to the cache
	writer, err := o.objectCacheRepo.CacheObjectWriter(link.Id())
	if err != nil {
		return err
	}
	defer writer.Close()
	//Write the decrypted object to the cache using the created writer and reader
	_, err = io.Copy(writer, decryptedReader)
	return err
}

func (o *Service) downloadToObject(link core.Link) error {
	//create the file in the repository
	file, _ := o.objectRepo.CreateFile(link)
	//download the object and store it in the repository
	err := o.downloader.Download(link.Id(), file)
	defer file.Close()
	if err != nil {
		return err
	}
	return nil
}

// encryptWriter encrypts the data written to the provided writer using the specified key and file ID.
// It returns a new io.WriteCloser that wraps the original writer and performs encryption.
// The returned writer should be closed after the writing process is done to flush the remaining data and finalize the encryption.
// If any error occurs during the process, it returns an error.
func (o *Service) encryptWriter(writer io.Writer, fileId string, key *core.KeyInfo) (write io.WriteCloser, err error) {
	return file_crypto.NewWriter(writer, key, fileId)
}

// decryptReader decrypts the data from the given reader using the provided key.
// It returns a new reader with the decrypted data and any error encountered.
func (o *Service) decryptReader(reader io.Reader, key *core.KeyInfo) (read io.Reader, err error) {
	// Parse the encrypted file and create an encrypted stream
	_, enc, err := file_crypto.Parse(reader)
	if err != nil {
		return nil, err
	}
	// Decrypt the encrypted stream using the key
	read, err = enc.Decrypt(key)
	return read, err
}

// GetKeyIdByObjectId retrieves the key ID associated with the given object ID.
// It opens the object from the repository, parses the encrypted file, and returns the key ID from the header.
// If any error occurs during the process, it returns an empty string and the error.
func (o *Service) GetKeyIdByObjectId(link core.Link) (string, error) {
	//open object from repo
	reader, err := o.objectRepo.OpenObject(link)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	//parse the encrypted file and get the header
	header, _, err := file_crypto.Parse(reader)
	if err != nil {
		return "", err
	}
	//return the key id from the header
	return header.KeyId, nil
}

// Commit adds the object to the encrypt channel queue.
// It takes a link and a key as parameters and returns an error if any.
func (o *Service) Commit(link core.Link, key *core.KeyInfo) error {
	o.objectCacheRepo.AdToCommitting(link.Id())
	// Encrypt the object
	o.encrypt(link, key)
	return nil
}

// RemoveFromCache removes the object with the specified ID from the cache.
// It returns an error if the removal operation fails.
// If the object is not in the cache, it returns nil (no error).
func (o *Service) RemoveFromCache(id string) error {
	return o.objectCacheRepo.FlushFromRead(id)
}

// IsOpenForWrite returns true if the object with the specified ID is open for writing.
func (o *Service) IsOpenForWrite(link core.Link) bool {
	return o.objectCacheRepo.IsOpenForWrite(link.Id())
}

// ValidateObject validates the object with the specified ID.
// It returns an error if the validation fails.
func (o *Service) ValidateObject(link core.Link, key *core.KeyInfo) error {
	//open object from repo
	openObject, _ := o.objectRepo.OpenObject(link)
	defer openObject.Close()
	//Create an unencrypted reader from encrypted file (reader interface) and the key
	decryptedReader, _ := o.decryptReader(openObject, key)
	//Create a hash object to calculate the MD5 hash
	hash1 := md5.New()
	//Copy the decrypted object to the hash object
	_, err := io.Copy(hash1, decryptedReader)
	if err != nil {
		return err
	}
	//Get the resulting hash sum
	hashSum1 := hash1.Sum(nil)

	// Create a hash object to calculate the MD5 hash in the cache
	hash2 := md5.New()
	// Read the decrypted object and update the hash
	buff := make([]byte, 1024*1024) // 1 MB
	nt := (int64)(0)
	for {
		n, err := o.objectCacheRepo.Read(link.Id(), buff, nt)
		if n > 0 {
			if _, err := hash2.Write(buff[:n]); err != nil {
				return err
			}
		}
		nt += int64(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	if nt != link.Data.Size {
		return errors.New("Data size mismatch")
	}
	// Get the resulting hash sum
	hashSum2 := hash2.Sum(nil)

	// Compare the hash sums
	if !bytes.Equal(hashSum1, hashSum2) {
		return errors.New("Hash sum mismatch")
	}
	return nil
}
