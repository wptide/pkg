package mongo

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/core/option"
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/bson"
)

type Client interface {
	Close() error
	Database(string) DataLayer
}

type MongoClient struct {
	ctx context.Context
	*mongo.Client
}

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
	return &MongoClient{ctx, client}, err
}

func (mc MongoClient) Database(name string) DataLayer {
	return &MongoDatabase{Database: mc.Client.Database(name)}
}

func (mc MongoClient) Close() error {
	return mc.Client.Disconnect(mc.ctx)
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
	InsertOne(ctx context.Context, document interface{}, opts ...option.InsertOneOptioner) (InsertOneResultLayer, error)
	FindOne(ctx context.Context, filter interface{}, opts ...option.FindOneOptioner) DocumentResultLayer
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...option.FindOneAndUpdateOptioner) DocumentResultLayer
	FindOneAndDelete(ctx context.Context, filter interface{}, opts ...option.FindOneAndDeleteOptioner) DocumentResultLayer
}

type MongoCollection struct {
	*mongo.Collection
}

func (c MongoCollection) InsertOne(ctx context.Context, document interface{}, opts ...option.InsertOneOptioner) (InsertOneResultLayer, error) {
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

func (c MongoCollection) FindOne(ctx context.Context, filter interface{}, opts ...option.FindOneOptioner) DocumentResultLayer {
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

func (c MongoCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...option.FindOneAndUpdateOptioner) DocumentResultLayer {
	// No need to panic check. Fails gracefully.
	return &MongoDocumentResult{c.Collection.FindOneAndUpdate(ctx, filter, update, opts...)}
}

func (c MongoCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...option.FindOneAndDeleteOptioner) DocumentResultLayer {
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
	Decode() (*bson.Document, error)
}

type MongoDocumentResult struct {
	*mongo.DocumentResult
}

func (d MongoDocumentResult) Decode() (*bson.Document, error) {
	elem := bson.NewDocument()
	err := d.DocumentResult.Decode(elem)
	return elem, err
}
