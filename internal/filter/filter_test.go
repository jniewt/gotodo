package filter

import (
	"testing"
	"time"

	"github.com/jniewt/gotodo/internal/core"
)

func TestPending(t *testing.T) {
	tests := []struct {
		name string
		task core.Task
		want bool
	}{
		{
			name: "pending",
			task: core.Task{Done: false},
			want: true,
		},
		{
			name: "done",
			task: core.Task{Done: true},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			filt := Pending()

			if got := filt.Evaluate(tt.task); got != tt.want {
				t.Errorf("Pending() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestNoDueDate(t *testing.T) {
	tests := []struct {
		name string
		task core.Task
		want bool
	}{
		{
			name: "no due date",
			task: core.Task{},
			want: true,
		},
		{
			name: "due by",
			task: core.Task{DueType: core.DueBy, Due: time.Now()},
			want: false,
		},
		{
			name: "due on",
			task: core.Task{DueType: core.DueOn, Due: time.Now()},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			filt := NoDueDate()

			if got := filt.Evaluate(tt.task); got != tt.want {
				t.Errorf("NoDueDate() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestDueByInDays(t *testing.T) {
	tests := []struct {
		name string
		task core.Task
		days int
		want bool
	}{
		{
			name: "due today",
			task: core.Task{DueType: core.DueBy, Due: time.Now()},
			days: 0,
			want: true,
		},
		{
			name: "due tomorrow",
			task: core.Task{DueType: core.DueBy, Due: time.Now().Add(24 * time.Hour)},
			days: 1,
			want: true,
		},
		{
			name: "due tomorrow want today",
			task: core.Task{DueType: core.DueBy, Due: time.Now().Add(24 * time.Hour)},
			days: 0,
			want: false,
		},
		{
			name: "due day after tomorrow, want 5 days",
			task: core.Task{DueType: core.DueBy, Due: time.Now().Add(48 * time.Hour)},
			days: 5,
			want: true,
		},
		{
			name: "overdue",
			task: core.Task{DueType: core.DueBy, Due: time.Now().Add(-24 * time.Hour)},
			days: 2,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			filt := DueByInDays(tt.days)

			if got := filt.Evaluate(tt.task); got != tt.want {
				t.Errorf("DueByInDays(%v) = %v, want %v", tt.days, got, tt.want)
			}

		})
	}
}

func TestDueOnToday(t *testing.T) {
	tests := []struct {
		name string
		task core.Task
		want bool
	}{
		{
			name: "due today",
			task: core.Task{DueType: core.DueOn, Due: time.Now()},
			want: true,
		},
		{
			name: "due today all day",
			task: core.Task{DueType: core.DueOn, AllDay: true, Due: time.Now()},
			want: true,
		},
		{
			name: "due tomorrow",
			task: core.Task{DueType: core.DueOn, Due: time.Now().Add(24 * time.Hour)},
			want: false,
		},
		{
			name: "no due date",
			task: core.Task{},
			want: false,
		},
		{
			name: "overdue",
			task: core.Task{DueType: core.DueOn, Due: time.Now().Add(-24 * time.Hour)},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			filt := DueOnToday()

			if got := filt.Evaluate(tt.task); got != tt.want {
				t.Errorf("DueOnToday() = %v, want %v", got, tt.want)
			}

		})
	}
}
