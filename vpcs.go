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
	addResource("vpcs", vpcs)
}

func vpcs(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading VPCs for account %s in %s\n", account, region)
	svc := ec2.New(s)
	input := &ec2.DescribeVpcsInput{}

	err := svc.DescribeVpcsPages(input, func(result *ec2.DescribeVpcsOutput, _ bool) bool {
		for _, v := range result.Vpcs {
			tmp := map[string]interface{}{
				"aws_account_id": account,
				"aws_region":     region,
				"cidr_block":     aws.StringValue(v.CidrBlock),
				"name":           getTagOrDefault(v.Tags, "Name", "default"),
				"vpc_id":         aws.StringValue(v.VpcId),
			}

			output = append(output, Record{
				File:  "aws-vpcs",
				Attrs: tmp,
			})
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeVpcs error: %s", err)
	}

	return output
}
