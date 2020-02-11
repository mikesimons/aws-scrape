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
	addResource("ec2-images", ec2Images)
}

func ec2Images(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading EC2 AMIs for account %s\n", account)
	svc := ec2.New(s)
	input := &ec2.DescribeImagesInput{
		Owners: []*string{aws.String("self")},
	}

	result, err := svc.DescribeImages(input)
	if err != nil {
		log.Fatalf("DescribeImages error: %s", err)
	}

	for _, i := range result.Images {
		arn := fmt.Sprintf("arn:aws:ec2:%s:%s:image/%s", region, account, aws.StringValue(i.ImageId))
		output = append(output,
			Record{
				File: "aws-ec2-images",
				Attrs: map[string]interface{}{
					"arn":                 arn,
					"aws_account_id":      account,
					"description":         aws.StringValue(i.Description),
					"aws_region":          region,
					"created_at":          aws.StringValue(i.CreationDate),
					"architecture":        aws.StringValue(i.Architecture),
					"name":                aws.StringValue(i.Name),
					"virtualization_type": aws.StringValue(i.VirtualizationType),
					"kernel_id":           aws.StringValue(i.KernelId),
				},
			},
		)

		for _, m := range i.BlockDeviceMappings {
			if m.Ebs != nil {
				output = append(output,
					Record{
						File: "aws-ec2-image-block-device",
						Attrs: map[string]interface{}{
							"aws_ec2_image_arn":   arn,
							"aws_ec2_snapshot_id": m.Ebs.SnapshotId,
						},
					})
			}
		}
	}

	return output
}
