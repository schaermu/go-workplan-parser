package interpreter

import (
	"time"
)

const TOLERANCE_X = 6
const TOLERANCE_Y = 6

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
