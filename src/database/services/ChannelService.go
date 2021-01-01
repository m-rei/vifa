package services

import (
	"database/sql"
	"strings"
	"visual-feed-aggregator/src/database/models"
)

type channelService struct {
	channelRepo models.ChannelRepository
}

// NewChannelService creates a new channel service with the necessary repository
func NewChannelService(channelRepo models.ChannelRepository) ChannelService {
	return &channelService{channelRepo: channelRepo}
}

func (s *channelService) CreateChannelIfNotExists(name, kind, profilePic, externalID string) (models.Channel, bool, error) {
	c, err := s.channelRepo.FindChannelByExternalID(externalID)
	if err == nil {
		return c, false, nil
	}
	c = models.Channel{Name: name, Kind: kind, ProfilePic: sql.NullString{String: profilePic, Valid: true}, ExternalID: externalID}
	err = s.channelRepo.CreateChannel(&c)
	return c, err == nil, err
}

func (s *channelService) FindChannelsByKind(kind string) ([]models.Channel, error) {
	return s.channelRepo.FindChannelsByKind(kind)
}

func (s *channelService) FindChannelsByAccountIDAndKind(accountID int64, kind string) ([]models.Channel, error) {
	ret, err := s.channelRepo.FindChannelsByAccountIDAndKind(accountID, kind)
	if err == nil {
		for idx := range ret {
			ret[idx].ExternalID = ChannelURL(ret[idx].ExternalID, kind)
		}
	}
	return ret, err
}

// ChannelURL returns the (full) URL for a given externalID (which is just the shortest possible ID to reconstruct the full URL)
func ChannelURL(externalID, kind string) string {
	switch kind {
	case models.KindYoutube:
		split := strings.Split(externalID, "=")
		cu := strings.Replace(split[0], "_id", "", -1)
		return "https://youtube.com/" + cu + "/" + split[1]
	case models.KindReddit:
		return "https://reddit.com/r/" + externalID + "/new"
	case models.KindTwitter:
		return "https://twitter.com/" + externalID
	case models.KindInstagram:
		return "https://instagram.com/" + externalID
	}
	return externalID
}

func (s *channelService) LoadContent(channel *models.Channel) error {
	return s.channelRepo.LoadContent(channel)
}

func (s *channelService) CleanupOrphanedChannels() (int64, error) {
	return s.channelRepo.CleanupOrphanedChannels()
}
