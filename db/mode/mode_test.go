package mode

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"thermostat/db"
	"thermostat/db/zone"
)

func TestGetMode(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)

	var modes []Mode
	for i := 0; i < 3; i++ {
		m, err := New(ctx, z.ID, fmt.Sprintf(t.Name()+"%d", i), 70, 80, 1)
		require.NoError(t, err)
		modes = append(modes, m)
	}

	assert.Len(t, modes, 3)
	for _, check := range modes {
		m, err := Get(ctx, check.ID)
		require.NoError(t, err)

		assert.Equal(t, check.ID, m.ID)
		assert.Equal(t, check.MinTemp, m.MinTemp)
		assert.Equal(t, check.MaxTemp, m.MaxTemp)
		assert.Equal(t, check.Correction, m.Correction)
	}
}

func TestNewMode(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)

	id, err := New(ctx, z.ID, t.Name(), 60, 80, 2)
	require.NoError(t, err)
	assert.NotZero(t, id)
}

func TestMode_Edit(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)

	m, err := New(ctx, z.ID, t.Name(), 71, 80, 1)
	require.NoError(t, err)

	m.Name = "foo"
	m.MinTemp = 70
	m.MaxTemp = 75
	err = m.Update(ctx)
	require.NoError(t, err)

	m, err = Get(ctx, m.ID)
	require.NoError(t, err)

	assert.Equal(t, "foo", m.Name)
	assert.Equal(t, float64(70), m.MinTemp)
	assert.Equal(t, float64(75), m.MaxTemp)
	assert.Equal(t, float64(1), m.Correction)
}

func modeExists(t testing.TB, ctx context.Context, id int64) bool {
	t.Helper()

	row := db.DB.QueryRowContext(ctx, "select id from mode where id=?", id)
	var check int
	_ = row.Scan(&check)

	return check != 0
}

func TestMode_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)

	m, err := New(ctx, z.ID, t.Name(), 0, 100, 1)
	require.NoError(t, err)
	assert.True(t, modeExists(t, ctx, m.ID))

	err = m.Delete(ctx)
	require.NoError(t, err)
	assert.False(t, modeExists(t, ctx, m.ID))
}
