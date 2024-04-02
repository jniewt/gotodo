package storage

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/jniewt/gotodo/internal/core"
	"github.com/jniewt/gotodo/internal/filter"
)

type File struct {
	Path string
}

type FileStore struct {
	Lists    []*core.List
	Filtered []*filter.List
}

func NewFile(path string) *File {
	// Ensure the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if errors.Is(err, os.ErrNotExist) {
			// Create the file
			if err := os.WriteFile(path, []byte{}, 0600); err != nil {
				panic(err)
			}
		}
	}

	return &File{Path: path}
}

func (f *File) load() (*FileStore, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}

	var store FileStore
	if err := yaml.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

func (f *File) save(store *FileStore) error {
	data, err := yaml.Marshal(store)
	if err != nil {
		return err
	}

	return os.WriteFile(f.Path, data, 0600)
}

// GetList retrieves a list by name from the YAML file.
func (f *File) GetList(name string) (*core.List, error) {
	store, err := f.load()
	if err != nil {
		return nil, err
	}

	for _, list := range store.Lists {
		if list.Name == name {
			return list, nil
		}
	}
	return nil, os.ErrNotExist // Or a custom error indicating the list was not found.
}

func (f *File) GetAllLists() ([]*core.List, error) {
	store, err := f.load()
	if err != nil {
		return nil, err
	}

	return store.Lists, nil
}

func (f *File) AddList(list *core.List) error {
	store, err := f.load()
	if err != nil {
		return err
	}

	store.Lists = append(store.Lists, list)

	return f.save(store)
}

func (f *File) UpdateList(name string, list *core.List) error {
	store, err := f.load()
	if err != nil {
		return err
	}

	for i, l := range store.Lists {
		if l.Name == name {
			store.Lists[i] = list
			break
		}
	}

	return f.save(store)
}

func (f *File) DeleteList(name string) error {
	store, err := f.load()
	if err != nil {
		return err
	}

	for i, list := range store.Lists {
		if list.Name == name {
			store.Lists = append(store.Lists[:i], store.Lists[i+1:]...)
			break
		}
	}

	return f.save(store)
}

func (f *File) GetFiltered(name string) (*filter.List, error) {
	store, err := f.load()
	if err != nil {
		return nil, err
	}

	for _, list := range store.Filtered {
		if list.Name == name {
			return list, nil
		}
	}

	return nil, os.ErrNotExist
}

func (f *File) GetAllFiltered() ([]*filter.List, error) {
	store, err := f.load()
	if err != nil {
		return nil, err
	}

	return store.Filtered, nil
}

func (f *File) AddFiltered(list *filter.List) error {
	store, err := f.load()
	if err != nil {
		return err
	}

	store.Filtered = append(store.Filtered, list)

	return f.save(store)
}

func (f *File) DeleteFiltered(name string) error {
	store, err := f.load()
	if err != nil {
		return err
	}

	for i, list := range store.Filtered {
		if list.Name == name {
			store.Filtered = append(store.Filtered[:i], store.Filtered[i+1:]...)
			break
		}
	}

	return f.save(store)
}
