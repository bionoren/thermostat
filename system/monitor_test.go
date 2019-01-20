package system

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"reflect"
	"testing"
	"thermostat/db"
	"time"
)

type testSchedule struct {
	ModeID      int64
	PriorityV   db.Priority
	DayOfWeekV  int
	StartDayV   time.Time
	EndDayV     time.Time
	StartTimeV  int
	EndTimeV    int
}

func (c testSchedule) ID() int64 {
	return 0
}

func (c testSchedule) SettingID() int64 {
	return c.ModeID
}

func (c testSchedule) Priority() db.Priority {
	return c.PriorityV
}

func (c testSchedule) DayOfWeek() int {
	return c.DayOfWeekV
}

func (c testSchedule) StartDay() time.Time {
	return c.StartDayV
}

func (c testSchedule) EndDay() time.Time {
	return c.EndDayV
}

func (c testSchedule) StartTime() int {
	return c.StartTimeV
}

func (c testSchedule) EndTime() int {
	return c.EndTimeV
}

func (c testSchedule) modify(t testing.TB, key string, value interface{}) testSchedule {
	field := reflect.ValueOf(&c).Elem().FieldByName(key)
	assert.True(t, field.IsValid(), "Field %s", key)
	assert.True(t, field.CanSet(), "Field %s", key)
	field.Set(reflect.ValueOf(value))

	return c
}

func TestCurrentSetting(t *testing.T) {
	t.Parallel()
	now := time.Now()

	id1, err := db.AddMode("test1", 60, 85, 2)
	require.NoError(t, err)
	id2, err := db.AddMode("test2", 65, 85, 2)
	require.NoError(t, err)
	id3, err := db.AddMode("test3", 70, 80, 1)
	require.NoError(t, err)

	getSetting := func(id int64) db.Setting {
		setting, err := db.GetSetting(id)
		require.NoError(t, err)
		return setting
	}

	modes := map[int64]db.Setting{
		id1: getSetting(id1),
		id2: getSetting(id2),
		id3: getSetting(id3),
	}

	defaultSetting := testSchedule{
		id1,
		db.DEFAULT,
		255,
		time.Unix(0, 0),
		time.Unix(7258118400, 0),
		0,
		86400,
	}
	defaultSettings := []db.Schedule{defaultSetting}

	tests := []struct {
		s        []map[string]interface{}
		expected int64
	}{
		{nil, id1},
		{[]map[string]interface{}{{"PriorityV": db.SCHEDULED, "ModeID": id2}}, id2},
		{[]map[string]interface{}{{"PriorityV": db.SCHEDULED, "ModeID": id3, "DayOfWeekV": 255 &^ weekdayMask(now.Weekday())}}, id1},
		{[]map[string]interface{}{{"PriorityV": db.SCHEDULED, "ModeID": id3, "DayOfWeekV": 255 &^ weekdayMask(now.Weekday())}}, id1},
		{[]map[string]interface{}{{"PriorityV": db.SCHEDULED, "ModeID": id3, "DayOfWeekV": weekdayMask(now.Weekday())}}, id3},
		{[]map[string]interface{}{{"PriorityV": db.SCHEDULED, "ModeID": id2}, {"PriorityV": db.OVERRIDE, "ModeID": id3}}, id3},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "StartDayV": now}}, id3},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "StartDayV": now.Add(time.Second)}}, id1},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "EndDayV": now.Add(time.Second)}}, id3},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "EndDayV": now.Add(-time.Second)}}, id1},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "StartTimeV": daySeconds(now)}}, id3},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "StartTimeV": daySeconds(now) + 1}}, id1},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "EndTimeV": daySeconds(now) + 1}}, id3},
		{[]map[string]interface{}{{"PriorityV": db.OVERRIDE, "ModeID": id3, "EndTimeV": daySeconds(now) - 1}}, id1},
	}

	for i, tt := range tests {
		settings := defaultSettings
		for _, s := range tt.s {
			setting := defaultSetting
			for k, v := range s {
				setting = setting.modify(t, k, v)
			}
			settings = append(settings, setting)
		}

		setting := currentSetting(settings)
		assert.Equal(t, modes[tt.expected].MinTemp(), setting.MinTemp(), "Test %d", i)
		assert.Equal(t, modes[tt.expected].MaxTemp(), setting.MaxTemp(), "Test %d", i)
		assert.Equal(t, modes[tt.expected].Correction(), setting.Correction(), "Test %d", i)
	}
}

func TestWeekdayMask(t *testing.T) {
	t.Parallel()

	for i := time.Sunday; i <= time.Saturday; i++ {
		assert.Equal(t, int(math.Pow(2, float64(i+1))), weekdayMask(i), "Day %s", i)
	}
}
