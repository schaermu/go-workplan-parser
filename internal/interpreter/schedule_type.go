package interpreter

import (
	"log"
	"path"
	"time"
)

type ScheduleType struct {
	Code          string
	TemplateImage string
	StartTime     time.Time
	EndTime       time.Time
	IsWholeDay    bool
}

/*
schedule_types = [
	ScheduleType('1', '07:45', '16:45'),
	ScheduleType('1_blau', '', ''),
	ScheduleType('a1_3', '', ''),
	ScheduleType('brille', '', ''),
	ScheduleType('buero', '', ''),
	ScheduleType('buero_p12', '', ''),
	ScheduleType('ferien', '', ''),
	ScheduleType('frei', '', ''),
	ScheduleType('k', '', ''),
	ScheduleType('k1', '', ''),
	ScheduleType('k2', '', ''),
	ScheduleType('k3', '', ''),
	ScheduleType('nacht', '', ''),
	ScheduleType('p12', '', ''),
	ScheduleType('p24', '', ''),
	ScheduleType('t2', '', '')
]
*/

var scheduleTypes = []ScheduleType{
	createScheduleType("fruehdienst", "t", "07:00", "15:30"),
	createScheduleType("spaetdienst", "1_blau", "15:00", "22:30"),
	createScheduleType("nachtdienst", "nacht", "22:30", "07:00"),
	createScheduleType("hosp", "1_gelb", "07:45", "16:45"),
	createScheduleType("tagspaet", "t2", "09:00", "18:00"),
	createScheduleType("aufnahme", "a13", "07:45", "16:45"),
	createScheduleType("aufnahme2", "a2", "07:45", "16:45"),
	createScheduleType("ferien", "ferien", "", ""),
	createScheduleType("frei", "frei", "", ""),
	createScheduleType("pikket_12h", "pikett_12", "", ""),
	createScheduleType("pikett_24h", "pikett_24", "", ""),
}

func createScheduleType(code string, image string, start string, end string) ScheduleType {
	imagePath := path.Join("assets", "icons", image+".png")

	if start == "" && end == "" {
		return ScheduleType{
			Code:          code,
			TemplateImage: imagePath,
			IsWholeDay:    true,
		}
	}

	startTime, err := time.Parse("15:04", start)
	if err != nil {
		log.Fatal(err)
	}
	endTime, err := time.Parse("15:04", end)
	if err != nil {
		log.Fatal(err)
	}

	return ScheduleType{
		Code:          code,
		TemplateImage: imagePath,
		StartTime:     startTime,
		EndTime:       endTime,
	}
}

func GetScheduleTypes() []ScheduleType {
	return scheduleTypes
}
