package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"job/lighblocks-ha/consumer"
	"job/lighblocks-ha/consumer/store"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func main() {
	var queueName string
	if queueName = os.Getenv("QUEUE_NAME"); queueName == "" {
		log.Fatal("[FATAL] Env variable QUEUE_NAME must be set")
	}
	rcvWaitTimeSec := flag.Int64("rcv-wait-time-sec", 20, "sqs long polling wait time (default 20 sec.)")
	flag.Parse()

	// Building a context with graceful shutdown handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// AWS client will go over default config provider chain - the env vars, ~/.aws folder etc.
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Panic(err)
	}
	client := sqs.NewFromConfig(cfg)

	queueURL, err := getOrCreateQueue(ctx, client, queueName)
	if err != nil {
		log.Panic(err)
	}

	c := consumer.New(client, queueURL, store.New())
	if err := c.Run(ctx, int32(*rcvWaitTimeSec)); err != nil {
		log.Panic(err)
	}

	log.Println("Consumer has been stopped.")
}

func getOrCreateQueue(ctx context.Context, client *sqs.Client, queueName string) (string, error) {
	getOutput, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: &queueName})
	if err != nil {
		var qneErr *types.QueueDoesNotExist
		if errors.As(err, &qneErr) {
			createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: &queueName})
			if err != nil {
				return "", fmt.Errorf("failed to create queue: %w", err)
			}
			return *createOutput.QueueUrl, nil
		}
		return "", fmt.Errorf("failed to get queue info: %w", err)
	}
	return *getOutput.QueueUrl, nil
}
