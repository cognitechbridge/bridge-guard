package share_service

import (
	"ctb-cli/core"
	"ctb-cli/repositories"
)

type Service struct {
	recipientRepository repositories.RecipientRepository
	linkRepository      *repositories.LinkRepository
	objectService       core.ObjectService
	keyStorer           core.KeyService
}

func NewService(
	recipientRepository repositories.RecipientRepository,
	keyStorer core.KeyService,
	linkRepository *repositories.LinkRepository,
	objectService core.ObjectService,
) *Service {
	return &Service{
		objectService:       objectService,
		recipientRepository: recipientRepository,
		keyStorer:           keyStorer,
		linkRepository:      linkRepository,
	}
}

func (s *Service) ShareByEmail(regex string, email string) error {
	rec, _ := s.recipientRepository.GetRecipientByEmail(email)
	files, _ := s.linkRepository.ListIdsByRegex(regex)
	for _, fileId := range files {
		keyId, err := s.objectService.GetKeyIdByObjectId(fileId)
		if err != nil {
			return err
		}
		publicBytes, err := rec.GetPublicBytes()
		if err != nil {
			return err
		}
		s.keyStorer.Share(keyId, publicBytes, rec.UserId)
	}
	return nil
}
