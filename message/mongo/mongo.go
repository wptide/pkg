package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/wptide/pkg/message"
	wrapper "github.com/wptide/pkg/wrapper/mongo"
)

const (
	// RetryAttempts sets the amount of default retries.
	RetryAttempts               = 3

	// LockDuration sets how long an item needs to be locked for.
	LockDuration  time.Duration = time.Minute * 10
)

// Provider implements the Provider interface.
type Provider struct {
	ctx        context.Context
	client     wrapper.Client
	database   string
	collection string
}

// SendMessage sends a message to MongoDB.
func (m Provider) SendMessage(msg *message.Message) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	_, err := collection.InsertOne(context.Background(), generateMessage(msg))
	return err
}

// GetNextMessage gets the next message from MongoDB.
func (m Provider) GetNextMessage() (*message.Message, error) {
	collection := m.client.Database(m.database).Collection(m.collection)

	// Query.
	filter := map[string]interface{}{
		"retry_available": true,
		"lock": map[string]interface{}{
			"$lt": time.Now().UnixNano(),
		},
	}

	//sort by 'created' ASC. DESC takes `-1` for the second argument.
	sort, _ := mongo.Opt.Sort(bson.NewDocument(bson.EC.Int32("created", 1)))

	result := collection.FindOne(m.ctx, filter, sort)
	qm, err := ResultToQueueMessage(result)
	if err != nil {
		return nil, err

	}

	itemID, _ := objectid.FromHex(*qm.Message.ExternalRef)

	// Lock and update.
	filter = map[string]interface{}{
		"_id": itemID,
	}

	// Get retries.
	retryAvailable := true
	retries := qm.Retries - 1
	if retries <= 0 {
		retryAvailable = false
	}

	// Update data.
	updateData := map[string]interface{}{
		"$set": map[string]interface{}{
			"retries":         int64(retries),
			"retry_available": retryAvailable,
			"lock":            int64(time.Now().Add(LockDuration).UnixNano()),
		},
	}

	// Update item and get new reference.
	uqm, err := ResultToQueueMessage(collection.FindOneAndUpdate(context.Background(), filter, updateData))
	if err != nil {
		return nil, errors.New("mongodb: could not set lock on item")
	}

	return uqm.Message, nil
}

// DeleteMessage deletes a Document from MongoDB.
func (m Provider) DeleteMessage(ref *string) error {
	collection := m.client.Database(m.database).Collection(m.collection)

	itemID, _ := objectid.FromHex(*ref)
	filter := map[string]interface{}{
		"_id": itemID,
	}

	collection.FindOneAndDelete(m.ctx, filter)

	return nil
}

// Close the MongoDB client.
func (m Provider) Close() error {
	return m.client.Close()
}

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

// ResultToQueueMessage converts a MongoDB result to a QueueMessage.
func ResultToQueueMessage(layer wrapper.DocumentResultLayer) (*message.QueueMessage, error) {

	elem, _ := layer.Decode()
	raw, _ := elem.MarshalBSON()
	js, err := bson.ToExtJSON(false, raw)

	if err != nil || js == "{}" {
		return nil, errors.New("mongodb: no document found")
	}

	extRef := elem.Lookup("_id").ObjectID().Hex()

	var qm *message.QueueMessage
	json.Unmarshal([]byte(js), &qm)

	qm.Message.ExternalRef = &[]string{extRef}[0]

	return qm, nil
}

// New creates a new MongoDB (UpdateChecker) with a default client.
func New(ctx context.Context, user string, pass string, host string, db string, collection string, opts *mongo.ClientOptions) (*Provider, error) {
	client, err := wrapper.NewMongoClient(ctx, user, pass, host, opts)
	if err != nil {
		return nil, err
	}

	return NewWithClient(ctx, db, collection, client)
}

// NewWithClient creates a new MongoDB (UpdateChecker) with a provided ClientInterface client.
func NewWithClient(ctx context.Context, db string, collection string, client wrapper.Client) (*Provider, error) {
	return &Provider{
		ctx:        ctx,
		client:     client,
		database:   db,
		collection: collection,
	}, nil
}
