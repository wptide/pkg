package firestore

import (
	"time"
	"github.com/wptide/pkg/wporg"
	"context"
	"fmt"
	"cloud.google.com/go/firestore"
	"sync"
	"encoding/json"
)

var (
	defaultClient *firestore.Client
	setData       = setDoc
	getData       = getDoc
	mutex         = sync.Mutex{}
)

type FirestoreSync struct {
	c        *firestore.Client
	ctx      context.Context
	rootPath string
}

func (f FirestoreSync) UpdateCheck(project wporg.RepoProject) bool {
	key := fmt.Sprintf("%s/%s/%s", f.rootPath, project.Type, project.Slug)

	data := getData(f.ctx, key, f.c)
	record, err := itop(data)
	if err != nil {
		return true
	}
	return record.LastUpdated != project.LastUpdated || record.Version != project.Version
}

func (f FirestoreSync) RecordUpdate(project wporg.RepoProject) error {
	mutex.Lock()
	defer mutex.Unlock()

	key := fmt.Sprintf("%s/%s/%s", f.rootPath, project.Type, project.Slug)

	data, _ := ptoi(project)
	return setData(f.ctx, key, f.c, data)
}

func (f FirestoreSync) SetSyncTime(event, projectType string, t time.Time) {
	key := fmt.Sprintf("%s-sync-%s", projectType, event)
	data := getData(f.ctx, f.rootPath, f.c)
	data[key] = t.UnixNano()
	setData(f.ctx, f.rootPath, f.c, data)
}

func (f FirestoreSync) GetSyncTime(event, projectType string) time.Time {
	key := fmt.Sprintf("%s-sync-%s", projectType, event)

	var t time.Time

	data := getData(f.ctx, f.rootPath, f.c)
	timestamp, ok := data[key].(int64)

	if !ok {
		t, _ = time.Parse(wporg.TimeFormat, "1970-01-01 12:00am MST")
	} else {
		t = time.Unix(0, timestamp)
	}

	return t
}

// itop converts a Firestore record into a wporg.RepoProject.
func itop(data map[string]interface{}) (wporg.RepoProject, error) {
	var project wporg.RepoProject
	var cErr error
	if temp, err := json.Marshal(data); err == nil {
		cErr = json.Unmarshal(temp, &project)
	}
	return project, cErr
}

// ptoi converts a wporg.RepoProject into a interface map for Firestore.
func ptoi(project wporg.RepoProject) (map[string]interface{}, error) {
	var data map[string]interface{}
	var cErr error
	if temp, err := json.Marshal(project); err == nil {
		cErr = json.Unmarshal(temp, &data)
		data["type"] = project.Type
	}
	return data, cErr
}

func New(ctx context.Context, projectId string, rootDocPath string) *FirestoreSync {
	if defaultClient == nil {
		c, _ := firestore.NewClient(ctx, projectId)
		defaultClient = c
	}

	return &FirestoreSync{
		ctx:      ctx,
		c:        defaultClient,
		rootPath: rootDocPath,
	}
}
