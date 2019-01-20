package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"thermostat/api/request"
	"thermostat/db"
	"thermostat/system"
	"time"
)

func status(_ context.Context, _ json.RawMessage) request.ApiResponse {
	var data struct {
		ScheduleID   int64
		ModeID int64
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

	config := system.Configuration()
	setting, err := db.GetSetting(config.SettingID())
	if err != nil {
		return request.ApiResponse{http.StatusInternalServerError, err.Error()}
	}
	sensor := system.Sensors()[0]
	sys := system.Systems()[0]

	data.ScheduleID = config.ID()
	data.ModeID = config.SettingID()
	data.Temperature = sensor.Temperature()
	data.Humidity = sensor.Humidity()
	data.HeatIndex = system.HeatIndex(data.Temperature, data.Humidity)
	data.Min = setting.MinTemp()
	data.Max = setting.MaxTemp()
	data.Correction = setting.Correction()
	data.Heat = sys.Heat()
	data.AC = sys.AC()
	data.Fan = sys.Fan()

	msg, err := json.Marshal(data)
	if err != nil {
		return request.ApiResponse{http.StatusInternalServerError, err.Error()}
	}

	return request.ApiResponse{http.StatusOK, string(msg)}
}

func schedules(_ context.Context, _ json.RawMessage) request.ApiResponse {
	settings, err := db.Settings()
	if err != nil {
		return request.ApiResponse{http.StatusInternalServerError, err.Error()}
	}

	msg, err := json.Marshal(settings)
	if err != nil {
		return request.ApiResponse{http.StatusInternalServerError, err.Error()}
	}

	return request.ApiResponse{http.StatusOK, string(msg)}
}

func modes(_ context.Context, _ json.RawMessage) request.ApiResponse {
	modes, err := db.Modes()
	if err != nil {
		return request.ApiResponse{http.StatusInternalServerError, err.Error()}
	}

	msg, err := json.Marshal(modes)
	if err != nil {
		return request.ApiResponse{http.StatusInternalServerError, err.Error()}
	}

	return request.ApiResponse{http.StatusOK, string(msg)}
}

func addSchedule(update chan<- []db.Schedule) handler {
	return func(_ context.Context, msg json.RawMessage) request.ApiResponse {
		var data struct {
			ModeID    int64
			Priority  db.Priority
			DayOfWeek int
			Start     int64
			End       int64
			StartTime int
			EndTime   int
		}

		if err := json.Unmarshal(msg, &data); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}

		start := time.Unix(data.Start, 0)
		end := time.Unix(data.End, 0)

		settings, err := db.Settings()
		if err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		setting, err := db.AddSetting(data.ModeID, data.Priority, data.DayOfWeek, start, end, data.StartTime, data.EndTime)
		if err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		update <- append(settings, setting)

		resp, err := json.Marshal(setting)
		if err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		return request.ApiResponse{http.StatusOK, string(resp)}
	}
}

func deleteSchedule(update chan<- []db.Schedule) handler {
	return func(_ context.Context, msg json.RawMessage) request.ApiResponse {
		var data struct {
			ID int64
		}

		if err := json.Unmarshal(msg, &data); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}

		if err := db.DeleteSetting(data.ID); err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		if err := updateSettings(update); err != nil {
			return request.ApiResponse{http.StatusAccepted, err.Error()}
		}

		return request.ApiResponse{http.StatusOK, `{}`}
	}
}

type modeData struct {
	ID         int64
	Name       string
	Min        float64
	Max        float64
	Correction float64
}

func (data modeData) validate() error {
	if len(data.Name) < 2 {
		return errors.New("name must be at least 2 characters long")
	}
	if data.Name == "custom" {
		return errors.New("\"custom\" is a reserved mode name")
	}
	if data.Name == "default" {
		return errors.New("\"default\" is a reserved mode name")
	}
	if data.Max - data.Min < 2 {
		return errors.New("max temperature must be at least 2 degrees higher than the min temperature")
	}
	if data.Correction < 0.5 {
		return errors.New("correction must be >= 0.5")
	}
	if data.Correction*2 > data.Max - data.Min {
		return errors.New("correction cannot be more than half the difference in temperature range")
	}

	return nil
}

func addMode(update chan<- []db.Schedule) handler {
	return func(_ context.Context, msg json.RawMessage) request.ApiResponse {
		var data modeData
		if err := json.Unmarshal(msg, &data); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}
		if err := data.validate(); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}

		id, err := db.AddMode(data.Name, data.Min, data.Max, data.Correction)
		if err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		if err := updateSettings(update); err != nil {
			return request.ApiResponse{http.StatusAccepted, err.Error()}
		}

		return request.ApiResponse{http.StatusOK, fmt.Sprintf(`{"id": %d}`, id)}
	}
}

func editMode(update chan<- []db.Schedule) handler {
	return func(_ context.Context, msg json.RawMessage) request.ApiResponse {
		var data modeData
		if err := json.Unmarshal(msg, &data); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}
		if err := data.validate(); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}

		if err := db.EditMode(data.ID, data.Name, data.Min, data.Max, data.Correction); err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		if err := updateSettings(update); err != nil {
			return request.ApiResponse{http.StatusAccepted, err.Error()}
		}

		return request.ApiResponse{}
	}
}

func deleteMode(update chan<- []db.Schedule) handler {
	return func(_ context.Context, msg json.RawMessage) request.ApiResponse {
		var data struct {
			ID int64
		}

		if err := json.Unmarshal(msg, &data); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}

		if err := db.DeleteMode(data.ID); err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		if err := updateSettings(update); err != nil {
			return request.ApiResponse{http.StatusAccepted, err.Error()}
		}

		return request.ApiResponse{http.StatusOK, `{}`}
	}
}

func editHandler(update chan<- []db.Schedule) handler {
	return func(_ context.Context, msg json.RawMessage) request.ApiResponse {
		var data struct {
			modeID int64
			delta float64
		}

		if err := json.Unmarshal(msg, &data); err != nil {
			return request.ApiResponse{http.StatusBadRequest, err.Error()}
		}

		current := system.Configuration()
		setting, err := db.GetSetting(current.SettingID())
		if err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		modeID := data.modeID
		if modeID == 0 {
			modeID, err = db.CustomMode()
			if err != nil {
				return request.ApiResponse{http.StatusNotImplemented, err.Error()}
			}
			if err := db.EditMode(modeID, "custom", setting.MinTemp()+data.delta, setting.MaxTemp()+data.delta, setting.Correction()); err != nil {
				return request.ApiResponse{http.StatusInternalServerError, err.Error()}
			}
		}

		next := system.NextConfigChange()
		if next.IsZero() {
			next = time.Now().Add(time.Hour*12)
		}
		if _, err := db.AddSetting(modeID, db.OVERRIDE, current.DayOfWeek(), time.Now(), next, 0, 86400); err != nil {
			return request.ApiResponse{http.StatusInternalServerError, err.Error()}
		}

		if err := updateSettings(update); err != nil {
			return request.ApiResponse{http.StatusAccepted, err.Error()}
		}

		return request.ApiResponse{http.StatusOK, `{}`}
	}
}

func updateSettings(update chan<- []db.Schedule) error {
	settings, err := db.Settings()
	if err != nil {
		logrus.WithError(err).Warn("error loading settings")
		return err
	}

	update <- settings
	return nil
}
