package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("asgs", asgs)
}

func asgs(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading ASGs for account %s in %s\n", account, region)
	svc := autoscaling.New(s)
	input := &autoscaling.DescribeAutoScalingGroupsInput{}

	err := svc.DescribeAutoScalingGroupsPages(input, func(result *autoscaling.DescribeAutoScalingGroupsOutput, _ bool) bool {
		for _, asg := range result.AutoScalingGroups {
			tmp := map[string]interface{}{
				"arn":                        aws.StringValue(asg.AutoScalingGroupARN),
				"aws_account_id":             account,
				"aws_region":                 region,
				"created_at":                 aws.TimeValue(asg.CreatedTime).UTC().Unix(),
				"desired_instances":          aws.Int64Value(asg.DesiredCapacity),
				"aws_ec2_launch_config_name": aws.StringValue(asg.LaunchConfigurationName),
				"max_instances":              aws.Int64Value(asg.MaxSize),
				"min_instances":              aws.Int64Value(asg.MinSize),
			}

			if asg.LaunchTemplate != nil {
				tmp["aws_ec2_launchtemplate_id"] = aws.StringValue(asg.LaunchTemplate.LaunchTemplateId)
			}

			output = append(output, Record{
				File:  "aws-asgs",
				Attrs: tmp,
			})

			if asg.VPCZoneIdentifier != nil {
				subnets := strings.Split(aws.StringValue(asg.VPCZoneIdentifier), ",")
				for _, subnetID := range subnets {
					output = append(output, Record{
						File: "aws-asg-subnets",
						Attrs: map[string]interface{}{
							"aws_asg_arn":       aws.StringValue(asg.AutoScalingGroupARN),
							"aws_vpc_subnet_id": subnetID,
						},
					})
				}
			}
		}

		return true
	})
	if err != nil {
		log.Fatalf("DescribeAutoScalingGroups error: %s", err)
	}

	return output
}
