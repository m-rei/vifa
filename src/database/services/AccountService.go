package services

import "visual-feed-aggregator/src/database/models"

type accountService struct {
	accountRepo models.AccountRepository
}

// NewAccountService creates a new user service with the necessary repository
func NewAccountService(accountRepo models.AccountRepository) AccountService {
	return &accountService{accountRepo: accountRepo}
}

func (s *accountService) AddAccount(userID int64, name, kind string) (models.Account, error) {
	acc := models.Account{Name: name, Kind: kind, UserID: userID}
	err := s.accountRepo.CreateAccount(&acc)
	return acc, err
}

func (s *accountService) RemoveAccount(accountID int64) error {
	return s.accountRepo.RemoveAccount(models.Account{ID: accountID})
}

func (s *accountService) GetAccount(id int64) (models.Account, error) {
	return s.accountRepo.GetAccount(id)
}

func (s *accountService) HasChannel(accountID, channelID int64) bool {
	return s.accountRepo.HasChannel(accountID, channelID)
}

func (s *accountService) AddChannel(account *models.Account, channel models.Channel) error {
	return s.accountRepo.AddChannel(account, channel)
}

func (s *accountService) RemoveAccountChannel(accountID, channelID int64) error {
	return s.accountRepo.RemoveChannel(accountID, channelID)
}
