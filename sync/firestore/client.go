package firestore

import (
	"context"
	"cloud.google.com/go/firestore"
)

type ClientInterface interface {
	GetDoc(path string) map[string]interface{}
	SetDoc(path string, data map[string]interface{}) error
	AddDoc(collection string, data map[string]interface{}) error
	Authenticated() bool
	Close() error
}

type Client struct {
	Firestore *firestore.Client
	Ctx       context.Context
}

func (c Client) GetDoc(path string) map[string]interface{} {
	doc := c.getDoc(path)
	docRef, _ := c.getDocRef(doc)
	return c.getDocData(docRef)
}

func (c Client) SetDoc(path string, data map[string]interface{}) error {
	doc := c.getDoc(path)
	// firestore.MergeAll "causes all the field paths ... to be overwritten." avoiding redundant read ops.
	_, err := doc.Set(c.Ctx, data, firestore.MergeAll)
	return err
}

func (c Client) AddDoc(collection string, data map[string]interface{}) error {
	_, _, err := c.Firestore.Collection(collection).Add(c.Ctx, data)
	return err
}

func (c Client) Close() error {
	return c.Firestore.Close()
}

func (c Client) getDoc(path string) *firestore.DocumentRef {
	return c.Firestore.Doc(path)
}

func (c Client) getDocRef(doc *firestore.DocumentRef) (*firestore.DocumentSnapshot, error) {
	return doc.Get(c.Ctx)
}

func (c Client) getDocData(ss *firestore.DocumentSnapshot) map[string]interface{} {
	return ss.Data()
}

func (c Client) Authenticated() bool {
	return c.Firestore != nil
}