package payload

import (
	"github.com/wptide/pkg/tide"
	"errors"
	"encoding/json"
	"github.com/wptide/pkg/message"
)

type TidePayload struct {
	Client tide.ClientInterface
}

func (t TidePayload) BuildPayload(msg message.Message, data map[string]interface{}) ([]byte, error) {

	codeInfo, ok := data["info"].(tide.CodeInfo)
	if !ok {
		return nil, errors.New("Code info not found")
	}

	simpleCodeInfo := tide.SimplifyCodeDetails(codeInfo.Details)

	// Loop through tc.Result to get all `AuditResult`s.
	results := make(map[string]tide.AuditResult)
	for key, result := range data {

		r, ok := result.(tide.AuditResult)
		if !ok {
			continue
		}

		results[key] = r
	}

	if len(results) == 0 {
		return nil, errors.New("no results to send to Tide API")
	}

	payloadItem := &tide.Item{
		fallbackValue(simpleCodeInfo.Name, msg.Title).(string),
		fallbackValue(simpleCodeInfo.Description, msg.Content).(string),
		simpleCodeInfo.Version,
		data["checksum"].(string),
		msg.Visibility,
		fallbackValue(codeInfo.Type, msg.ProjectType).(string),
		msg.SourceURL,
		msg.SourceType,
		codeInfo,
		results,
	}

	return json.Marshal(payloadItem)
}

func fallbackValue(value ...interface{}) interface{} {

	for _, val := range value {

		switch val.(type) {
		case tide.CodeInfo:
			if val.(tide.CodeInfo).Type != "" {
				return val.(tide.CodeInfo)
			}
		case int64:
			if val.(int64) != 0 {
				return val.(int64)
			}
		case int32:
			if val.(int32) != 0 {
				return val.(int32)
			}
		case int:
			if val.(int) != 0 {
				return val.(int)
			}
		case string:
			if val.(string) != "" {
				return val.(string)
			}
		case float64:
			if val.(float64) != 0.0 {
				return val.(float64)
			}
		case float32:
			if val.(float32) != 0.0 {
				return val.(float32)
			}
		default:
			return val
		}
	}

	return nil
}

func (t TidePayload) SendPayload(destination string, payload []byte) ([]byte, error) {

	reply, err := t.Client.SendPayload("POST", destination, string(payload))

	if err != nil {
		return nil, err
	}

	return []byte(reply), err
}
