package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("rds-instances", rdsInstances)
}

func rdsInstances(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading RDS databases for account %s\n", account)
	svc := rds.New(s)
	input := &rds.DescribeDBInstancesInput{}

	err := svc.DescribeDBInstancesPages(input, func(result *rds.DescribeDBInstancesOutput, _ bool) bool {
		for _, db := range result.DBInstances {
			subnetMap := make(map[string]string)
			var subnetGroupName string
			var vpcID string
			if db.DBSubnetGroup != nil {
				vpcID = aws.StringValue(db.DBSubnetGroup.VpcId)
				subnetGroupName = aws.StringValue(db.DBSubnetGroup.DBSubnetGroupName)
				for _, subnet := range db.DBSubnetGroup.Subnets {
					subnetMap[aws.StringValue(subnet.SubnetAvailabilityZone.Name)] = aws.StringValue(subnet.SubnetIdentifier)
				}
			}

			tmp := map[string]interface{}{
				"allocated_storage":           aws.Int64Value(db.AllocatedStorage),
				"arn":                         aws.StringValue(db.DBInstanceArn),
				"aws_account_id":              account,
				"aws_availability_zone":       aws.StringValue(db.AvailabilityZone),
				"aws_vpc_subnet_id":           subnetMap[aws.StringValue(db.AvailabilityZone)],
				"aws_vpc_id":                  vpcID,
				"created_at":                  aws.TimeValue(db.InstanceCreateTime).UTC().Unix(),
				"encrypted":                   aws.BoolValue(db.StorageEncrypted),
				"endpoint":                    aws.StringValue(db.Endpoint.Address),
				"engine":                      aws.StringValue(db.Engine),
				"engine_version":              aws.StringValue(db.EngineVersion),
				"iam_emabled":                 aws.BoolValue(db.IAMDatabaseAuthenticationEnabled),
				"instance_class":              aws.StringValue(db.DBInstanceClass),
				"kms_key_id":                  aws.StringValue(db.KmsKeyId),
				"multi_az":                    aws.BoolValue(db.MultiAZ),
				"name":                        aws.StringValue(db.DBInstanceIdentifier),
				"port":                        aws.Int64Value(db.Endpoint.Port),
				"public":                      aws.BoolValue(db.PubliclyAccessible),
				"region":                      region,
				"secondary_availability_zone": aws.StringValue(db.SecondaryAvailabilityZone),
				"secondary_subnet_id":         subnetMap[aws.StringValue(db.SecondaryAvailabilityZone)],
				"status":                      aws.StringValue(db.DBInstanceStatus),
				"storage_type":                aws.StringValue(db.StorageType),
				"subnet_group_name":           subnetGroupName,
			}

			output = append(output, Record{
				File:  "aws-rds-instances",
				Attrs: tmp,
			})

			// TODO db.DBSecurityGroups
			// TODO db.DBParameterGroups

			for _, sg := range db.VpcSecurityGroups {
				tmp := map[string]interface{}{
					"aws_ec2_security_group_id": aws.StringValue(sg.VpcSecurityGroupId),
					"aws_rds_instance_arn":      aws.StringValue(db.DBInstanceArn),
				}
				output = append(output, Record{
					File:  "aws-rds-instance-security-groups",
					Attrs: tmp,
				})

			}
		}
		return true
	})
	if err != nil {
		log.Fatalf("DescribeDBInstances error: %s", err)
	}

	return output
}
