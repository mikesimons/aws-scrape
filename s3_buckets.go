package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func init() {
	addResource("s3_buckets", s3Buckets)
}

func s3Buckets(s *session.Session, _ string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading S3 buckets for account %s\n", account)
	svc := s3.New(s)
	input := &s3.ListBucketsInput{}
	result, err := svc.ListBuckets(input)
	if err != nil {
		log.Fatalf("ListBuckets error: %s", err)
	}

	var output []Record
	for _, b := range result.Buckets {
		output = append(output,
			Record{
				File: "aws-s3-buckets",
				Attrs: map[string]interface{}{
					"aws_account_id": account,
					"created_at":     aws.TimeValue(b.CreationDate).UTC().Unix(),
					"name":           aws.StringValue(b.Name),
				},
			},
		)
	}
	return output
}
