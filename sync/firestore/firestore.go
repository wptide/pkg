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
	mutex = sync.Mutex{}
)

type FirestoreSync struct {
	ctx      context.Context
	client   ClientInterface
	rootPath string
}

func (f FirestoreSync) UpdateCheck(project wporg.RepoProject) bool {
	key := fmt.Sprintf("%s/%s/%s", f.rootPath, project.Type, project.Slug)

	data := f.client.GetDoc(key)
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

	return f.client.SetDoc(key, data)
}

func (f FirestoreSync) SetSyncTime(event, projectType string, t time.Time) {
	key := fmt.Sprintf("%s-sync-%s", projectType, event)

	data := f.client.GetDoc(f.rootPath)
	if len(data) == 0 {
		data = make(map[string]interface{})
	}
	data[key] = t.UnixNano()
	f.client.SetDoc(f.rootPath, data)
}


func (f FirestoreSync) GetSyncTime(event, projectType string) time.Time {
	key := fmt.Sprintf("%s-sync-%s", projectType, event)

	var t time.Time

	data := f.client.GetDoc(f.rootPath)
	timestamp, ok := data[key].(int64)

	if !ok {
		t, _ = time.Parse(wporg.TimeFormat, "1970-01-01 12:00am UTC")
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

// New creates a new FirestoreSync (UpdateSyncChecker) with a default client
// using Firestore.
func New(ctx context.Context, projectId string, rootDocPath string) *FirestoreSync {

	fireClient, _ := firestore.NewClient(ctx, projectId)
	client := Client{
		Firestore: fireClient,
		Ctx:       ctx,
	}

	return NewWithClient(ctx, projectId, rootDocPath, client)
}

// New creates a new FirestoreSync (UpdateSyncChecker) with a provided ClientInterface client.
// Note: Use this one for the tests with a mock ClientInterface.
func NewWithClient(ctx context.Context, projectId string, rootDocPath string, client ClientInterface) *FirestoreSync {
	return &FirestoreSync{
		ctx:      ctx,
		client:   client,
		rootPath: rootDocPath,
	}
}
