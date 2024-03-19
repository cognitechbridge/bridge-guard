package share_service

import (
	"ctb-cli/core"
	"ctb-cli/repositories"
)

type Service struct {
	linkRepository *repositories.LinkRepository
	objectService  core.ObjectService
	keyStorer      core.KeyService
}

func NewService(
	keyStorer core.KeyService,
	linkRepository *repositories.LinkRepository,
	objectService core.ObjectService,
) *Service {
	return &Service{
		objectService:  objectService,
		keyStorer:      keyStorer,
		linkRepository: linkRepository,
	}
}

func (s *Service) ShareByPublicKey(regex string, publicKeyEncoded string) error {
	publicKeyBytes, err := core.DecodePublic(publicKeyEncoded)
	if err != nil {
		return err
	}
	files, _ := s.linkRepository.ListIdsByRegex(regex)
	for _, fileId := range files {
		keyId, err := s.objectService.GetKeyIdByObjectId(fileId)
		if err != nil {
			return err
		}
		err = s.keyStorer.Share(keyId, publicKeyBytes, publicKeyEncoded)
		if err != nil {
			return err
		}

	}
	return nil
}
