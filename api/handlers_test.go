package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"net/http"
	"testing"
	"thermostat/db/mode"
	"thermostat/db/setting"
	"thermostat/db/zone"
	"thermostat/system"
	"time"
)

type nopController struct {
}

func (c nopController) Fan() bool {
	return false
}

func (c nopController) SetFan(_ bool) bool {
	return false
}

func (c nopController) AC() bool {
	return false
}

func (c nopController) SetAC(_ bool) bool {
	return false
}

func (c nopController) Heat() bool {
	return false
}

func (c nopController) SetHeat(_ bool) bool {
	return false
}

var _ system.Controller = &nopController{}

type constantSensor float64

func (s constantSensor) Temperature() float64 {
	return float64(s)
}

func (s constantSensor) Humidity() float64 {
	return 0
}

var _ system.Sensor = constantSensor(0)

func TestEditHandler(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	for {
		z, err := zone.Get(ctx, t.Name())
		if err != nil {
			break
		}
		err = z.Delete(ctx)
		require.NoError(t, err)
	}

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)

	defaultMode, err := mode.New(ctx, z.ID, "default", 70, 75, 2)
	require.NoError(t, err)
	_, err = mode.New(ctx, z.ID, "custom", 70, 75, 2)
	require.NoError(t, err)
	mode1, err := mode.New(ctx, z.ID, "mode1", 70, 75, 2)
	require.NoError(t, err)

	systemZone, err := system.NewZone(ctx, t.Name(), nopController{}, constantSensor(72))
	require.NoError(t, err)
	systemZone.Startup()

	settings, err := setting.All(ctx, z.ID)
	require.NoError(t, err)
	for _, s := range settings {
		err := s.Delete(ctx)
		require.NoError(t, err)
	}

	defaultSetting, err := setting.New(ctx, z.ID, defaultMode.ID, setting.DEFAULT, math.MaxInt32, time.Time{}, time.Now().Add(time.Hour*24*365), 0, 86400)
	require.NoError(t, err)

	tests := []struct {
		name  string
		mode  int64
		delta float64
	}{
		{"change mode", mode1.ID, 0},
		{"delta up", 0, 1},
		{"delta down", 0, -1},
	}

	for i, tt := range tests {
		if i != 1 {
			continue
		}
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			settings, err := setting.All(ctx, systemZone.ID())
			require.NoError(t, err)
			for _, s := range settings {
				if s.Priority != setting.DEFAULT {
					err := s.Delete(ctx)
					require.NoError(t, err)
				}
			}

			systemZone.Update([]setting.Setting{defaultSetting})

			for i := 1; i <= 5; i++ {
				if systemZone.Setting().ModeID != 0 {
					break
				}
				time.Sleep(time.Millisecond * time.Duration(5*i))
			}
			require.NotZero(t, systemZone.Setting().ModeID)

			data := struct {
				ZoneID int64
				ModeID int64
				Delta  float64
			}{
				z.ID,
				tt.mode,
				tt.delta,
			}
			msg, err := json.Marshal(data)
			require.NoError(t, err)

			response := editHandler(ctx, msg)
			assert.Equal(t, http.StatusOK, response.Code)

			settings, err = setting.All(ctx, z.ID)
			require.NoError(t, err)
			assert.NotEmpty(t, settings)

			var custom setting.Setting
			for _, s := range settings {
				if s.Priority == setting.CUSTOM {
					assert.Zero(t, custom.ID, "found multiple custom settings")
					custom = s
				}
			}
			assert.NotZero(t, custom.ID)

			if tt.mode != 0 {
				assert.Equal(t, tt.mode, custom.ModeID)
			} else {
				require.NotZero(t, custom.ModeID)
				m := custom.Mode(ctx)
				assert.NotZero(t, m.ID)
				assert.NotEqual(t, defaultMode.ID, m.ID)
				assert.Equal(t, defaultMode.MinTemp+tt.delta, m.MinTemp)
				assert.Equal(t, defaultMode.MaxTemp+tt.delta, m.MaxTemp)
				assert.Equal(t, float64(1), m.Correction)
			}

			assert.Equal(t, custom.ID, systemZone.Setting().ID)
		})
	}
}
