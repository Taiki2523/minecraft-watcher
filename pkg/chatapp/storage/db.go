package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	DB *sql.DB
}

type User struct {
	ID        int64  `json:"id"`
	GoogleID  string `json:"google_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type Channel struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type Message struct {
	ID        int64     `json:"id"`
	ChannelID int64     `json:"channel_id"`
	UserID    int64     `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UserName  string    `json:"user_name"`
	AvatarURL string    `json:"avatar_url"`
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) UpsertUser(user User) (User, error) {
	result, err := s.DB.Exec(`
		INSERT INTO users (google_id, email, name, avatar_url)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			email = VALUES(email),
			name = VALUES(name),
			avatar_url = VALUES(avatar_url)
	`, user.GoogleID, user.Email, user.Name, user.AvatarURL)
	if err != nil {
		return User{}, fmt.Errorf("upsert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil && id != 0 {
		user.ID = id
		return user, nil
	}

	row := s.DB.QueryRow(`SELECT id, google_id, email, name, avatar_url FROM users WHERE google_id = ?`, user.GoogleID)
	if err := row.Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.AvatarURL); err != nil {
		return User{}, fmt.Errorf("select user: %w", err)
	}

	return user, nil
}

func (s *Store) GetUser(id int64) (User, error) {
	var user User
	row := s.DB.QueryRow(`SELECT id, google_id, email, name, avatar_url FROM users WHERE id = ?`, id)
	if err := row.Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.AvatarURL); err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *Store) ListChannels() ([]Channel, error) {
	rows, err := s.DB.Query(`SELECT id, name, display_name FROM channels ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var channel Channel
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.DisplayName); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (s *Store) ListMessages(channelID int64, limit int) ([]Message, error) {
	rows, err := s.DB.Query(`
		SELECT m.id, m.channel_id, m.user_id, m.body, m.created_at, u.name, u.avatar_url
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.channel_id = ?
		ORDER BY m.created_at DESC
		LIMIT ?
	`, channelID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Body, &msg.CreatedAt, &msg.UserName, &msg.AvatarURL); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *Store) CreateMessage(channelID, userID int64, body string) (Message, error) {
	result, err := s.DB.Exec(`INSERT INTO messages (channel_id, user_id, body) VALUES (?, ?, ?)`, channelID, userID, body)
	if err != nil {
		return Message{}, fmt.Errorf("insert message: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("message id: %w", err)
	}

	var msg Message
	row := s.DB.QueryRow(`
		SELECT m.id, m.channel_id, m.user_id, m.body, m.created_at, u.name, u.avatar_url
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = ?
	`, id)
	if err := row.Scan(&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Body, &msg.CreatedAt, &msg.UserName, &msg.AvatarURL); err != nil {
		return Message{}, fmt.Errorf("select message: %w", err)
	}

	return msg, nil
}
