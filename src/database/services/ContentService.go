package services

import (
	"time"
	"visual-feed-aggregator/src/database/models"
)

type contentService struct {
	contentRepo models.ContentRepository
}

// NewContentService creates a new content service with the necessary repository
func NewContentService(contentRepo models.ContentRepository) ContentService {
	return &contentService{contentRepo: contentRepo}
}

func (s *contentService) CreateContent(content *models.Content) error {
	return s.contentRepo.CreateContent(content)
}

func (s *contentService) LoadMedia(content *models.Content) error {
	return s.contentRepo.LoadMedia(content)
}
func (s *contentService) CleanupOldContent(time *time.Time) (int64, error) {
	return s.contentRepo.CleanupOldContent(time)
}

func (s *contentService) LoadContentFor(userID int64, kind string, accID int64, offset, count int64) ([]models.Content, error) {
	return s.contentRepo.LoadContentFor(userID, kind, accID, offset, count)
}

func (s *contentService) CountAllContentFor(userID int64, kind string, accID int64) (int64, error) {
	return s.contentRepo.CountAllContentFor(userID, kind, accID)
}
