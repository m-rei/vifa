package repos

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"visual-feed-aggregator/src/database/models"
	"visual-feed-aggregator/src/util/logging"

	"github.com/jmoiron/sqlx"
)

type mySQLUserRepository struct {
	db *sqlx.DB
}

type mySQLAccountRepository struct {
	db *sqlx.DB
}

type mySQLChannelRepository struct {
	db *sqlx.DB
}

type mySQLContentRepository struct {
	db *sqlx.DB
}

type mySQLMediaRepository struct {
	db *sqlx.DB
}

// NewMySQLUserRepository ...
func NewMySQLUserRepository(db *sqlx.DB) models.UserRepository {
	return &mySQLUserRepository{db: db}
}

// CreateUser creates a new user
func (r *mySQLUserRepository) CreateUser(user *models.User) error {
	query := `
	INSERT INTO user (email, picture_url) 
	VALUES (:email, :picture_url)	
	`
	res, err := r.db.NamedExec(query, &user)
	if err == nil {
		user.ID, err = res.LastInsertId()
	}
	return err
}

// GetUser loads an user by email
func (r *mySQLUserRepository) GetUser(email string) (models.User, error) {
	query := `
	SELECT *
	FROM user
	WHERE email = ?
	`
	user := models.User{}
	err := r.db.Get(&user, query, email)
	return user, err
}

// UpdateUser updates the user
func (r *mySQLUserRepository) UpdateUser(user models.User) error {
	query := `
	UPDATE user
	SET email = :email, picture_url = :picture_url
	WHERE id=:id
	`
	_, err := r.db.NamedExec(query, user)
	return err
}

// RemoveUser removes an user
func (r *mySQLUserRepository) RemoveUser(user models.User) error {
	query := `
	DELETE FROM user
	WHERE id = :id
	`
	_, err := r.db.NamedExec(query, &user)
	return err
}

// LoadUserAccounts loads all the related accounts of user
func (r *mySQLUserRepository) LoadUserAccounts(user *models.User) error {
	query := `
	SELECT *
	FROM account
	WHERE user_id = ?
	`
	accounts := []models.Account{}
	err := r.db.Select(&accounts, query, user.ID)
	if err != nil {
		return err
	}
	user.Accounts = accounts
	return nil
}

// LoadUserAccounts loads all the related accounts of user
func (r *mySQLUserRepository) LoadUserAccountsForKind(user *models.User, kind string) error {
	query := `
	SELECT *
	FROM account
	WHERE user_id = ? AND kind = ?
	`
	accounts := []models.Account{}
	err := r.db.Select(&accounts, query, user.ID, kind)
	if err != nil {
		return err
	}
	user.Accounts = accounts
	return nil
}

// NewMySQLAccountRepository ...
func NewMySQLAccountRepository(db *sqlx.DB) models.AccountRepository {
	return &mySQLAccountRepository{db: db}
}

// CreateAccount creates a new account
func (r *mySQLAccountRepository) CreateAccount(account *models.Account) error {
	query := `
	INSERT INTO account (name, kind, user_id)
	VALUES (:name, :kind, :user_id)
	`
	res, err := r.db.NamedExec(query, &account)
	if err == nil {
		account.ID, err = res.LastInsertId()
	}
	return err
}

// GetAccount ...
func (r *mySQLAccountRepository) GetAccount(id int64) (models.Account, error) {
	query := `
	SELECT *
	FROM account
	WHERE id = ?
	`
	account := models.Account{}
	err := r.db.Get(&account, query, id)
	return account, err
}

// UpdateAccount updates the account
func (r *mySQLAccountRepository) UpdateAccount(account models.Account) error {
	query := `
	UPDATE account
	SET name = :name, kind = :kind, user_id = :user_id
	WHERE id = :id
	`
	_, err := r.db.NamedExec(query, &account)
	return err
}

// RemoveAccount removes an account
func (r *mySQLAccountRepository) RemoveAccount(account models.Account) error {
	query := `
	DELETE FROM account
	WHERE id = :id
	`
	_, err := r.db.NamedExec(query, &account)
	return err
}

// AddPerson ...
func (r *mySQLAccountRepository) AddChannel(account *models.Account, channel models.Channel) error {
	if channel.ID == 0 {
		return errors.New("channel empty")
	}
	query := `
	INSERT INTO account_channel (account_id, channel_id)
	VALUES (?, ?)
	`
	_, err := r.db.Exec(query, account.ID, channel.ID)
	if err == nil {
		account.Channels = append(account.Channels, channel)
	}
	return err
}

// RemoveAccountChannel ...
func (r *mySQLAccountRepository) RemoveChannel(accountID, channelID int64) error {
	query := `
	DELETE FROM account_channel
	WHERE account_id = ? AND channel_id = ?
	`
	_, err := r.db.Exec(query, accountID, channelID)
	return err
}

// HasChannel ...
func (r *mySQLAccountRepository) HasChannel(accountID, channelID int64) bool {
	query := `
	SELECT channel_id
	FROM account_channel
	WHERE account_id = ? AND channel_id = ?
	`
	row := r.db.QueryRow(query, accountID, channelID)
	var cID int64
	err := row.Scan(&cID)
	return err != sql.ErrNoRows
}

// LoadUser ...
func (r *mySQLAccountRepository) LoadUser(account *models.Account) error {
	query := `
	SELECT *
	FROM user
	WHERE id = ?
	`
	user := &models.User{}
	err := r.db.Get(user, query, account.UserID)
	account.User = user
	return err
}

// LoadFollowing ...
func (r *mySQLAccountRepository) LoadPeople(account *models.Account) error {
	query := `
	SELECT channel.*
	FROM channel
	INNER JOIN account_channel
	ON account_channel.channel_id = id AND account_channel.account_id = ?
	`
	channels := []models.Channel{}
	err := r.db.Select(&channels, query, account.ID)
	if err != nil {
		return err
	}
	account.Channels = channels
	return nil
}

// NewMySQLChannelRepository ...
func NewMySQLChannelRepository(db *sqlx.DB) models.ChannelRepository {
	return &mySQLChannelRepository{db: db}
}

func (r *mySQLChannelRepository) CreateChannel(Channel *models.Channel) error {
	query := `
	INSERT INTO channel (name, kind, profile_pic, external_id) 
	VALUES (:name, :kind, :profile_pic, :external_id)
	`
	res, err := r.db.NamedExec(query, &Channel)
	if err == nil {
		Channel.ID, err = res.LastInsertId()
	}
	return err
}
func (r *mySQLChannelRepository) GetChannel(id int64) (models.Channel, error) {
	query := `
	SELECT *
	FROM channel
	WHERE id = ?
	`
	channel := models.Channel{}
	err := r.db.Get(&channel, query, id)
	return channel, err
}

func (r *mySQLChannelRepository) FindChannelByExternalID(externalID string) (models.Channel, error) {
	query := `
	SELECT *
	FROM channel
	WHERE external_id = ?
	`
	channel := models.Channel{}
	err := r.db.Get(&channel, query, externalID)
	return channel, err
}

func (r *mySQLChannelRepository) FindChannelsByKind(kind string) ([]models.Channel, error) {
	query := `
	SELECT *
	FROM channel
	WHERE kind = ?
	`
	channels := []models.Channel{}
	err := r.db.Select(&channels, query, kind)
	return channels, err
}

func (r *mySQLChannelRepository) FindChannelsByAccountIDAndKind(accountID int64, kind string) ([]models.Channel, error) {
	query := `
	SELECT channel.*
	FROM channel
	INNER JOIN account_channel
	ON account_channel.channel_id = channel.id AND account_channel.account_id = ?
	WHERE kind = ?
	`
	channels := []models.Channel{}
	err := r.db.Select(&channels, query, accountID, kind)
	return channels, err
}

func (r *mySQLChannelRepository) UpdateChannel(Channel models.Channel) error {
	query := `
	UPDATE channel
	SET name = :name, external_id = :external_id
	WHERE id = :id
	`
	_, err := r.db.NamedExec(query, Channel)
	return err
}

func (r *mySQLChannelRepository) RemoveChannel(Channel models.Channel) error {
	query := `
	DELETE FROM channel
	WHERE id = :id
	`
	_, err := r.db.NamedExec(query, &Channel)
	return err
}

func (r *mySQLChannelRepository) LoadFollowers(Channel *models.Channel) error {
	query := `
	SELECT account.*
	FROM account
	INNER JOIN account_channel
	ON account_channel.channel_id = ? AND account_channel.account_id = id
	`
	accounts := []models.Account{}
	err := r.db.Select(&accounts, query, Channel.ID)
	if err != nil {
		return err
	}
	Channel.Followers = accounts
	return nil
}

func (r *mySQLChannelRepository) LoadContent(channel *models.Channel) error {
	query := `
	SELECT *
	FROM content
	WHERE Channel_id = ?
	`
	contents := []models.Content{}
	err := r.db.Select(&contents, query, channel.ID)
	if err != nil {
		return err
	}
	for idx := range contents {
		contents[idx].Channel = channel
	}
	channel.Contents = contents
	return nil
}

func (r *mySQLChannelRepository) CleanupOrphanedChannels() (int64, error) {
	query := `
	DELETE channel
	FROM channel
	LEFT JOIN account_channel
	ON channel.id = account_channel.channel_id
	WHERE account_channel.account_id IS NULL
	`
	res, err := r.db.Exec(query)
	if err == nil {
		cnt, err := res.RowsAffected()
		if err != nil {
			return 0, nil // nil because RowsAffected is an optional feature db dependent, not indicative of an error
		}
		return cnt, nil
	}
	return 0, err
}

// NewMySQLContentRepository ...
func NewMySQLContentRepository(db *sqlx.DB) models.ContentRepository {
	return &mySQLContentRepository{db: db}
}

func (r *mySQLContentRepository) CreateContent(content *models.Content) error {
	query := `
	INSERT INTO content (title, date, external_id, channel_id) 
	VALUES (:title, :date, :external_id, :channel_id)
	`
	res, err := r.db.NamedExec(query, &content)
	if err == nil {
		content.ID, err = res.LastInsertId()
	}
	return err
}

func (r *mySQLContentRepository) GetContent(id int64) (models.Content, error) {
	query := `
	SELECT *
	FROM content
	WHERE id = ?
	`
	content := models.Content{}
	err := r.db.Get(&content, query, id)
	return content, err

}

func (r *mySQLContentRepository) UpdateContent(content models.Content) error {
	query := `
	UPDATE content
	SET title = :title, date = :date, external_id = :external_id, channel_id = :channel_id
	WHERE id=:id
	`
	_, err := r.db.NamedExec(query, content)
	return err

}

func (r *mySQLContentRepository) RemoveContent(content models.Content) error {
	query := `
	DELETE FROM content
	WHERE id = :id
	`
	_, err := r.db.NamedExec(query, &content)
	return err
}

func (r *mySQLContentRepository) LoadChannel(content *models.Content) error {
	query := `
	SELECT *
	FROM channel
	WHERE id = ?
	`
	channel := models.Channel{}
	err := r.db.Get(&channel, query, content.ChannelID)
	return err
}

func (r *mySQLContentRepository) LoadMedia(content *models.Content) error {
	query := `
	SELECT *
	FROM media
	WHERE content_id = ?
	`
	allMedia := []models.Media{}
	err := r.db.Select(&allMedia, query, content.ID)
	if err != nil {
		return err
	}
	for idx := range allMedia {
		allMedia[idx].Content = content
	}
	content.AllMedia = allMedia
	return nil
}

func (r *mySQLContentRepository) CleanupOldContent(time *time.Time) (int64, error) {
	query := `
	DELETE FROM content
	WHERE date < ?
	`
	res, err := r.db.Exec(query, *time)
	if err == nil {
		cnt, err := res.RowsAffected()
		if err != nil {
			return 0, nil // nil because RowsAffected is an optional feature db dependent, not indicative of an error
		}
		return cnt, nil
	}
	return 0, err
}

func (r *mySQLContentRepository) LoadContentFor(userID int64, kind string, accID int64, offset, count int64) ([]models.Content, error) {
	query := `
	SELECT
		ch.id, ch.name, ch.kind, ch.profile_pic, ch.external_id,
		c.id, c.title, c.date, c.external_id, c.channel_id,
		m.id, m.url, m.content_id
	FROM (
		SELECT c2.* 
		FROM content c2
		INNER JOIN channel ch2 ON c2.channel_id = ch2.id
		INNER JOIN account_channel ac2 ON ac2.channel_id = ch2.id
		INNER JOIN account a2 ON a2.id =  ac2.account_id    
		%s
		ORDER BY c2.date DESC
		%s
	) as c
		LEFT JOIN channel ch ON ch.id = c.channel_id
		LEFT JOIN account_channel ac ON ch.id = ac.account_id    
		LEFT JOIN account a ON a.id= ac.account_id
		LEFT JOIN media m ON m.content_id = c.id
	;`
	args := []interface{}{}
	sqlWhere := ""
	if accID > 0 {
		sqlWhere = "WHERE a2.user_id = ? AND a2.kind = ? AND a2.id = ?"
		args = append(args, userID, kind, accID)
	} else {
		sqlWhere = "WHERE a2.user_id = ? AND a2.kind = ?"
		args = append(args, userID, kind)
	}
	sqlLimit := ""
	if offset >= 0 && count > 0 {
		sqlLimit = "LIMIT ?, ?"
		args = append(args, offset, count)
	}
	query = fmt.Sprintf(query, sqlWhere, sqlLimit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		logging.Println(logging.Error, err)
		return nil, err
	}

	var contents []models.Content
	var contentMap = make(map[int64]*models.Content)
	for rows.Next() {
		var ch models.Channel
		var c models.Content
		var m models.Media
		err := rows.Scan(&ch.ID, &ch.Name, &ch.Kind, &ch.ProfilePic, &ch.ExternalID,
			&c.ID, &c.Title, &c.Date, &c.ExternalID, &c.ChannelID,
			&m.ID, &m.URL, &m.ContentID)
		if err != nil {
			logging.Println(logging.Debug, err)
			// media can be null and it will throw conversion error -- some content may not have any associated media!
		}

		content, haveIt := contentMap[c.ID]
		if haveIt {
			if m.ID > 0 {
				content.AllMedia = append(content.AllMedia, m)
			}
		} else {
			c.Channel = &ch
			if m.ID > 0 {
				c.AllMedia = append(c.AllMedia, m)
			}
			contents = append(contents, c)
			contentMap[c.ID] = &contents[len(contents)-1]
		}
	}

	return contents, nil
}

func (r *mySQLContentRepository) CountAllContentFor(userID int64, kind string, accID int64) (int64, error) {
	query := `
	SELECT count(*) as count
	FROM account a
		INNER JOIN account_channel ac
		ON id = ac.account_id
		
		INNER JOIN channel ch
		ON channel_id = ch.id
		
		INNER JOIN content c
		ON c.channel_id = ch.id
	%s
	ORDER BY c.date DESC
	;`
	var row *sql.Row
	if accID > 0 {
		query = fmt.Sprintf(query, "WHERE a.user_id = ? AND a.kind = ? AND a.id = ?")
		row = r.db.QueryRow(query, 1, kind, accID)
	} else {
		query = fmt.Sprintf(query, "WHERE a.user_id = ? AND a.kind = ?")
		row = r.db.QueryRow(query, 1, kind)
	}
	var count int64
	err := row.Scan(&count)
	if err != nil {
		logging.Println(logging.Error, err)
		return -1, err
	}
	return count, nil
}

// NewMySQLMediaRepository ...
func NewMySQLMediaRepository(db *sqlx.DB) models.MediaRepository {
	return &mySQLMediaRepository{db: db}
}

func (r *mySQLMediaRepository) CreateMedia(media *models.Media) error {
	createMediaQuery := `
	INSERT INTO media (url, content_id) 
	VALUES (:url, :content_id)
	`
	res, err := r.db.NamedExec(createMediaQuery, &media)
	if err == nil {
		media.ID, err = res.LastInsertId()
	}
	return err
}

func (r *mySQLMediaRepository) GetMedia(id int64) (models.Media, error) {
	getMediaQuery := `
	SELECT *
	FROM media
	WHERE id = ?
	`
	media := models.Media{}
	err := r.db.Get(&media, getMediaQuery, id)
	return media, err
}

func (r *mySQLMediaRepository) UpdateMedia(media models.Media) error {
	updateMediaQuery := `
	UPDATE media
	SET url = :url, content_id = :content_id
	WHERE id = :id
	`
	_, err := r.db.NamedExec(updateMediaQuery, media)
	return err
}

func (r *mySQLMediaRepository) RemoveMedia(media models.Media) error {
	removeMediaQuery := `
	DELETE FROM media
	WHERE id = :id
	`
	_, err := r.db.NamedExec(removeMediaQuery, &media)
	return err
}
