package repository

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
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

type userRepo struct {
	*repo
}

func NewUserRepository(dsn string) IUserRepository {
	repo := NewMysqlRepository(dsn)
	return &userRepo{repo: repo}
}

type IUserRepository interface {
	GetDB() *sql.DB
	GetAll() ([]*User, error)
	Get(id int64) (*User, error)
	GetByIDs(ids []int64) ([]*User, error)
	Create(*User) error
	Update(user *User) error
	IsLoginExist(login string) bool
	FindByLoginAndPassword(login string, password string) (*User, error)
	FindByNamePrefix(prefix string, limit int, minId int64) ([]*User, error)
	BulkCreate(users []*User)
}

func (r *repo) GetDB() *sql.DB {
	return r.db
}

func (r *userRepo) GetAll() ([]*User, error) {
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

func (r *userRepo) Get(id int64) (*User, error) {
	row := r.db.QueryRow("SELECT id, login, name, last_name, description, photo_file, created_at FROM users WHERE id = ?", id)

	user := new(User)
	err := row.Scan(&user.ID, &user.Login, &user.Name, &user.LastName, &user.Description, &user.PhotoFile, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepo) GetByIDs(ids []int64) ([]*User, error) {
	idsIn := make([]interface{}, len(ids), len(ids))
	for i := range ids {
		idsIn[i] = ids[i]
	}
	rows, err := r.db.Query("SELECT id, login, name, last_name, description, photo_file, created_at FROM users WHERE id in(?"+strings.Repeat(",? ", len(ids)-1)+")", idsIn...)
	if err != nil {
		return nil, err
	}

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

func (r *userRepo) Update(user *User) error {
	_, err := r.db.Exec("UPDATE users set description = ?, photo_file = ? where id = ?", user.Description, user.PhotoFile, user.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) IsLoginExist(login string) bool {
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

func (r *userRepo) FindByLoginAndPassword(login string, password string) (*User, error) {
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

func (r *userRepo) FindByNamePrefix(prefix string, limit int, minId int64) ([]*User, error) {

	rows, err := r.db.Query("(select id, name, last_name from users where id>? and name like ? limit 1000) "+
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

func (r *userRepo) Create(user *User) error {
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

func (r *userRepo) BulkCreate(users []*User) {
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

func (r *userRepo) bulkCreate(users []*User) error {
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
