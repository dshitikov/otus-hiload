package repository

import (
	"database/sql"
	"log"
	"strings"
)

type repo struct {
	db *sql.DB
}

type IRepository interface {
	IUserRepository
}

func NewMysqlRepository(dsn string) *repo {
	dsn = strings.Replace(dsn, "mysql://", "", 1) + "?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	return &repo{db: db}
}
