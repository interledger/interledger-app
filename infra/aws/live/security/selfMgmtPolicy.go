package main

import (
	"encoding/json"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateSlfMgmtPolicy() pulumi.String {

	type Condition struct {
		BoolIfExists map[string]string `json:"BoolIfExists,omitempty"`
	}

	type Statement struct {
		Sid       string     `json:"Sid"`
		Effect    string     `json:"Effect"`
		Action    []string   `json:"Action,omitempty"`
		NotAction []string   `json:"NotAction,omitempty"`
		Resource  string     `json:"Resource"`
		Condition *Condition `json:"Condition,omitempty"`
	}

	type Policy struct {
		Version    string      `json:"Version"`
		ID         string      `json:"Id"`
		Statements []Statement `json:"Statement"`
	}

	rawPolicy := &Policy{
		Version: "2012-10-17",
		ID:      "IAM User self management policy",
		Statements: []Statement{
			{
				Sid:    "AllowViewAccountInfo",
				Effect: "Allow",
				Action: []string{
					"iam:ListVirtualMFADevices",
				},
				Resource: "*",
			},
			{
				Sid:    "AllowManageOwnVirtualMFADevice",
				Effect: "Allow",
				Action: []string{
					"iam:CreateVirtualMFADevice",
					"iam:DeleteVirtualMFADevice",
				},
				Resource: "arn:aws:iam::*:mfa/${aws:username}",
			},
			{
				Sid:    "AllowManageOwnUserMFA",
				Effect: "Allow",
				Action: []string{
					"iam:DeactivateMFADevice",
					"iam:EnableMFADevice",
					"iam:GetUser",
					"iam:ListMFADevices",
					"iam:ResyncMFADevice",
				},
				Resource: "arn:aws:iam::*:user/${aws:username}",
			},
			{
				Sid:    "AllowManageOwnAccessKeys",
				Effect: "Allow",
				Action: []string{
					"iam:CreateAccessKey",
					"iam:DeleteAccessKey",
					"iam:ListAccessKeys",
					"iam:UpdateAccessKey",
				},
				Resource: "arn:aws:iam::*:user/${aws:username}",
			},
			{
				Sid:    "DenyAllExceptListedIfNoMFA",
				Effect: "Deny",
				NotAction: []string{
					"iam:CreateVirtualMFADevice",
					"iam:EnableMFADevice",
					"iam:GetUser",
					"iam:ListMFADevices",
					"iam:ListVirtualMFADevices",
					"iam:ResyncMFADevice",
					"sts:GetSessionToken",
				},
				Resource: "*",
				Condition: &Condition{
					BoolIfExists: map[string]string{
						"aws:MultiFactorAuthPresent": "false",
					},
				},
			},
		},
	}

	policy, _ := json.Marshal(rawPolicy)

	return pulumi.String(policy)
}
