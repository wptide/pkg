package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
)

func getDoc(ctx context.Context, path string, client *firestore.Client) map[string]interface{} {
	doc, _ := getDocRef(ctx, path, client)
	return getDocData(doc)
}

func getDocRef(ctx context.Context, path string, client *firestore.Client) (*firestore.DocumentSnapshot, error) {
	return client.Doc(path).Get(ctx)
}

func getDocData(ss *firestore.DocumentSnapshot) map[string]interface{} {
	return ss.Data()
}

func setDoc(ctx context.Context, path string, client *firestore.Client, data map[string]interface{}) error {
	_, err := client.Doc(path).Set(ctx, data)
	return err
}

func addDoc(ctx context.Context, collection string, client *firestore.Client, data map[string]interface{}) error {
	_, _, err := client.Collection(collection).Add(ctx, data)
	return err
}
