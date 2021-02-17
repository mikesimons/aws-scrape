# aws-scrape

Example:
```
./aws-scrape \
	--accounts='12345678901: { role: "arn:aws:iam::%s:role/TerraformRole" }' \
	--scrape='account_ids: [12345678901], regions: [us-east-1], resources: [accounts,regions]' \
	--scrape='account_ids: [12345678901], regions: [us-east-*], resources: [albs, asgs, ec2_instances, ec2_security_groups, ec2_volumes, eks_clusters, iam_users, rds_instances, s3_buckets, sns_topics, sqs_queues, vpcs, vpc_subnets]'
```
