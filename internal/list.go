package internal

import (
	"encoding/json"
	"fmt"
	"time"
)

// Repository provides access to the todo list storage.
type Repository struct {
	lists []*List
}

// NewRepository creates a new repository.
func NewRepository() *Repository {
	return &Repository{}
}

var ErrListNotFound = fmt.Errorf("list not found")

var ErrListExists = fmt.Errorf("list already exists")

// GetList returns a list by name.
func (r *Repository) GetList(name string) (List, error) {
	list, err := r.getList(name)
	if err != nil {
		return List{}, err
	}
	return *list, err
}

// Lists returns the names of all lists.
func (r *Repository) Lists() []string {
	lists := make([]string, 0, len(r.lists))
	for _, list := range r.lists {
		lists = append(lists, list.Name)
	}
	return lists
}

// AddList adds a new list.
func (r *Repository) AddList(name string) (List, error) {
	for _, list := range r.lists {
		if list.Name == name {
			return List{}, ErrListExists
		}
	}
	l := List{Name: name}
	r.lists = append(r.lists, &l)
	return l, nil
}

// DelList deletes a list.
func (r *Repository) DelList(name string) error {
	for i, list := range r.lists {
		if list.Name == name {
			r.lists = append(r.lists[:i], r.lists[i+1:]...)
			return nil
		}
	}
	return ErrListNotFound
}

func (r *Repository) AddItem(list string, task TaskAdd) (Task, error) {
	l, err := r.getList(list)
	if err != nil {
		return Task{}, err
	}

	if task.Title == "" {
		return Task{}, fmt.Errorf("missing task title")
	}

	item := Task{
		ID:      r.NewID(),
		Title:   task.Title,
		List:    list,
		AllDay:  task.AllDay,
		Created: time.Now(),
	}

	// set dueOn or dueBy if set
	switch {
	case !task.DueBy.IsZero() && !task.DueOn.IsZero():
		return Task{}, fmt.Errorf("only one of dueOn or dueBy can be set")
	case !task.DueBy.IsZero():
		item.DueBy = task.DueBy
	case !task.DueOn.IsZero():
		item.DueOn = task.DueOn
	}

	l.Items = append(l.Items, &item)
	return item, nil
}

func (r *Repository) DelItem(id int) error {
	for _, list := range r.lists {
		for i, item := range list.Items {
			if item.ID == id {
				list.Items = append(list.Items[:i], list.Items[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("item not found")
}

func (r *Repository) GetTask(id int) (Task, error) {
	t, err := r.getTask(id)
	if err != nil {
		return Task{}, err
	}
	return *t, nil
}

func (r *Repository) UpdateTask(id int, change TaskChange) (Task, error) {
	t, err := r.getTask(id)
	if err != nil {
		return Task{}, err
	}

	if t.Done != change.Done {
		_, err = r.MarkDone(id, change.Done)
		if err != nil {
			return Task{}, err
		}
	}
	if t.List != change.List {
		_, err = r.MoveTask(id, change.List)
		if err != nil {
			return Task{}, err
		}
	}

	t.Title = change.Title
	t.AllDay = change.AllDay
	t.DueBy = change.DueBy
	t.DueOn = change.DueOn

	return *t, nil
}

func (r *Repository) MoveTask(id int, list string) (Task, error) {
	task, err := r.getTask(id)
	if err != nil {
		return Task{}, err
	}

	listFrom, err := r.getList(task.List)
	if err != nil {
		panic(fmt.Sprintf("list %v for task %d not found", task.List, id))
	}

	listTo, err := r.getList(list)
	if err != nil {
		return Task{}, err
	}

	// remove task from listFrom
	for i, item := range listFrom.Items {
		if item.ID == id {
			listFrom.Items = append(listFrom.Items[:i], listFrom.Items[i+1:]...)
			break
		}
	}

	// add task to listTo
	listTo.Items = append(listTo.Items, task)

	// update task list
	task.List = list

	return *task, nil
}

func (r *Repository) MarkDone(id int, done bool) (Task, error) {
	task, err := r.getTask(id)
	if err != nil {
		return Task{}, err
	}
	task.Done = done
	if done {
		task.DoneOn = time.Now()
	} else {
		task.DoneOn = time.Time{}
	}
	return *task, nil
}

func (r *Repository) getList(name string) (*List, error) {
	for _, list := range r.lists {
		if list.Name == name {
			return list, nil
		}
	}
	return nil, ErrListNotFound
}

func (r *Repository) getTask(id int) (*Task, error) {
	for _, list := range r.lists {
		for _, item := range list.Items {
			if item.ID == id {
				return item, nil
			}
		}
	}
	return nil, fmt.Errorf("item not found")
}

// NewID returns a new unique ID. This is a naive implementation that iterates over all items to find the highest ID.
func (r *Repository) NewID() int {
	id := 0
	for _, list := range r.lists {
		for _, item := range list.Items {
			if item.ID > id {
				id = item.ID
			}
		}
	}
	return id + 1
}

type List struct {
	Name  string  `json:"name"`
	Items []*Task `json:"items"`
}

type Task struct {
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

// MarshalJSON overwrites JSON marshalling to not send zero-value time fields
func (t Task) MarshalJSON() ([]byte, error) {
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
	type Alias Task
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

type TaskChange struct {
	Title  string    `json:"title"`
	List   string    `json:"list"`
	Done   bool      `json:"done"`
	AllDay bool      `json:"all_day"`
	DueBy  time.Time `json:"due_by,omitempty"`
	DueOn  time.Time `json:"due_on,omitempty"`
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
// TODO need to rethink handling of dueBy and dueOn since changing them through a REST request is a bit tricky
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

	format := "2006-01-02T15:04"
	if t.AllDay {
		format = "2006-01-02"
	}

	var dueByRaw, dueOnRaw interface{}
	var dueByExists, dueOnExists bool
	var dueBy, dueOn time.Time
	var err error
	if dueByRaw, dueByExists = input["due_by"]; dueByExists {
		dueBy, err = time.ParseInLocation(format, dueByRaw.(string), time.Local)
		if err != nil {
			return err
		}
	}
	if dueOnRaw, dueOnExists = input["due_on"]; dueOnExists {
		dueOn, err = time.ParseInLocation(format, dueOnRaw.(string), time.Local)
		if err != nil {
			return err
		}
	}

	if dueByExists {
		t.DueBy = dueBy
		if !dueBy.IsZero() {
			t.DueOn = time.Time{}
		}
	}
	if dueOnExists {
		t.DueOn = dueOn
		if !dueOn.IsZero() {
			t.DueBy = time.Time{}
		}
	}

	return nil
}
