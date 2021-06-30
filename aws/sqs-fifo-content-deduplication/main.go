package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("specify SQS URL\n")
		os.Exit(1)
	}

	sqsURL := os.Args[1]
	msg := os.Args[2]

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		panic(err)
	}
	client := sqs.New(sess)

	sendInput := &sqs.SendMessageInput{
		MessageGroupId: aws.String("test"),
		MessageBody:    &msg,
		QueueUrl:       &sqsURL,
	}

	sendOutput, err := client.SendMessage(sendInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("send message: %v\n", sendOutput)

	receiveOutputCh := make(chan *sqs.ReceiveMessageOutput, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		receiveInput := &sqs.ReceiveMessageInput{
			QueueUrl:          &sqsURL,
			VisibilityTimeout: aws.Int64(10),
			WaitTimeSeconds:   aws.Int64(3),
		}
		receiveOutput, err := client.ReceiveMessage(receiveInput)
		if err != nil {
			panic(err)
		}
		receiveOutputCh <- receiveOutput
	}()

	select {
	case <-ctx.Done():
		fmt.Printf("message receive timeout\n")
		os.Exit(1)
	case receiveOutput := <-receiveOutputCh:
		fmt.Printf("received message: %v\n", receiveOutput)
		if len(receiveOutput.Messages) == 0 {
			fmt.Printf("received no messages\n")
			os.Exit(1)
		}

		for _, m := range receiveOutput.Messages {
			deleteInput := &sqs.DeleteMessageInput{
				QueueUrl:      &sqsURL,
				ReceiptHandle: m.ReceiptHandle,
			}
			_, err = client.DeleteMessage(deleteInput)
			if err != nil {
				fmt.Printf("failed to delete message: %s\n", *m.MessageId)
				continue
			}
			fmt.Printf("deleted message: %s\n", *m.MessageId)
		}
	}
}
