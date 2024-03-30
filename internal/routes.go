package internal

import (
	"net/http"

	"github.com/jniewt/gotodo/cors"
)

func (s *Server) routes() {
	// allowCors is a middleware that allows CORS requests from any origin.
	allowCors := cors.New([]string{"*"}).HandleFunc

	s.router.Handle("GET /", http.FileServer(http.FS(s.staticFS)))

	// allow CORS preflight requests for all routes under /api
	s.router.Handle("OPTIONS /api/*", allowCors(func(_ http.ResponseWriter, _ *http.Request) {}))

	// get names of all lists TODO shouldn't it return the lists themselves?
	// returns JSON: {lists: [string]}
	s.router.HandleFunc("GET /api/list", allowCors(s.handleListGetAll))

	// get a list
	// returns JSON: {list: List}
	s.router.HandleFunc("GET /api/list/{name}", allowCors(s.handleListGet))

	// create a new list
	// accepts JSON: ListAdd, returns JSON: {list: List}
	s.router.HandleFunc("POST /api/list", allowCors(s.handleListPost))

	// delete a list
	s.router.HandleFunc("DELETE /api/list/{name}", allowCors(s.handleListDel))

	// add a task
	// accepts JSON: TaskAdd, returns JSON: {task: Task}
	s.router.HandleFunc("POST /api/list/{name}", allowCors(s.handleTaskPost))

	// delete a task
	s.router.HandleFunc("DELETE /api/items/{id}", allowCors(s.handleTaskDel))

	// change a task, e.g. mark item as done
	// accepts JSON: TaskChange, returns JSON: {task: Task}
	s.router.HandleFunc("PATCH /api/items/{id}", allowCors(s.handleTaskChange))
}
