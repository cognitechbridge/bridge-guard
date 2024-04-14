package share_service

import (
	"ctb-cli/core"
	"ctb-cli/repositories"
)

type Service struct {
	linkRepository  *repositories.LinkRepository
	vaultRepository repositories.VaultRepository
	objectService   core.ObjectService
	keyService      core.KeyService
}

func NewService(
	keyService core.KeyService,
	linkRepository *repositories.LinkRepository,
	vaultRepository repositories.VaultRepository,
	objectService core.ObjectService,
) *Service {
	return &Service{
		objectService:   objectService,
		keyService:      keyService,
		linkRepository:  linkRepository,
		vaultRepository: vaultRepository,
	}
}

// ShareByPublicKey shares a file or directory located at the specified path with the given public key.
// It retrieves the key ID associated with the path, decodes the provided public key, and then calls the Share method of the key service.
// If any error occurs during the process, it is returned.
func (s *Service) ShareByPublicKey(path string, publicKeyEncoded string) error {
	keyId, startVaultId, err := s.GetKeyIdByPath(path)
	if err != nil {
		return err
	}
	publicKeyBytes, err := core.DecodePublic(publicKeyEncoded)
	if err != nil {
		return err
	}
	err = s.keyService.Share(keyId, startVaultId, publicKeyBytes, publicKeyEncoded)
	if err != nil {
		return err
	}

	return nil
}

// GetKeyIdByPath retrieves the key ID associated with the given path.
// If the path represents a directory, it retrieves the key ID from the vault link associated with the path.
// If the path represents a file, it retrieves the key ID from the object service using the object ID associated with the path.
// The retrieved key ID is returned along with any error encountered during the process.
func (s *Service) GetKeyIdByPath(path string) (keyId string, startVaultId string, err error) {
	isDir := s.linkRepository.IsDir(path)
	if isDir {
		link, err := s.linkRepository.GetVaultLinkByPath(path)
		if err != nil {
			return "", "", err
		}
		vault, err := s.vaultRepository.GetVault(link.VaultId)
		if err != nil {
			return "", "", err
		}
		keyId = link.KeyId
		startVaultId = vault.ParentId
	} else {
		link, err := s.linkRepository.GetByPath(path)
		if err != nil {
			return "", "", err
		}
		keyId, err = s.objectService.GetKeyIdByObjectId(link.ObjectId)
		if err != nil {
			return "", "", err
		}
		vault, err := s.linkRepository.GetFileVaultLink(path)
		if err != nil {
			return "", "", err
		}
		startVaultId = vault.VaultId
	}
	return keyId, startVaultId, nil
}

// GetAccessList retrieves the access list for a given path.
// It returns the key access list and an error if any.
func (s *Service) GetAccessList(path string) (core.KeyAccessList, error) {
	isValid := s.linkRepository.IsValidPath(path)
	if !isValid {
		return nil, core.ErrInvalidPath
	}

	// Get the key ID and start vault ID associated with the path
	keyId, startVaultId, err := s.GetKeyIdByPath(path)
	if err != nil {
		return nil, err
	}

	return s.keyService.GetKeyAccessList(keyId, startVaultId)
}

// Unshare removes the sharing of a file or directory specified by the given path
// with the public key provided. It returns an error if the operation fails.
func (s *Service) Unshare(path string, publicKeyEncoded string) error {
	keyId, _, err := s.GetKeyIdByPath(path)
	if err != nil {
		return err
	}
	err = s.keyService.Unshare(keyId, publicKeyEncoded)
	if err != nil {
		return err
	}

	return nil
}
