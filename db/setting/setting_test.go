package setting

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
	"thermostat/db"
	"thermostat/db/mode"
	"thermostat/db/zone"
	"time"
)

func TestAddSetting(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	now := time.Now()
	later := now.Add(time.Minute)

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)

	m, err := mode.New(ctx, z.ID, t.Name(), 60, 80, 1)
	require.NoError(t, err)
	s, err := New(ctx, z.ID, m.ID, DEFAULT, 1, now, later, 1, 86400)
	require.NoError(t, err)

	assert.Equal(t, DEFAULT, s.Priority)
	assert.Equal(t, 1, s.DayOfWeek)
	assert.Equal(t, now, s.StartDay)
	assert.Equal(t, later, s.EndDay)
	assert.Equal(t, 1, s.StartTime)
	assert.Equal(t, 86400, s.EndTime)
}

func settingExists(t testing.TB, ctx context.Context, id int64) bool {
	t.Helper()

	row := db.DB.QueryRowContext(ctx, "select id from setting where id=?", id)
	var check int
	_ = row.Scan(&check)

	return check != 0
}

func TestSetting_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	z, err := zone.New(ctx, t.Name())
	require.NoError(t, err)
	m, err := mode.New(ctx, z.ID, t.Name(), 70, 80, 1)
	require.NoError(t, err)

	s, err := New(ctx, z.ID, m.ID, SCHEDULED, 1, time.Now(), time.Now().Add(time.Minute), 0, 50)
	require.NoError(t, err)

	assert.True(t, settingExists(t, ctx, s.ID))

	err = s.Delete(ctx)
	require.NoError(t, err)

	assert.False(t, settingExists(t, ctx, s.ID))
}

func TestSetting_Runtime(t *testing.T) {
	t.Parallel()

	now := time.Date(2019, 12, 31, 12, 12, 13, 0, time.Local) // Tuesday
	tests := []struct {
		name               string
		start, end         time.Time
		startTime, endTime int
		dayOfWeek          []time.Weekday
		expected           time.Time
	}{
		{
			name:      "before start date",
			start:     now.AddDate(0, 0, 10),
			end:       now.AddDate(0, 0, 11),
			startTime: 0,
			endTime:   86400,
			dayOfWeek: []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday},
			expected:  now.AddDate(0, 0, 10),
		},
		{
			name:      "after end date",
			start:     now.AddDate(0, 0, -10),
			end:       now.AddDate(0, 0, -1),
			startTime: 0,
			endTime:   86400,
			dayOfWeek: []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday},
			expected:  time.Time{},
		},
		{
			name:      "today now",
			start:     now.AddDate(0, 0, -10),
			end:       now.AddDate(0, 0, 10),
			startTime: 0,
			endTime:   86400,
			dayOfWeek: []time.Weekday{time.Tuesday},
			expected:  now,
		},
		{
			name:      "today but later",
			start:     now.AddDate(0, 0, -10),
			end:       now.AddDate(0, 0, 10),
			startTime: 60 * 60 * 14,
			endTime:   60 * 60 * 18,
			dayOfWeek: []time.Weekday{time.Tuesday},
			expected:  time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, now.Location()),
		},
		{
			name:      "today but earlier",
			start:     now.AddDate(0, 0, -10),
			end:       now.AddDate(0, 0, 10),
			startTime: 60 * 60 * 4,
			endTime:   60 * 60 * 8,
			dayOfWeek: []time.Weekday{time.Tuesday},
			expected:  time.Date(now.Year(), now.Month(), now.Day()+7, 4, 0, 0, 0, now.Location()),
		},
		{
			name:      "tomorrow",
			start:     now.AddDate(0, 0, -10),
			end:       now.AddDate(0, 0, 10),
			startTime: 60 * 60 * 4,
			endTime:   60 * 60 * 8,
			dayOfWeek: []time.Weekday{time.Sunday, time.Monday, time.Wednesday, time.Thursday, time.Friday, time.Saturday},
			expected:  time.Date(now.Year(), now.Month(), now.Day()+1, 4, 0, 0, 0, now.Location()),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			var dayOfWeek int
			for _, d := range tt.dayOfWeek {
				dayOfWeek |= 2 << uint(d)
			}

			s := Setting{
				DayOfWeek: dayOfWeek,
				StartDay:  tt.start,
				EndDay:    tt.end,
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}

			runtime := s.Runtime(now)
			// strip monotonic clock readings with Round(0) so == will work
			assert.Equal(t, tt.expected.Round(0), runtime.Round(0))
		})
	}
}

func TestOverlaps(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	now := time.Now()

	z1, err := zone.New(ctx, t.Name()+"1")
	require.NoError(t, err)
	z2, err := zone.New(ctx, t.Name()+"2")
	require.NoError(t, err)
	m1, err := mode.New(ctx, z1.ID, t.Name()+"1", 70, 80, 1)
	require.NoError(t, err)
	m2, err := mode.New(ctx, z1.ID, t.Name()+"2", 71, 79, 2)
	require.NoError(t, err)
	m3, err := mode.New(ctx, z2.ID, t.Name()+"1", 70, 80, 1)
	require.NoError(t, err)
	existing, err := New(ctx, z1.ID, m1.ID, SCHEDULED, WeekdayMask(time.Monday)|WeekdayMask(time.Wednesday), now, now.Add(time.Hour*24*30), 32400, 61200) // 9 to 5 monday and wednesday for the next 30 days
	require.NoError(t, err)

	tests := []struct {
		name      string
		zone      zone.Zone
		priority  Priority
		weekdays  []time.Weekday
		start     time.Time
		end       time.Time
		startTime int
		endTime   int
		overlap   bool
	}{
		{
			name:    "different zone",
			zone:    z2,
			overlap: false,
		},
		{
			name:     "different priority",
			priority: OVERRIDE,
			overlap:  false,
		},
		{
			name:    "span is before",
			start:   existing.StartDay.Add(-time.Hour * 24),
			end:     existing.StartDay.Add(-time.Second),
			overlap: false,
		},
		{
			name:    "span is after",
			start:   existing.EndDay.Add(time.Second),
			end:     existing.EndDay.Add(time.Hour * 24),
			overlap: false,
		},
		{
			name:     "different weekday",
			weekdays: []time.Weekday{time.Sunday, time.Tuesday, time.Thursday, time.Friday, time.Saturday},
			overlap:  false,
		},
		{
			name:      "before",
			startTime: 0,
			endTime:   existing.StartTime - 1,
			overlap:   false,
		},
		{
			name:      "after",
			startTime: existing.EndTime + 1,
			endTime:   86400,
			overlap:   false,
		},
		{
			name:      "overlaps start time",
			startTime: 0,
			endTime:   existing.StartTime,
			overlap:   true,
		},
		{
			name:      "overlaps end time",
			startTime: existing.EndTime,
			endTime:   86400,
			overlap:   true,
		},
		{
			name:    "overlaps start span",
			start:   existing.StartDay.Add(-time.Hour * 24),
			end:     existing.StartDay,
			overlap: true,
		},
		{
			name:    "overlaps end span",
			start:   existing.EndDay,
			end:     existing.EndDay.Add(time.Hour * 24),
			overlap: true,
		},
		{
			name:    "covers span",
			start:   existing.StartDay.Add(-time.Second),
			end:     existing.EndDay.Add(time.Second),
			overlap: true,
		},
		{
			name:    "inside span",
			start:   existing.StartDay.Add(time.Second),
			end:     existing.EndDay.Add(-time.Second),
			overlap: true,
		},
		{
			name:      "covers time",
			startTime: existing.StartTime - 1,
			endTime:   existing.EndTime + 1,
			overlap:   true,
		},
		{
			name:      "inside time",
			startTime: existing.StartTime + 1,
			endTime:   existing.EndTime - 1,
			overlap:   true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			var m mode.Mode
			switch tt.zone.ID {
			case z2.ID:
				m = m3
			case z1.ID:
				fallthrough
			default:
				m = m2
			}
			sched := Setting{
				ZoneID:    z1.ID,
				ModeID:    m.ID,
				Priority:  existing.Priority,
				DayOfWeek: existing.DayOfWeek,
				StartDay:  existing.StartDay,
				EndDay:    existing.EndDay,
				StartTime: existing.StartTime,
				EndTime:   existing.EndTime,
			}

			if tt.zone.ID != 0 {
				sched.ZoneID = tt.zone.ID
			}
			if tt.priority != 0 {
				sched.Priority = tt.priority
			}
			if tt.weekdays != nil {
				sched.DayOfWeek = 0
				for _, d := range tt.weekdays {
					sched.DayOfWeek |= WeekdayMask(d)
				}
			}
			if !tt.start.IsZero() {
				sched.StartDay = tt.start
			}
			if !tt.end.IsZero() {
				sched.EndDay = tt.end
			}
			if tt.startTime != 0 {
				sched.StartTime = tt.startTime
			}
			if tt.endTime != 0 {
				sched.EndTime = tt.endTime
			}

			overlap := Overlaps(sched, existing)
			assert.Equal(t, tt.overlap, overlap)
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	now := time.Now()

	z1, err := zone.New(ctx, t.Name()+"1")
	require.NoError(t, err)
	m1, err := mode.New(ctx, z1.ID, t.Name()+"1", 70, 80, 1)
	require.NoError(t, err)
	m2, err := mode.New(ctx, z1.ID, t.Name()+"2", 71, 79, 2)
	require.NoError(t, err)
	existing, err := New(ctx, z1.ID, m1.ID, SCHEDULED, WeekdayMask(time.Monday)|WeekdayMask(time.Wednesday), now, now.Add(time.Hour*24*30), 32400, 61200) // 9 to 5 monday and wednesday for the next 30 days
	require.NoError(t, err)

	tests := []struct {
		name      string
		weekdays  []time.Weekday
		start     time.Time
		end       time.Time
		startTime int
		endTime   int
		err       string
	}{
		{
			name:     "valid",
			weekdays: []time.Weekday{time.Tuesday},
		},
		{
			name:     "backward span",
			weekdays: []time.Weekday{time.Tuesday},
			start:    existing.EndDay,
			end:      existing.StartDay,
			err:      "setting start must be before setting end",
		},
		{
			name:      "backward time",
			weekdays:  []time.Weekday{time.Tuesday},
			startTime: existing.EndTime,
			endTime:   existing.StartTime,
			err:       "setting end time must be after start time",
		},
		{
			name: "no days",
			err:  "setting must be active on at least one day of the week",
		},
		{
			name:     "overlapping",
			weekdays: []time.Weekday{time.Monday},
			err:      fmt.Sprintf("new setting overlaps with setting %d", existing.ID),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			sched := Setting{
				ZoneID:    z1.ID,
				ModeID:    m2.ID,
				Priority:  existing.Priority,
				StartDay:  existing.StartDay,
				EndDay:    existing.EndDay,
				StartTime: existing.StartTime,
				EndTime:   existing.EndTime,
			}

			for _, d := range tt.weekdays {
				sched.DayOfWeek |= WeekdayMask(d)
			}

			if !tt.start.IsZero() {
				sched.StartDay = tt.start
			}
			if !tt.end.IsZero() {
				sched.EndDay = tt.end
			}
			if tt.startTime != 0 {
				sched.StartTime = tt.startTime
			}
			if tt.endTime != 0 {
				sched.EndTime = tt.endTime
			}

			err := Validate(ctx, sched)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWeekdayMask(t *testing.T) {
	t.Parallel()

	for i := time.Sunday; i <= time.Saturday; i++ {
		assert.Equal(t, int(math.Pow(2, float64(i+1))), WeekdayMask(i), "Day %s", i)
	}
}
