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
	addResource("route53-zones", route53Zones)
}

func route53Zones(s *session.Session, _ string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading Route53 zones for account %s\n", account)

	svc := route53.New(s)
	input := &route53.ListHostedZonesInput{}
	err := svc.ListHostedZonesPages(input, func(result *route53.ListHostedZonesOutput, _ bool) bool {
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
		return true
	})

	if err != nil {
		log.Fatalf("ListHostedZones error: %s", err)
	}

	return output
}

func route53RecordSets(s *session.Session, hostedZoneID string, account string) []Record {
	var output []Record

	fmt.Fprintf(os.Stderr, "Loading Route53 record sets for zone %s in account %s\n", hostedZoneID, account)

	svc := route53.New(s)
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &hostedZoneID,
	}

	err := svc.ListResourceRecordSetsPages(input, func(result *route53.ListResourceRecordSetsOutput, _ bool) bool {
		for _, v := range result.ResourceRecordSets {
			for _, rr := range v.ResourceRecords {
				tmp := map[string]interface{}{
					"aws_account_id":             account,
					"aws_route53_hosted_zone_id": hostedZoneID,
					"name":                       strings.Trim(aws.StringValue(v.Name), "."),
					"type":                       aws.StringValue(v.Type),
					"record":                     aws.StringValue(rr.Value),
				}

				output = append(output, Record{
					File:  "aws-route53-records",
					Attrs: tmp,
				})
			}
		}
		return true
	})

	if err != nil {
		log.Fatalf("ListResourceRecordSets error: %s", err)
	}

	return output
}
