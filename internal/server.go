package internal

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
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
	lists := s.orga.Lists()

	type response struct {
		Lists []string `json:"lists"`
	}

	s.jsonResponse(w, http.StatusOK, response{Lists: lists})
}

func (s *Server) handleListGet(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing list name"))
		return
	}

	type response struct {
		List List `json:"list"`
	}
	l, err := s.orga.GetList(name)
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}
	s.jsonResponse(w, http.StatusOK, response{List: l})
}

func (s *Server) handleListPost(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Name string `json:"name"`
	}
	var req request
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
		List List `json:"list"`
	}

	s.jsonResponse(w, http.StatusCreated, response{List: l})
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

	request := struct {
		Done bool `json:"done"`
	}{}
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
	}
	if err = s.orga.MarkDone(id, request.Done); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTaskPost(w http.ResponseWriter, r *http.Request) {
	list := r.PathValue("name")
	if list == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing list name"))
		return
	}

	request := struct {
		Title string `json:"title"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	if request.Title == "" {
		s.httpError(w, http.StatusBadRequest, errors.New("missing task title"))
		return
	}

	t, err := s.orga.AddItem(list, request.Title)
	if err != nil {
		s.httpError(w, http.StatusBadRequest, err)
		return
	}

	resp := struct {
		Task Task `json:"task"`
	}{Task: t}

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
	Lists() []string
	GetList(name string) (List, error)
	AddList(name string) (List, error)
	DelList(name string) error
	AddItem(list, title string) (Task, error)
	DelItem(id int) error
	MarkDone(taskID int, done bool) error
}
