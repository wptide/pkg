package firestore

import (
	"errors"

	"github.com/wptide/pkg/message"
	fsClient "github.com/wptide/pkg/wrapper/firestore"
)

type mockClient struct {
}

func (m mockClient) GetDoc(path string) map[string]interface{} {
	return nil
}

func (m mockClient) SetDoc(path string, data map[string]interface{}) error {
	return nil
}

func (m mockClient) AddDoc(collection string, data interface{}) error {

	switch collection {
	case "test-fail":
		return errors.New("something went wrong")
	default:
		return nil
	}
}

func (m mockClient) Authenticated() bool {
	return true
}

func (m mockClient) Close() error {
	return nil
}

func (m mockClient) QueryItems(collection string, conditions []fsClient.Condition, ordering []fsClient.Order, limit int, updateFunc fsClient.UpdateFunc) ([]interface{}, error) {

	simpleMessage := func(retries int, id string) []interface{} {
		docData := map[string]interface{}{
			"retries": int64(retries),
			"message": message.Message{
				Title: "Simple Message",
			},
		}
		data, _ := updateFunc(docData)
		for key, val := range data {
			docData[key] = val
		}

		if id != "" {
			docData["_id"] = id
		}

		return []interface{}{
			docData,
		}
	}

	type fake struct {
		test interface{}
	}
	type fakeToo struct {
		test interface{}
	}

	switch collection {
	case "simple-message":
		return simpleMessage(5, ""), nil
	case "last-retry":
		return simpleMessage(1, ""), nil
	case "with-id":
		return simpleMessage(1, "ABC123"), nil
	default:
		return nil, nil
	}
}

func (m mockClient) DeleteDoc(path string) error {
	return nil
}
