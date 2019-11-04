package repository

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"strings"
	"sync/atomic"
)

type User struct {
	ID           int64
	Login        string
	Name         string
	LastName     string
	Password     string
	PasswordHash string
	Description  string
	PhotoFile    string
	CreatedAt    sql.NullTime
}

type IUserRepository interface {
	GetMasterDB() *sql.DB
	GetRoDB() *sql.DB
	GetAll() ([]*User, error)
	Get(id int64) (*User, error)
	Create(*User) error
	Update(user *User) error
	IsLoginExist(login string) bool
	FindByLoginAndPassword(login string, password string) (*User, error)
	FindByNamePrefix(prefix string, limit int, minId int64) ([]*User, error)
	BulkCreate(users []*User)
}

func (r *repo) GetMasterDB() *sql.DB {
	return r.db
}

func (r *repo) GetRoDB() *sql.DB {
	idx := rand.Intn(len(r.readReplicas))
	if idx == 0 {
		atomic.AddInt32(&r.masterCnt, 1)
	} else {
		atomic.AddInt32(&r.slaveCnt, 1)
	}
	// fmt.Printf("GetRoDB: idx=%d, masterCnt=%d, slaveCnt=%d\n", idx, r.masterCnt, r.slaveCnt)
	return r.readReplicas[idx]
}

func (r *repo) GetAll() ([]*User, error) {
	rows, err := r.db.Query("SELECT id, login, name, last_name, description, photo_file, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		user := new(User)
		err := rows.Scan(&user.ID, &user.Login, &user.Name, &user.LastName, &user.Description, &user.PhotoFile, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *repo) Get(id int64) (*User, error) {
	row := r.db.QueryRow("SELECT id, login, name, last_name, description, photo_file, created_at FROM users WHERE id = ?", id)

	user := new(User)
	err := row.Scan(&user.ID, &user.Login, &user.Name, &user.LastName, &user.Description, &user.PhotoFile, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *repo) Update(user *User) error {
	_, err := r.db.Exec("UPDATE users set description = ?, photo_file = ? where id = ?", user.Description, user.PhotoFile, user.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *repo) IsLoginExist(login string) bool {
	row := r.db.QueryRow("SELECT id FROM users WHERE login = ?", login)

	user := new(User)
	err := row.Scan(&user.ID)

	if err == sql.ErrNoRows {
		return false
	}

	if err != nil {
		log.Printf("IsLoginExist error: %s", err.Error())
	}

	return true
}

func (r *repo) FindByLoginAndPassword(login string, password string) (*User, error) {
	row := r.db.QueryRow("SELECT id, login, name, last_name, password_hash, description, photo_file, created_at FROM users WHERE login = ?", login)

	user := new(User)
	err := row.Scan(&user.ID, &user.Login, &user.Name, &user.LastName, &user.PasswordHash, &user.Description, &user.PhotoFile, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	user.PasswordHash = ""

	return user, nil
}

func (r *repo) FindByNamePrefix(prefix string, limit int, minId int64) ([]*User, error) {
	rows, err := r.GetRoDB().Query("(select id, name, last_name from users where id>? and name like ? limit 1000) "+
		"union (select id, name, last_name from users where id>? and last_name like ? limit 1000) "+
		"order by id asc limit ?", minId, prefix+"%", minId, prefix+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0, limit)
	for rows.Next() {
		user := new(User)
		err := rows.Scan(&user.ID, &user.Name, &user.LastName)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *repo) Create(user *User) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	res, err := r.db.Exec("INSERT INTO users(login, name, last_name, password_hash, created_at) VALUES(?, ?, ?, ?, NOW())",
		user.Login, user.Name, user.LastName, passwordHash)

	if err != nil {
		return err
	}

	userID, err := res.LastInsertId()

	if err != nil {
		return err
	}

	user.ID = userID

	return nil
}

func (r *repo) BulkCreate(users []*User) {
	size := 500
	for min := 0; min < len(users); min = min + size {
		max := min + size
		if max > len(users) {
			max = len(users)
		}
		log.Printf("bulk: %d - %d", min, max)
		batch := users[min:max]
		err := r.bulkCreate(batch)
		if err != nil {
			log.Printf("bulkCreate error: %s", err.Error())
		}
	}
}

func (r *repo) bulkCreate(users []*User) error {
	valueStrings := make([]string, 0, len(users))
	valueArgs := make([]interface{}, 0, len(users)*5)
	for _, user := range users {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, NOW())")
		valueArgs = append(valueArgs, user.Login)
		valueArgs = append(valueArgs, user.Name)
		valueArgs = append(valueArgs, user.LastName)
		valueArgs = append(valueArgs, user.PasswordHash)
		valueArgs = append(valueArgs, user.Description)
	}
	stmt := fmt.Sprintf("INSERT INTO users(login, name, last_name, password_hash, description, created_at) VALUES %s", strings.Join(valueStrings, ","))
	_, err := r.db.Exec(stmt, valueArgs...)
	return err
}
