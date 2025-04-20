package utils

import (
	"fmt"
	"strings"
	"time"
)

var monthMap = map[string]string{
	"Jan": "01",
	"Feb": "02",
	"Mar": "03",
	"Apr": "04",
	"May": "05",
	"Jun": "06",
	"Jul": "07",
	"Aug": "08",
	"Sep": "09",
	"Oct": "10",
	"Nov": "11",
	"Dec": "12",
}

// Steam timestamp format: Jan 02 2006 15: +0
func ParseSteamTimestamp(timestamp string) (time.Time, error) {
	res := strings.Split(strings.Replace(timestamp, ": +0", "", 1), " ")
	if len(res) != 4 {
		return time.Now(), fmt.Errorf("fail to parse steam timestamp: %s", timestamp)
	}
	month, ok := monthMap[res[0]]
	if !ok {
		return time.Now(), fmt.Errorf("fail to parse steam timestamp: %s", timestamp)
	}
	day := res[1]
	year := res[2]
	hour := res[3]
	timestamp = year + "-" + month + "-" + day + "T" + hour + ":00:00Z"
	return time.Parse("2006-01-02T15:04:05Z", timestamp)
}

// Calculate delta day between two timestamp, t2-t1
func DeltaDay(t1 time.Time, t2 time.Time) float64 {
	detlaSec := t2.Unix() - t1.Unix()
	res := float64(detlaSec) / 24 / 60 / 60
	return res
}
