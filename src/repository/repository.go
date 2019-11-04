package repository

import (
	"database/sql"
	"log"
	"strings"
)

type repo struct {
	db           *sql.DB
	slaveDb      *sql.DB
	readReplicas []*sql.DB
	masterCnt    int32
	slaveCnt     int32
}

type IRepository interface {
	IUserRepository
}

func NewMysqlRepository(dsn string, slaveDsn string) IRepository {
	dsn = strings.Replace(dsn, "mysql://", "", 1) + "?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)

	slaveDsn = strings.Replace(slaveDsn, "mysql://", "", 1) + "?parseTime=true"
	slaveDb, err := sql.Open("mysql", slaveDsn)
	if err != nil {
		log.Fatal(err)
	}
	slaveDb.SetConnMaxLifetime(0)
	replicas := []*sql.DB{db, slaveDb}
	return &repo{db: db, slaveDb: slaveDb, readReplicas: replicas}
}
