package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func CheckUpAndDown(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+*migrationFilePath,
		"postgres", driver)
	if err != nil {
		return err
	}

	for i := 0; i < 2; i++ {
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			return err
		}
		version, dirty, err := m.Version()
		if err != nil {
			return err
		}
		if dirty {
			return errors.New("it s dirty")
		}
		if version == 0 {
			return errors.New("version should not be 0")
		}
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			return err
		}
		version, dirty, err = m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return err
		}
		if dirty {
			return errors.New("it s dirty")
		}
		if version != 0 {
			return errors.New("version should be 0")
		}
	}
	return nil
}

func CheckEmpty(db *sql.DB) error {
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' and table_name != 'schema_migrations';")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	tablesFound := make([]string, 0)
	for rows.Next() {
		var current string
		if err := rows.Scan(&current); err != nil {
			return err
		}
		tablesFound = append(tablesFound, current)
	}

	if len(tablesFound) == 0 {
		return nil
	}
	return fmt.Errorf("database got more than 0 table found: %s", tablesFound)
}
