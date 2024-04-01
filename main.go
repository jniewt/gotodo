package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jniewt/gotodo/api"
	"github.com/jniewt/gotodo/internal/filter"
	"github.com/jniewt/gotodo/internal/repository"
	"github.com/jniewt/gotodo/internal/rest"
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

	repo := repository.NewRepository()

	// TODO remove this in production
	addTestData(repo)

	server := rest.NewServer(staticFS, repo, log.NewEntry(logger))

	log.WithField("addr", web).Info("Server started.")
	srv := http.Server{Handler: server, Addr: web}
	if err := srv.ListenAndServe(); err != nil {
		logger.WithError(err).Error("Failed to start server")
		os.Exit(1)
	}
}

func addTestData(repo *repository.Repository) {
	tasks := []api.TaskAdd{
		{Title: "Buy avocados", AllDay: true, DueBy: timeAtHourInDays(0, 2)},
		{Title: "Walk the cat", DueOn: time.Now().Add(2 * time.Hour)},
		{Title: "Write task app", DueBy: time.Now().Add(24 * time.Hour)},
		{Title: "Learn JS"},
		{Title: "Cook dinner", DueOn: todayAtHour(18)},
		{Title: "Wash the dishes", AllDay: true, DueBy: today()},
		{Title: "Something overdue", AllDay: true, DueBy: time.Now().Add(-24 * time.Hour)},
	}

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

	tasks = []api.TaskAdd{
		{Title: "Write report"},
		{Title: "Prepare presentation", AllDay: true, DueBy: timeAtHourInDays(0, 5)},
		{Title: "Call client", DueOn: timeAtHourInDays(9, 3)},
	}

	_, err = repo.AddList("work")
	if err != nil {
		panic(err)
	}

	for _, t := range tasks {
		_, err = repo.AddItem("work", t)
		if err != nil {
			panic(err)
		}
	}

	// create filtered list that contains all undone tasks from all lists that are either due on today, due by in the
	// next n days or have no due date
	soonFilters := []filter.Node{
		filter.PendingOrDoneToday(),
		filter.Due(
			filter.DueOnToday(),
			filter.DueByInDays(14),
			filter.NoDueDate(),
		),
	}
	_, err = repo.AddFilteredList("soon", soonFilters...)
	if err != nil {
		panic(err)
	}
}

// timeAtHourInDays takes a time and number of days from now and returns a time at that hour of the day that many days from now.
func timeAtHourInDays(hh int, days int) time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day()+days, hh, 0, 0, 0, time.Local)
}

func today() time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func todayAtHour(hh int) time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), hh, 0, 0, 0, time.Local)
}
