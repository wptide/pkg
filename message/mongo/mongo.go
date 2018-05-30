package mongo

import (
	"github.com/wptide/pkg/message"
	"context"
	"github.com/mongodb/mongo-go-driver/mongo"
	"encoding/json"
	"time"
	"errors"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
)

const RetryAttemps = 3
const LockDuration time.Duration = time.Minute * 5

type MongoProvider struct {
	ctx        context.Context
	client     Client
	database   string
	collection string
}

func (m MongoProvider) SendMessage(msg *message.Message) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	_, err := collection.InsertOne(context.Background(), generateMessage(msg))
	return err
}

func (m MongoProvider) GetNextMessage() (*message.Message, error) {
	collection := m.client.Database(m.database).Collection(m.collection)

	// Query.
	filter := map[string]interface{}{
		"retry_available": true,
		"lock": map[string]interface{}{
			"$lt": time.Now().UnixNano(),
		},
	}

	//sort by 'created' ASC. DESC takes `-1` for the second argument.
	sort, _ := mongo.Opt.Sort(map[string]interface{}{
		"created": int32(1),
	})

	var ptr interface{}

	collection.FindOne(context.Background(), filter, sort).Decode(&ptr)
	item, ok := ptr.(map[string]interface{})
	if !ok {
		return nil, errors.New("No matching documents.")
	}

	itemID := item["_id"].(objectid.ObjectID)

	// Lock and update.
	filter = map[string]interface{}{
		"_id": itemID,
	}

	// Get retries.
	retryAvailable := true
	retries := item["retries"].(int64) - 1
	if (retries <= 0) {
		retryAvailable = false
	}

	//
	// Update data.
	updateData := map[string]interface{}{
		"$set": map[string]interface{}{
			"retries":         int64(retries),
			"retry_available": retryAvailable,
			"lock":            int64(time.Now().Add(LockDuration).UnixNano()),
		},
	}

	// Update item and get new reference.
	var updatePtr interface{}
	collection.FindOneAndUpdate(context.Background(), filter, updateData).Decode(&updatePtr)

	updatedItem, ok := updatePtr.(map[string]interface{})
	if !ok {
		return nil, errors.New("MongoDB: Could not set lock on item.")
	}

	var msg *message.Message

	messageRaw, ok := updatedItem["message"].(map[string]interface{})
	if ! ok {
		return nil, errors.New("MongoDB: Could not retrieve message.")
	}

	messageJson, _ := json.Marshal(messageRaw)
	json.Unmarshal(messageJson, &msg)

	msg.ExternalRef = &[]string{itemID.Hex()}[0]

	return msg, nil
}

func (m MongoProvider) DeleteMessage(ref *string) error {
	collection := m.client.Database(m.database).Collection(m.collection)

	itemID, _ := objectid.FromHex(*ref)
	filter := map[string]interface{}{
		"_id": itemID,
	}

	collection.FindOneAndDelete(m.ctx, filter)

	return nil
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
		"retries":         int64(RetryAttemps),
		"message":         msgMap,
		"status":          "pending",
		"retry_available": true,
	}
}

func New(ctx context.Context, user string, pass string, host string, db string, collection string, opts *mongo.ClientOptions) (*MongoProvider, error) {
	client, err := NewMongoClient(user, pass, host, opts)
	if err != nil {
		return nil, err
	}

	return NewWithClient(ctx, db, collection, client)
}

func NewWithClient(ctx context.Context, db string, collection string, client Client) (*MongoProvider, error) {
	return &MongoProvider{
		ctx:        ctx,
		client:     client,
		database:   db,
		collection: collection,
	}, nil
}
