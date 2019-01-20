package log

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

var logger *csv.Writer

func Setup(file string) error {
	f, err := os.OpenFile(file, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 660)
	if err != nil {
		return err
	}
	logger = csv.NewWriter(f)

	return nil
}

func Log(fan, ac, heat bool, temp, hum float64) {
	if logger == nil {
		return
	}

	data := []string{
		time.Now().Format(time.Stamp),
		strconv.FormatBool(fan),
		strconv.FormatBool(ac),
		strconv.FormatBool(heat),
		strconv.FormatFloat(temp, 'f', 6, 64),
		strconv.FormatFloat(hum, 'f', 6, 64),
	}

	_ = logger.Write(data)
	logger.Flush()
}
