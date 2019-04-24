package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("asgs", asgs)
}

func asgs(s *session.Session, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading ASGs for account %s in %s\n", account, region)
	svc := autoscaling.New(s)
	input := &autoscaling.DescribeAutoScalingGroupsInput{}
	result, err := svc.DescribeAutoScalingGroups(input)
	if err != nil {
		log.Fatalf("DescribeAutoScalingGroups error: %s", err)
	}

	var output []Record
	for _, asg := range result.AutoScalingGroups {
		tmp := map[string]interface{}{
			"arn":                aws.StringValue(asg.AutoScalingGroupARN),
			"aws_account_id":     account,
			"aws_region":         region,
			"aws_vpc_id":         aws.StringValue(asg.VPCZoneIdentifier),
			"created_at":         aws.TimeValue(asg.CreatedTime).UTC().Unix(),
			"desired_instances":  aws.Int64Value(asg.DesiredCapacity),
			"launch_config_name": aws.StringValue(asg.LaunchConfigurationName),
			"max_instances":      aws.Int64Value(asg.MaxSize),
			"min_instances":      aws.Int64Value(asg.MinSize),
		}

		output = append(output, Record{
			File:  "aws-asgs",
			Attrs: tmp,
		})

		if asg.LaunchTemplate != nil {
			output = append(output, Record{
				File: "aws-asg-launch-templates",
				Attrs: map[string]interface{}{
					"aws_asg_arn":               aws.StringValue(asg.AutoScalingGroupARN),
					"aws_ec2_launchtemplate_id": aws.StringValue(asg.LaunchTemplate.LaunchTemplateId),
				},
			})
		}
	}
	return output
}
