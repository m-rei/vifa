package models

import (
	"database/sql"
	"time"
)

// Account represents a social media account, like a youtube account
type Account struct {
	ID     int64
	Name   string
	Kind   string
	UserID int64 `db:"user_id"`

	User     *User
	Channels []Channel
}

// User represents a VIFA user, who has different social media accounts
type User struct {
	ID         int64
	Email      string
	PictureURL string `db:"picture_url"`

	Accounts []Account
}

// Channel represents a social media Channel, which is being followed by an "account"
type Channel struct {
	ID         int64
	Name       string
	Kind       string
	ProfilePic sql.NullString `db:"profile_pic"`
	ExternalID string         `db:"external_id"`

	Followers []Account
	Contents  []Content
}

// Content represents a single social media publication, like a reddit post, a youtube video or a tweet
type Content struct {
	ID         int64
	Title      string
	Date       time.Time
	ExternalID string `db:"external_id"` // 255 chars
	ChannelID  int64  `db:"channel_id"`

	Channel  *Channel
	AllMedia []Media
}

// Media represents all the linked media to one publication (content)
type Media struct {
	ID        int64
	URL       string
	ContentID int64 `db:"content_id"`

	Content *Content
}
