package models

import "time"

// UserRepository defines the basic CRUD functionality for any concrete implementations (mysql, mongodb, etc.)
type UserRepository interface {
	CreateUser(user *User) error
	GetUser(email string) (User, error)
	UpdateUser(user User) error
	RemoveUser(user User) error
	LoadUserAccounts(user *User) error
	LoadUserAccountsForKind(user *User, kind string) error
}

// AccountRepository defines the basic CRUD functionality for any concrete implementations (mysql, mongodb, etc.)
type AccountRepository interface {
	CreateAccount(account *Account) error
	GetAccount(id int64) (Account, error)
	UpdateAccount(account Account) error
	RemoveAccount(account Account) error

	AddChannel(account *Account, channel Channel) error
	RemoveChannel(accountID, channelID int64) error
	HasChannel(accountID, channelID int64) bool
	LoadUser(account *Account) error
	LoadPeople(account *Account) error
}

// ChannelRepository ...
type ChannelRepository interface {
	CreateChannel(channel *Channel) error
	GetChannel(id int64) (Channel, error)
	FindChannelByExternalID(externalID string) (Channel, error)
	FindChannelsByKind(kind string) ([]Channel, error)
	FindChannelsByAccountIDAndKind(accountID int64, kind string) ([]Channel, error)
	UpdateChannel(channel Channel) error
	RemoveChannel(channel Channel) error
	LoadFollowers(channel *Channel) error
	LoadContent(channel *Channel) error
	CleanupOrphanedChannels() (int64, error)
}

// ContentRepository ...
type ContentRepository interface {
	CreateContent(content *Content) error
	GetContent(id int64) (Content, error)
	UpdateContent(content Content) error
	RemoveContent(content Content) error
	LoadChannel(content *Content) error
	LoadMedia(content *Content) error
	CleanupOldContent(time *time.Time) (int64, error)
	LoadContentFor(userID int64, kind string, accID int64, offset, count int64) ([]Content, error)
	CountAllContentFor(userID int64, kind string, accID int64) (int64, error)
}

// MediaRepository ...
type MediaRepository interface {
	CreateMedia(media *Media) error
	GetMedia(id int64) (Media, error)
	UpdateMedia(media Media) error
	RemoveMedia(media Media) error
}
