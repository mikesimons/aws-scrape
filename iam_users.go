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

		output = append(output, iamGroupsForUser(s, u.Arn, u.UserName, account)...)
	}
	return output
}

func iamGroupsForUser(s *session.Session, userArn *string, username *string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading IAM groups for %s for account %s\n", aws.StringValue(username), account)

	svc := iam.New(s)
	input := &iam.ListGroupsForUserInput{
		UserName: username,
	}
	result, err := svc.ListGroupsForUser(input)
	if err != nil {
		log.Fatalf("ListGroupsForUser %s error: %s", aws.StringValue(username), err)
	}

	var output []Record
	for _, g := range result.Groups {
		output = append(output, Record{
			File: "aws-iam-user-groups",
			Attrs: map[string]interface{}{
				"aws_iam_user_arn":  aws.StringValue(userArn),
				"aws_iam_group_arn": aws.StringValue(g.Arn),
			},
		})
	}

	return output
}
