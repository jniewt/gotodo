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
