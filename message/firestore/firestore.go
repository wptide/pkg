package firestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/wptide/pkg/message"
	fsClient "github.com/wptide/pkg/wrapper/firestore"
)

const (
	// RetryAttempts sets the amount of default retries.
	RetryAttempts               = 3

	// LockDuration sets how long an item needs to be locked for.
	LockDuration  time.Duration = time.Minute * 10
)

// Provider implements the Provider interface.
type Provider struct {
	ctx      context.Context
	client   fsClient.ClientInterface
	rootPath string
}

// SendMessage sends a message to Firestore.
func (fs Provider) SendMessage(msg *message.Message) error {
	return fs.client.AddDoc(fs.rootPath, generateMessage(msg))
}

// GetNextMessage gets the next message from Firestore.
//
// This uses Firestore transactions to update the lock time and
// available retries for an item.
func (fs Provider) GetNextMessage() (*message.Message, error) {
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

// DeleteMessage deletes a Document from Firestore.
func (fs Provider) DeleteMessage(ref *string) error {
	return fs.client.DeleteDoc(fmt.Sprintf("%s/%s", fs.rootPath, *ref))
}

// Close the Firestore client.
func (fs Provider) Close() error {
	if fs.client != nil {
		return fs.client.Close()
	}
	return nil
}

// itom converts a Firestore Document into a QueueMessage.
func itom(data map[string]interface{}) *message.QueueMessage {

	var msg *message.QueueMessage

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
		"retries":         int64(RetryAttempts),
		"message":         msgMap,
		"status":          "pending",
		"retry_available": true,
	}
}

// New creates a new Sync (UpdateSyncChecker) with a default client
// using Firestore.
func New(ctx context.Context, projectID string, rootDocPath string) (*Provider, error) {

	fireClient, _ := firestore.NewClient(ctx, projectID)
	client := fsClient.Client{
		Firestore: fireClient,
		Ctx:       ctx,
	}

	return NewWithClient(ctx, projectID, rootDocPath, client)
}

// NewWithClient creates a new Sync (UpdateSyncChecker) with a provided ClientInterface client.
// Note: Use this one for the tests with a mock ClientInterface.
func NewWithClient(ctx context.Context, projectID string, rootDocPath string, client fsClient.ClientInterface) (*Provider, error) {
	if client == nil || !client.Authenticated() {
		return nil, errors.New("could not authenticate sync client")
	}

	return &Provider{
		ctx:      ctx,
		client:   client,
		rootPath: rootDocPath,
	}, nil
}
