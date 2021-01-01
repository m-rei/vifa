package services

import (
	"visual-feed-aggregator/src/database/models"
)

type userService struct {
	userRepo models.UserRepository
}

// NewUserService creates a new user service with the necessary repository
func NewUserService(userRepo models.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUser(email string) (models.User, error) {
	return s.userRepo.GetUser(email)
}

func (s *userService) LoadUserAccounts(user *models.User) error {
	return s.userRepo.LoadUserAccounts(user)
}

func (s *userService) LoadUserAccountsForSocialMedia(user *models.User, kind string) error {
	return s.userRepo.LoadUserAccountsForKind(user, kind)
}

func (s *userService) CreateUserIfNotExists(email, pictureURL string) (models.User, bool, error) {
	user, err := s.userRepo.GetUser(email)
	if err == nil {
		return user, false, nil
	}
	user = models.User{Email: email, PictureURL: pictureURL}
	err = s.userRepo.CreateUser(&user)
	return user, err == nil, err
}
