package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
	}

	// Ensure the file exists
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		// Create the file
		if err := os.WriteFile(path, []byte{}, 0600); err != nil {
			panic(err)
		}
	}

	return &File{Path: path}
}

func (f *File) load() (*FileStore, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}

	return unmarshalStore(data)
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

// YAML can't handle the Node interface, so we need to unmarshal the store manually.
func unmarshalStore(data []byte) (*FileStore, error) {

	storeRaw := &yamlStore{}
	if err := yaml.Unmarshal(data, storeRaw); err != nil {
		return nil, err
	}

	store := yamlToStore(storeRaw)

	return store, nil
}

type yamlStore struct {
	Lists    []*core.List
	Filtered []*filteredList
}

func yamlToStore(raw *yamlStore) *FileStore {
	store := &FileStore{
		Lists: raw.Lists,
	}
	for _, l := range raw.Filtered {
		store.Filtered = append(store.Filtered, &filter.List{
			Name:   l.Name,
			Filter: l.Filter,
		})
	}
	return store
}

type filteredList struct {
	Name   string
	Filter filter.Node
}

func (l *filteredList) UnmarshalYAML(node *yaml.Node) error {

	type rawList struct {
		Name   string
		Filter map[string]interface{}
	}

	var r rawList
	if err := node.Decode(&r); err != nil {
		return err
	}

	l.Name = r.Name
	if r.Filter != nil {
		filt, err := unmarshalFilter(r.Filter)
		if err != nil {
			return err
		}
		l.Filter = filt
	}

	return nil
}

func unmarshalFilter(node map[string]interface{}) (filter.Node, error) {
	_, ok := node["operator"]
	if !ok {
		// it's a leaf node
		return unmarshalComparison(node)
	}
	return unmarshalLogical(node)
}

func unmarshalComparison(node map[string]interface{}) (filter.Node, error) {
	field, ok := node["field"]
	if !ok {
		return nil, fmt.Errorf("missing field in comparison")
	}
	op, ok := node["op"]
	if !ok {
		return nil, fmt.Errorf("missing operator in comparison")
	}
	value, ok := node["value"]
	if !ok {
		return nil, fmt.Errorf("missing value in comparison")
	}
	compOp, err := filter.NewComparisonOperator(field.(string), op.(string), value.(string))
	if err != nil {
		return nil, err
	}
	return compOp, nil
}

func unmarshalLogical(node map[string]interface{}) (filter.Node, error) {
	op := filter.LogicalOperator{}
	switch node["operator"] {
	case "AND":
		op.Operator = filter.OpAnd
	case "OR":
		op.Operator = filter.OpOr
	default:
		return nil, fmt.Errorf("unknown logical operator: %s", node["operator"])
	}
	for _, child := range node["children"].([]interface{}) {
		childNode, err := unmarshalFilter(child.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		op.Children = append(op.Children, childNode)
	}
	return &op, nil
}
