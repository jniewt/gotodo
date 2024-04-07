// Package filter provides a way to view tasks from different lists and filter them by different criteria.
//
// It creates a tree of filters that are evaluated for a task and return true if all were successful. Since Evaluate()
// is a recursive function, only the root node must be stored directly.
//
// NewFilter creates a new root node from given nodes by combining them with a logical operator AND.
package filter

import (
	"fmt"
	"strconv"
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
	compOp, err := NewComparisonOperator("done", string(OpEq), "false")
	if err != nil {
		panic(err)
	}
	return compOp
}

// PendingOrDoneToday creates a filter that accepts undone tasks, and done tasks if they were done today.
func PendingOrDoneToday() Node {
	doneToday, err := NewComparisonOperator("done_on", string(OpOnDay), "0")
	if err != nil {
		panic(err)
	}
	return &LogicalOperator{
		Operator: OpOr,
		Children: []Node{
			Pending(),
			doneToday,
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

func Overdue() Node {
	overdue, err := NewComparisonOperator("due_on", string(OpLt), "today")
	if err != nil {
		panic(err)
	}
	return overdue
}

// NoDueDate creates a filter that checks if a task has no due date.
func NoDueDate() Node {
	noDueBy, err := NewComparisonOperator("due_by", string(OpUnset), "")
	if err != nil {
		panic(err)
	}
	noDueOn, err := NewComparisonOperator("due_on", string(OpUnset), "")
	if err != nil {
		panic(err)
	}
	return &LogicalOperator{
		Operator: OpAnd,
		Children: []Node{noDueBy, noDueOn},
	}
}

// DueByInDays creates a filter that checks if a task is due within n days from now. Overdue tasks are included.
func DueByInDays(n int) Node {
	if n < 0 {
		panic(fmt.Sprint("n must be non-negative, got ", n))
	}
	compOp, err := NewComparisonOperator("due_by", string(OpNextDays), strconv.Itoa(n))
	if err != nil {
		panic(err)
	}
	return compOp
}

// DueOnToday creates a filter that checks if a task is due today. Overdue tasks are included.
func DueOnToday() Node {
	dueToday, err := NewComparisonOperator("due_on", string(OpOnDay), "today")
	if err != nil {
		panic(err)
	}
	return &LogicalOperator{
		Operator: OpOr,
		Children: []Node{
			Overdue(),
			dueToday,
		},
	}
}

// Node is a node in a filter tree. It can be either a logical operator or a comparison operator. The root node is always
// a logical operator. The tree is evaluated recursively.
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
	compareTo        FieldComparisonFunc
	Field, Op, Value string
}

func NewComparisonOperator(field, op, value string) (*ComparisonOperator, error) {
	compareTo, err := newComparison(field, op, value)
	if err != nil {
		return nil, err
	}
	return &ComparisonOperator{
		compareTo: compareTo,
		Field:     field,
		Op:        op,
		Value:     value,
	}, nil
}

func (op *ComparisonOperator) Evaluate(task core.Task) bool {
	if op.compareTo == nil {
		panic(fmt.Sprint("comparison function not set for ", op.Field, op.Op, op.Value))
	}
	return op.compareTo(task)
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

func newComparison(field, op, value string) (FieldComparisonFunc, error) {
	switch field {
	case "done":
		return parseDone(op, value)
	case "due_by":
		return parseDueBy(op, value)
	case "due_on":
		return parseDueOn(op, value)
	case "done_on":
		return parseDoneOn(op, value)
	default:
		return nil, fmt.Errorf("unsupported field for filter: %s", field)
	}
}

func parseDone(op, value string) (FieldComparisonFunc, error) {

	if op != string(OpEq) {
		return nil, fmt.Errorf("unsupported operator for done: %s", op)
	}

	var done bool
	switch value {
	case "true":
		done = true
	case "false":
		done = false
	default:
		return nil, fmt.Errorf("unknown value for done: %s", op)
	}

	return func(task core.Task) bool {
		return task.Done == done
	}, nil
}

func parseDueBy(op, value string) (FieldComparisonFunc, error) {

	if op == string(OpNextDays) {
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value for due_by: %s", value)
		}
		return func(task core.Task) bool {
			return isWithinDays(task.DueBy, n, time.Now())
		}, nil
	}

	if op == string(OpUnset) {
		return func(task core.Task) bool {
			return !task.HasDueBy()
		}, nil
	}

	date, err := ParseDate(value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for due_by: %s", value)
	}

	switch op {
	case string(OpOnDay):
		return func(task core.Task) bool {
			return isOnDay(task.DueOn, 0, date)
		}, nil
	case string(OpEq):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DueBy.Equal(date)
		}, nil
	case string(OpNeq):
		return func(task core.Task) bool {
			return !task.HasDueBy() || !task.DueBy.Equal(date)
		}, nil
	case string(OpGt):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DueBy.After(date)
		}, nil
	case string(OpGte):
		return func(task core.Task) bool {
			return task.HasDueBy() && !task.DueBy.Before(date)
		}, nil
	case string(OpLt):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DueBy.Before(date)
		}, nil
	case string(OpLte):
		return func(task core.Task) bool {
			return task.HasDueBy() && !task.DueBy.After(date)
		}, nil
	default:
		return nil, fmt.Errorf("unknown operator for due_by: %s", op)
	}
}

func parseDueOn(op, value string) (FieldComparisonFunc, error) {

	if op == string(OpNextDays) {
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value for due_on: %s", value)
		}
		return func(task core.Task) bool {
			return isWithinDays(task.DueOn, n, time.Now())
		}, nil
	}

	if op == string(OpUnset) {
		return func(task core.Task) bool {
			return !task.HasDueOn()
		}, nil
	}

	date, err := ParseDate(value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for due_on: %s", value)
	}

	switch op {
	case string(OpOnDay):
		return func(task core.Task) bool {
			return isOnDay(task.DueOn, 0, date)
		}, nil
	case string(OpEq):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DueOn.Equal(date)
		}, nil
	case string(OpNeq):
		return func(task core.Task) bool {
			return !task.HasDueBy() || !task.DueOn.Equal(date)
		}, nil
	case string(OpGt):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DueOn.After(date)
		}, nil
	case string(OpGte):
		return func(task core.Task) bool {
			return task.HasDueBy() && !task.DueOn.Before(date)
		}, nil
	case string(OpLt):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DueOn.Before(date)
		}, nil
	case string(OpLte):
		return func(task core.Task) bool {
			return task.HasDueBy() && !task.DueOn.After(date)
		}, nil
	default:
		return nil, fmt.Errorf("unknown operator for due_on: %s", op)
	}
}

func parseDoneOn(op, value string) (FieldComparisonFunc, error) {

	if op == string(OpNextDays) {
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value for done_on: %s", value)
		}
		return func(task core.Task) bool {
			return isWithinDays(task.DoneOn, n, time.Now())
		}, nil
	}

	if op == string(OpOnDay) {
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value for done_on: %s", value)
		}
		return func(task core.Task) bool {
			return isOnDay(task.DoneOn, n, time.Now())
		}, nil
	}

	date, err := ParseDate(value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for done_on: %s", value)
	}

	switch op {
	case string(OpEq):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DoneOn.Equal(date)
		}, nil
	case string(OpNeq):
		return func(task core.Task) bool {
			return !task.HasDueBy() || !task.DoneOn.Equal(date)
		}, nil
	case string(OpGt):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DoneOn.After(date)
		}, nil
	case string(OpGte):
		return func(task core.Task) bool {
			return task.HasDueBy() && !task.DoneOn.Before(date)
		}, nil
	case string(OpLt):
		return func(task core.Task) bool {
			return task.HasDueBy() && task.DoneOn.Before(date)
		}, nil
	case string(OpLte):
		return func(task core.Task) bool {
			return task.HasDueBy() && !task.DoneOn.After(date)
		}, nil
	default:
		return nil, fmt.Errorf("unknown operator for done_on: %s", op)
	}
}

type comparisonOp string

const (
	OpEq       comparisonOp = "=="
	OpNeq      comparisonOp = "!="
	OpGt       comparisonOp = ">"
	OpGte      comparisonOp = ">="
	OpLt       comparisonOp = "<"
	OpLte      comparisonOp = "<="
	OpNextDays comparisonOp = "next_days" // checks if the date is within n days from TODAY
	OpOnDay    comparisonOp = "on_day"    // compares any date to a specific day
	OpUnset    comparisonOp = "unset"
)

// ParseDate parses a date in the format "2006-01-02" or "2006-01-02T15:04" and returns a time.Time. String "today" is
// also accepted as a date.
func ParseDate(date string) (time.Time, error) {
	// if date is today return now at midnight
	if date == "today" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
	}
	// most specific format first
	formats := []string{"2006-01-02T15:04", "2006-01-02"}
	for _, format := range formats {
		t, err := time.Parse(format, date)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", date)
}
