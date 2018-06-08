package sync

import (
	"github.com/wptide/pkg/wporg"
	"time"
)

// Dispatcher describes an interface for dispatching RepoProjects to
// a queue service. The service will need to implement this interface.
type Dispatcher interface {
	Dispatch(project wporg.RepoProject) error
	Init() error
	Close() error
}

// UpdateChecker describes an interface to determine and record the currency
// of the last dispatched RepoProject.
type UpdateChecker interface {
	UpdateCheck(project wporg.RepoProject) bool
	RecordUpdate(project wporg.RepoProject) error
}

// Syncer records and retrieves the last sync times
type SyncChecker interface {
	SetSyncTime(event, projectType string, t time.Time)
	GetSyncTime(event, projectType string) time.Time
}

type UpdateSyncChecker interface {
	UpdateChecker
	SyncChecker
}
