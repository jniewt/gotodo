package core

import (
	"time"
)

type List struct {
	Name  string
	Items []*Task
}

type Task struct {
	ID      int
	Title   string
	List    string
	Done    bool
	AllDay  bool
	DueBy   time.Time
	DueOn   time.Time
	Created time.Time
	DoneOn  time.Time
}

// IsOverdueOn returns true if the task is overdue based on the "due on" date.
func (t Task) IsOverdueOn() bool {
	return t.HasDueOn() && t.DueOn.Before(time.Now())
}

// IsOverdueBy returns true if the task is overdue based on the "due by" date.
func (t Task) IsOverdueBy() bool {
	return t.HasDueBy() && t.DueBy.Before(time.Now())
}

// HasDueDate returns true if the task has "due on" or "due by" date set.
func (t Task) HasDueDate() bool {
	return t.HasDueOn() || t.HasDueBy()
}

// HasDueOn returns true if the task has a "due on" date set.
func (t Task) HasDueOn() bool {
	return !t.DueOn.IsZero()
}

// HasDueBy returns true if the task has a "due by" date set.
func (t Task) HasDueBy() bool {
	return !t.DueBy.IsZero()
}
