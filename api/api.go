package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jniewt/gotodo/internal/core"
)

type ListResponse struct {
	Name  string          `json:"name"`
	Items []*TaskResponse `json:"items"`
}

func FromList(l core.List) ListResponse {
	tasks := make([]*TaskResponse, len(l.Items))
	for i, task := range l.Items {
		t := FromTask(*task)
		tasks[i] = &t
	}
	return ListResponse{
		Name:  l.Name,
		Items: tasks,
	}
}

type TaskResponse struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	List    string    `json:"list"`
	Done    bool      `json:"done"`
	AllDay  bool      `json:"all_day"`
	DueBy   time.Time `json:"due_by,omitempty"`
	DueOn   time.Time `json:"due_on,omitempty"`
	Created time.Time `json:"created"`
	DoneOn  time.Time `json:"done_on,omitempty"`
}

func FromTask(t core.Task) TaskResponse {
	return TaskResponse{
		ID:      t.ID,
		Title:   t.Title,
		List:    t.List,
		Done:    t.Done,
		AllDay:  t.AllDay,
		DueBy:   t.DueBy,
		DueOn:   t.DueOn,
		Created: t.Created,
		DoneOn:  t.DoneOn,
	}
}

// MarshalJSON overwrites JSON marshalling to not send zero-value time fields
func (t TaskResponse) MarshalJSON() ([]byte, error) {
	var dueBy, dueOn, doneOn string
	if !t.DueBy.IsZero() {
		dueBy = t.DueBy.Format(time.RFC3339)
	}
	if !t.DueOn.IsZero() {
		dueOn = t.DueOn.Format(time.RFC3339)
	}
	if !t.DoneOn.IsZero() {
		doneOn = t.DoneOn.Format(time.RFC3339)
	}
	type Alias TaskResponse
	return json.Marshal(&struct {
		Alias
		DueBy  string `json:"due_by,omitempty"`
		DueOn  string `json:"due_on,omitempty"`
		DoneOn string `json:"done_on,omitempty"`
	}{
		Alias:  Alias(t),
		DueBy:  dueBy,
		DueOn:  dueOn,
		DoneOn: doneOn,
	})
}

type ListAdd struct {
	Name string `json:"name"`
}

type TaskAdd struct {
	Title  string    `json:"title"`
	AllDay bool      `json:"all_day"`
	DueBy  time.Time `json:"due_by"`
	DueOn  time.Time `json:"due_on"`
}

// UnmarshalJSON overwrites JSON unmarshalling to parse time fields properly
// TODO this is not fail-safe, it will fall apart if JS sends a different format
func (t *TaskAdd) UnmarshalJSON(data []byte) error {
	type Alias TaskAdd
	aux := struct {
		DueBy string `json:"due_by"`
		DueOn string `json:"due_on"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	format := "2006-01-02T15:04"
	if aux.AllDay {
		format = "2006-01-02"
	}
	if aux.DueBy != "" {
		dueBy, err := time.ParseInLocation(format, aux.DueBy, time.Local)
		if err != nil {
			return err
		}
		t.DueBy = dueBy
	}
	if aux.DueOn != "" {
		dueOn, err := time.ParseInLocation(format, aux.DueOn, time.Local)
		if err != nil {
			return err
		}
		t.DueOn = dueOn
	}
	return nil
}

// TaskChange is used to change a task. It is used in PATCH requests. Only fields that are set will be changed.
// DueType must be set to one of TypeDueOn, TypeDueBy or TypeDueNone in requests to change the due date, otherwise
// the due date supplied in the request will be ignored.
type TaskChange struct {
	Title  string `json:"title"`
	List   string `json:"list"`
	Done   bool   `json:"done"`
	AllDay bool   `json:"all_day"`

	// DueType must be set to one of TypeDueOn, TypeDueBy or TypeDueNone in requests to change the due date.
	DueType dueType   `json:"due_type"`
	DueBy   time.Time `json:"due_by,omitempty"`
	DueOn   time.Time `json:"due_on,omitempty"`
}

func (t *TaskChange) getDueType() dueType {
	if !t.DueBy.IsZero() {
		return TypeDueBy
	}
	if !t.DueOn.IsZero() {
		return TypeDueOn
	}
	return TypeDueNone
}

func (t *TaskChange) Validate() error {
	if t.Title == "" {
		return fmt.Errorf("missing task title")
	}
	if t.List == "" {
		return fmt.Errorf("missing list name")
	}

	if !t.DueBy.IsZero() && !t.DueOn.IsZero() {
		return fmt.Errorf("only one of dueOn or dueBy can be set")
	}

	return nil
}

// UnmarshalJSON overwrites JSON unmarshalling to parse time fields properly
// TODO this is not fail-safe, it will fall apart if JS sends a different format
func (t *TaskChange) UnmarshalJSON(data []byte) error {

	var input map[string]interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	if title, ok := input["title"]; ok {
		t.Title = title.(string)
	}

	if list, ok := input["list"]; ok {
		t.List = list.(string)
	}

	if done, ok := input["done"]; ok {
		t.Done = done.(bool)
	}

	if allDay, ok := input["all_day"]; ok {
		t.AllDay = allDay.(bool)
	}

	if dueTypRaw, ok := input["due_type"]; ok {
		if t.getDueType() != dueTypRaw.(dueType) {
			if err := t.overwriteDueFields(input); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *TaskChange) overwriteDueFields(input map[string]interface{}) error {

	dueTyp := input["due_type"].(dueType)

	format := "2006-01-02T15:04"
	if t.AllDay {
		format = "2006-01-02"
	}

	switch dueTyp {
	case TypeDueNone:
		t.DueBy = time.Time{}
		t.DueOn = time.Time{}
	case TypeDueOn:
		dueOnRaw, ok := input["due_on"]
		if !ok {
			return fmt.Errorf("missing due_on field")
		}
		dueOn, err := time.ParseInLocation(format, dueOnRaw.(string), time.Local)
		if err != nil {
			return err
		}
		t.DueOn = dueOn
		t.DueBy = time.Time{}
	case TypeDueBy:
		dueByRaw, ok := input["due_by"]
		if !ok {
			return fmt.Errorf("missing due_by field")
		}
		dueBy, err := time.ParseInLocation(format, dueByRaw.(string), time.Local)
		if err != nil {
			return err
		}
		t.DueBy = dueBy
		t.DueOn = time.Time{}
	default:
		return fmt.Errorf("invalid due_type")
	}

	return nil
}

type dueType string

const (
	TypeDueOn   dueType = "on"
	TypeDueBy   dueType = "by"
	TypeDueNone dueType = "none"
)
