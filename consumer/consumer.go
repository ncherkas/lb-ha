package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	// 10 is max number supported by SQS
	MaxNumberOfMessages = 10

	// Supported commands
	AddCommand    = "addItem"
	DeleteCommand = "deleteItem"
	GetCommand    = "getItem"
	GetAllCommand = "getAllItems"
)

// commandMsg is a structure unmarshalled from JSON messages received over SQS
type commandMsg struct {
	Command string `json:"command"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

func (msg commandMsg) validate() (err error) {
	if msg.Command == "" {
		err = errors.New("command is empty")
	} else if msg.Command != GetAllCommand && msg.Key == "" {
		err = errors.New("key is empty")
	}
	return
}

// SQS lists API methods used by the consumer
type SQS interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// Store is a common store interface
type Store interface {
	Add(key, val string, timestamp int64)
	Delete(key string)
	Get(key string) (found bool, val string, timestamp int64)
	DumpAll()
}

// Consumer continuosly receives and executes messages from the queue
type Consumer struct {
	sqs      SQS
	queueURL string
	store    Store
}

func New(sqs SQS, queueURL string, store Store) *Consumer {
	return &Consumer{
		sqs:      sqs,
		queueURL: queueURL,
		store:    store,
	}
}

// Run messages consumption given the context ctx and SQS long polling time rcvWaitTime specified in seconds
func (c *Consumer) Run(ctx context.Context, rcvWaitTime int32) error {
	for {
		log.Printf("[INFO] Receiving messages (wait time %d sec.)...\n", rcvWaitTime)
		rcvInput := &sqs.ReceiveMessageInput{QueueUrl: &c.queueURL, MaxNumberOfMessages: MaxNumberOfMessages, WaitTimeSeconds: rcvWaitTime}
		rcvOutput, err := c.sqs.ReceiveMessage(ctx, rcvInput)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			return fmt.Errorf("failed to receive messages: %w", err)
		}
		log.Println("[INFO] Received", len(rcvOutput.Messages), "messages")

		for _, msg := range rcvOutput.Messages {
			body := *msg.Body
			cm := new(commandMsg)
			if err := json.Unmarshal([]byte(body), cm); err != nil {
				log.Println("[ERROR] Unsupported message format:", err.Error())
				continue
			}
			log.Printf("[DEBUG] Message: %#v\n", cm)
			if err := cm.validate(); err != nil {
				log.Println("[ERROR] Message is invalid:", err.Error())
				continue
			}
			key := cm.Key
			timestamp := time.Now().UnixMicro()
			go func(receiptHandle string) {
				switch cm.Command {
				case AddCommand:
					c.store.Add(key, cm.Value, timestamp)
					log.Println("[DEBUG] Stored value under key", key)
				case DeleteCommand:
					c.store.Delete(key)
					log.Println("[DEBUG] Deleted value under key", key)
				case GetCommand:
					if ok, v, _ := c.store.Get(key); ok {
						log.Printf("[DEBUG] Got value %s under key %s\n", v, key)
					} else {
						log.Println("[DEBUG] Value under key", key, "not found")
					}
				case GetAllCommand:
					log.Println("[DEBUG] Getting all entries...")
					c.store.DumpAll()
				default:
					log.Println("[WARN] Unsupported command:", cm.Command)
				}

				// A separate context because we want it to delete the message despite a shutdown
				_, err := c.sqs.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{QueueUrl: &c.queueURL, ReceiptHandle: &receiptHandle})
				if err != nil {
					log.Printf("[ERROR] failed to delete message %s: %v", receiptHandle, err)
				}
			}(*msg.ReceiptHandle)
		}
	}
	return nil
}
