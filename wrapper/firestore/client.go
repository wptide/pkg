package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"strings"
)

// Condition describes a Query condition.
type Condition struct {
	Path     string
	Operator string
	Value    interface{}
}

// Order describes a Query order.
type Order struct {
	Field     string
	Direction string
}

// UpdateFunc describes a method signature for an update function.
type UpdateFunc func(data map[string]interface{}) (map[string]interface{}, error)

// ClientInterface is the Firestore client interface.
type ClientInterface interface {
	GetDoc(path string) map[string]interface{}
	SetDoc(path string, data map[string]interface{}) error
	AddDoc(collection string, data interface{}) error
	Authenticated() bool
	Close() error
	QueryItems(collection string, conditions []Condition, ordering []Order, limit int, updateFunc UpdateFunc) ([]interface{}, error)
	DeleteDoc(path string) error
}

// Client wraps the Firestore client.
type Client struct {
	Firestore *firestore.Client
	Ctx       context.Context
}

// GetDoc gets a Firestore document.
func (c Client) GetDoc(path string) map[string]interface{} {
	docRef, _ := c.Firestore.Doc(path).Get(c.Ctx)
	return c.getDocData(docRef)
}

// SetDoc writes a Firestore document.
func (c Client) SetDoc(path string, data map[string]interface{}) error {
	// firestore.MergeAll "causes all the field paths ... to be overwritten." avoiding redundant read ops.
	_, err := c.Firestore.Doc(path).Set(c.Ctx, data, firestore.MergeAll)
	return err
}

// AddDoc creates a Firestore document.
func (c Client) AddDoc(collection string, data interface{}) error {
	_, _, err := c.Firestore.Collection(collection).Add(c.Ctx, data)
	return err
}

// Close the Firestore client.
func (c Client) Close() error {
	return c.Firestore.Close()
}

func (c Client) getDocData(ss *firestore.DocumentSnapshot) map[string]interface{} {
	if ss == nil {
		return nil
	}
	return ss.Data()
}

// Authenticated returns true if client is authenticated.
func (c Client) Authenticated() bool {
	return c.Firestore != nil
}

// QueryItems queries the Firestore database.
func (c Client) QueryItems(
	collection string,
	conditions []Condition,
	ordering []Order,
	limit int,
	updateFunc UpdateFunc,
) ([]interface{}, error) {
	if len(conditions) == 0 {
		return nil, errors.New("firestore query: must provide conditions")
	}

	colRef := c.Firestore.Collection(collection)
	var query firestore.Query
	querySet := false
	for _, condition := range conditions {
		if !querySet {
			query = colRef.Where(condition.Path, condition.Operator, condition.Value)
			querySet = true
		} else {
			query = query.Where(condition.Path, condition.Operator, condition.Value)
		}
	}

	for _, order := range ordering {
		var direction = firestore.Asc
		if strings.ToLower(order.Direction) == "desc" {
			direction = firestore.Desc
		}
		query = query.OrderBy(order.Field, direction)
	}

	if limit != 0 {
		query = query.Limit(limit)
	}

	var items []interface{}
	txErr := c.Firestore.RunTransaction(c.Ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		// Pass query to transaction.
		docs, err := tx.Documents(query).GetAll()
		if err != nil {
			return err
		}

		// Don't error, but still exit.
		if len(docs) == 0 {
			return nil
		}

		// Iterate over each document.
		for _, doc := range docs {
			docData := doc.Data()

			// If update function is provided then set the doc with new data.
			if updateFunc != nil {
				data, err := updateFunc(docData)
				if err != nil {
					return err
				}
				err = tx.Set(doc.Ref, data, firestore.MergeAll)
				if err != nil {
					return err
				}

				// ... and update the already loaded data.
				for key, val := range data {
					docData[key] = val
				}
			}

			docData["_id"] = doc.Ref.ID

			// Add fetched item to the items array.
			items = append(items, docData)
		}

		return nil
	})

	return items, txErr
}

// DeleteDoc deletes a Firestore document.
func (c Client) DeleteDoc(path string) error {
	_, err := c.Firestore.Doc(path).Delete(c.Ctx)
	return err
}
