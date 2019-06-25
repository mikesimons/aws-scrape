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
	addResource("s3-buckets", s3Buckets)
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
		location, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: b.Name})
		if err != nil {
			log.Fatalf("GetBucketLocation error: %s", err)
		}

		regionPart := ""
		if location.LocationConstraint != nil {
			regionPart = fmt.Sprintf("%s.", aws.StringValue(location.LocationConstraint))
		}

		output = append(output,
			Record{
				File: "aws-s3-buckets",
				Attrs: map[string]interface{}{
					"aws_account_id": account,
					"created_at":     aws.TimeValue(b.CreationDate).UTC().Unix(),
					"domain":         fmt.Sprintf("%s.s3.%samazonaws.com", aws.StringValue(b.Name), regionPart),
					"name":           aws.StringValue(b.Name),
				},
			},
		)
	}
	return output
}
