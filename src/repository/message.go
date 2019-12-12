package repository

import (
	"time"
)

type Message struct {
	ID        int64
	ChatID    int64
	UserID    int64
	Text      string
	CreatedAt time.Time
}

type IMessageRepository interface {
	GetByChat(chatID int64, minID int64, limit int) ([]*Message, error)
	Create(chatID int64, userID int64, text string) error
}

type messageRepo struct {
	*repo
}

func NewMessageRepository(dsn string) IMessageRepository {
	repo := NewMysqlRepository(dsn)
	return &messageRepo{repo: repo}
}

func (r *messageRepo) GetByChat(chatID int64, minID int64, limit int) ([]*Message, error) {
	sql := "SELECT id, chat_id, user_id, text, created_at FROM messages where chat_id=? and id>? order by id asc limit ?"
	rows, err := r.db.Query(sql, chatID, minID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*Message, 0)
	for rows.Next() {
		msg := new(Message)
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.UserID, &msg.Text, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepo) Create(chatID int64, userID int64, text string) error {
	_, err := r.db.Exec("INSERT INTO messages(chat_id, user_id, text, created_at) VALUES(?, ?, ?, NOW())", chatID, userID, text)
	if err != nil {
		return err
	}

	return nil
}
