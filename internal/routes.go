package internal

import "net/http"

func (s *Server) routes() {
	s.router.Handle("GET /", http.FileServer(http.FS(s.staticFS)))
	// get names of all lists TODO shouldn't it return the lists themselves?
	s.router.HandleFunc("GET /api/list", s.handleListGetAll)
	// get a list
	s.router.HandleFunc("GET /api/list/{name}", s.handleListGet)
	// create a new list
	s.router.HandleFunc("POST /api/list", s.handleListPost)
	// delete a list
	s.router.HandleFunc("DELETE /api/list/{name}", s.handleListDel)
	// add a task
	s.router.HandleFunc("POST /api/list/{name}", s.handleTaskPost)
	// delete a task
	s.router.HandleFunc("DELETE /api/items/{id}", s.handleTaskDel)
	// mark item as done
	s.router.HandleFunc("PATCH /api/items/{id}", s.handleTaskChange)
}
