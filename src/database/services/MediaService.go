package services

import "visual-feed-aggregator/src/database/models"

type mediaService struct {
	mediaRepo models.MediaRepository
}

// NewMediaService creates a new user service with the necessary repository
func NewMediaService(mediaRepo models.MediaRepository) MediaService {
	return &mediaService{mediaRepo: mediaRepo}
}

func (s *mediaService) CreateMedia(media *models.Media) error {
	return s.mediaRepo.CreateMedia(media)
}
