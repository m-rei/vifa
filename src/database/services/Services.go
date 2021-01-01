package services

import (
	"time"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/database/repos"

	"github.com/jmoiron/sqlx"
)

// UserService defines all the necessary business logic
type UserService interface {
	GetUser(email string) (models.User, error)
	LoadUserAccounts(user *models.User) error
	LoadUserAccountsForSocialMedia(user *models.User, kind string) error
	CreateUserIfNotExists(email, pictureURL string) (models.User, bool, error)
}

// AccountService defines all the necessary business logic
type AccountService interface {
	AddAccount(userID int64, name, kind string) (models.Account, error)
	RemoveAccount(accountID int64) error
	GetAccount(id int64) (models.Account, error)
	HasChannel(accountID, channelID int64) bool
	AddChannel(account *models.Account, channel models.Channel) error
	RemoveAccountChannel(accountID, channelID int64) error
}

// ChannelService ...
type ChannelService interface {
	CreateChannelIfNotExists(name, kind, profilePic, externalID string) (models.Channel, bool, error)
	FindChannelsByKind(kind string) ([]models.Channel, error)
	FindChannelsByAccountIDAndKind(accountID int64, kind string) ([]models.Channel, error)
	LoadContent(channel *models.Channel) error
	CleanupOrphanedChannels() (int64, error)
}

// ContentService ...
type ContentService interface {
	CreateContent(content *models.Content) error
	LoadMedia(content *models.Content) error
	CleanupOldContent(time *time.Time) (int64, error)
	LoadContentFor(userID int64, kind string, accID int64, offset, count int64) ([]models.Content, error)
	CountAllContentFor(userID int64, kind string, accID int64) (int64, error)
}

// MediaService ...
type MediaService interface {
	CreateMedia(media *models.Media) error
}

// ServiceCollection ...
type ServiceCollection struct {
	UserService    UserService
	AccountService AccountService
	ChannelService ChannelService
	ContentService ContentService
	MediaService   MediaService
}

// NewMySQLServiceCollection ...
func NewMySQLServiceCollection(db *sqlx.DB) ServiceCollection {
	return ServiceCollection{
		UserService:    NewUserService(repos.NewMySQLUserRepository(db)),
		AccountService: NewAccountService(repos.NewMySQLAccountRepository(db)),
		ChannelService: NewChannelService(repos.NewMySQLChannelRepository(db)),
		ContentService: NewContentService(repos.NewMySQLContentRepository(db)),
		MediaService:   NewMediaService(repos.NewMySQLMediaRepository(db)),
	}
}
