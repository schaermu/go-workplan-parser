package interpreter

import (
	"fmt"
	"sort"
	"time"
)

const TOLERANCE_X = 8
const TOLERANCE_Y = 8

// TODO: we have to figure out a better way to determine the correct day...
const SCHEDULE_ITEM_WIDTH = 87

type ScheduleEntry struct {
	Code          string
	Description   string
	IsAllDayEvent bool
	Start         time.Time
	End           time.Time
	SourceX       int
	SourceY       int
}

func (s *ScheduleEntry) GetWorktime() string {
	return fmt.Sprintf("%02d.%02d.%d (%02d:%02d - %02d:%02d)", s.Start.Day(), s.Start.Month(), s.Start.Year(), s.Start.Hour(), s.Start.Minute(), s.End.Hour(), s.End.Minute())
}

type ScheduleEntries struct {
	Month   time.Time
	Entries []ScheduleEntry
}

func NewScheduleEntries(startDate time.Time) ScheduleEntries {
	return ScheduleEntries{
		Entries: []ScheduleEntry{},
		Month:   startDate,
	}
}

func (se *ScheduleEntries) AddEntry(scheduleType ScheduleType, date time.Time, x int, y int) bool {
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

	// determine day based on x offset
	day := (x-SCHEDULE_PADDING_X)/SCHEDULE_ITEM_WIDTH + 1

	// make sure overnight shifts are properly reflected
	endDayOffset := 0
	if scheduleType.StartTime.Hour() > scheduleType.EndTime.Hour() {
		endDayOffset = 1
	}

	newEntry := ScheduleEntry{
		Code:          scheduleType.Code,
		Description:   scheduleType.Description,
		IsAllDayEvent: scheduleType.IsAllDay,
		Start:         time.Date(date.Year(), date.Month(), day, scheduleType.StartTime.Hour(), scheduleType.StartTime.Minute(), 0, 0, date.Location()),
		End:           time.Date(date.Year(), date.Month(), day+endDayOffset, scheduleType.EndTime.Hour(), scheduleType.EndTime.Minute(), 0, 0, date.Location()),
		SourceX:       x,
		SourceY:       y,
	}
	se.Entries = append(se.Entries, newEntry)
	return true
}

func (se *ScheduleEntries) SortEntriesByDate() {
	sort.Slice(se.Entries, func(i, j int) bool {
		return se.Entries[i].Start.Before(se.Entries[j].Start)
	})
}

func (se *ScheduleEntries) RemoveDuplicates() {
	keys := make(map[time.Time]bool)
	newEntries := []ScheduleEntry{}
	for _, entry := range se.Entries {
		if _, ok := keys[entry.Start]; !ok {
			keys[entry.Start] = true
			newEntries = append(newEntries, entry)
		}
	}
	se.Entries = newEntries
}
