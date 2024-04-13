package core

import (
	"time"
)

type List struct {
	Name   string
	Colour RGB
	Items  []*Task
}

type RGB struct {
	R uint8
	G uint8
	B uint8
}

type Task struct {
	ID       int
	Title    string
	List     string
	Done     bool
	Priority int
	AllDay   bool
	DueType  DueType
	Due      time.Time
	Created  time.Time
	DoneOn   time.Time
}

// IsOverdue returns true if the task is overdue. A task is overdue if it is not done and the due date is in the past.
func (t Task) IsOverdue() bool {
	if t.Done {
		return false
	}
	if !t.HasDueDate() {
		return false
	}
	now := time.Now()
	if t.AllDay {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return t.Due.Before(today)
	}
	return t.Due.Before(time.Now())
}

// HasDueDate returns true if the task has a due date set.
func (t Task) HasDueDate() bool {
	return t.DueType != DueNone
}

func (t Task) HasDueOnDate() bool {
	return t.DueType == DueOn
}

func (t Task) HasDueByDate() bool {
	return t.DueType == DueBy
}

func (t Task) IsDone() bool {
	return t.Done
}

type DueType string

const (
	DueOn   DueType = "due_on"
	DueBy   DueType = "due_by"
	DueNone DueType = ""
)

// Priorities
const (
	PrioLowest  = -2
	PrioLow     = -1
	PrioNormal  = 0
	PrioHigh    = 1
	PrioHighest = 2
)
