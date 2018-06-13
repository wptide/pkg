package sync

import (
	"time"

	"github.com/wptide/pkg/wporg"
)

// Dispatcher describes an interface for dispatching RepoProjects to
// a queue service. The service will need to implement this interface.
type Dispatcher interface {
	Dispatch(project wporg.RepoProject) error
	Init() error
	Close() error
}

// Updater describes an interface to determine and record the currency
// of the last dispatched RepoProject.
type Updater interface {
	UpdateCheck(project wporg.RepoProject) bool
	RecordUpdate(project wporg.RepoProject) error
}

// Checker records and retrieves the last sync times.
type Checker interface {
	SetSyncTime(event, projectType string, t time.Time)
	GetSyncTime(event, projectType string) time.Time
}

// UpdateChecker describes a combined Updater and Checker.
type UpdateChecker interface {
	Updater
	Checker
}
