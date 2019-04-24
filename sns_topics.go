package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func init() {
	addResource("sns_topics", snsTopics)
}

func snsTopics(s *session.Session, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading SNS topics for account %s in %s\n", account, region)
	svc := sns.New(s)
	input := &sns.ListTopicsInput{}
	result, err := svc.ListTopics(input)
	if err != nil {
		log.Fatalf("ListTopics error: %s", err)
	}

	var output []Record
	for _, t := range result.Topics {
		splitARN := strings.Split(*t.TopicArn, ":")
		tmp := map[string]interface{}{
			"arn":            aws.StringValue(t.TopicArn),
			"aws_account_id": account,
			"aws_region":     region,
			"name":           splitARN[len(splitARN)-1],
		}
		output = append(output, Record{
			File:  "aws-sns-topics",
			Attrs: tmp,
		})
	}
	return output
}
