package interpreter

import (
	"math"
	"time"
)

const TOLERANCE_X = 6
const TOLERANCE_Y = 6

const SCHEDULE_ITEM_WIDTH = 84

type ScheduleEntry struct {
	Code    string
	Start   time.Time
	End     time.Time
	SourceX int
	SourceY int
}

type ScheduleEntries struct {
	Entries []ScheduleEntry
}

func NewScheduleEntries() ScheduleEntries {
	return ScheduleEntries{
		Entries: []ScheduleEntry{},
	}
}

func (se *ScheduleEntries) AddEntry(scheduleType ScheduleType, x int, y int) bool {
	// basic sanity checks
	if scheduleType.Code == "" {
		return false
	}

	// fuzzy check for already existing entry
	for _, entry := range se.Entries {
		if entry.Code != scheduleType.Code {
			continue
		}

		if (x >= entry.SourceX-TOLERANCE_X && x <= entry.SourceX+TOLERANCE_X) &&
			(y >= entry.SourceY-TOLERANCE_Y && y <= entry.SourceY+TOLERANCE_Y) {
			return false
		}
	}

	newEntry := ScheduleEntry{
		Code:    scheduleType.Code,
		Start:   scheduleType.StartTime,
		End:     scheduleType.EndTime,
		SourceX: x,
		SourceY: y,
	}
	se.Entries = append(se.Entries, newEntry)
	return true
}

func (se *ScheduleEntries) SetCorrectDates(startDate time.Time) {
	for _, entry := range se.Entries {
		offsetX := entry.SourceX - SCHEDULE_PADDING
		day := math.Ceil(float64(entry.SourceX) / SCHEDULE_ITEM_WIDTH)
		if offsetX < SCHEDULE_ITEM_WIDTH {
			day = 1
		}
		entry.Start = time.Date(
			startDate.Year(), startDate.Month(), int(day),
			entry.Start.Hour(), entry.Start.Minute(), 0, 0,
			startDate.Location())
	}
}
