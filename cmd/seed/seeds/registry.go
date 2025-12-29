package seeds

import (
	"time"
	"youdoyou-server/model"
)

type SeedData struct {
	Thread   model.ChatThread
	Messages []model.ChatMessage
}

var Registry = make(map[string]SeedData)

var JST = time.FixedZone("JST", 9*60*60)

func Register(name string, data SeedData) {
	Registry[name] = data
}

func JSTTime(year, month, day, hour, min int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, JST)
}
