package sqs

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/wptide/pkg/message"
)

// Provider represents an SQS queue.
type Provider struct {
	session *session.Session
	// Use sqsiface instead of sqs.SQS to benefit from the interface.
	sqs       sqsiface.SQSAPI
	QueueURL  *string
	QueueName *string
}

// SendMessage implements the required interface method to be a Provider.
// This method sends a new SQS SendMessageInput message to SQS.
func (mgr Provider) SendMessage(msg *message.Message) error {

	// Encode the task to send as the message body.
	taskEncoded, _ := json.Marshal(msg)

	// Create the message object.
	messageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(taskEncoded)),
		QueueUrl:    mgr.QueueURL,
	}

	// Change message if .fifo queue
	var messageGroupID = fmt.Sprintf("%s-%s", msg.RequestClient, msg.Slug)
	if strings.HasSuffix(*mgr.QueueName, ".fifo") {
		messageInput.MessageGroupId = &messageGroupID
	} else {
		messageInput.DelaySeconds = aws.Int64(10)
	}

	// Send the message and check for errors.
	_, err := mgr.sqs.SendMessage(messageInput)

	if err != nil {
		return err
	}

	return nil
}

// GetNextMessage implements the required interface method to be a Provider.
// This method sends a ReceiveMessageInput message to SQS and converts the message into a *task.Task object.
func (mgr Provider) GetNextMessage() (*message.Message, error) {
	var returnMessage message.Message

	// Prepare the message
	messageInput := &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            mgr.QueueURL,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   aws.Int64(600), // 600 seconds : 10 minutes
		WaitTimeSeconds:     aws.Int64(0),
	}

	// Retrieve the message from SQS
	result, err := mgr.sqs.ReceiveMessage(messageInput)

	if err != nil {
		// If we get a critical AWS error, issue a new provider error.
		if awsErr, ok := err.(awserr.Error); ok {

			pErr := message.NewProviderError(awsErr.Message())
			if awsErr.Message() != sqs.ErrCodeOverLimit {
				pErr.Type = message.ErrCritcal
			} else {
				pErr.Type = message.ErrOverQuota
			}
			return nil, pErr
		}
		return nil, err
	}

	// Attempt to unmarshal the message body into the returnTask.
	if len(result.Messages) != 0 {
		body := result.Messages[0].Body
		err = json.Unmarshal([]byte(*body), &returnMessage)

		// Return the queue receipt so that the message can be deleted.
		returnMessage.ExternalRef = result.Messages[0].ReceiptHandle
		return &returnMessage, err
	}

	return nil, errors.New("could not retrieve message")
}

// DeleteMessage implements the required interface method to be a Provider.
// This method deletes a message from the queue.
func (mgr Provider) DeleteMessage(reference *string) error {
	_, err := mgr.sqs.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      mgr.QueueURL,
		ReceiptHandle: reference,
	})

	if err != nil {
		return err
	}

	return nil
}

// Close implemented to satisfy Provider interface.
func (mgr Provider) Close() error {
	return nil
}

// getQueueURL requests the queueURL from SQS.
func getQueueURL(svc sqsiface.SQSAPI, name string) (string, error) {
	result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})

	if err != nil {
		return "", err
	}
	return *result.QueueUrl, nil
}

// getSession establishes a new SQS session.
func getSession(region, key, secret string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	})

	return sess, err
}

// NewSqsProvider is a convenience method to return a new *Provider instance.
func NewSqsProvider(region, key, secret, queue string) *Provider {

	sess, _ := getSession(region, key, secret)
	svc := sqs.New(sess)
	queueURL, _ := getQueueURL(svc, queue)

	return &Provider{
		session:   sess,
		sqs:       svc,
		QueueURL:  &queueURL,
		QueueName: &queue,
	}
}
