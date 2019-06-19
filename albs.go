package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("albs", albs)
}

func albs(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading ALBs for account %s in %s\n", account, region)
	svc := elbv2.New(s)
	input := &elbv2.DescribeLoadBalancersInput{}

	err := svc.DescribeLoadBalancersPages(input, func(result *elbv2.DescribeLoadBalancersOutput, _ bool) bool {
		for _, a := range result.LoadBalancers {
			tmp := map[string]interface{}{
				"arn":            aws.StringValue(a.LoadBalancerArn),
				"aws_account_id": account,
				"aws_region":     region,
				"aws_vpc_id":     aws.StringValue(a.VpcId),
				"name":           aws.StringValue(a.LoadBalancerName),
				"scheme":         aws.StringValue(a.Scheme),
			}

			for _, sg := range a.SecurityGroups {
				output = append(output,
					Record{
						File: "aws-alb-security-groups",
						Attrs: map[string]interface{}{
							"aws_alb_arn":               aws.StringValue(a.LoadBalancerArn),
							"aws_ec2_security_group_id": aws.StringValue(sg),
						},
					},
				)
			}

			for _, az := range a.AvailabilityZones {
				output = append(output,
					Record{
						File: "aws-alb-subnets",
						Attrs: map[string]interface{}{
							"aws_alb_arn":       aws.StringValue(a.LoadBalancerArn),
							"aws_vpc_subnet_id": aws.StringValue(az.SubnetId),
						},
					},
				)
			}

			output = append(output, Record{
				File:  "aws-albs",
				Attrs: tmp,
			})
			//output = append(output, albTargetGroups(s, aws.StringValue(a.LoadBalancerArn), region, account)...)
			//output = append(output, albListeners(s, aws.StringValue(a.LoadBalancerArn), region, account)...)

		}

		return true
	})
	if err != nil {
		log.Fatalf("DescribeLoadBalancers error: %s", err)
	}

	return output
}

func albTargetGroups(s *session.Session, ARN string, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading ALB target groups for %s in account %s in %s\n", ARN, account, region)
	svc := elbv2.New(s)
	input := &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: &ARN,
	}

	err := svc.DescribeTargetGroupsPages(input, func(result *elbv2.DescribeTargetGroupsOutput, _ bool) bool {
		for _, tg := range result.TargetGroups {
			tmp := map[string]interface{}{
				"arn":            aws.StringValue(tg.TargetGroupArn),
				"aws_account_id": account,
				"aws_alb_arn":    ARN,
				"aws_region":     region,
				"name":           aws.StringValue(tg.TargetGroupName),
			}
			output = append(output, Record{
				File:  "aws-alb-target-groups",
				Attrs: tmp,
			})
			output = append(output, albTargets(s, aws.StringValue(tg.TargetGroupArn), aws.StringValue(tg.TargetType), region, account)...)
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeTargetGroups error: %s", err)
	}

	return output
}

func albListeners(s *session.Session, ARN string, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading ALB listeners for %s in account %s in %s\n", ARN, account, region)
	svc := elbv2.New(s)
	input := &elbv2.DescribeListenersInput{
		LoadBalancerArn: &ARN,
	}

	err := svc.DescribeListenersPages(input, func(result *elbv2.DescribeListenersOutput, _ bool) bool {
		for _, l := range result.Listeners {
			output = append(output, Record{
				File: "aws-alb-listeners",
				Attrs: map[string]interface{}{
					"arn":            aws.StringValue(l.ListenerArn),
					"aws_account_id": account,
					"aws_alb_arn":    aws.StringValue(l.LoadBalancerArn),
					"aws_region":     region,
					"port":           aws.Int64Value(l.Port),
					"protocol":       aws.StringValue(l.Protocol),
					"ssl_policy":     aws.StringValue(l.SslPolicy),
				},
			})
			output = append(output, albListenerRules(s, aws.StringValue(l.ListenerArn), region, account)...)
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeListeners error: %s", err)
	}

	return output
}

func albListenerRules(s *session.Session, ARN string, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading ALB listener rules for %s in account %s in %s\n", ARN, account, region)
	svc := elbv2.New(s)
	input := &elbv2.DescribeRulesInput{
		ListenerArn: &ARN,
	}

	result, err := svc.DescribeRules(input)
	if err != nil {
		log.Fatalf("DescribeRules error: %s", err)
	}

	var output []Record
	for _, r := range result.Rules {
		for _, a := range r.Actions {
			// TODO r.Conditions
			if *a.Type == "forward" {
				tmp := map[string]interface{}{
					"aws_alb_listener_arn":     ARN,
					"aws_alb_target_group_arn": aws.StringValue(a.TargetGroupArn),
				}
				output = append(output, Record{
					File:  "aws-alb-listener-target-groups",
					Attrs: tmp,
				})
			}
		}
	}

	return output
}

func albTargets(s *session.Session, ARN string, targetType string, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading ALB targets for %s in account %s in %s\n", ARN, account, region)
	svc := elbv2.New(s)
	input := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: &ARN,
	}

	result, err := svc.DescribeTargetHealth(input)
	if err != nil {
		log.Fatalf("DescribeTargetHealth error: %s", err)
	}

	var output []Record
	for _, t := range result.TargetHealthDescriptions {
		tmp := map[string]interface{}{
			"aws_account_id":           account,
			"aws_alb_target_group_arn": ARN,
			"aws_region":               region,
			"port":                     aws.Int64Value(t.Target.Port),
			"type":                     targetType,
		}

		if targetType == "instance" {
			tmp["aws_ec2_instance_id"] = aws.StringValue(t.Target.Id)
		} else if targetType == "ip" {
			tmp["ip"] = aws.StringValue(t.Target.Id)
		} else if targetType == "lambda" {
			tmp["aws_lambda_arn"] = aws.StringValue(t.Target.Id)
		}

		output = append(output, Record{
			File:  "aws-alb-target-group-targets",
			Attrs: tmp,
		})
	}

	return output
}
