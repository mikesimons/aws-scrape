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
	addResource("vpc-subnets", vpcSubnets)
}

func vpcSubnets(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading VPC subnets for account %s in %s\n", account, region)
	svc := ec2.New(s)
	input := &ec2.DescribeSubnetsInput{}

	err := svc.DescribeSubnetsPages(input, func(result *ec2.DescribeSubnetsOutput, _ bool) bool {
		for _, s := range result.Subnets {
			tmp := map[string]interface{}{
				"available_ips":         aws.Int64Value(s.AvailableIpAddressCount),
				"aws_account_id":        account,
				"aws_region":            region,
				"aws_availability_zone": aws.StringValue(s.AvailabilityZone),
				"cidr_block":            aws.StringValue(s.CidrBlock),
				"name":                  getTagOrDefault(s.Tags, "Name", aws.StringValue(s.SubnetId)),
				"subnet_id":             aws.StringValue(s.SubnetId),
				"tier":                  getTagOrDefault(s.Tags, "Tier", "unknown"),
				"vpc_id":                aws.StringValue(s.VpcId),
			}
			output = append(output, Record{
				File:  "aws-vpc-subnets",
				Attrs: tmp,
			})
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeSubnets error: %s", err)
	}

	return output
}
