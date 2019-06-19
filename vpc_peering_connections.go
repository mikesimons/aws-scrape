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
	addResource("vpc-peering-connections", vpcPeeringConnections)
}

func vpcPeeringConnections(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading VPC peering connections for account %s in %s\n", account, region)
	svc := ec2.New(s)
	input := &ec2.DescribeVpcPeeringConnectionsInput{}

	err := svc.DescribeVpcPeeringConnectionsPages(input, func(result *ec2.DescribeVpcPeeringConnectionsOutput, _ bool) bool {
		for _, v := range result.VpcPeeringConnections {
			tmp := map[string]interface{}{
				"requester_vpc_id": aws.StringValue(v.RequesterVpcInfo.VpcId),
				"accepter_vpc_id":  aws.StringValue(v.AccepterVpcInfo.VpcId),
			}

			output = append(output, Record{
				File:  "aws-vpc-peering-connections",
				Attrs: tmp,
			})
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeVpcPeeringsConnections error: %s", err)
	}

	return output
}
