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
		Title:         fallbackValue(simpleCodeInfo.Name, msg.Title).(string),
		Description:   fallbackValue(simpleCodeInfo.Description, msg.Content).(string),
		Version:       simpleCodeInfo.Version,
		Checksum:      data["checksum"].(string),
		Visibility:    msg.Visibility,
		ProjectType:   fallbackValue(codeInfo.Type, msg.ProjectType).(string),
		SourceUrl:     msg.SourceURL,
		SourceType:    msg.SourceType,
		CodeInfo:      codeInfo,
		Results:       results,
		Standards:     msg.Standards,
		RequestClient: msg.RequestClient,
	}

	if msg.Slug != "" {
		payloadItem.Project = []string{msg.Slug}
	}

	return json.Marshal(payloadItem)
}

func fallbackValue(value ...interface{}) interface{} {

	for i, val := range value {

		switch val.(type) {
		case tide.CodeInfo:
			if val.(tide.CodeInfo).Type != "" {
				return val.(tide.CodeInfo)
			}
			if i == (len(value) - 1) {
				return nil
			}
		case int64:
			if val.(int64) != 0 {
				return val.(int64)
			}
			if i == (len(value) - 1) {
				return int64(0)
			}
		case int32:
			if val.(int32) != 0 {
				return val.(int32)
			}
			if i == (len(value) - 1) {
				return int32(0)
			}
		case int:
			if val.(int) != 0 {
				return val.(int)
			}
			if i == (len(value) - 1) {
				return 0
			}
		case string:
			if val.(string) != "" {
				return val.(string)
			}
			if i == (len(value) - 1) {
				return ""
			}
		case float64:
			if val.(float64) != 0.0 {
				return val.(float64)
			}
			if i == (len(value) - 1) {
				return float64(0.0)
			}
		case float32:
			if val.(float32) != 0.0 {
				return val.(float32)
			}
			if i == (len(value) - 1) {
				return float32(0.0)
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
