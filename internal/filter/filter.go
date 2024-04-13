package filter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jniewt/gotodo/internal/core"
)

type List struct {
	Name   string
	Filter Filter
}

// Filter is a collection of rule sets. If any rule set is true, the filter is true.
type Filter struct {
	RuleSets []RuleSet
}

func (f Filter) Evaluate(task core.Task) bool {
	for _, ruleSet := range f.RuleSets {
		if ruleSet.Evaluate(task) {
			return true
		}
	}
	return false
}

// RuleSet is a set of rules. All rules in the set must be true for the set to be true.
type RuleSet struct {
	Rules []Rule
}

func (s RuleSet) Evaluate(task core.Task) bool {
	for _, rule := range s.Rules {
		if !rule.Evaluate(task) {
			return false
		}
	}
	return true
}

// Rule is a filter rule that can be evaluated against a task.
type Rule struct {
	Field   string
	Value   interface{}
	compare comparisonFunc
}

func NewRule(field, value string) (Rule, error) {
	compare, err := newComparison(field, value)
	if err != nil {
		return Rule{}, err
	}
	return Rule{
		Field:   field,
		Value:   value,
		compare: compare,
	}, nil
}

func (r Rule) Evaluate(task core.Task) bool {
	return r.compare(task)
}

func (r *Rule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw struct {
		Field string
		Value string
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	rule, err := NewRule(raw.Field, raw.Value)
	if err != nil {
		return err
	}
	*r = rule
	return nil
}

type comparisonFunc func(core.Task) bool

func newComparison(field, value string) (comparisonFunc, error) {

	switch field {
	case "list": // value is comma separated string, empty string means all lists
		return newComparisonList(value)
	case "done": // value is boolean
		return func(task core.Task) bool {
			return task.Done == (value == "true")
		}, nil
	case "done_on": // value is days past today (0 or negative)
		return newComparisonDoneOn(value)
	case "due_by": // value is days from today, doesn't include overdue
		return newComparisonDue(core.DueBy, value)
	case "due_on": // value is days from today, doesn't include overdue
		return newComparisonDue(core.DueOn, value)
	case "due_none": // value is ignored
		return func(task core.Task) bool {
			return task.DueType == core.DueNone
		}, nil
	case "overdue": // value is boolean
		return func(task core.Task) bool {
			return task.IsOverdue() == (value == "true")
		}, nil
	}

	return nil, fmt.Errorf("invalid field: %s", field)
}

func newComparisonDoneOn(value string) (comparisonFunc, error) {
	// value is days from today, should only be 0 or negative
	days, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for done_on: %s", value)
	}
	if days > 0 {
		return nil, fmt.Errorf("value for done_on must be 0 or negative")
	}
	return func(task core.Task) bool {
		if !task.Done {
			return false
		}
		// compare only date
		now := truncateToDay(time.Now())
		doneOn := truncateToDay(task.DoneOn)
		limit := now.AddDate(0, 0, days)
		if doneOn.Equal(limit) {
			return true
		}
		// doneOn is always in the past
		return doneOn.After(limit)
	}, nil
}

func newComparisonDue(dueType core.DueType, value string) (comparisonFunc, error) {
	// value is days from today, should only be positive
	days, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for due: %s", value)
	}
	if days < 0 {
		return nil, fmt.Errorf("value for due_by must be positive")
	}
	return func(task core.Task) bool {
		if task.DueType != dueType {
			return false
		}
		if task.IsOverdue() {
			return false
		}
		// compare only date
		now := truncateToDay(time.Now())
		due := truncateToDay(task.Due)
		limit := now.AddDate(0, 0, days)
		if due.Equal(limit) {
			return true
		}
		// due (which is not overdue) is always in the future
		return due.Before(limit)
	}, nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func newComparisonList(value string) (comparisonFunc, error) {
	if value == "" {
		return func(_ core.Task) bool {
			return true
		}, nil
	}
	// split lists by the comma
	lists := strings.Split(value, ",")
	return func(task core.Task) bool {
		for _, list := range lists {
			if list == task.List {
				return true
			}
		}
		return false
	}, nil
}
