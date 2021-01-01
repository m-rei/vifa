package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"visual-feed-aggregator/src/server"
	"visual-feed-aggregator/src/util/logging"
)

// ChannelMetaDataProvider given a channel id, this function retrieves all the necessary
// meta data for a specific channel
// returns:
// author, kind, profilePic, externalID
type ChannelMetaDataProvider func(channelID string) (string, string, string, string)

// ChannelMetaDataProviderFactory returns the appropriate ChannelMetaDataProvider for a given social media kind
type ChannelMetaDataProviderFactory func(kind string) ChannelMetaDataProvider

// ChannelDataValidator validates meta data for a channel
type ChannelDataValidator func(data string) bool

// ChannelDataValidatorFactory spawns the appropriate ChannelDataValidator for a given social media kind
type ChannelDataValidatorFactory func(kind string) ChannelDataValidator

// ValidateChannel checks if the channel id is a valid one
// {
//		"ChannelID": "<ChannelID>",
//		"kind": "<kind>",
// }
func ValidateChannel(s *server.Server, f ChannelDataValidatorFactory) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		channelID := r.URL.Query().Get("channelID")
		kind := r.URL.Query().Get("kind")
		v := f(kind)

		if v != nil && v(channelID) {
			rw.WriteHeader(http.StatusAccepted)
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
	}
}

// AddChannel consumes
// {
//		"ChannelID": "<ChannelID>",
//		"AccountID": "<AccountID>",
//		"kind": "<kind>",
// }
// to create a new association
func AddChannel(s *server.Server, f ChannelMetaDataProviderFactory) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var channelData struct {
			ChannelID string
			AccountID string
			Kind      string
		}
		err := json.NewDecoder(r.Body).Decode(&channelData)
		if err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		accountID, _ := strconv.ParseInt(channelData.AccountID, 10, 64)

		if err := DoAddChannel(s, accountID, channelData.ChannelID, f(channelData.Kind)); err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}

// DoAddChannel ...
func DoAddChannel(s *server.Server, accountID int64, channelID string, f ChannelMetaDataProvider) error {
	author, kind, profilePic, externalID := f(channelID)
	if author == "" {
		return errors.New("author not found")
	}

	c, _, err := s.Services.ChannelService.CreateChannelIfNotExists(author, kind, profilePic, externalID)
	if err != nil {
		return err
	}

	hasToAddChannel := !s.Services.AccountService.HasChannel(accountID, c.ID)

	if !hasToAddChannel {
		return nil
	}

	acc, err := s.Services.AccountService.GetAccount(accountID)
	if err != nil {
		return err
	}
	err = s.Services.AccountService.AddChannel(&acc, c)
	if err != nil {
		return err
	}

	return nil
}

// DeleteChannel ...
func DeleteChannel(s *server.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var channelData struct {
			ChannelID string
			AccountID string
			Kind      string
		}
		err := json.NewDecoder(r.Body).Decode(&channelData)
		if err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		accountID, err := strconv.ParseInt(channelData.AccountID, 10, 64)
		if err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		channelID, err := strconv.ParseInt(channelData.ChannelID, 10, 64)
		if err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = s.Services.AccountService.RemoveAccountChannel(accountID, channelID)
		if err != nil {
			logging.Println(logging.Error, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}
