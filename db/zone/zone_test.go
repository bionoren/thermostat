package zone

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"thermostat/db"
)

func TestNewZone(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	id, err := New(ctx, t.Name())
	require.NoError(t, err)
	assert.NotZero(t, id)
}

func TestZone_Edit(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	z, err := New(ctx, t.Name())
	require.NoError(t, err)

	z.Name = "foo"
	err = z.Update(ctx)
	require.NoError(t, err)

	z, err = Get(ctx, z.Name)
	require.NoError(t, err)

	assert.Equal(t, "foo", z.Name)
}

func zoneExists(t testing.TB, ctx context.Context, id int64) bool {
	t.Helper()

	row := db.DB.QueryRowContext(ctx, "select id from zone where id=?", id)
	var check int
	_ = row.Scan(&check)

	return check != 0
}

func TestZone_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	z, err := New(ctx, t.Name())
	require.NoError(t, err)
	assert.True(t, zoneExists(t, ctx, z.ID))

	err = z.Delete(ctx)
	require.NoError(t, err)
	assert.False(t, zoneExists(t, ctx, z.ID))
}
