package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("elbs", elbs)
}

func elbs(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading ELBs for account %s in %s\n", account, region)
	svc := elb.New(s)
	input := &elb.DescribeLoadBalancersInput{}

	err := svc.DescribeLoadBalancersPages(input, func(result *elb.DescribeLoadBalancersOutput, _ bool) bool {
		for _, v := range result.LoadBalancerDescriptions {
			arn := fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s", region, account, aws.StringValue(v.LoadBalancerName))
			tmp := map[string]interface{}{
				"arn":            arn,
				"aws_account_id": account,
				"aws_region":     region,
				"name":           aws.StringValue(v.LoadBalancerName),
				"dns_name":       aws.StringValue(v.DNSName),
				"scheme":         aws.StringValue(v.Scheme),
				"created_at":     aws.TimeValue(v.CreatedTime).UTC().Unix(),
				"vpc_id":         aws.StringValue(v.VPCId),
				// policies
			}

			output = append(output, Record{
				File:  "aws-elbs",
				Attrs: tmp,
			})

			for _, subnet := range v.Subnets {
				output = append(output, Record{
					File: "aws-elb-subnets",
					Attrs: map[string]interface{}{
						"aws_elb_arn":       arn,
						"aws_vpc_subnet_id": aws.StringValue(subnet),
					},
				})
			}

			for _, sg := range v.SecurityGroups {
				output = append(output, Record{
					File: "aws-elb-security-groups",
					Attrs: map[string]interface{}{
						"aws_elb_arn":               arn,
						"aws_ec2_security_group_id": aws.StringValue(sg),
					},
				})
			}

			for _, i := range v.Instances {
				output = append(output, Record{
					File: "aws-elb-instances",
					Attrs: map[string]interface{}{
						"aws_elb_arn":         arn,
						"aws_ec2_instance_id": aws.StringValue(i.InstanceId),
					},
				})
			}
		}

		return true
	})

	if err != nil {
		log.Fatalf("DescribeLoadBalancers (%s) error: %s", account, err)
	}

	return output
}
