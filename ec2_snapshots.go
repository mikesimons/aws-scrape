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
	addResource("ec2-snapshots", ec2Snapshots)
}

func ec2Snapshots(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading EC2 snapshots for account %s\n", account)
	svc := ec2.New(s)
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{aws.String("self")},
	}

	err := svc.DescribeSnapshotsPages(input, func(result *ec2.DescribeSnapshotsOutput, _ bool) bool {
		for _, s := range result.Snapshots {
			output = append(output,
				Record{
					File: "aws-ec2-snapshots",
					Attrs: map[string]interface{}{
						"arn":               fmt.Sprintf("arn:aws:ec2:%s:%s:snapshot/%s", region, account, aws.StringValue(s.SnapshotId)),
						"aws_account_id":    account,
						"aws_region":        region,
						"snapshot_id":       aws.StringValue(s.SnapshotId),
						"aws_ec2_volume_id": aws.StringValue(s.VolumeId),
						"owner_id":          aws.StringValue(s.OwnerId),
						"description":       aws.StringValue(s.Description),
					},
				},
			)
		}
		return true
	})

	if err != nil {
		log.Fatalf("DescribeSnapshots error: %s", err)
	}

	return output
}
