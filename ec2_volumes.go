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
	addResource("ec2_volumes", ec2Volumes)
}

func ec2Volumes(s *session.Session, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading EC2 volumes for account %s\n", account)
	svc := ec2.New(s)
	input := &ec2.DescribeVolumesInput{}
	result, err := svc.DescribeVolumes(input)
	if err != nil {
		log.Fatalf("DescribeVolumes error: %s", err)
	}

	var output []Record
	for _, v := range result.Volumes {
		output = append(output,
			Record{
				File: "aws-ec2-volumes",
				Attrs: map[string]interface{}{
					"aws_account_id":        account,
					"aws_availability_zone": aws.StringValue(v.AvailabilityZone),
					"aws_ec2_snapshot_id":   aws.StringValue(v.SnapshotId),
					"aws_region":            region,
					"created_at":            aws.TimeValue(v.CreateTime).UTC().Unix(),
					"encrypted":             aws.BoolValue(v.Encrypted),
					"kms_key_id":            aws.StringValue(v.KmsKeyId),
					"name":                  getTagOrDefault(v.Tags, "Name", aws.StringValue(v.VolumeId)),
					"size":                  aws.Int64Value(v.Size),
					"state":                 aws.StringValue(v.State),
					"volume_id":             aws.StringValue(v.VolumeId),
					"volume_type":           aws.StringValue(v.VolumeType),
				},
			},
		)
	}
	return output
}
