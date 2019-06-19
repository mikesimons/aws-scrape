package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

func init() {
	addResource("eks-clusters", eksClusters)
}

func eksClusters(s *session.Session, region string, account string) []Record {
	fmt.Fprintf(os.Stderr, "Loading EKS clusters for account %s\n", account)
	svc := eks.New(s)
	input := &eks.ListClustersInput{}
	result, err := svc.ListClusters(input)

	if err != nil {
		log.Fatalf("ListClusters error: %s", err)
	}

	var output []Record
	for _, clusterName := range result.Clusters {
		describeInput := &eks.DescribeClusterInput{
			Name: clusterName,
		}

		c, err := svc.DescribeCluster(describeInput)
		if err != nil {
			log.Fatalf("DescribeClusters error: %s", err)
		}

		output = append(output,
			Record{
				File: "aws-eks-clusters",
				Attrs: map[string]interface{}{
					"arn":              aws.StringValue(c.Cluster.Arn),
					"aws_account_id":   account,
					"aws_region":       region,
					"cluster_role_arn": aws.StringValue(c.Cluster.RoleArn),
					"eks_version":      aws.StringValue(c.Cluster.PlatformVersion),
					"k8s_version":      aws.StringValue(c.Cluster.Version),
					"name":             aws.StringValue(c.Cluster.Name),
				},
			},
		)
	}
	return output
}
