package repository

import (
	"math"
	"math/rand"
	"time"
)

type Chat struct {
	ID        int64
	User1ID   int64
	User2ID   int64
	UpdatedAt time.Time
}

type IChatRepository interface {
	GetAllByUser(userID int64) ([]*Chat, error)
	// IsExists(user1ID int64, user2ID int64) (bool, error)
	Get(user1ID int64, user2ID int64) (*Chat, error)
	GetByID(id int64) (*Chat, error)
	Start(user1ID int64, user2ID int64) (int64, error)
	UpdateDate(ID int64) error
}

type chatRepo struct {
	*repo
}

func NewChatRepository(dsn string) IChatRepository {
	repo := NewMysqlRepository(dsn)
	return &chatRepo{repo: repo}
}

func (r *chatRepo) GetAllByUser(userID int64) ([]*Chat, error) {
	rows, err := r.db.Query("select id,user1_id,user2_id,updated_at from chats where user1_id = ? or user2_id = ? order by updated_at desc", userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]*Chat, 0)
	for rows.Next() {
		chat := new(Chat)
		err := rows.Scan(&chat.ID, &chat.User1ID, &chat.User2ID, &chat.UpdatedAt)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

// func (r *chatRepo) IsExists(user1ID int64, user2ID int64) (bool, error) {
// 	u1ID := math.Min(float64(user1ID), float64(user2ID))
// 	u2ID := math.Max(float64(user1ID), float64(user2ID))
//
// 	var count int64
// 	row := r.db.QueryRow("select count(*) from chats where user1_id=? and user2_id=?", u1ID, u2ID)
// 	err := row.Scan(&count)
//
// 	if err != nil {
// 		return false, err
// 	}
// 	if count > 0 {
// 		return true, nil
// 	}
// 	return false, nil
// }

func (r *chatRepo) Get(user1ID int64, user2ID int64) (*Chat, error) {
	u1ID := math.Min(float64(user1ID), float64(user2ID))
	u2ID := math.Max(float64(user1ID), float64(user2ID))

	row := r.db.QueryRow("select id,user1_id,user2_id,updated_at from chats where user1_id = ? and user2_id = ?", u1ID, u2ID)

	var chat = new(Chat)
	err := row.Scan(&chat.ID, &chat.User1ID, &chat.User2ID, &chat.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (r *chatRepo) GetByID(id int64) (*Chat, error) {
	row := r.db.QueryRow("select id,user1_id,user2_id,updated_at from chats where id = ?", id)

	var chat = new(Chat)
	err := row.Scan(&chat.ID, &chat.User1ID, &chat.User2ID, &chat.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (r *chatRepo) Start(user1ID int64, user2ID int64) (int64, error) {
	u1ID := math.Min(float64(user1ID), float64(user2ID))
	u2ID := math.Max(float64(user1ID), float64(user2ID))

	chatID := rand.Int63n(math.MaxInt64)
	_, err := r.db.Exec("INSERT INTO chats(id, user1_id, user2_id, updated_at) VALUES(?, ?, ?, NOW())", chatID, u1ID, u2ID)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}

func (r *chatRepo) UpdateDate(ID int64) error {
	_, err := r.db.Exec("UPDATE chats set updated_at = NOW() where id = ?", ID)

	if err != nil {
		return err
	}

	return nil
}
