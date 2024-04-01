// Package filter provides a way to view tasks from different lists and filter them by different criteria.
//
// It creates a tree of filters that are evaluated for a task and return true if all were successful. Since Evaluate()
// is a recursive function, only the root node must be stored directly.
//
// NewFilter creates a new root node from given nodes by combining them with a logical operator AND.
package filter

import (
	"fmt"
	"time"

	"github.com/jniewt/gotodo/internal/core"
)

type List struct {
	Name   string
	Filter Node
}

func NewFilter(nodes ...Node) Node {
	return &LogicalOperator{
		Operator: OpAnd,
		Children: nodes,
	}
}

// Pending creates a filter that checks if a task is not yet done.
func Pending() Node {
	return &ComparisonOperator{
		CompareTo: func(task core.Task) bool {
			return !task.Done
		},
	}
}

// PendingOrDoneToday creates a filter that accepts undone tasks, and done tasks if they were done today.
func PendingOrDoneToday() Node {
	return &LogicalOperator{
		Operator: OpOr,
		Children: []Node{
			Pending(),
			&ComparisonOperator{
				CompareTo: func(task core.Task) bool {
					return task.Done && isOnDay(task.DoneOn, 0, time.Now())
				},
			},
		},
	}
}

// Due combines multiple due date filters with a logical OR operator.
func Due(nodes ...Node) Node {
	return &LogicalOperator{
		Operator: OpOr,
		Children: nodes,
	}
}

// NoDueDate creates a filter that checks if a task has no due date.
func NoDueDate() Node {
	return &ComparisonOperator{
		CompareTo: func(task core.Task) bool {
			return !task.HasDueDate()
		},
	}
}

// DueByInDays creates a filter that checks if a task is due within n days from now. Overdue tasks are included.
func DueByInDays(n int) Node {
	if n < 0 {
		panic(fmt.Sprint("n must be non-negative, got ", n))
	}
	return &ComparisonOperator{
		CompareTo: func(task core.Task) bool {
			return isWithinDays(task.DueBy, n, time.Now())
		},
	}
}

// DueOnToday creates a filter that checks if a task is due today. Overdue tasks are included.
func DueOnToday() Node {
	return &ComparisonOperator{
		CompareTo: func(task core.Task) bool {
			return isOnDay(task.DueOn, 0, time.Now()) || task.IsOverdueOn()
		},
	}
}

type Node interface {
	Evaluate(task core.Task) bool
}

// LogicalOperator is a node that combines its children with a logical operator (AND, OR).
type LogicalOperator struct {
	Operator logicalOp
	Children []Node
}

func (op *LogicalOperator) Evaluate(task core.Task) bool {
	switch op.Operator {
	case OpAnd:
		for _, child := range op.Children {
			if !child.Evaluate(task) {
				return false
			}
		}
		return true
	case OpOr:
		for _, child := range op.Children {
			if child.Evaluate(task) {
				return true
			}
		}
		return false
	default:
		panic(fmt.Sprint("unknown logical operator: ", op.Operator))
	}
}

// FieldComparisonFunc is a function that compares the field of a task to a value.
type FieldComparisonFunc func(core.Task) bool

type ComparisonOperator struct {
	CompareTo FieldComparisonFunc
}

func (op *ComparisonOperator) Evaluate(task core.Task) bool {
	return op.CompareTo(task)
}

// checks if a date is on a given day relative to today
func isOnDay(date time.Time, daysFromNow int, now time.Time) bool {
	return date.Year() == now.Year() &&
		date.YearDay() == now.YearDay()+daysFromNow
}

// checks if a date is within n days from now, considering only full days
func isWithinDays(date time.Time, daysFromNow int, now time.Time) bool {
	return date.Year() == now.Year() &&
		date.YearDay() <= now.YearDay()+daysFromNow
}

type logicalOp string

const (
	OpAnd logicalOp = "AND"
	OpOr  logicalOp = "OR"
)
