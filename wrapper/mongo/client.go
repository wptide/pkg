package mongo

import (
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/core/option"
	"github.com/mongodb/mongo-go-driver/mongo"
)

// Client interface describes a MongoDB client (a wrapper).
type Client interface {
	Close() error
	Database(string) DataLayer
}

// Wrapper wraps the client.
// The client gets shadowed when tested.
type Wrapper struct {
	ctx context.Context
	*mongo.Client
}

// NewMongoClient creates a new Wrapper.
func NewMongoClient(ctx context.Context, user string, pass string, host string, opts *mongo.ClientOptions) (Client, error) {

	var creds string

	if user != "" {
		creds = user
		if pass != "" {
			creds += ":" + pass
		}
		creds += "@"
	}

	client, err := mongo.NewClientWithOptions("mongodb://"+creds+host, opts)
	if err == nil {
		client.Connect(ctx)
	}
	return &Wrapper{ctx, client}, err
}

// Database returns a Database object. DataLayer shadows the mongo.Database.
func (mc Wrapper) Database(name string) DataLayer {
	return &WrapperDatabase{Database: mc.Client.Database(name)}
}

// Close closes the MongoDB connection.
func (mc Wrapper) Close() error {
	return mc.Client.Disconnect(mc.ctx)
}

// DataLayer wraps a CollectionLayer.
type DataLayer interface {
	Collection(name string) CollectionLayer
}

// WrapperDatabase wraps the mongo.Database.
type WrapperDatabase struct {
	*mongo.Database
}

// Collection returns a mongo collection. CollectionLayer shadows the mongo.Collection.
func (d WrapperDatabase) Collection(name string) CollectionLayer {
	return &WrapperCollection{Collection: d.Database.Collection(name)}
}

// CollectionLayer abstracts the required functions for MongoDB.
type CollectionLayer interface {
	InsertOne(ctx context.Context, document interface{}, opts ...option.InsertOneOptioner) (InsertOneResultLayer, error)
	FindOne(ctx context.Context, filter interface{}, opts ...option.FindOneOptioner) DocumentResultLayer
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...option.FindOneAndUpdateOptioner) DocumentResultLayer
	FindOneAndDelete(ctx context.Context, filter interface{}, opts ...option.FindOneAndDeleteOptioner) DocumentResultLayer
}

// WrapperCollection wraps mongo.Collection.
type WrapperCollection struct {
	*mongo.Collection
}

// InsertOne inserts a document.
func (c WrapperCollection) InsertOne(ctx context.Context, document interface{}, opts ...option.InsertOneOptioner) (InsertOneResultLayer, error) {
	var insertResult *WrapperInsertOneResult
	var err error

	// Recover on panic() from mongo driver.
	defer func() {
		if r := recover(); r != nil {
			insertResult = nil
			err = errors.New("mongodb: collection insert error")
		}
	}()

	res, err := c.Collection.InsertOne(ctx, document, opts...)
	insertResult = &WrapperInsertOneResult{res}
	return insertResult, err
}

// FindOne finds a document given filters and options.
func (c WrapperCollection) FindOne(ctx context.Context, filter interface{}, opts ...option.FindOneOptioner) DocumentResultLayer {
	var docResult *WrapperDocumentResult

	// Recover on panic() from mongo driver.
	defer func() {
		if r := recover(); r != nil {
			docResult = nil
		}
	}()

	docResult = &WrapperDocumentResult{c.Collection.FindOne(ctx, filter, opts...)}
	return docResult
}

// FindOneAndUpdate finds a document and updates it. Closest we have to a transaction.
func (c WrapperCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...option.FindOneAndUpdateOptioner) DocumentResultLayer {
	// No need to panic check. Fails gracefully.
	return &WrapperDocumentResult{c.Collection.FindOneAndUpdate(ctx, filter, update, opts...)}
}

// FindOneAndDelete finds a document and removes it.
func (c WrapperCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...option.FindOneAndDeleteOptioner) DocumentResultLayer {
	var docResult *WrapperDocumentResult

	// Recover on panic() from mongo driver.
	defer func() {
		if r := recover(); r != nil {
			docResult = nil
		}
	}()

	docResult = &WrapperDocumentResult{c.Collection.FindOneAndDelete(ctx, filter, opts...)}
	return docResult
}

// InsertOneResultLayer is an empty interface. No methods are required for this.
// Everything implements this.
type InsertOneResultLayer interface{}

// WrapperInsertOneResult wraps mongo.InsertOneResult.
type WrapperInsertOneResult struct {
	*mongo.InsertOneResult
}

// DocumentResultLayer abstracts the Decode() method.
type DocumentResultLayer interface {
	Decode() (*bson.Document, error)
}

// WrapperDocumentResult wraps mongo.DocumentResult.
type WrapperDocumentResult struct {
	*mongo.DocumentResult
}

// Decode decodes an element into a bson document.
// Note: Mongo's Decode() accepts a document as a parameter. This method uses
// Mongo's Decode(), but makes it simpler to test.
func (d WrapperDocumentResult) Decode() (*bson.Document, error) {
	elem := bson.NewDocument()
	err := d.DocumentResult.Decode(elem)
	return elem, err
}
