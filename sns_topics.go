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
	addResource("sns-topics", snsTopics)
}

func snsTopics(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading SNS topics for account %s in %s\n", account, region)
	svc := sns.New(s)
	input := &sns.ListTopicsInput{}

	err := svc.ListTopicsPages(input, func(result *sns.ListTopicsOutput, _ bool) bool {
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
		return true
	})

	if err != nil {
		log.Fatalf("ListTopics error: %s", err)
	}

	return output
}
