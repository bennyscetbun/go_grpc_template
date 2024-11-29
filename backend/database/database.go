package database

import (
	"database/sql"
	"flag"
	"fmt"

	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/environment"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ztrue/tracerr"
)

var migrationFilePath = flag.String("migration_file_path", "./resources/database/migrations", "migration files path")

func OpenPSQL() (*sql.DB, error) {
	dbHost, err := environment.GetenvString("DBHOST", "localhost")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	dbPort, err := environment.GetenvInt("DBPORT", 5432)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	dbPasswd, err := environment.GetenvString("DBPASSWD", "")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	dbUser, err := environment.GetenvString("DBUSER", "postgres")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	dbName, err := environment.GetenvString("DBNAME", "postgres")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	con, err := sql.Open("pgx", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPasswd, dbName))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return con, tracerr.Wrap(con.Ping())
}

func MigratePSQL(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return tracerr.Wrap(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+*migrationFilePath,
		"postgres", driver)
	if err != nil {
		return tracerr.Wrap(err)
	}
	err = m.Up()
	if err == nil || err == migrate.ErrNoChange {
		return nil
	}

	switch v := err.(type) {
	case migrate.ErrDirty:
		if err := m.Force(v.Version); err != nil {
			return tracerr.Wrap(err)
		}
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			return tracerr.Wrap(err)
		}
		err = m.Up()
		if err == nil || err == migrate.ErrNoChange {
			return nil
		}
	}

	return tracerr.Wrap(err)
}
