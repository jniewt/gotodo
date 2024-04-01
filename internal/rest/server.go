package rest

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/jniewt/gotodo/api"
	"github.com/jniewt/gotodo/internal/core"
	"github.com/jniewt/gotodo/internal/repository"
)

type Server struct {
	orga Organiser

	router   *http.ServeMux
	staticFS fs.FS

	log *log.Entry
}

func NewServer(static fs.FS, orga Organiser, logger *log.Entry) *Server {

	s := &Server{
		orga:     orga,
		router:   http.NewServeMux(),
		staticFS: static,
		log:      logger,
	}

	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// httpError is a helper that writes an JSON error response to the response writer in a standard way.
func (s *Server) httpError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	_, wErr := w.Write([]byte(`{"error":"` + err.Error() + `"}`))
	if wErr != nil {
		// use slog to log the error
		s.log.WithError(wErr).Warn("Failed to write error response")
	}
}

// helper that writes a JSON response to the response writer in a standard way.
func (s *Server) jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.httpError(w, http.StatusInternalServerError, err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func (s *Server) handleListGetAll(w http.ResponseWriter, _ *http.Request) {
	lists, filtered := s.orga.Lists()

	type response struct {
		Lists         []string `json:"lists"`
		FilteredLists []string `json:"filtered_lists"`
	}

	s.jsonResponse(w, http.StatusOK, response{Lists: lists, FilteredLists: filtered})
}

func (s *Server) handleListGet(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing list name"))
		return
	}

	type response struct {
		List     api.ListResponse `json:"list"`
		Filtered bool             `json:"filtered"`
	}
	l, err := s.orga.GetList(name)
	if err == nil {
		s.jsonResponse(w, http.StatusOK, response{List: api.FromList(l)})
	} else if !errors.Is(err, repository.ErrListNotFound) {
		s.httpError(w, http.StatusInternalServerError, err)
	}

	// check if the list is a filtered list
	// TODO this is not a particularly intuitive flow
	s.handleFilteredGet(w, r)
}

// handleFilteredGet returns a virtual list containing the tasks matching the given filter. The response looks exactly
// like a normal list response but has the "filtered" field set to true.
func (s *Server) handleFilteredGet(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing list name"))
		return
	}

	type response struct {
		List     api.ListResponse `json:"list"`
		Filtered bool             `json:"filtered"`
	}

	tasks, err := s.orga.GetFilteredTasks(name)
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	l := core.List{
		Name:  name,
		Items: tasks,
	}

	s.jsonResponse(w, http.StatusOK, response{List: api.FromList(l), Filtered: true})

}

func (s *Server) handleListPost(w http.ResponseWriter, r *http.Request) {
	var req api.ListAdd
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	if req.Name == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing list name"))
		return
	}

	l, err := s.orga.AddList(req.Name)
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	type response struct {
		List api.ListResponse `json:"list"`
	}

	s.jsonResponse(w, http.StatusCreated, response{List: api.FromList(l)})
}

func (s *Server) handleListDel(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing list name"))
		return
	}

	if err := s.orga.DelList(name); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTaskChange(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	change, err := s.initTaskChange(id)
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&change); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
	}

	// validate the input
	if _, err = s.orga.GetTask(id); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}
	if err = change.Validate(); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	res, err := s.orga.UpdateTask(id, change)
	if err != nil {
		s.httpError(w, http.StatusInternalServerError, err)
		return
	}

	// respond with the updated task
	resp := struct {
		Task api.TaskResponse `json:"task"`
	}{
		Task: api.FromTask(res),
	}

	s.jsonResponse(w, http.StatusAccepted, resp)
}

// initTaskChange initializes a TaskChange struct from the existing task with the given ID.
func (s *Server) initTaskChange(id int) (api.TaskChange, error) {
	t, err := s.orga.GetTask(id)
	if err != nil {
		return api.TaskChange{}, err
	}

	return api.TaskChange{
		Title:  t.Title,
		Done:   t.Done,
		List:   t.List,
		AllDay: t.AllDay,
		DueOn:  t.DueOn,
		DueBy:  t.DueBy,
	}, nil
}

func (s *Server) handleTaskPost(w http.ResponseWriter, r *http.Request) {
	list := r.PathValue("name")
	if list == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing list name"))
		return
	}

	request := api.TaskAdd{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	t, err := s.orga.AddItem(list, request)
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	resp := struct {
		Task api.TaskResponse `json:"task"`
	}{Task: api.FromTask(t)}

	s.jsonResponse(w, http.StatusCreated, resp)
}

func (s *Server) handleTaskDel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.orga.DelItem(id); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type Organiser interface {
	Lists() ([]string, []string)
	GetList(name string) (core.List, error)
	AddList(name string) (core.List, error)
	DelList(name string) error
	AddItem(list string, item api.TaskAdd) (core.Task, error)
	DelItem(id int) error
	MarkDone(taskID int, done bool) (core.Task, error)
	GetTask(id int) (core.Task, error)
	GetFilteredTasks(name string) ([]*core.Task, error)
	UpdateTask(id int, request api.TaskChange) (core.Task, error)
}
