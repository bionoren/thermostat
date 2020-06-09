package system

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"thermostat/db/mode"
	"thermostat/db/setting"
	"thermostat/db/zone"
	"time"
)

func TestCurrentSetting(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	now := time.Now()

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)

	m1, err := mode.New(ctx, z.ID, "test1", 60, 85, 2)
	require.NoError(t, err)
	m2, err := mode.New(ctx, z.ID, "test2", 65, 85, 2)
	require.NoError(t, err)
	m3, err := mode.New(ctx, z.ID, "test3", 70, 80, 1)
	require.NoError(t, err)

	defaultSetting, err := setting.New(ctx, z.ID, m1.ID, setting.DEFAULT, 255, time.Unix(0, 0), time.Unix(7258118400, 0), 0, 86400)
	require.NoError(t, err)
	defaultSettings := []setting.Setting{defaultSetting}

	tests := []struct {
		s        []map[string]interface{}
		expected mode.Mode
	}{
		{nil, m1},
		{[]map[string]interface{}{{"priority": setting.SCHEDULED, "modeID": m2.ID}}, m2},
		{[]map[string]interface{}{{"priority": setting.SCHEDULED, "modeID": m3.ID, "dayOfWeek": 255 &^ setting.WeekdayMask(now.Weekday())}}, m1},
		{[]map[string]interface{}{{"priority": setting.SCHEDULED, "modeID": m3.ID, "dayOfWeek": 255 &^ setting.WeekdayMask(now.Weekday())}}, m1},
		{[]map[string]interface{}{{"priority": setting.SCHEDULED, "modeID": m3.ID, "dayOfWeek": setting.WeekdayMask(now.Weekday())}}, m3},
		{[]map[string]interface{}{{"priority": setting.SCHEDULED, "modeID": m2.ID}, {"priority": setting.OVERRIDE, "modeID": m3.ID}}, m3},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "startDay": now.Add(-time.Second)}}, m3},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "startDay": now.Add(3*time.Second)}}, m1},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "endDay": now.Add(3*time.Second)}}, m3},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "endDay": now.Add(-time.Second)}}, m1},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "startTime": daySeconds(now)}}, m3},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "startTime": daySeconds(now) + 3}}, m1},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "endTime": daySeconds(now) + 3}}, m3},
		{[]map[string]interface{}{{"priority": setting.OVERRIDE, "modeID": m3.ID, "endTime": daySeconds(now) - 3}}, m1},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			settings := defaultSettings
			for _, s := range tt.s {
				current := defaultSetting
				data, err := json.Marshal(current)
				require.NoError(t, err)

				var settingData map[string]interface{}
				err = json.Unmarshal(data, &settingData)
				require.NoError(t, err)

				for k, v := range s {
					settingData[k] = v
				}

				data, err = json.Marshal(settingData)
				require.NoError(t, err)

				var s setting.Setting
				err = json.Unmarshal(data, &s)
				require.NoError(t, err)

				settings = append(settings, s)
			}

			current := currentSetting(settings)
			m := current.Mode(ctx)
			assert.Equal(t, tt.expected.MinTemp, m.MinTemp)
			assert.Equal(t, tt.expected.MaxTemp, m.MaxTemp)
			assert.Equal(t, tt.expected.Correction, m.Correction)
		})
	}
}
