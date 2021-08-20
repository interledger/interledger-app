package main

import (
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateKMSPolicy(logAccount *organizations.Account, org *organizations.Organization, accountIds []pulumi.IDOutput) pulumi.StringOutput {
	type Principal struct {
		AWS     []string `json:"AWS,omitempty"`
		Service string   `json:"Service,omitempty"`
	}

	type Condition struct {
		StringLike map[string][]string `json:"StringLike,omitempty"`
	}

	type Statement struct {
		Sid       string     `json:"Sid"`
		Effect    string     `json:"Effect"`
		Principal Principal  `json:"Principal"`
		Action    []string   `json:"Action"`
		Resource  string     `json:"Resource"`
		Condition *Condition `json:"Condition,omitempty"`
	}

	type KeyPolicy struct {
		Version    string      `json:"Version"`
		ID         string      `json:"Id"`
		Statements []Statement `json:"Statement"`
	}

	var inputs []interface{}

	inputs = append(inputs, logAccount.ID())
	inputs = append(inputs, org.MasterAccountId)
	for _, a := range accountIds {
		inputs = append(inputs, a)
	}

	policy := pulumi.All(inputs...).ApplyT(func(args []interface{}) (string, error) {
		logAccountId := args[0].(pulumi.ID)
		masterAccId := args[1].(string)

		var encryptCondition []string
		encryptCondition = append(encryptCondition, fmt.Sprintf("arn:aws:cloudtrail:*:%s:trail/*", masterAccId))

		for i := 2; i < len(args); i++ {
			accId := args[i].(pulumi.ID)
			encryptCondition = append(encryptCondition, fmt.Sprintf("arn:aws:cloudtrail:*:%s:trail/*", accId))
		}

		rawKeyPolicy := &KeyPolicy{
			Version: "2012-10-17",
			ID:      "Key policy for CloudTrail",
			Statements: []Statement{
				{
					Sid:    "Enable IAM User Permissions",
					Effect: "Allow",
					Action: []string{
						"kms:*",
					},
					Resource: "*",
					Principal: Principal{
						AWS: []string{
							fmt.Sprintf("arn:aws:iam::%s:root", logAccountId),
						},
					},
				},
				{
					Sid:    "Enable CloudTrail Encrypt Permissions",
					Effect: "Allow",
					Action: []string{
						"kms:GenerateDataKey*",
					},
					Resource: "*",
					Principal: Principal{
						Service: "cloudtrail.amazonaws.com",
					},
					Condition: &Condition{
						StringLike: map[string][]string{
							"kms:EncryptionContext:aws:cloudtrail:arn": encryptCondition,
						},
					},
				},
				{
					Sid:    "Allow CloudTrail to describe key",
					Effect: "Allow",
					Action: []string{
						"kms:DescribeKey",
					},
					Resource: "*",
					Principal: Principal{
						Service: "cloudtrail.amazonaws.com",
					},
				},
			},
		}

		keyPolicy, err := json.Marshal(rawKeyPolicy)

		return string(keyPolicy), err
	}).(pulumi.StringOutput)

	return policy
}
