package db

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	"testing"
	"time"
)

func addMode(t testing.TB, name string, min, max float64) mode {
	t.Helper()

	m := mode{
		Name:       name,
		MinTempv:    min,
		MaxTempv:    max,
		Correctionv: 1,
	}

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		id, _ := b.NextSequence()
		m.ID = int64(id)
		buf, err := json.Marshal(m)
		if err != nil {
			return err
		}
		return b.Put(itob(m.ID), buf)
	})
	require.NoError(t, err)

	return m
}

func TestAddSetting(t *testing.T) {
	t.Parallel()
	now := time.Now()

	m := addMode(t, t.Name(), 60, 80)
	s, err := AddSetting(m.ID, DEFAULT, 1, now, now, 1, 86400)
	require.NoError(t, err)
	assert.Equal(t, DEFAULT, s.Priority())
	assert.Equal(t, 1, s.DayOfWeek())
	assert.Equal(t, now, s.StartDay())
	assert.Equal(t, now, s.EndDay())
	assert.Equal(t, 1, s.StartTime())
	assert.Equal(t, 86400, s.EndTime())
}

func settingExists(t testing.TB, id int64) bool {
	t.Helper()

	var exists bool
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))

		exists = b.Get(itob(id)) != nil
		return nil
	})
	require.NoError(t, err)

	return exists
}

func TestDeleteSetting(t *testing.T) {
	t.Parallel()

	var s setting
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))

		id, _ := b.NextSequence()
		s.IDv = int64(id)

        buf, err := json.Marshal(s)
        require.NoError(t, err)

        return b.Put(itob(s.IDv), buf)
	})
	require.NoError(t, err)

	assert.True(t, settingExists(t, s.IDv))

	err = DeleteSetting(s.IDv)
	require.NoError(t, err)

	assert.False(t, settingExists(t, s.IDv))
}
