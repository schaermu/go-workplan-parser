package interpreter

import (
	"log"
	"path"
	"time"
)

type ScheduleType struct {
	Code          string
	Description   string
	TemplateImage string
	StartTime     time.Time
	EndTime       time.Time
	IsAllDay      bool
}

var scheduleTypes = []ScheduleType{
	createScheduleType("fruehdienst", "Frühdienst", "t", "07:00", "15:30"),
	createScheduleType("spaetdienst", "Spätdienst", "1_blau", "15:00", "22:30"),
	createScheduleType("nachtdienst", "Nachtdienst", "nacht", "22:30", "07:00"),
	createScheduleType("hosp", "Hosp-Dienst", "1_gelb", "07:45", "16:45"),
	createScheduleType("tagspaet", "Tagdienst (spät)", "t2", "09:00", "18:00"),
	createScheduleType("aufnahme", "Aufnahme", "a13", "07:45", "16:45"),
	createScheduleType("aufnahme2", "Aufnahme", "a2", "07:45", "16:45"),
	createScheduleType("ferien", "Ferien", "ferien", "", ""),
	createScheduleType("frei", "Frei", "frei", "", ""),
	createScheduleType("pikket_12h", "Pikket", "pikett_12", "", ""),
	createScheduleType("pikett_24h", "Pikket", "pikett_24", "", ""),
}

func createScheduleType(code string, description string, image string, start string, end string) ScheduleType {
	imagePath := path.Join("assets", "icons", image+".png")

	if start == "" && end == "" {
		return ScheduleType{
			Code:          code,
			Description:   description,
			TemplateImage: imagePath,
			IsAllDay:      true,
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
		Description:   description,
		TemplateImage: imagePath,
		StartTime:     startTime,
		EndTime:       endTime,
	}
}

func GetScheduleTypes() []ScheduleType {
	return scheduleTypes
}
