package repository

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"log"
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
	GetDB() *sql.DB
	GetAll() ([]*User, error)
	Get(id int64) (*User, error)
	Create(*User) error
	Update(user *User) error
	IsLoginExist(login string) bool
	FindByLoginAndPassword(login string, password string) (*User, error)
	FindByNameAndLastNamePrefix(name string, lastName string) ([]*User, error)
}

func (r *repo) GetDB() *sql.DB {
	return r.db
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

func (r *repo) FindByNameAndLastNamePrefix(name string, lastName string) ([]*User, error) {
	count, err := r.countByNameAndLastNamePrefix(name, lastName)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query("SELECT id, login, name, last_name, description, photo_file, created_at FROM users where name like ? and last_name like ?",
		name+"%", lastName+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0, count)
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

func (r *repo) countByNameAndLastNamePrefix(name string, lastName string) (int64, error) {
	var count int64
	rows := r.db.QueryRow("SELECT count(*) FROM users where name like ? and last_name like ?", name+"%", lastName+"%")
	err := rows.Scan(&count)
	return count, err
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
