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
	addResource("iam_groups", iamGroups)
}

func iamGroups(s *session.Session, _ string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading IAM groups for account %s\n", account)
	svc := iam.New(s)
	input := &iam.ListGroupsInput{}
	result, err := svc.ListGroups(input)
	if err != nil {
		log.Fatalf("ListGroups error: %s", err)
	}

	var output []Record
	for _, g := range result.Groups {
		output = append(output,
			Record{
				File: "aws-iam-groups",
				Attrs: map[string]interface{}{
					"arn":            aws.StringValue(g.Arn),
					"aws_account_id": account,
					"created_at":     aws.TimeValue(g.CreateDate).UTC().Unix(),
					"id":             g.GroupId,
					"name":           g.GroupName,
				},
			},
		)
	}
	return output
}
