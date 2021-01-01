-- used for session management
CREATE TABLE IF NOT EXISTS sessions (
	sid VARCHAR(36) NOT NULL,
	k VARCHAR(255) NOT NULL,
	v TEXT NOT NULL,

	UNIQUE(sid, k)
);

INSERT IGNORE INTO sessions (sid, k, v)
VALUES ("session", "key", TO_BASE64(RANDOM_BYTES(32)))
;

-- represents VIFA user
CREATE TABLE IF NOT EXISTS user (
	id INT AUTO_INCREMENT PRIMARY KEY,
	email VARCHAR(255) NOT NULL UNIQUE,
	picture_url TEXT
);

-- represents VIFA user's  social media account
CREATE TABLE IF NOT EXISTS account (
	id INT AUTO_INCREMENT PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	kind VARCHAR(50) NOT NULL, -- "youtube", "instagram", "reddit", "twitter", ...
	user_id INT NOT NULL,

	UNIQUE(name, kind),
	FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

-- represents something like a youtube channel or other social media account, like a reddit account
CREATE TABLE IF NOT EXISTS channel (
	id INT AUTO_INCREMENT PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	kind VARCHAR(50) NOT NULL, -- "youtube", "instagram", "reddit", "twitter", ...
	profile_pic TEXT,
	external_id VARCHAR(255) NOT NULL
);

-- one youtube account can have many channels and one channel can have relations to many accounts
create table IF NOT EXISTS account_channel (
    account_id int not null,
    channel_id int not null,
	FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
	FOREIGN KEY (channel_id) REFERENCES channel(id) ON DELETE CASCADE
);

-- represents a single social media publication, like a post, a video (youtube) or a tweet etc.
CREATE TABLE IF NOT EXISTS content (
	id INT AUTO_INCREMENT PRIMARY KEY,
	title TEXT NOT NULL,
	date DATETIME NOT NULL,
	external_id VARCHAR(256) NOT NULL, -- main url, to the whole content
	channel_id INT NOT NULL,

	UNIQUE(external_id, channel_id),
	FOREIGN KEY (channel_id) REFERENCES channel(id) ON DELETE CASCADE
);

-- each publication can have many media files associated, e.g. one tweet could contain 3-4 images
CREATE TABLE IF NOT EXISTS media (
	id INT AUTO_INCREMENT PRIMARY KEY,
	url TEXT NOT NULL, -- thumbnails, etc.
	content_id INT NOT NULL,
	FOREIGN KEY (content_id) REFERENCES content(id) ON DELETE CASCADE
);