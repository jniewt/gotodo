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

func (r *Repository) AddItem(list, title string) (Task, error) {
	l, err := r.getList(list)
	if err != nil {
		return Task{}, err
	}
	item := Task{
		ID:      r.NewID(),
		Title:   title,
		List:    list,
		Created: time.Now(),
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

func (r *Repository) MarkDone(id int, done bool) (Task, error) {
	task, err := r.getItem(id)
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

func (r *Repository) getItem(id int) (*Task, error) {
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
