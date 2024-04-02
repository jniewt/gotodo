package storage

import (
	"errors"

	"github.com/jniewt/gotodo/internal/core"
	"github.com/jniewt/gotodo/internal/filter"
)

// Fake is a fake storage implementation that is used for testing purposes.
type Fake struct {
	Lists    []*core.List
	Filtered []*filter.List
}

func (f *Fake) GetList(name string) (*core.List, error) {
	for _, l := range f.Lists {
		if l.Name == name {
			return l, nil
		}
	}
	return nil, errors.New("list not found")
}

func (f *Fake) GetAllLists() ([]*core.List, error) {
	return f.Lists, nil
}

func (f *Fake) AddList(list *core.List) error {
	f.Lists = append(f.Lists, list)
	return nil
}

func (f *Fake) UpdateList(name string, list *core.List) error {
	for i, l := range f.Lists {
		if l.Name == name {
			f.Lists[i] = list
			return nil
		}
	}
	return errors.New("list not found")
}

func (f *Fake) DeleteList(name string) error {
	for i, l := range f.Lists {
		if l.Name == name {
			f.Lists = append(f.Lists[:i], f.Lists[i+1:]...)
			return nil
		}
	}
	return errors.New("list not found")
}

func (f *Fake) GetFiltered(name string) (*filter.List, error) {
	for _, l := range f.Filtered {
		if l.Name == name {
			return l, nil
		}
	}
	return nil, errors.New("filtered list not found")
}

func (f *Fake) GetAllFiltered() ([]*filter.List, error) {
	return f.Filtered, nil
}

func (f *Fake) AddFiltered(list *filter.List) error {
	f.Filtered = append(f.Filtered, list)
	return nil
}

func (f *Fake) DeleteFiltered(name string) error {
	for i, l := range f.Filtered {
		if l.Name == name {
			f.Filtered = append(f.Filtered[:i], f.Filtered[i+1:]...)
			return nil
		}
	}
	return errors.New("filtered list not found")
}
