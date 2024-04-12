package core

import (
	"testing"
	"time"
)

func TestHasDueDate(t *testing.T) {
	var tests = []struct {
		name string
		task Task
		want bool
	}{
		{"No due date", Task{DueType: DueNone}, false},
		{"Due on specific date", Task{DueType: DueOn}, true},
		{"Due by specific date", Task{DueType: DueBy}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.HasDueDate(); got != tt.want {
				t.Errorf("Task.HasDueDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsOverdue(t *testing.T) {
	now := time.Now()
	var tests = []struct {
		name string
		task Task
		want bool
	}{
		{"Overdue task", Task{DueType: DueOn, Due: now.Add(-24 * time.Hour)}, true},
		{"Task due today", Task{DueType: DueOn, Due: now, AllDay: true}, false},
		{"Future task", Task{DueType: DueOn, Due: now.Add(24 * time.Hour)}, false},
		{"Due by task overdue", Task{DueType: DueBy, Due: now.Add(-48 * time.Hour)}, true},
		{"Due by task not overdue", Task{DueType: DueBy, Due: now.Add(24 * time.Hour)}, false},
		{"No due date", Task{DueType: DueNone, Due: now.Add(-24 * time.Hour)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.IsOverdue(); got != tt.want {
				t.Errorf("Task.IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}
