package main

import (
	"encoding/base64"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
	"os"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		vpcId := vpcStack.GetStringOutput(pulumi.String("vpcId"))
		publicSubnet := utils.ValueFromStringArrayOutput(vpcStack.GetOutput(pulumi.String("publicSubnets")), 1)
		sdmToken := os.Getenv("SDM_TOKEN")

		sg, err := ec2.NewSecurityGroup(ctx, "strongdm-sg", &ec2.SecurityGroupArgs{
			Name:        pulumi.String("strongdm-gateway"),
			VpcId:       vpcId,
			Description: pulumi.String("Strongdm Gateway security Group"),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{ // controller api endpoint
					FromPort: pulumi.Int(5000),
					ToPort:   pulumi.Int(5000),
					Protocol: pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"), // allow from anywhere as users will use cli on their local machines. i.e. network load balancer does not adjust src ip address.
					},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				// Allow outbound communication to everywhere.
				&ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			MostRecent: pulumi.BoolRef(true),
			Owners: []string{
				"amazon",
			},
			Filters: []ec2.GetAmiFilter{
				{
					Name: "name",
					Values: []string{
						"amzn2-ami-hvm-*-x86_64-gp2",
					},
				},
			},
		})
		if err != nil {
			return err
		}

		instanceUserData := fmt.Sprintf(`
		#!/bin/bash -xe
		# set environement varibles
		export TARGET_USER=ec2-user
		export SDM_LISTEN_PORT=5000
		export SDM_GATEWAY_NAME=aws-shared-eu-west-1-$(date +%%s)
		export SDM_HOSTNAME="$(curl http://169.254.169.254/latest/meta-data/public-hostname)"
		# downloads sdm binary
		yum update -y && yum install -y unzip curl
		curl -J -O -L https://app.strongdm.com/releases/cli/linux && unzip sdmcli* && rm sdmcli*
		# Generate a gateway token
		export SDM_RELAY_TOKEN="$(./sdm --admin-token=%s relay create-gateway --name $SDM_GATEWAY_NAME $SDM_HOSTNAME:$SDM_LISTEN_PORT 0.0.0.0:$SDM_LISTEN_PORT)"
		# Install SDM
		sudo ./sdm install --relay --token=$SDM_RELAY_TOKEN
		`, sdmToken)

		_, err = ec2.NewInstance(ctx, "strongdm-gateway", &ec2.InstanceArgs{
			Ami:          pulumi.String(ami.Id),
			InstanceType: pulumi.String("t3.medium"),
			SubnetId:     publicSubnet,
			VpcSecurityGroupIds: pulumi.StringArray{
				sg.ID(),
			},
			UserDataBase64:           pulumi.String(base64.StdEncoding.EncodeToString([]byte(instanceUserData))),
			Tags:                     pulumi.StringMap{"Name": pulumi.String("StrongDM Gateway")},
			AssociatePublicIpAddress: pulumi.BoolPtr(true),
		}, pulumi.ReplaceOnChanges([]string{"*"}))
		if err != nil {
			return err
		}

		return nil
	})
}
