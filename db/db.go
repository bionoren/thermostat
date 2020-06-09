package db

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sync"
	"thermostat/config"
)

var DB *sql.DB

var connectLock sync.Once
var migrateLock sync.Once

func init() {
	config.Ready()

	connectLock.Do(func() {
		if err := connect(); err != nil {
			panic(err)
		}
		if err := migrateDB(); err != nil {
			panic(err)
		}
	})
}

func connect() error {
	path := viper.GetString("db.file")
	logrus.WithField("path", path).Info("Opening database")

	var err error
	if DB, err = sql.Open("sqlite3", "file:"+path+"?_fk=true&_sync=3"); err != nil {
		return err
	}
	if err := DB.Ping(); err != nil {
		return err
	}
	return nil
}

func migrateDB() error {
	var err error
	migrateLock.Do(func() {
		var driver database.Driver
		if driver, err = sqlite3.WithInstance(DB, &sqlite3.Config{}); err != nil {
			return
		}
		var m *migrate.Migrate
		if m, err = migrate.NewWithDatabaseInstance( "file://"+viper.GetString("db.migrations"), "sqlite3", driver); err != nil {
			return
		}
		if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return
		}
		err = nil
	})

	return err
}
