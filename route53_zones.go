package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/awslabs/aws-sdk-go/service/route53"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	addResource("route53_zones", route53Zones)
}

func route53Zones(s *session.Session, _ string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading Route53 zones for account %s\n", account)

	svc := route53.New(s)
	input := &route53.ListHostedZonesInput{}
	result, err := svc.ListHostedZones(input)
	if err != nil {
		log.Fatalf("ListHostedZones error: %s", err)
	}

	var output []Record
	for _, v := range result.HostedZones {
		tmp := map[string]interface{}{
			"aws_account_id": account,
			"id":             strings.ReplaceAll(aws.StringValue(v.Id), "/hostedzone/", ""),
			"name":           aws.StringValue(v.Name),
			"private":        aws.BoolValue(v.Config.PrivateZone),
		}

		output = append(output, Record{
			File:  "aws-route53-zones",
			Attrs: tmp,
		})

		output = append(output, route53RecordSets(s, tmp["id"].(string), account)...)
	}
	return output
}

func route53RecordSets(s *session.Session, hostedZoneID string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading Route53 record sets for zone %s in account %s\n", hostedZoneID, account)

	svc := route53.New(s)
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &hostedZoneID,
	}
	result, err := svc.ListResourceRecordSets(input)
	if err != nil {
		log.Fatalf("ListResourceRecordSets error: %s", err)
	}

	var output []Record
	for _, v := range result.ResourceRecordSets {
		for _, rr := range v.ResourceRecords {
			tmp := map[string]interface{}{
				"aws_account_id":             account,
				"aws_route53_hosted_zone_id": hostedZoneID,
				"name":                       aws.StringValue(v.Name),
				"type":                       aws.StringValue(v.Type),
				"record":                     aws.StringValue(rr.Value),
			}

			output = append(output, Record{
				File:  "aws-route53-records",
				Attrs: tmp,
			})
		}
	}
	return output
}
