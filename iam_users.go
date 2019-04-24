package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("iam_users", iamUsers)
}

func iamUsers(s *session.Session, _ string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading IAM users for account %s\n", account)
	svc := iam.New(s)
	input := &iam.ListUsersInput{}
	result, err := svc.ListUsers(input)
	if err != nil {
		log.Fatalf("ListUsers error: %s", err)
	}

	var output []Record
	for _, u := range result.Users {
		output = append(output,
			Record{
				File: "aws-iam-users",
				Attrs: map[string]interface{}{
					"arn":            aws.StringValue(u.Arn),
					"aws_account_id": account,
					"created_at":     aws.TimeValue(u.CreateDate).UTC().Unix(),
					"id":             aws.StringValue(u.UserId),
					"name":           aws.StringValue(u.UserName),
				},
			},
		)
	}
	return output
}
