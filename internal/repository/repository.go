package repository

import (
	"fmt"
	"time"

	"github.com/jniewt/gotodo/api"
	"github.com/jniewt/gotodo/internal/core"
)

// Repository provides access to the todo list storage.
type Repository struct {
	lists []*core.List
}

// NewRepository creates a new repository.
func NewRepository() *Repository {
	return &Repository{}
}

var ErrListNotFound = fmt.Errorf("list not found")

var ErrListExists = fmt.Errorf("list already exists")

// GetList returns a list by name.
func (r *Repository) GetList(name string) (core.List, error) {
	list, err := r.getList(name)
	if err != nil {
		return core.List{}, err
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
func (r *Repository) AddList(name string) (core.List, error) {
	for _, list := range r.lists {
		if list.Name == name {
			return core.List{}, ErrListExists
		}
	}
	l := core.List{Name: name}
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

func (r *Repository) AddItem(list string, task api.TaskAdd) (core.Task, error) {
	l, err := r.getList(list)
	if err != nil {
		return core.Task{}, err
	}

	if task.Title == "" {
		return core.Task{}, fmt.Errorf("missing task title")
	}

	item := core.Task{
		ID:      r.NewID(),
		Title:   task.Title,
		List:    list,
		AllDay:  task.AllDay,
		Created: time.Now(),
	}

	// set dueOn or dueBy if set
	switch {
	case !task.DueBy.IsZero() && !task.DueOn.IsZero():
		return core.Task{}, fmt.Errorf("only one of dueOn or dueBy can be set")
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

func (r *Repository) GetTask(id int) (core.Task, error) {
	t, err := r.getTask(id)
	if err != nil {
		return core.Task{}, err
	}
	return *t, nil
}

func (r *Repository) UpdateTask(id int, change api.TaskChange) (core.Task, error) {
	t, err := r.getTask(id)
	if err != nil {
		return core.Task{}, err
	}

	if t.Done != change.Done {
		_, err = r.MarkDone(id, change.Done)
		if err != nil {
			return core.Task{}, err
		}
	}
	if t.List != change.List {
		_, err = r.MoveTask(id, change.List)
		if err != nil {
			return core.Task{}, err
		}
	}

	t.Title = change.Title
	t.AllDay = change.AllDay
	t.DueBy = change.DueBy
	t.DueOn = change.DueOn

	return *t, nil
}

func (r *Repository) MoveTask(id int, list string) (core.Task, error) {
	task, err := r.getTask(id)
	if err != nil {
		return core.Task{}, err
	}

	listFrom, err := r.getList(task.List)
	if err != nil {
		panic(fmt.Sprintf("list %v for task %d not found", task.List, id))
	}

	listTo, err := r.getList(list)
	if err != nil {
		return core.Task{}, err
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

func (r *Repository) MarkDone(id int, done bool) (core.Task, error) {
	task, err := r.getTask(id)
	if err != nil {
		return core.Task{}, err
	}
	task.Done = done
	if done {
		task.DoneOn = time.Now()
	} else {
		task.DoneOn = time.Time{}
	}
	return *task, nil
}

func (r *Repository) getList(name string) (*core.List, error) {
	for _, list := range r.lists {
		if list.Name == name {
			return list, nil
		}
	}
	return nil, ErrListNotFound
}

func (r *Repository) getTask(id int) (*core.Task, error) {
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