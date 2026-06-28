package telegram

import (
	"log"
	"time"
	_ "time/tzdata"
)

const moscowTimezone = "Europe/Moscow"

var moscowLocation = loadMoscowLocation()

func loadMoscowLocation() *time.Location {
	location, err := time.LoadLocation(moscowTimezone)
	if err != nil {
		log.Printf("can't load %s timezone, using fixed UTC+3: %v", moscowTimezone, err)
		return time.FixedZone("MSK", 3*60*60)
	}
	return location
}

func nowInMoscow() time.Time {
	return time.Now().In(moscowLocation)
}
