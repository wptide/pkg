package firestore

import (
	"github.com/wptide/pkg/message"
	fsClient "github.com/wptide/pkg/wrapper/firestore"
	"context"
	"errors"
	"cloud.google.com/go/firestore"
	"encoding/json"
	"time"
	"fmt"
)

const RetryAttemps = 3
const LockDuration time.Duration = time.Minute * 5

// QueueMessage defines the schema for how messages are stored in Firestore.
type QueueMessage struct {
	Created        int64            `json:"created" firestore:"created"`
	Lock           int64            `json:"lock" firestore:"lock"`
	Message        *message.Message `json:"message" firestore:"message"`
	Retries        int64            `json:"retries" firestore:"retries"`
	Status         string           `json:"status" firestore:"status"`
	RetryAvailable bool             `json:"retry_available" firestore:"retry_available"`
}

// FirestoreProvider implements the MessageProvider interface.
type FirestoreProvider struct {
	ctx      context.Context
	client   fsClient.ClientInterface
	rootPath string
}

// SendMessage sends a message to Firestore.
func (fs FirestoreProvider) SendMessage(msg *message.Message) error {
	return fs.client.AddDoc(fs.rootPath, generateMessage(msg))
}

// Get message gets the next message from Firestore.
//
// This uses Firestore transactions to update the lock time and
// available retries for an item.
func (fs FirestoreProvider) GetNextMessage() (*message.Message, error) {
	items, err := fs.client.QueryItems(
		// Collection to get the message from.
		fs.rootPath,
		// Conditions provided to the client query.
		[]fsClient.Condition{
			{"retry_available", "==", true},
			{"lock", "<", time.Now().UnixNano()},
		},
		// Order parameters for the results.
		[]fsClient.Order{
			{"lock", "asc"},
			{"created", "asc"},
		},
		// Number of Documents to fetch. 1 only for next available message.
		1,
		// Update callback. This updates the given data map with new values
		// to update the Document during the transaction.
		func(data map[string]interface{}) (map[string]interface{}, error) {

			// Decrease retries and setting to false if required.
			retryAvailable := true
			retries := data["retries"].(int64) - 1
			if retries == 0 {
				retryAvailable = false
			}

			// Update retry fields and lock.
			out := map[string]interface{}{
				"retries":         retries,
				"retry_available": retryAvailable,
				"lock":            time.Now().Add(LockDuration).UnixNano(),
			}
			return out, nil
		},
	)

	var msg *message.Message

	if len(items) > 0 {
		// Convert the data (interface map) to a QueueMessage object.
		qmsg := itom(items[0].(map[string]interface{}))

		// Get the message to process.
		msg = qmsg.Message

		// If an "_id" is set, which it should, this becomes an ExternalRef.
		if ref, ok := items[0].(map[string]interface{})["_id"].(string); ok {
			msg.ExternalRef = &ref
		}
	}

	return msg, err
}

// Delete a Document from Firestore.
func (fs FirestoreProvider) DeleteMessage(ref *string) error {
	return fs.client.DeleteDoc(fmt.Sprintf("%s/%s", fs.rootPath, *ref))
}

// Close the Firestore client.
func (fs FirestoreProvider) Close() error {
	if fs.client != nil {
		return fs.client.Close()
	}
	return nil
}

// itom converts a Firestore Document into a QueueMessage.
func itom(data map[string]interface{}) *QueueMessage {

	var msg *QueueMessage

	if temp, err := json.Marshal(data); err == nil {
		json.Unmarshal(temp, &msg)
	}

	return msg
}

// generateMessages generates a new interface map given *message.Message.
func generateMessage(in *message.Message) map[string]interface{} {

	// Convert the struct into an interface map.
	var msgMap map[string]interface{}
	msg, _ := json.Marshal(in)
	json.Unmarshal(msg, &msgMap)

	// Return the QueueMessage as an interface map.
	return map[string]interface{}{
		"created":         time.Now().UnixNano(),
		"lock":            int64(0),
		"retries":         int64(RetryAttemps),
		"message":         msgMap,
		"status":          "pending",
		"retry_available": true,
	}
}

// New creates a new FirestoreSync (UpdateSyncChecker) with a default client
// using Firestore.
func New(ctx context.Context, projectId string, rootDocPath string) (*FirestoreProvider, error) {

	fireClient, _ := firestore.NewClient(ctx, projectId)
	client := fsClient.Client{
		Firestore: fireClient,
		Ctx:       ctx,
	}

	return NewWithClient(ctx, projectId, rootDocPath, client)
}

// New creates a new FirestoreSync (UpdateSyncChecker) with a provided ClientInterface client.
// Note: Use this one for the tests with a mock ClientInterface.
func NewWithClient(ctx context.Context, projectId string, rootDocPath string, client fsClient.ClientInterface) (*FirestoreProvider, error) {
	if ! client.Authenticated() {
		return nil, errors.New("Could not authenticate sync client.")
	}

	return &FirestoreProvider{
		ctx:      ctx,
		client:   client,
		rootPath: rootDocPath,
	}, nil
}