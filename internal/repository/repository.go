package repository

import (
	"fmt"
	"time"

	"github.com/jniewt/gotodo/api"
	"github.com/jniewt/gotodo/internal/core"
	"github.com/jniewt/gotodo/internal/filter"
)

// Storage defines the interface for task list storage operations.
type Storage interface {
	GetList(name string) (*core.List, error)
	GetAllLists() ([]*core.List, error)
	AddList(list *core.List) error
	UpdateList(name string, list *core.List) error
	DeleteList(name string) error
	GetFiltered(name string) (*filter.List, error)
	GetAllFiltered() ([]*filter.List, error)
	AddFiltered(list *filter.List) error
	DeleteFiltered(name string) error
}

// Repository provides access to the task list storage. It keeps a cache of all lists and filtered lists to avoid
// unnecessary reads.
type Repository struct {
	lists    []*core.List
	filtered []*filter.List
	store    Storage
}

// NewRepository creates a new repository.
func NewRepository(store Storage) *Repository {
	lists, err := store.GetAllLists()
	if err != nil {
		panic(fmt.Sprintf("failed to load lists: %v", err))
	}
	filtered, err := store.GetAllFiltered()
	if err != nil {
		panic(fmt.Sprintf("failed to load filtered lists: %v", err))
	}
	return &Repository{
		lists:    lists,
		filtered: filtered,
		store:    store,
	}

}

// GetList returns a list by name.
func (r *Repository) GetList(name string) (core.List, error) {
	list, err := r.getList(name)
	if err != nil {
		return core.List{}, err
	}
	return *list, err
}

// Lists returns the names of all lists and filtered lists.
func (r *Repository) Lists() ([]string, []string) {
	lists := make([]string, 0, len(r.lists))
	for _, list := range r.lists {
		lists = append(lists, list.Name)
	}
	filtered := make([]string, 0, len(r.filtered))
	for _, list := range r.filtered {
		filtered = append(filtered, list.Name)
	}
	return lists, filtered
}

// AddList adds a new list.
func (r *Repository) AddList(name string) (core.List, error) {
	for _, list := range r.lists {
		if list.Name == name {
			return core.List{}, ErrListExists
		}
	}
	l := core.List{Name: name}
	err := r.store.AddList(&l)
	if err != nil {
		return core.List{}, err
	}

	err = r.updateListCache()
	if err != nil {
		return core.List{}, fmt.Errorf("failed to update list cache: %w", err)
	}
	return l, nil
}

// DelList deletes a list.
func (r *Repository) DelList(name string) error {
	err := r.store.DeleteList(name)
	if err != nil {
		return err
	}
	err = r.updateListCache()
	if err != nil {
		return fmt.Errorf("failed to update list cache: %w", err)
	}
	return nil
}

// AddFilteredList adds a new virtual list.
func (r *Repository) AddFilteredList(name string, filters ...filter.Node) (filter.List, error) {
	for _, l := range r.filtered {
		if l.Name == name {
			return filter.List{}, ErrListExists
		}
	}
	fl := &filter.List{Name: name, Filter: filter.NewFilter(filters...)}
	err := r.store.AddFiltered(fl)
	if err != nil {
		return filter.List{}, err
	}

	err = r.updateFilteredListCache()
	if err != nil {
		return filter.List{}, fmt.Errorf("failed to update filtered list cache: %w", err)
	}

	return *fl, nil
}

// GetFilteredTasks returns the tasks for a virtual list.
func (r *Repository) GetFilteredTasks(name string) ([]*core.Task, error) {
	for _, l := range r.filtered {
		if l.Name == name {
			return r.filterTasks(l.Filter), nil
		}
	}
	return nil, ErrListNotFound
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
		ID:      r.newID(),
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

	err = r.store.UpdateList(list, l)
	if err != nil {
		return core.Task{}, err
	}

	err = r.updateListCache()
	if err != nil {
		return core.Task{}, fmt.Errorf("failed to update list cache: %w", err)
	}

	return item, nil
}

func (r *Repository) DelItem(id int) error {
	for _, list := range r.lists {
		for i, item := range list.Items {
			if item.ID != id {
				continue
			}
			list.Items = append(list.Items[:i], list.Items[i+1:]...)
			err := r.store.UpdateList(list.Name, list)
			if err != nil {
				return err
			}

			err = r.updateListCache()
			if err != nil {
				return fmt.Errorf("failed to update list cache: %w", err)
			}
			return nil
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

	list, err := r.getList(t.List)
	if err != nil {
		panic(fmt.Sprintf("list %v for task %d not found", t.List, id))
	}
	err = r.store.UpdateList(list.Name, list)
	if err != nil {
		return core.Task{}, err
	}

	err = r.updateListCache()
	if err != nil {
		return core.Task{}, fmt.Errorf("failed to update list cache: %w", err)
	}

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

	err = r.store.UpdateList(listFrom.Name, listFrom)
	if err != nil {
		return core.Task{}, err
	}
	err = r.store.UpdateList(listTo.Name, listTo)
	if err != nil {
		return core.Task{}, err
	}

	err = r.updateListCache()
	if err != nil {
		return core.Task{}, fmt.Errorf("failed to update list cache: %w", err)
	}

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

	list, err := r.getList(task.List)
	if err != nil {
		panic(fmt.Sprintf("list %v for task %d not found", task.List, id))
	}
	err = r.store.UpdateList(list.Name, list)
	if err != nil {
		return core.Task{}, err
	}

	err = r.updateListCache()
	if err != nil {
		return core.Task{}, fmt.Errorf("failed to update list cache: %w", err)
	}

	return *task, nil
}

// newID returns a new unique ID. This is a naive implementation that iterates over all items to find the highest ID.
func (r *Repository) newID() int {
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

func (r *Repository) filterTasks(f filter.Node) []*core.Task {
	tasks := make([]*core.Task, 0)
	for _, list := range r.lists {
		for _, task := range list.Items {
			if f.Evaluate(*task) {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
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

func (r *Repository) updateListCache() error {
	lists, err := r.store.GetAllLists()
	if err != nil {
		return err
	}
	r.lists = lists
	return nil
}

func (r *Repository) updateFilteredListCache() error {
	filtered, err := r.store.GetAllFiltered()
	if err != nil {
		return err
	}
	r.filtered = filtered
	return nil
}

var (
	ErrListNotFound = fmt.Errorf("list not found")
	ErrListExists   = fmt.Errorf("list already exists")
)
