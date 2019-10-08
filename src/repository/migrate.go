package repository

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"strconv"
)

func Migrate(connUri string, migrationsDir string) error {
	log.Print("Migration started")

	fileSource, err := (&file.File{}).Open("file://" + migrationsDir)
	if err != nil {
		log.Printf("Migration file open failed: %s", err.Error())
		return err
	}
	log.Print("Opened migration store")

	log.Print("create NewSourceWithInstance")
	m, err := migrate.NewWithSourceInstance("file", fileSource, connUri)
	if err != nil {
		log.Printf("NewWithSourceInstance creation failed: %s", err.Error())
		return err
	}
	log.Print("Migrator initialized")

	version, dirty, err := m.Version()
	if err != nil {
		log.Print("No current migration version found")
	}

	log.Print("Migrating: current schema version ",
		strconv.FormatInt(int64(version), 10),
		" dirty: ", strconv.FormatBool(dirty))
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Print("Migration: already up to date (no change)")
		} else {
			log.Printf("Migration: ERROR %s", err.Error())
			return err
		}
	}

	log.Print("Migration: SUCCESS")
	srcErr, dbErr := m.Close()

	if srcErr != nil {
		log.Printf("Migration: ERROR while closing source - %v", srcErr.Error())
	}

	if dbErr != nil {
		log.Printf("Migration: ERROR while closing database - %v", dbErr.Error())
	}

	return nil
}
