package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func init() {
	addResource("sqs_queues", sqsQueues)
}

func sqsQueues(s *session.Session, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading SQS queues for account %s in %s\n", account, region)

	svc := sqs.New(s)
	input := &sqs.ListQueuesInput{}
	result, err := svc.ListQueues(input)
	if err != nil {
		log.Fatalf("ListQueues error: %s", err)
	}

	var output []Record
	for _, u := range result.QueueUrls {
		parsedQueueURL, _ := url.Parse(aws.StringValue(u))
		tmp := map[string]interface{}{
			"aws_account_id": account,
			"aws_region":     region,
			"name":           strings.Split(parsedQueueURL.Path, "/")[2],
			"url":            aws.StringValue(u),
		}
		output = append(output, Record{
			File:  "aws-sqs-queues",
			Attrs: tmp,
		})
	}
	return output
}
