package firestore

import (
	"context"
	"cloud.google.com/go/firestore"
)

type ClientInterface interface {
	GetDoc(path string) map[string]interface{}
	SetDoc(path string, data map[string]interface{}) error
	AddDoc(collection string, data map[string]interface{}) error
	Close() error
}

type Client struct {
	Firestore *firestore.Client
	Ctx       context.Context
}

func (c Client) GetDoc(path string) map[string]interface{} {
	doc, _ := c.getDocRef(path)
	return c.getDocData(doc)
}

func (c Client) SetDoc(path string, data map[string]interface{}) error {
	_, err := c.Firestore.Doc(path).Set(c.Ctx, data)
	return err
}

func (c Client) AddDoc(collection string, data map[string]interface{}) error {
	_, _, err := c.Firestore.Collection(collection).Add(c.Ctx, data)
	return err
}

func (c Client) Close() error {
	return c.Firestore.Close()
}

func (c Client) getDocRef(path string) (*firestore.DocumentSnapshot, error) {
	return c.Firestore.Doc(path).Get(c.Ctx)
}

func (c Client) getDocData(ss *firestore.DocumentSnapshot) map[string]interface{} {
	return ss.Data()
}
