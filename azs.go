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
	addResource("azs", azs)
}

func azs(s *session.Session, _ string, _ string) []Record {
	fmt.Fprintf(os.Stderr, "Loading AZ list\n")
	svc := ec2.New(s)
	input := &ec2.DescribeAvailabilityZonesInput{}
	result, err := svc.DescribeAvailabilityZones(input)
	if err != nil {
		log.Fatalf("DescribeAvailabilityZones error: %s", err)
	}

	var output []Record
	for _, az := range result.AvailabilityZones {
		tmp := map[string]interface{}{
			"region": aws.StringValue(az.RegionName),
			"zone":   aws.StringValue(az.ZoneName),
		}
		output = append(output, Record{
			File:  "aws-azs",
			Attrs: tmp,
		})
	}
	return output
}
