package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"thermostat/api/request"
	"thermostat/db/mode"
	"thermostat/db/setting"
	"thermostat/db/zone"
	"thermostat/sensor"
	"thermostat/system"
	"time"
)

func zones(ctx context.Context, _ json.RawMessage) request.ApiResponse {
	zones, err := zone.All(ctx)
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	msg, err := json.Marshal(zones)
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	return request.NewResponse(http.StatusOK, string(msg))
}

func status(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var input struct {
		ZoneID int64
	}
	if err := json.Unmarshal(msg, &input); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	z, err := system.GetZone(input.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	var data struct {
		ScheduleID  int64
		ModeID      int64
		Temperature float64
		Humidity    float64
		HeatIndex   float64
		Min         float64
		Max         float64
		Correction  float64
		Heat        bool
		AC          bool
		Fan         bool
	}

	config := z.Setting()
	m := config.Mode(ctx)
	sens := z.Sensor()
	sys := z.Controller()

	data.ScheduleID = config.ID
	data.ModeID = config.ModeID
	data.Temperature = sens.Temperature()
	data.Humidity = sens.Humidity()
	data.HeatIndex = sensor.HeatIndex(data.Temperature, data.Humidity)
	data.Min = m.MinTemp
	data.Max = m.MaxTemp
	data.Correction = m.Correction
	data.Heat = sys.Heat()
	data.AC = sys.AC()
	data.Fan = sys.Fan()

	if msg, err = json.Marshal(data); err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	return request.NewResponse(http.StatusOK, string(msg))
}

func schedules(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var input struct {
		ZoneID int64
	}
	if err := json.Unmarshal(msg, &input); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	z, err := system.GetZone(input.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	settings, err := setting.All(ctx, z.ID())
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	userSettings := settings[:0]
	for _, s := range settings {
		if s.Priority != setting.CUSTOM && s.Priority != setting.DEFAULT {
			userSettings = append(userSettings, s)
		}
	}

	msg, err = json.Marshal(userSettings)
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	return request.NewResponse(http.StatusOK, string(msg))
}

func modes(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var input struct {
		ZoneID int64
	}
	if err := json.Unmarshal(msg, &input); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	z, err := system.GetZone(input.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	modes, err := mode.All(ctx, z.ID())
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	msg, err = json.Marshal(modes)
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	return request.NewResponse(http.StatusOK, string(msg))
}

func addSchedule(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var s setting.Setting

	if err := json.Unmarshal(msg, &s); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	if s.Priority == setting.DEFAULT {
		return request.NewResponse(http.StatusBadRequest, "you may not add a schedule with default priority")
	}
	if s.Priority == setting.CUSTOM {
		return request.NewResponse(http.StatusBadRequest, "you may not add a schedule with custom priority")
	}

	z, err := system.GetZone(s.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	if s, err = setting.New(ctx, z.ID(), s.ModeID, s.Priority, s.DayOfWeek, s.StartDay, s.EndDay, s.StartTime, s.EndTime); err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	settings, err := setting.All(ctx, z.ID())
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	z.Update(append(settings, s))

	resp, err := json.Marshal(s)
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	return request.NewResponse(http.StatusOK, string(resp))
}

func deleteSchedule(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var data struct {
		ZoneID int64
		ID     int64
	}

	if err := json.Unmarshal(msg, &data); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	z, err := system.GetZone(data.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	settings, err := setting.All(ctx, z.ID())
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	for i, s := range settings {
		if s.ID == data.ID {
			if err := s.Delete(ctx); err != nil {
				return request.NewResponse(http.StatusInternalServerError, err.Error())
			}
			settings = append(settings[:i], settings[i+1:]...)
			break
		}
	}

	z.Update(settings)

	return request.NewResponse(http.StatusOK, `{}`)
}

func addMode(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var data mode.Mode
	if err := json.Unmarshal(msg, &data); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}
	if err := data.Validate(); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	m, err := mode.New(ctx, data.ZoneID, data.Name, data.MinTemp, data.MaxTemp, data.Correction)
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	return request.NewResponse(http.StatusOK, fmt.Sprintf(`{"id": %d}`, m.ID))
}

func editMode(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var m mode.Mode
	if err := json.Unmarshal(msg, &m); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}
	if err := m.Validate(); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	z, err := system.GetZone(m.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	if err := m.Update(ctx); err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	settings, err := setting.All(ctx, z.ID())
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}
	z.Update(settings)

	return request.NewResponse(http.StatusOK, "")
}

func deleteMode(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var m mode.Mode
	if err := json.Unmarshal(msg, &m); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	z, err := system.GetZone(m.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	settings, err := setting.All(ctx, z.ID())
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	for _, s := range settings {
		if s.ModeID == m.ID {
			return request.NewResponse(http.StatusBadRequest, fmt.Sprintf("a schedule (%d) depends on mode %d", s.ID, m.ID))
		}
	}

	if err := m.Delete(ctx); err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	return request.NewResponse(http.StatusOK, `{}`)
}

func editHandler(ctx context.Context, msg json.RawMessage) request.ApiResponse {
	var data struct {
		ZoneID int64
		ModeID int64
		Delta  float64
	}

	if err := json.Unmarshal(msg, &data); err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	z, err := system.GetZone(data.ZoneID)
	if err != nil {
		return request.NewResponse(http.StatusBadRequest, err.Error())
	}

	current := z.Setting()
	currentMode, err := mode.Get(ctx, current.ModeID)
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	var custom mode.Mode
	if data.ModeID != 0 {
		if custom, err = mode.Get(ctx, data.ModeID); err != nil {
			return request.NewResponse(http.StatusInternalServerError, err.Error())
		}
	} else {
		modes, err := mode.All(ctx, z.ID())
		if err != nil {
			return request.NewResponse(http.StatusInternalServerError, err.Error())
		}

		for _, m := range modes {
			if m.Name == "custom" {
				custom = m
				break
			}
		}
		custom.MinTemp = currentMode.MinTemp + data.Delta
		custom.MaxTemp = currentMode.MaxTemp + data.Delta
		custom.Correction = 1
		if err := custom.Update(ctx); err != nil {
			return request.NewResponse(http.StatusInternalServerError, err.Error())
		}
	}

	if current.Priority == setting.CUSTOM {
		if err := current.Delete(ctx); err != nil {
			return request.NewResponse(http.StatusInternalServerError, err.Error())
		}
	}
	settings, err := setting.All(ctx, z.ID())
	if err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	}

	now := time.Now()
	next := nextConfigChange(now, settings, z.Setting().Priority)
	if next.IsZero() {
		next = now.Add(time.Hour * 12)
	}
	if s, err := setting.New(ctx, z.ID(), custom.ID, setting.CUSTOM, current.DayOfWeek, now.Add(-time.Second), next, 0, 86400); err != nil {
		return request.NewResponse(http.StatusInternalServerError, err.Error())
	} else {
		settings = append(settings, s)
	}

	z.Update(settings)

	return request.NewResponse(http.StatusOK, `{}`)
}

func nextConfigChange(now time.Time, settings []setting.Setting, current setting.Priority) time.Time {
	type sched struct {
		runtime time.Time
		setting setting.Setting
	}

	starts := make([]sched, len(settings))
	for i, s := range settings {
		starts[i] = sched{
			runtime: s.Runtime(now),
			setting: s,
		}
	}

	sort.Slice(starts, func(i, j int) bool {
		return starts[i].runtime.Before(starts[j].runtime)
	})

	for _, s := range starts {
		if s.setting.Priority >= current && !s.runtime.IsZero() && s.runtime.After(now) {
			return s.runtime
		}
	}

	return time.Time{}
}
