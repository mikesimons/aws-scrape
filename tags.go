package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("tags", tags)
}

func tags(s *session.Session, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading tags for account %s in %s\n", account, region)

	svc := resourcegroupstaggingapi.New(s)
	input := &resourcegroupstaggingapi.GetResourcesInput{}
	result, err := svc.GetResources(input)
	if err != nil {
		log.Fatalf("GetResources error: %s", err)
	}

	var output []Record
	for _, list := range result.ResourceTagMappingList {
		for _, tag := range list.Tags {
			tmp := map[string]interface{}{
				"aws_account_id": account,
				"aws_region":     region,
				"name":           aws.StringValue(tag.Key),
				"value":          aws.StringValue(tag.Value),
				"resource_arn":   aws.StringValue(list.ResourceARN),
			}

			output = append(output, Record{
				File:  "aws-tag",
				Attrs: tmp,
			})
		}
	}

	return output
}
