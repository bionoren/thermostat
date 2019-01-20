package db

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	"testing"
)

func TestGetSetting(t *testing.T) {
	var modes []mode
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		for i := 0; i < 3; i++ {
			var m mode

			id, _ := b.NextSequence()
			m.ID = int64(id)
			buf, err := json.Marshal(m)
			if err != nil {
				return err
			}
			if err := b.Put(itob(m.ID), buf); err != nil {
				return err
			}

			modes = append(modes, m)
		}

		return nil
	})
	require.NoError(t, err)

	assert.Len(t, modes, 3)
	for _, check := range modes {
		m, err := GetSetting(check.ID)
		require.NoError(t, err)
		assert.Equal(t, check.ID, m.(mode).ID)
		assert.Equal(t, check.MinTempv, m.(mode).MinTempv)
		assert.Equal(t, check.MaxTempv, m.(mode).MaxTempv)
		assert.Equal(t, check.Correctionv, m.(mode).Correctionv)
	}
}

func TestAddMode(t *testing.T) {
	t.Parallel()

	id, err := AddMode(t.Name(), 60, 80, 2)
	require.NoError(t, err)
	assert.NotZero(t, id)
}

func TestEditMode(t *testing.T) {
	t.Parallel()

	var m mode
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		id, _ := b.NextSequence()
		m.ID = int64(id)

		buf, err := json.Marshal(m)
		require.NoError(t, err)

		return b.Put(itob(m.ID), buf)
	})
	require.NoError(t, err)

	err = EditMode(m.ID, "foo", 70, 75, 1)
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		data := b.Get(itob(m.ID))
		return json.Unmarshal(data, &m)
	})
	require.NoError(t, err)

	assert.Equal(t, "foo", m.Name)
	assert.Equal(t, float64(70), m.MinTempv)
	assert.Equal(t, float64(75), m.MaxTempv)
	assert.Equal(t, float64(1), m.Correctionv)
}

func modeExists(t testing.TB, id int64) bool {
	t.Helper()

	var exists bool
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		exists = b.Get(itob(id)) != nil
		return nil
	})
	require.NoError(t, err)

	return exists
}

func TestDeleteMode(t *testing.T) {
	t.Parallel()

	var m mode
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("modes"))

		id, _ := b.NextSequence()
		m.ID = int64(id)

		buf, err := json.Marshal(m)
		require.NoError(t, err)

		return b.Put(itob(m.ID), buf)
	})
	require.NoError(t, err)

	assert.True(t, modeExists(t, m.ID))

	err = DeleteMode(m.ID)
	require.NoError(t, err)

	assert.False(t, modeExists(t, m.ID))
}
