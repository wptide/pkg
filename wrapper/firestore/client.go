package firestore

import (
	"context"
	"cloud.google.com/go/firestore"
	"errors"
	"strings"
)

type Condition struct {
	Path     string
	Operator string
	Value    interface{}
}

type Order struct {
	Field     string
	Direction string
}

type UpdateFunc func(data map[string]interface{}) (map[string]interface{}, error)

type ClientInterface interface {
	GetDoc(path string) map[string]interface{}
	SetDoc(path string, data map[string]interface{}) error
	AddDoc(collection string, data interface{}) error
	Authenticated() bool
	Close() error
	QueryItems(collection string, conditions []Condition, ordering []Order, limit int, updateFunc UpdateFunc) ([]interface{}, error)
	DeleteDoc(path string) error
}

type Client struct {
	Firestore *firestore.Client
	Ctx       context.Context
}

func (c Client) GetDoc(path string) map[string]interface{} {
	docRef, _ := c.Firestore.Doc(path).Get(c.Ctx)
	return c.getDocData(docRef)
}

func (c Client) SetDoc(path string, data map[string]interface{}) error {
	// firestore.MergeAll "causes all the field paths ... to be overwritten." avoiding redundant read ops.
	_, err := c.Firestore.Doc(path).Set(c.Ctx, data, firestore.MergeAll)
	return err
}

func (c Client) AddDoc(collection string, data interface{}) error {
	_, _, err := c.Firestore.Collection(collection).Add(c.Ctx, data)
	return err
}

func (c Client) Close() error {
	return c.Firestore.Close()
}

func (c Client) getDocData(ss *firestore.DocumentSnapshot) map[string]interface{} {
	if ss == nil {
		return nil
	}
	return ss.Data()
}

func (c Client) Authenticated() bool {
	return c.Firestore != nil
}

func (c Client) QueryItems(
	collection string,
	conditions []Condition,
	ordering []Order,
	limit int,
	updateFunc UpdateFunc,
) ([]interface{}, error) {
	if len(conditions) == 0 {
		return nil, errors.New("Must provide conditions.")
	}

	colRef := c.Firestore.Collection(collection)
	var query firestore.Query
	querySet := false
	for _, condition := range conditions {
		if ! querySet {
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

func (c Client) DeleteDoc(path string) error {
	_, err := c.Firestore.Doc(path).Delete(c.Ctx)
	return err
}