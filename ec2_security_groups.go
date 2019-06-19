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
	addResource("ec2-security-groups", ec2SecurityGroups)
}

func ec2SecurityGroups(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading EC2 security groups for account %s in %s\n", account, region)
	svc := ec2.New(s)
	input := &ec2.DescribeSecurityGroupsInput{}

	err := svc.DescribeSecurityGroupsPages(input, func(result *ec2.DescribeSecurityGroupsOutput, _ bool) bool {
		for _, sg := range result.SecurityGroups {
			tmp := map[string]interface{}{
				"aws_account_id":    account,
				"aws_region":        region,
				"description":       aws.StringValue(sg.Description),
				"name":              aws.StringValue(sg.GroupName),
				"security_group_id": aws.StringValue(sg.GroupId),
			}
			output = append(output, Record{File: "aws-ec2-security-groups", Attrs: tmp})
			output = append(output, ec2SecurityGroupRules(sg.IpPermissions, aws.StringValue(sg.GroupId), "ingress")...)
			output = append(output, ec2SecurityGroupRules(sg.IpPermissionsEgress, aws.StringValue(sg.GroupId), "egress")...)
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeSecurityGroups error: %s", err)
	}

	return output
}

func ec2SecurityGroupRules(rules []*ec2.IpPermission, securityGroupID string, direction string) []Record {
	var output []Record

	for _, r := range rules {
		for _, x := range r.IpRanges {
			output = append(output,
				Record{
					File: "aws-ec2-security-group-rules",
					Attrs: map[string]interface{}{
						"aws_ec2_security_group_id": securityGroupID,
						"cidr":                      aws.StringValue(x.CidrIp),
						"description":               aws.StringValue(x.Description),
						"direction":                 direction,
						"from_port":                 aws.Int64Value(r.FromPort),
						"protocol":                  aws.StringValue(r.IpProtocol),
						"to_port":                   aws.Int64Value(r.ToPort),
					},
				},
			)
		}

		for _, x := range r.UserIdGroupPairs {
			output = append(output,
				Record{
					File: "aws-ec2-security-group-rules",
					Attrs: map[string]interface{}{
						"allowed_aws_account_id":            aws.StringValue(x.UserId),
						"allowed_aws_ec2_security_group_id": aws.StringValue(x.GroupId),
						"aws_ec2_security_group_id":         securityGroupID,
						"description":                       aws.StringValue(x.Description),
						"direction":                         direction,
						"from_port":                         aws.Int64Value(r.FromPort),
						"protocol":                          aws.StringValue(r.IpProtocol),
						"to_port":                           aws.Int64Value(r.ToPort),
					},
				},
			)
		}
	}

	return output
}
