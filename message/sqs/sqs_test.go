package sqs

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/wptide/pkg/message"
)

type mockSqs struct {
	sqsiface.SQSAPI
	sendMessageOutput   *sqs.SendMessageOutput
	deleteMessageOutput *sqs.DeleteMessageOutput
}

var (
	// Provider to mock fifo queue.
	testQueue    = "test.fifo"
	testQueueURL = "http://sqsurl/test.fifo"
	testProvider = Provider{
		session:   &session.Session{},
		sqs:       &mockSqs{},
		QueueName: &testQueue,
		QueueURL:  &testQueueURL,
	}

	// Provider to mock non-fifo queue.
	testNonFifoQueue    = "test"
	testNonFifoQueueURL = "http://sqsurl/test"
	testNonFifoProvider = Provider{
		session:   &session.Session{},
		sqs:       &mockSqs{},
		QueueName: &testNonFifoQueue,
		QueueURL:  &testNonFifoQueueURL,
	}

	// Provider to mock an error response.
	failQueue    = "fail.fifo"
	failQueueURL = "http://sqsurl/fail.fifo"
	failProvider = Provider{
		session:   &session.Session{},
		sqs:       &mockSqs{},
		QueueName: &failQueue,
		QueueURL:  &failQueueURL,
	}

	// Provider to mock an empty response.
	emptyQueue    = "empty.fifo"
	emptyQueueURL = "http://sqsurl/empty.fifo"
	emptyProvider = Provider{
		session:   &session.Session{},
		sqs:       &mockSqs{},
		QueueName: &emptyQueue,
		QueueURL:  &emptyQueueURL,
	}

	// Provider to mock an over limit response.
	limitQueueURL = "http://sqsurl/limit.fifo"
	limitProvider = Provider{
		session:   &session.Session{},
		sqs:       &mockSqs{},
		QueueName: &emptyQueue,
		QueueURL:  &limitQueueURL,
	}

	// Provider to mock an over limit response.
	errorQueueURL = "http://sqsurl/error.fifo"
	errorProvider = Provider{
		session:   &session.Session{},
		sqs:       &mockSqs{},
		QueueName: &emptyQueue,
		QueueURL:  &errorQueueURL,
	}
)

func (m mockSqs) DeleteMessage(in *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {

	if *in.ReceiptHandle == "fail-id" {
		return m.deleteMessageOutput, errors.New("something went wrong")
	}
	// contains filtered or unexported fields - so will be an empty struct of sqs.DeleteMessageOutput
	// https://docs.aws.amazon.com/sdk-for-go/api/service/sqs/#DeleteMessageOutput
	return m.deleteMessageOutput, nil
}

func (m mockSqs) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {

	var messages []*sqs.Message

	switch *in.QueueUrl {
	case failQueueURL:
		return nil, awserr.New("Provider Error", "Provider Error", errors.New("provider error"))
	case errorQueueURL:
		return nil, errors.New("other error")
	case emptyQueueURL:
		// Do nothing here.
	case limitQueueURL:
		return nil, awserr.New(sqs.ErrCodeOverLimit, sqs.ErrCodeOverLimit, errors.New(sqs.ErrCodeOverLimit))
	default:
		fake := message.Message{
			Title: "Success!",
		}

		bodyBytes, _ := json.Marshal(fake)
		body := string(bodyBytes)
		msg := &sqs.Message{
			Body: &body,
		}
		messages = append(messages, msg)
	}

	out := &sqs.ReceiveMessageOutput{
		Messages: messages,
	}

	return out, nil
}

func (m mockSqs) SendMessage(in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {

	var msg *message.Message
	err := json.Unmarshal([]byte(*in.MessageBody), &msg)
	if err != nil {
		return m.sendMessageOutput, err
	}

	if msg.Title == "FAIL" {
		return m.sendMessageOutput, errors.New("something went wrong")
	}

	return m.sendMessageOutput, nil
}

// Note: Must be GetQueueUrl to implement sqsiface.SQSAPI.
//       DO NOT change to GetQueueURL.
//       Run golint with `golint -min_confidence=0.9`
func (m mockSqs) GetQueueUrl(in *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	url := "http://sqsurl/" + *in.QueueName
	return &sqs.GetQueueUrlOutput{
		QueueUrl: &url,
	}, nil
}

func TestSqsProvider_SendMessage(t *testing.T) {
	type args struct {
		msg *message.Message
	}
	tests := []struct {
		name    string
		mgr     Provider
		args    args
		wantErr bool
	}{
		{
			name: "Test Simple Message - FIFO",
			mgr:  testProvider,
			args: args{
				&message.Message{},
			},
			wantErr: false,
		},
		{
			name: "Test Simple Message - Non-FIFO",
			mgr:  testNonFifoProvider,
			args: args{
				&message.Message{},
			},
			wantErr: false,
		},
		{
			name: "Test Failed Message",
			mgr:  testNonFifoProvider,
			args: args{
				&message.Message{
					Title: "FAIL",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mgr.SendMessage(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Provider.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSqsProvider_GetNextMessage(t *testing.T) {
	tests := []struct {
		name    string
		mgr     Provider
		want    *message.Message
		wantErr bool
	}{
		{
			name: "Get Simple Message",
			mgr:  testProvider,
			want: &message.Message{
				Title: "Success!",
			},
			wantErr: false,
		},
		{
			name:    "Failed Message",
			mgr:     failProvider,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty Message",
			mgr:     emptyProvider,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Provider Over Limits",
			mgr:     limitProvider,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Other Error",
			mgr:     errorProvider,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mgr.GetNextMessage()
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.GetNextMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.GetNextMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqsProvider_DeleteMessage(t *testing.T) {
	type args struct {
		reference *string
	}

	successID := "success-id"
	failID := "fail-id"

	tests := []struct {
		name    string
		mgr     Provider
		args    args
		wantErr bool
	}{
		{
			name: "Test Delete Message",
			mgr:  testProvider,
			args: args{
				reference: &successID,
			},
			wantErr: false,
		},
		{
			name: "Test Delete Message Error",
			mgr:  testProvider,
			args: args{
				reference: &failID,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mgr.DeleteMessage(tt.args.reference); (err != nil) != tt.wantErr {
				t.Errorf("Provider.DeleteMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getSession(t *testing.T) {
	type args struct {
		region string
		key    string
		secret string
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		{
			name: "Dummy Session",
			args: args{},
			want: reflect.TypeOf(&session.Session{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := getSession(tt.args.region, tt.args.key, tt.args.secret)
			if !reflect.DeepEqual(reflect.TypeOf(got), tt.want) {
				t.Errorf("getSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSqsProvider(t *testing.T) {
	type args struct {
		region string
		key    string
		secret string
		queue  string
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		{
			name: "Test Provider Client",
			args: args{
				region: "us-west-2",
				key:    "random-key",
				secret: "so-secret",
				queue:  "test.fifo",
			},
			want: reflect.TypeOf(&Provider{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSqsProvider(tt.args.region, tt.args.key, tt.args.secret, tt.args.queue); !reflect.DeepEqual(reflect.TypeOf(got), tt.want) {
				t.Errorf("NewSqsProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getQueueUrl(t *testing.T) {
	type args struct {
		svc  sqsiface.SQSAPI
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test getQueueURL",
			args: args{
				svc:  &mockSqs{},
				name: "test.fifo",
			},
			want: "http://sqsurl/test.fifo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := getQueueURL(tt.args.svc, tt.args.name)
			if got != tt.want {
				t.Errorf("getQueueURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqsProvider_Close(t *testing.T) {
	type fields struct {
		session   *session.Session
		sqs       sqsiface.SQSAPI
		QueueURL  *string
		QueueName *string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Close()",
			fields{
				nil,
				nil,
				nil,
				nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := Provider{
				session:   tt.fields.session,
				sqs:       tt.fields.sqs,
				QueueURL:  tt.fields.QueueURL,
				QueueName: tt.fields.QueueName,
			}
			if err := mgr.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Provider.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
