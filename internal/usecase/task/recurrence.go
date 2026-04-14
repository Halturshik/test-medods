package task

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func GenerateInstances(template *taskdomain.Task, from time.Time, now time.Time, horizonDays int) []*taskdomain.Task {
	rc := template.RepeatConfig
	if rc == nil || rc.Type == taskdomain.RepeatNone {
		return nil
	}

	var instances []*taskdomain.Task

	cursor := truncateDay(from)

	if template.GeneratedUntil != nil {
		next := truncateDay(*template.GeneratedUntil).AddDate(0, 0, 1)
		if next.After(cursor) {
			cursor = next
		}
	}

	for i := 0; i < horizonDays; i++ {
		current := cursor.AddDate(0, 0, i)

		if shouldStop(rc, current) {
			break
		}

		if !matches(rc, current, truncateDay(template.DueDate)) {
			continue
		}

		due := combineDateTime(current, template.DueDate)

		instances = append(instances, &taskdomain.Task{
			ParentID:    &template.ID,
			Title:       template.Title,
			Description: template.Description,
			Status:      taskdomain.StatusNew,
			DueDate:     due,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return instances
}

func shouldStop(rc *taskdomain.RepeatConfig, current time.Time) bool {
	if rc.EndType == taskdomain.EndNever {
		return false
	}

	if rc.EndType == taskdomain.EndByDate && rc.EndDate != nil {
		if current.After(truncateDay(*rc.EndDate)) {
			return true
		}
	}

	return false
}

func matches(rc *taskdomain.RepeatConfig, date, anchor time.Time) bool {
	switch rc.Type {
	case taskdomain.RepeatDaily:
		return true
	case taskdomain.RepeatEveryN:
		if rc.Interval <= 0 {
			return false
		}
		days := int(date.Sub(anchor).Hours() / 24)
		return days%rc.Interval == 0
	case taskdomain.RepeatEven:
		return date.Day()%2 == 0
	case taskdomain.RepeatOdd:
		return date.Day()%2 == 1
	case taskdomain.RepeatMonthly:
		for _, d := range rc.Days {
			if date.Day() == d {
				return true
			}
		}
	}
	return false
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func combineDateTime(date, src time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(),
		src.Hour(), src.Minute(), src.Second(), 0, time.UTC)
}
