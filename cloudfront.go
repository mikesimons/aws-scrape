package main

import (
	"fmt"
	"log"
	"os"

	"github.com/awslabs/aws-sdk-go/service/cloudfront"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("cloudfront-cdns", cloudfrontCdns)
}

func cloudfrontCdns(s *session.Session, region string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading cloudfront CDNs list\n")
	svc := cloudfront.New(s)
	input := &cloudfront.ListDistributionsInput{}

	err := svc.ListDistributionsPages(input, func(result *cloudfront.ListDistributionsOutput, _ bool) bool {
		for _, dist := range result.DistributionList.Items {
			output = append(output, Record{
				File: "aws-cloudfront-cdns",
				Attrs: map[string]interface{}{
					"arn":    aws.StringValue(dist.ARN),
					"domain": aws.StringValue(dist.DomainName),
					"id":     aws.StringValue(dist.Id),
				},
			})

			for _, origin := range dist.Origins.Items {
				output = append(output, Record{
					File: "aws-cloudfront-cdn-origins",
					Attrs: map[string]interface{}{
						"aws_cloudfront_cdn_id": aws.StringValue(dist.Id),
						"domain":                aws.StringValue(origin.DomainName),
					},
				})
			}
		}

		return true
	})

	if err != nil {
		log.Fatalf("ListDistributions error: %s", err)
	}

	return output
}
