package utils

import "time"

func FloorTimeToHours(t time.Time) int64 {
	return (t.Unix() / 3600) * 3600
}
