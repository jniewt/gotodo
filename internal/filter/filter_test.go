package filter

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/jniewt/gotodo/internal/core"
)

// testCase defines the structure for each test case
type testCase struct {
	name    string
	field   string
	value   string
	task    core.Task
	wantErr bool
	want    bool
}

func TestNewComparison(t *testing.T) {
	tests := []testCase{
		{
			field: "list",
			value: "Shopping,Work",
			task:  core.Task{List: "Work"},
			want:  true,
		},
		{
			name:  "list any",
			field: "list",
			value: "",
			task:  core.Task{List: "Work"},
			want:  true,
		},
		{
			name:  "list not in",
			field: "list",
			value: "Shopping,Work",
			task:  core.Task{List: "Home"},
			want:  false,
		},
		{
			field: "done",
			value: "true",
			task:  core.Task{Done: true},
			want:  true,
		},
		{
			field: "done_on",
			value: "0",
			task:  core.Task{Done: true, DoneOn: time.Now()},
			want:  true,
		},
		{
			field:   "done_on",
			value:   "abc",
			wantErr: true,
		},
		{
			field: "due_by",
			value: "3",
			task:  core.Task{DueType: core.DueBy, Due: time.Now().Add(48 * time.Hour)},
			want:  true,
		},
		{
			name:    "due_by invalid",
			field:   "due_by",
			value:   "abc",
			task:    core.Task{DueType: core.DueBy, Due: time.Now().Add(48 * time.Hour)},
			wantErr: true,
		},
		{
			field:   "invalid",
			value:   "any",
			wantErr: true,
		},
		// Tests for due_on
		{
			field: "due_on",
			value: "0",
			task:  core.Task{DueType: core.DueOn, AllDay: true, Due: time.Now()},
			want:  true,
		},
		{
			name:  "due_on overdue today",
			field: "due_on",
			value: "0",
			task:  core.Task{DueType: core.DueOn, Due: time.Now()},
			want:  false,
		},
		{
			field: "due_on",
			value: "1",
			task:  core.Task{DueType: core.DueOn, Due: time.Now().Add(24 * time.Hour)},
			want:  true,
		},
		{
			name:  "due_on allday tomorrow, want today",
			field: "due_on",
			value: "0",
			task:  core.Task{DueType: core.DueOn, AllDay: true, Due: time.Now().Add(24 * time.Hour)},
			want:  false,
		},
		{
			field:   "due_on",
			value:   "-1",
			wantErr: true, // Negative values are not valid for due_on
		},
		{
			name:  "due_on due_by",
			field: "due_on",
			value: "0",
			task:  core.Task{DueType: core.DueBy, Due: time.Now()},
			want:  false,
		},

		// Tests for due_none
		{
			field: "due_none",
			value: "",
			task:  core.Task{DueType: core.DueNone},
			want:  true,
		},
		{
			field: "due_none",
			value: "anyvalue", // value is ignored for due_none
			task:  core.Task{DueType: core.DueNone},
			want:  true,
		},
		{
			field: "due_none",
			value: "",
			task:  core.Task{DueType: core.DueOn, Due: time.Now()},
			want:  false,
		},

		// Tests for overdue
		{
			field: "overdue",
			value: "true",
			task:  core.Task{DueType: core.DueBy, Due: time.Now().Add(-2 * time.Hour)},
			want:  true,
		},
		{
			field: "overdue",
			value: "true",
			task:  core.Task{DueType: core.DueOn, Due: time.Now().Add(-24 * time.Hour)},
			want:  true,
		},
		{
			field: "overdue",
			value: "true",
			task:  core.Task{DueType: core.DueNone},
			want:  false,
		},
	}

	for _, tc := range tests {
		name := tc.name
		if name == "" {
			name = fmt.Sprintf("%s_%s", tc.field, tc.value)
		}
		t.Run(fmt.Sprintf(name, tc.field, tc.value), func(t *testing.T) {
			compFunc, err := newComparison(tc.field, tc.value)
			if (err != nil) != tc.wantErr {
				t.Errorf("Expected error: %v, got %v", tc.wantErr, err)
			}
			if err == nil {
				got := compFunc(tc.task)
				if got != tc.want {
					t.Errorf("Comparing field %s with value %s: got %v, want %v", tc.field, tc.value, got, tc.want)
				}
			}
		})
	}
}

func TestFilter_Evaluate(t *testing.T) {
	type fieldValue struct {
		field string
		value string
	}
	type ruleSet struct {
		rules []fieldValue
	}

	tests := []struct {
		name   string
		filter []ruleSet
		task   core.Task
		want   bool
	}{
		{
			name: "0",
			filter: []ruleSet{
				{
					rules: []fieldValue{
						{field: "list", value: "Work"},
						{field: "done", value: "true"},
					},
				},
				{
					rules: []fieldValue{
						{field: "due_by", value: "0"},
					},
				},
			},
			task: core.Task{List: "Work", Done: true},
			want: true,
		},
		{
			name: "none of the rules match",
			filter: []ruleSet{
				{
					rules: []fieldValue{
						{field: "list", value: "Home"},
					},
				},
				{
					rules: []fieldValue{
						{field: "done", value: "false"},
					},
				},
			},
			task: core.Task{List: "Work", Done: true},
			want: false,
		},
		{
			name: "All rules match in one ruleset",
			filter: []ruleSet{
				{
					rules: []fieldValue{
						{field: "list", value: "Work"},
						{field: "done", value: "true"},
					},
				},
			},
			task: core.Task{List: "Work", Done: true},
			want: true,
		},
		{
			name: "Rules match in none of the rulesets",
			filter: []ruleSet{
				{
					rules: []fieldValue{
						{field: "list", value: "Home"},
					},
				},
				{
					rules: []fieldValue{
						{field: "done", value: "false"},
					},
				},
			},
			task: core.Task{List: "Work", Done: true},
			want: false,
		},
		{
			name: "Rules match in second ruleset",
			filter: []ruleSet{
				{
					rules: []fieldValue{
						{field: "list", value: "Home"},
					},
				},
				{
					rules: []fieldValue{
						{field: "done", value: "true"},
					},
				},
			},
			task: core.Task{List: "Work", Done: true},
			want: true,
		},
		{
			name: "Mixed results across rulesets",
			filter: []ruleSet{
				{
					rules: []fieldValue{
						{field: "list", value: "Home"},
					},
				},
				{
					rules: []fieldValue{
						{field: "list", value: "Work"},
						{field: "done", value: "false"},
					},
				},
			},
			task: core.Task{List: "Work", Done: true},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ruleSets []RuleSet
			for _, rs := range tc.filter {
				var rules []Rule
				for _, rv := range rs.rules {
					rule, err := NewRule(rv.field, rv.value)
					if err != nil {
						t.Fatalf("Failed to create rule for field '%s' with value '%s': %s", rv.field, rv.value, err)
					}
					rules = append(rules, rule)
				}
				ruleSets = append(ruleSets, RuleSet{Rules: rules})
			}

			filter := Filter{
				RuleSets: ruleSets,
			}

			got := filter.Evaluate(tc.task)
			if got != tc.want {
				t.Errorf("Filter.Evaluate() got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilter_UnmarshalYAML(t *testing.T) {
	overdue, _ := NewRule("overdue", "true")
	pending, _ := NewRule("done", "false")
	doneToday, _ := NewRule("done_on", "0")
	dueOnToday, _ := NewRule("due_on", "0")
	dueBy7Days, _ := NewRule("due_by", "7")
	noDueDate, _ := NewRule("due_none", "")
	filter := Filter{
		RuleSets: []RuleSet{
			{Rules: []Rule{pending, dueOnToday}},
			{Rules: []Rule{pending, dueBy7Days}},
			{Rules: []Rule{pending, noDueDate}},
			{Rules: []Rule{overdue}},
			{Rules: []Rule{doneToday}},
		},
	}
	list := List{
		Name:   "Soon",
		Filter: filter,
	}
	out, err := yaml.Marshal(list)
	if err != nil {
		t.Fatalf("Failed to marshal list: %s", err)
	}
	t.Log(string(out))
	in := List{}
	err = yaml.Unmarshal(out, &in)
	if err != nil {
		t.Fatalf("Failed to unmarshal list: %s", err)
	}

	t.Log(in)
}
