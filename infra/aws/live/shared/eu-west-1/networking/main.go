package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/networking"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		err := networking.NewVpc(ctx, &networking.VpcArgs{
			Name:      "shared",
			CidrBlock: "10.100.0.0/16",
			AvailabilityZones: []string{
				"eu-west-1a",
				"eu-west-1b",
				"eu-west-1c",
			},
			// Public allocated to 10.100.128.0/18
			PublicSubnets: []string{
				"10.100.128.0/20",
				"10.100.144.0/20",
				"10.100.160.0/20",
			},
			// Private allocated to 10.100.0.0/17
			PrivateSubnets: []string{
				"10.100.0.0/19",
				"10.100.32.0/19",
				"10.100.64.0/19",
			},
			// Intra allocated to 10.100.192.0/18
			IntraSubnets: []string{
				"10.100.192.0/20",
				"10.100.208.0/20",
				"10.100.224.0/20",
			},
			EnableNatGateway: true,
			SingleNatGateway: true,
			PublicInboundNacls: ec2.NetworkAclIngressArray{
				// HTTP Access
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(100),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(80),
					ToPort:    pulumi.Int(80),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// HTTPS Access
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(110),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(443),
					ToPort:    pulumi.Int(443),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// SSH Access
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(120),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(22),
					ToPort:    pulumi.Int(22),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// NAT Gateway Access
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(130),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(1024),
					ToPort:    pulumi.Int(65535),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
			},
			PublicOutboundNacls: ec2.NetworkAclEgressArray{
				// HTTP Access
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(100),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(80),
					ToPort:    pulumi.Int(80),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// HTTPS Access
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(110),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(443),
					ToPort:    pulumi.Int(443),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// NAT Gateway Access
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(130),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(1024),
					ToPort:    pulumi.Int(65535),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// SSH Access to Private subnet
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(120),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(22),
					ToPort:    pulumi.Int(22),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("10.100.0.0/17"),
				},
			},
			PrivateInboundNacls: ec2.NetworkAclIngressArray{
				// NAT Gateway allow returning traffic
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(140),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(1024),
					ToPort:    pulumi.Int(65535),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// SSH Access from public subnet
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(120),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(22),
					ToPort:    pulumi.Int(22),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("10.100.128.0/18"),
				},
				// Allow all comm within Private subnet
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(130),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.100.0.0/17"),
				},
			},
			PrivateOutboundNacls: ec2.NetworkAclEgressArray{
				// HTTP Access
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(100),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(80),
					ToPort:    pulumi.Int(80),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// HTTPS Access
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(110),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(443),
					ToPort:    pulumi.Int(443),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				// Allow all comm within Private subnet
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(130),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.100.0.0/17"),
				},
				// NAT Gateway Access (May not be needed)
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(140),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(1024),
					ToPort:    pulumi.Int(65535),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("10.100.128.0/18"),
				},
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
