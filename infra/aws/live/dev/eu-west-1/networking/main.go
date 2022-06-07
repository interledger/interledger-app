package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/networking"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		vpc, err := networking.NewVpc(ctx, &networking.VpcArgs{
			Name:      "dev",
			CidrBlock: "10.10.0.0/16",
			AvailabilityZones: []string{
				"eu-west-1a",
				"eu-west-1b",
				"eu-west-1c",
			},
			// Public allocated to 10.10.128.0/18
			PublicSubnets: []string{
				"10.10.128.0/20",
				"10.10.144.0/20",
				"10.10.160.0/20",
			},
			// Private allocated to 10.10.0.0/17
			PrivateSubnets: []string{
				"10.10.0.0/19",
				"10.10.32.0/19",
				"10.10.64.0/19",
			},
			// Intra allocated to 10.10.192.0/18
			IntraSubnets: []string{
				"10.10.192.0/20",
				"10.10.208.0/20",
				"10.10.224.0/20",
			},
			PublicSubnetTags: pulumi.StringMap{
				"kubernetes.io/role/elb": pulumi.String("1"),
			},
			PrivateSubnetTags: pulumi.StringMap{
				"kubernetes.io/role/internal-elb": pulumi.String("1"),
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
					CidrBlock: pulumi.String("10.10.0.0/17"),
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
					CidrBlock: pulumi.String("10.10.128.0/18"),
				},
				// Allow all comm within Private subnet
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(130),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.10.0.0/17"),
				},
				// Allow comm to intra subnet
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(150),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.10.192.0/18"),
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
					CidrBlock: pulumi.String("10.10.0.0/17"),
				},
				// NAT Gateway Access (May not be needed)
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(140),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(1024),
					ToPort:    pulumi.Int(65535),
					Protocol:  pulumi.String("tcp"),
					CidrBlock: pulumi.String("0.0.0.0/0"),
				},
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(150),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.10.192.0/18"),
				},
			},
			IntraInboundNacls: ec2.NetworkAclIngressArray{
				// Allow all traffic from private subnet
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(120),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.10.0.0/17"),
				},
				// Allow all traffic within intra
				ec2.NetworkAclIngressArgs{
					RuleNo:    pulumi.Int(130),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.10.192.0/18"),
				},
			},
			IntraOutboundNacls: ec2.NetworkAclEgressArray{
				// Allow all outbound to private subnet
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(120),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.10.0.0/17"),
				},
				// Allow all traffic within intra
				ec2.NetworkAclEgressArgs{
					RuleNo:    pulumi.Int(130),
					Action:    pulumi.String("allow"),
					FromPort:  pulumi.Int(0),
					ToPort:    pulumi.Int(0),
					Protocol:  pulumi.String("-1"),
					CidrBlock: pulumi.String("10.10.192.0/18"),
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("vpcId", vpc.ID())

		return nil
	})
}
