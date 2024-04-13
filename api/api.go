package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jniewt/gotodo/internal/core"
)

type ListResponse struct {
	Name   string          `json:"name"`
	Colour RGB             `json:"colour"`
	Items  []*TaskResponse `json:"items"`
}

func FromList(l core.List) ListResponse {
	tasks := make([]*TaskResponse, len(l.Items))
	for i, task := range l.Items {
		t := FromTask(*task)
		tasks[i] = &t
	}
	return ListResponse{
		Name: l.Name,
		Colour: RGB{
			R: l.Colour.R,
			G: l.Colour.G,
			B: l.Colour.B,
		},
		Items: tasks,
	}
}

type TaskResponse struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	List     string    `json:"list"`
	Done     bool      `json:"done"`
	Priority int       `json:"priority"`
	AllDay   bool      `json:"all_day"`
	DueType  string    `json:"due_type"`
	Due      time.Time `json:"due,omitempty"`
	Created  time.Time `json:"created"`
	DoneOn   time.Time `json:"done_on,omitempty"`
}

func FromTask(t core.Task) TaskResponse {
	resp := TaskResponse{
		ID:       t.ID,
		Title:    t.Title,
		List:     t.List,
		Done:     t.Done,
		Priority: t.Priority,
		DueType:  string(t.DueType),
		Due:      t.Due,
		AllDay:   t.AllDay,
		Created:  t.Created,
		DoneOn:   t.DoneOn,
	}
	return resp
}

// MarshalJSON overwrites JSON marshalling to not send zero-value time fields
func (t TaskResponse) MarshalJSON() ([]byte, error) {
	var due, doneOn string

	if !t.Due.IsZero() {
		due = t.Due.Format(time.RFC3339)
	}

	if !t.DoneOn.IsZero() {
		doneOn = t.DoneOn.Format(time.RFC3339)
	}
	type Alias TaskResponse
	return json.Marshal(&struct {
		Alias
		Due    string `json:"due,omitempty"`
		DoneOn string `json:"done_on,omitempty"`
	}{
		Alias:  Alias(t),
		Due:    due,
		DoneOn: doneOn,
	})
}

type ListAdd struct {
	Name   string `json:"name"`
	Colour RGB    `json:"colour"`
}

type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

type TaskAdd struct {
	Title    string       `json:"title"`
	Priority int          `json:"priority"`
	AllDay   bool         `json:"all_day"`
	DueType  core.DueType `json:"due_type"`
	Due      time.Time    `json:"due"`
}

// UnmarshalJSON overwrites JSON unmarshalling to parse time fields properly
// TODO this is not fail-safe, it will fall apart if JS sends a different format
func (t *TaskAdd) UnmarshalJSON(data []byte) error {
	type Alias TaskAdd
	aux := struct {
		Due string `json:"due"`
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
	if string(aux.DueType) != "none" || aux.DueType != core.DueNone {
		due, err := time.ParseInLocation(format, aux.Due, time.Local)
		if err != nil {
			return err
		}
		t.Due = due
	}
	return nil
}

// TaskChange is used to change a task. It is used in PATCH requests. Only fields that are set will be changed.
// DueType must be set to one of TypeDueOn, TypeDueBy or TypeDueNone in requests to change the due date, otherwise
// the due date supplied in the request will be ignored.
type TaskChange struct {
	Title    string `json:"title"`
	List     string `json:"list"`
	Done     bool   `json:"done"`
	Priority int    `json:"priority"`
	AllDay   bool   `json:"all_day"`

	// DueType must be set to one of TypeDueOn, TypeDueBy or TypeDueNone in requests to change the due date.
	DueType core.DueType `json:"due_type"`
	Due     time.Time    `json:"due,omitempty"`
}

func (t *TaskChange) Validate() error {
	if t.Title == "" {
		return fmt.Errorf("missing task title")
	}
	if t.List == "" {
		return fmt.Errorf("missing list name")
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

	if priority, ok := input["priority"]; ok {
		t.Priority = priority.(int)
	}

	if allDay, ok := input["all_day"]; ok {
		t.AllDay = allDay.(bool)
	}

	// due_type must be set on all requests to change the due date
	if _, ok := input["due_type"]; ok {
		if err := t.overwriteDueFields(input); err != nil {
			return err
		}
	} else {
		// if due_type is not set, but due is, inform the user that due_type is required
		if _, ok = input["due"]; ok {
			return fmt.Errorf("due_type must be set (to due_on, due_by or none) when changing the due date")
		}
	}

	return nil
}

func (t *TaskChange) overwriteDueFields(input map[string]interface{}) error {

	// gracefully handle bad input
	var dueTyp core.DueType
	v, ok := input["due_type"].(string)
	if !ok {
		return fmt.Errorf("missing due_type field")
	}
	switch v {
	case string(core.DueBy), string(core.DueOn), string(core.DueNone):
		dueTyp = core.DueType(v)
	case "none":
		dueTyp = core.DueNone
	default:
		return fmt.Errorf("invalid due_type")
	}

	format := "2006-01-02T15:04"
	if t.AllDay {
		format = "2006-01-02"
	}

	switch dueTyp {
	case core.DueNone:
		t.Due = time.Time{}
		t.DueType = core.DueNone
	case core.DueOn, core.DueBy:
		dueRaw, ok := input["due"]
		if !ok {
			return fmt.Errorf("missing due field")
		}
		due, err := time.ParseInLocation(format, dueRaw.(string), time.Local)
		if err != nil {
			return err
		}
		t.Due = due
		t.DueType = dueTyp
	default:
		panic(fmt.Sprintf("invalid due type %v", dueTyp))
	}

	return nil
}
