package db

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.etcd.io/bbolt"
	"sync"
	"time"
)

var db *bbolt.DB

var connectLock sync.Once

func Connect() error {
	var err error
	connectLock.Do(func() {
		path := viper.GetString("db.file")
		logrus.WithField("path", path).Info("Opening database")
		db, err = bbolt.Open(path, 0666, &bbolt.Options{
			Timeout: 5 * time.Second,
		})
	})
	return err
}

var migrateLock sync.Once

func Migrate() error {
	var err error
	migrateLock.Do(func() {
		err = db.Update(func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("migration"))
			if err != nil {
				return err
			}

			if err := init1(tx, b); err != nil {
				return err
			}

			return nil
		})
	})

	return err
}

// itob returns an 8-byte big endian representation of v.
func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
