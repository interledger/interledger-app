package main

import (
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/organizations"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateS3Policy(bucket *s3.Bucket, org *organizations.Organization, accountIds []pulumi.IDOutput) pulumi.StringOutput {
	type Principal struct {
		AWS     []string `json:"AWS,omitempty"`
		Service string   `json:"Service,omitempty"`
	}
	type Condition struct {
		StringEquals map[string]string `json:"StringEquals,omitempty"`
	}

	type Statement struct {
		Sid       string     `json:"Sid"`
		Effect    string     `json:"Effect"`
		Principal Principal  `json:"Principal"`
		Action    []string   `json:"Action"`
		Resource  []string   `json:"Resource"`
		Condition *Condition `json:"Condition,omitempty"`
	}

	type Policy struct {
		Version    string      `json:"Version"`
		ID         string      `json:"Id,omitempty"`
		Statements []Statement `json:"Statement"`
	}

	var inputs []interface{}

	inputs = append(inputs, bucket.ID())
	inputs = append(inputs, org.MasterAccountId)
	for _, a := range accountIds {
		inputs = append(inputs, a)
	}

	return pulumi.All(inputs...).ApplyT(func(args []interface{}) (string, error) {
		bucketId := args[0].(pulumi.ID)
		masterAccId := args[1].(string)

		var writeResources []string
		orgResource := fmt.Sprintf("arn:aws:s3:::%s/AWSLogs/%s/*", bucketId, masterAccId)
		writeResources = append(writeResources, orgResource)

		for i := 2; i < len(args); i++ {
			accId := args[i].(pulumi.ID)
			resource := fmt.Sprintf("arn:aws:s3:::%s/AWSLogs/%s/*", bucketId, accId)
			writeResources = append(writeResources, resource)
		}

		rawPolicy := &Policy{
			Version: "2012-10-17",
			Statements: []Statement{
				{
					Sid:    "AWSCloudTrailAclCheck",
					Effect: "Allow",
					Principal: Principal{
						Service: "cloudtrail.amazonaws.com",
					},
					Action: []string{
						"s3:GetBucketAcl",
					},
					Resource: []string{
						fmt.Sprintf("arn:aws:s3:::%s", bucketId),
					},
				},
				{
					Sid:    "AWSCloudTrailWrite",
					Effect: "Allow",
					Principal: Principal{
						Service: "cloudtrail.amazonaws.com",
					},
					Action: []string{
						"s3:PutObject",
					},
					Resource: writeResources,
					Condition: &Condition{
						StringEquals: map[string]string{
							"s3:x-amz-acl": "bucket-owner-full-control",
						},
					},
				},
			},
		}

		policy, err := json.Marshal(rawPolicy)

		return string(policy), err
	}).(pulumi.StringOutput)
}
