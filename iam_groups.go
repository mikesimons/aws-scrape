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
	addResource("iam-groups", iamGroups)
}

func iamGroups(s *session.Session, _ string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading IAM groups for account %s\n", account)
	svc := iam.New(s)
	input := &iam.ListGroupsInput{}

	err := svc.ListGroupsPages(input, func(result *iam.ListGroupsOutput, _ bool) bool {
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

			output = append(output, iamUsersInGroup(s, g.GroupName, account)...)
		}
		return true
	})

	if err != nil {
		log.Fatalf("ListGroups error: %s", err)
	}

	return output
}

func iamUsersInGroup(s *session.Session, groupName *string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading users in IAM group %s for account %s\n", aws.StringValue(groupName), account)

	svc := iam.New(s)
	input := &iam.GetGroupInput{
		GroupName: groupName,
	}
	result, err := svc.GetGroup(input)

	if err != nil {
		log.Fatalf("GetGroup %s error: %s", aws.StringValue(groupName), err)
	}

	var output []Record
	for _, user := range result.Users {
		output = append(output, Record{
			File: "aws-iam-user-groups",
			Attrs: map[string]interface{}{
				"aws_iam_user_arn":  aws.StringValue(user.Arn),
				"aws_iam_group_arn": aws.StringValue(result.Group.Arn),
			},
		})
	}

	return output
}
