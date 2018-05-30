package mongo

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"context"
	"github.com/mongodb/mongo-go-driver/core/options"
	"errors"
)

type Client interface {
	Database(string) DataLayer
}

type MongoClient struct {
	*mongo.Client
}

func NewMongoClient(user string, pass string, host string, opts *mongo.ClientOptions) (Client, error) {

	var creds string

	if user != "" {
		creds = user
		if pass != "" {
			creds += ":" + pass
		}
		creds += "@"
	}

	client, err := mongo.NewClientWithOptions("mongodb://"+creds+host, opts)
	return &MongoClient{client}, err
}

func (mc MongoClient) Database(name string) DataLayer {
	return &MongoDatabase{Database: mc.Client.Database(name)}
}

type DataLayer interface {
	Collection(name string) CollectionLayer
}

type MongoDatabase struct {
	*mongo.Database
}

func (d MongoDatabase) Collection(name string) CollectionLayer {
	return &MongoCollection{Collection: d.Database.Collection(name)}
}

type CollectionLayer interface {
	InsertOne(ctx context.Context, document interface{}, opts ...options.InsertOneOptioner) (InsertOneResultLayer, error)
	FindOne(ctx context.Context, filter interface{}, opts ...options.FindOneOptioner) DocumentResultLayer
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.FindOneAndUpdateOptioner) DocumentResultLayer
	FindOneAndDelete(ctx context.Context, filter interface{}, opts ...options.FindOneAndDeleteOptioner) DocumentResultLayer
}

type MongoCollection struct {
	*mongo.Collection
}

func (c MongoCollection) InsertOne(ctx context.Context, document interface{}, opts ...options.InsertOneOptioner) (InsertOneResultLayer, error) {
	var insertResult *MongoInsertOneResult
	var err error

	// Recover on panic() from mongo driver.
	defer func() {
		if r := recover(); r!= nil {
			insertResult = nil
			err = errors.New("Collection insert error.")
		}
	}()

	res, err := c.Collection.InsertOne(ctx, document, opts...)
	insertResult = &MongoInsertOneResult{res}
	return insertResult, err
}

func (c MongoCollection) FindOne(ctx context.Context, filter interface{}, opts ...options.FindOneOptioner) DocumentResultLayer {
	var docResult *MongoDocumentResult

	// Recover on panic() from mongo driver.
	defer func() {
		if r := recover(); r!= nil {
			docResult = nil
		}
	}()

	docResult = &MongoDocumentResult{c.Collection.FindOne(ctx, filter, opts...)}
	return docResult
}

func (c MongoCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.FindOneAndUpdateOptioner) DocumentResultLayer {
	// No need to panic check. Fails gracefully.
	return &MongoDocumentResult{c.Collection.FindOneAndUpdate(ctx, filter, update, opts...)}
}

func (c MongoCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...options.FindOneAndDeleteOptioner) DocumentResultLayer {
	var docResult *MongoDocumentResult

	// Recover on panic() from mongo driver.
	defer func() {
		if r := recover(); r!= nil {
			docResult = nil
		}
	}()

	docResult = &MongoDocumentResult{c.Collection.FindOneAndDelete(ctx, filter, opts...)}
	return docResult
}

type InsertOneResultLayer interface{}

type MongoInsertOneResult struct {
	*mongo.InsertOneResult
}

type DocumentResultLayer interface {
	Decode(v interface{}) error
}

type MongoDocumentResult struct {
	*mongo.DocumentResult
}

func (d MongoDocumentResult) Decode(v interface{}) error {
	return d.DocumentResult.Decode(v)
}