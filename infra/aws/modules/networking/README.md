# AWS VPC Module

Pulumi module which creates VPC resources on AWS.

## Usage

```golang
err := networking.NewVpc(ctx, &networking.VpcArgs{
			Name:      "my-vpc",
			CidrBlock: "10.0.0.0/16",
			AvailabilityZones: []string{
				"eu-west-1a",
				"eu-west-1b",
				"eu-west-1c",
			},
			// Public allocated to 10.0.128.0/18
			PublicSubnets: []string{
				"10.0.128.0/20",
				"10.0.144.0/20",
				"10.0.160.0/20",
			},
			// Public allocated to 10.0.0.0/17
			PrivateSubnets: []string{
				"10.0.0.0/19",
				"10.0.32.0/19",
				"10.0.64.0/19",
			},
			// Public allocated to 10.0.192.0/18
			IntraSubnets: []string{
				"10.0.192.0/20",
				"10.0.208.0/20",
				"10.0.224.0/20",
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
```
## Subnet Types

This module supports three different Subnet types, `public`, `private` and `intra`

### Public

Public subnets are subnets that can be accessible from the Internet. Resources created in these can be assigned public IP
addresses and are also able to route traffic out to `0.0.0.0/0` via the Internet Gateway

Typically, these subnets will host your load balancers and bastions servers.

### Private

Private subnets are subnets which are not directly accessible by the internet. Resources created within here can not get
a public IP address. Further to be able to route traffic to the internet, you need to create NAT Gateways. All outbound
traffic will then route through the NAT Gateways.

Typically, these subnets will host your application servers that require outside internet connections. Such as, calling 
external API's like Stripe.

### Intra
Private subnets are subnets which are not directly accessible by the internet and can also not access `0.0.0.0/0`. Resources
in this subnet are only accessible from within your VPC and can also only. These needs to be configured via Network ACL's

Typically, these subnets will host your persistence stores such as databases and secrets stores.

## NAT Gateways
This module supports two scenarios for creating NAT gateways. Each will be explained in further detail in the 
corresponding sections. To ensure NAT Gateways are created ensure to enable them with `EnableNatGateway=true`

* One NAT Gateway per availability zone
  * EnableNatGateway = true
  * SingleNatGateway = false
* Single NAT Gateway
  * EnableNatGateway = true
  * SingleNatGateway = true
  
This module will provision new Elastic IPs for the VPC's NAT Gateways.

### Single NAT Gateway

If `SingleNatGateway = true`, then all private subnets will route their Internet traffic through this single NAT gateway. 
The NAT gateway will be placed in the first public subnet in your `publicSubnets` block.

### One NAT Gateway per availability zone

If `EnableNatGateway = true` and `SingleNatGateway = false`, then the module will place one NAT gateway in each 
availability zone you specify.

## Network Access Control Lists (ACL or NACL)

This module does NOT provide any default Network ACL's. Above is an example of a configuration, but users are required
to add their own Network ACL's to ensure correct flow of traffic.