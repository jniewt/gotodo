package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	gotasks "github.com/jniewt/gotodo/internal"
)

//go:embed static
var staticFiles embed.FS

func main() {

	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	var web string
	flag.StringVar(&web, "addr", ":8080", "address and port to listen on (<addr>:<port>)")
	flag.Parse()

	// Create a subdirectory in the embedded filesystem
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		logger.WithError(err).Error("Failed to create sub filesystem")
		os.Exit(1)
	}

	repo := gotasks.NewRepository()

	// TODO remove this in production
	addTestData(repo)

	server := gotasks.NewServer(staticFS, repo, log.NewEntry(logger))

	log.WithField("addr", web).Info("Server started.")
	srv := http.Server{Handler: server, Addr: web}
	if err := srv.ListenAndServe(); err != nil {
		logger.WithError(err).Error("Failed to start server")
		os.Exit(1)
	}
}

func addTestData(repo *gotasks.Repository) {
	tasks := []string{"Buy avocados", "Walk the cat", "Write task app", "Learn JS"}

	_, err := repo.AddList("home")
	if err != nil {
		panic(err)
	}

	for _, todo := range tasks {
		_, err = repo.AddItem("home", todo)
		if err != nil {
			panic(err)
		}
	}

	tasks = []string{"Write report"}

	_, err = repo.AddList("work")
	if err != nil {
		panic(err)
	}

	for _, todo := range tasks {
		_, err = repo.AddItem("work", todo)
		if err != nil {
			panic(err)
		}
	}
}
