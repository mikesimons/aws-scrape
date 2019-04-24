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
	addResource("ec2_instances", ec2Instances)
}

func ec2Instances(s *session.Session, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading EC2 instances for account %s\n", account)
	svc := ec2.New(s)
	input := &ec2.DescribeInstancesInput{}
	result, err := svc.DescribeInstances(input)
	if err != nil {
		log.Fatalf("DescribeInstances error: %s", err)
	}

	var output []Record
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			tmp := map[string]interface{}{
				"aws_account_id":     account,
				"aws_region":         region,
				"aws_vpc_id":         aws.StringValue(instance.VpcId),
				"created_at":         aws.TimeValue(instance.LaunchTime).UTC().Unix(),
				"ebs_optimized":      aws.BoolValue(instance.EbsOptimized),
				"image_id":           aws.StringValue(instance.ImageId),
				"instance_id":        aws.StringValue(instance.InstanceId),
				"instance_type":      aws.StringValue(instance.InstanceType),
				"kernel_id":          aws.StringValue(instance.KernelId),
				"key_name":           aws.StringValue(instance.KeyName),
				"name":               getTagOrDefault(instance.Tags, "Name", aws.StringValue(instance.InstanceId)),
				"private_dns_name":   aws.StringValue(instance.PrivateDnsName),
				"private_ip_address": aws.StringValue(instance.PrivateIpAddress),
				"public_dns_name":    aws.StringValue(instance.PublicDnsName),
				"public_ip_address":  aws.StringValue(instance.PublicIpAddress),
				"source_dest_check":  aws.BoolValue(instance.SourceDestCheck),
				"state":              aws.StringValue(instance.State.Name),
				"aws_vpc_subnet_id":  aws.StringValue(instance.SubnetId),
			}

			if instance.IamInstanceProfile != nil {
				tmp["aws_iam_instance_profile_arn"] = aws.StringValue(instance.IamInstanceProfile.Arn)
			}

			output = append(output, Record{
				File:  "aws-ec2-instances",
				Attrs: tmp,
			})

			// TODO instance.NetworkInterfaces

			for _, bd := range instance.BlockDeviceMappings {
				tmp := map[string]interface{}{
					"attach_time":           aws.TimeValue(bd.Ebs.AttachTime).UTC().Unix(),
					"aws_ec2_instance_id":   aws.StringValue(instance.InstanceId),
					"aws_ec2_volume_id":     aws.StringValue(bd.Ebs.VolumeId),
					"delete_on_termination": aws.BoolValue(bd.Ebs.DeleteOnTermination),
					"device":                aws.StringValue(bd.DeviceName),
					"status":                aws.StringValue(bd.Ebs.Status),
				}
				output = append(output, Record{
					File:  "aws-ec2-instance-block-devices",
					Attrs: tmp,
				})
			}

			for _, sg := range instance.SecurityGroups {
				tmp := map[string]interface{}{
					"aws_ec2_instance_id":       aws.StringValue(instance.InstanceId),
					"aws_ec2_security_group_id": aws.StringValue(sg.GroupId),
				}
				output = append(output, Record{
					File:  "aws-ec2-instance-security-groups",
					Attrs: tmp,
				})
			}
		}

	}
	return output
}
