package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func init() {
	addResource("regions", regions)
}

func regions(s *session.Session, _ string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading region list\n")
	svc := ec2.New(s)
	input := &ec2.DescribeRegionsInput{}
	result, err := svc.DescribeRegions(input)

	if err != nil {
		log.Fatalf("DescribeRegions error: %s", err)
	}

	var output []Record
	for _, r := range result.Regions {
		tmp := map[string]interface{}{
			"endpoint": aws.StringValue(r.Endpoint),
			"name":     aws.StringValue(r.RegionName),
		}
		output = append(output, Record{
			File:  "aws-regions",
			Attrs: tmp,
		})
	}
	return output
}
