package networking

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VpcArgs struct {
	Name                 string
	CidrBlock            string
	AvailabilityZones    []string
	PublicSubnets        []string
	PrivateSubnets       []string
	IntraSubnets         []string
	EnableNatGateway     bool
	SingleNatGateway     bool
	PublicInboundNacls   ec2.NetworkAclIngressArray
	PublicOutboundNacls  ec2.NetworkAclEgressArray
	PrivateInboundNacls  ec2.NetworkAclIngressArray
	PrivateOutboundNacls ec2.NetworkAclEgressArray
	IntraInboundNacls    ec2.NetworkAclIngressArray
	IntraOutboundNacls   ec2.NetworkAclEgressArray
}

func NewVpc(ctx *pulumi.Context, args *VpcArgs) error {

	vpc, err := ec2.NewVpc(ctx, args.Name, &ec2.VpcArgs{
		CidrBlock:                    pulumi.String(args.CidrBlock),
		AssignGeneratedIpv6CidrBlock: pulumi.Bool(false),
		EnableDnsHostnames:           pulumi.Bool(true),
		EnableDnsSupport:             pulumi.Bool(true),
		InstanceTenancy:              pulumi.String("default"),
	})
	if err != nil {
		return err
	}
	//  Remove all rules from default security group
	_, err = ec2.NewDefaultSecurityGroup(ctx, "_default", &ec2.DefaultSecurityGroupArgs{
		VpcId: vpc.ID(),
	})
	if err != nil {
		return err
	}
	//  Remove all rules from default NACL
	_, err = ec2.NewDefaultNetworkAcl(ctx, "_default", &ec2.DefaultNetworkAclArgs{
		DefaultNetworkAclId: vpc.DefaultNetworkAclId,
		SubnetIds:           nil,
	})
	if err != nil {
		return err
	}

	// Create public subnets
	var publicSubnets []*ec2.Subnet
	for i, s := range args.PublicSubnets {
		az := args.AvailabilityZones[i]
		name := fmt.Sprintf("%s-public-%s", args.Name, az)
		subnet, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(s),
			AvailabilityZone: pulumi.String(args.AvailabilityZones[i]),
		})
		if err != nil {
			return err
		}
		publicSubnets = append(publicSubnets, subnet)
	}
	publicRouteTable, err := ec2.NewRouteTable(ctx, fmt.Sprintf("%s-public", args.Name), &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
	})
	if err != nil {
		return err
	}
	// Associate all public subnets to route table
	for i, s := range publicSubnets {
		az := args.AvailabilityZones[i]
		name := fmt.Sprintf("%s-public-%s", args.Name, az)
		_, err = ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			SubnetId:     s.ID(),
			RouteTableId: publicRouteTable.ID(),
		})
		if err != nil {
			return err
		}
	}
	// Add internet gateway for public subnets
	ig, err := ec2.NewInternetGateway(ctx, fmt.Sprintf("%s-ig", args.Name), &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
	})
	_, err = ec2.NewRoute(ctx, "public-route", &ec2.RouteArgs{
		RouteTableId:         publicRouteTable.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		GatewayId:            ig.ID(),
	})

	// Create private subnets
	var privateSubnets []*ec2.Subnet
	for i, s := range args.PrivateSubnets {
		az := args.AvailabilityZones[i]
		name := fmt.Sprintf("%s-private-%s", args.Name, az)
		subnet, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(s),
			AvailabilityZone: pulumi.String(args.AvailabilityZones[i]),
		})
		if err != nil {
			return err
		}
		privateSubnets = append(privateSubnets, subnet)
	}

	// Setup NAT Gateway
	natGatewayCount := 0
	if args.EnableNatGateway && args.SingleNatGateway {
		natGatewayCount = 1
	} else if args.EnableNatGateway && !args.SingleNatGateway {
		natGatewayCount = len(args.AvailabilityZones)
	}

	var natGateways []*ec2.NatGateway
	for i := 0; i < natGatewayCount; i++ {
		az := args.AvailabilityZones[i]
		ip, err := ec2.NewEip(ctx, fmt.Sprintf("%s-%s", args.Name, az), &ec2.EipArgs{
			Vpc: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}
		natGateway, err := ec2.NewNatGateway(ctx, fmt.Sprintf("%s-%s", args.Name, az), &ec2.NatGatewayArgs{
			SubnetId:     publicSubnets[i].ID(),
			AllocationId: ip.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{ig}))
		if err != nil {
			return err
		}
		natGateways = append(natGateways, natGateway)
	}
	// Create a route table per NAT Gateway for the private layer
	var privateRouteTables []*ec2.RouteTable
	for i, ng := range natGateways {
		az := args.AvailabilityZones[i]
		name := fmt.Sprintf("%s-private-%s", args.Name, az)
		rt, err := ec2.NewRouteTable(ctx, name, &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewRoute(ctx, "nat-route", &ec2.RouteArgs{
			RouteTableId:         rt.ID(),
			DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			NatGatewayId:         ng.ID(),
		})
		if err != nil {
			return err
		}

		privateRouteTables = append(privateRouteTables, rt)
	}
	// Associate Private Route Tables
	for i, sn := range privateSubnets {
		az := args.AvailabilityZones[i]
		name := fmt.Sprintf("%s-private-%s", args.Name, az)
		var routeTableIndex = i

		// If we only have a single nat gateway use the first
		if args.SingleNatGateway {
			routeTableIndex = 0
		}

		_, err = ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			SubnetId:     sn.ID(),
			RouteTableId: privateRouteTables[routeTableIndex].ID(),
		})
	}

	// Create intra subnets
	var intraSubnets []*ec2.Subnet
	if len(args.IntraSubnets) > 0 {
		for i, s := range args.IntraSubnets {
			az := args.AvailabilityZones[i]
			name := fmt.Sprintf("%s-intra-%s", args.Name, az)
			subnet, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
				VpcId:            vpc.ID(),
				CidrBlock:        pulumi.String(s),
				AvailabilityZone: pulumi.String(args.AvailabilityZones[i]),
			})
			if err != nil {
				return err
			}
			intraSubnets = append(intraSubnets, subnet)
		}
		intraRouteTable, err := ec2.NewRouteTable(ctx, fmt.Sprintf("%s-intra", args.Name), &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		})
		if err != nil {
			return err
		}

		// Associate all intra subnets to route table
		for i, s := range intraSubnets {
			az := args.AvailabilityZones[i]
			name := fmt.Sprintf("%s-intra-%s", args.Name, az)
			_, err = ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
				SubnetId:     s.ID(),
				RouteTableId: intraRouteTable.ID(),
			})
			if err != nil {
				return err
			}
		}
	}

	// Setup network ACL's
	publicSubnetIds := subnetArrayToPulumiIDArray(publicSubnets)
	privateSubnetIds := subnetArrayToPulumiIDArray(privateSubnets)
	intraSubnetIds := subnetArrayToPulumiIDArray(intraSubnets)

	_, err = ec2.NewNetworkAcl(ctx, fmt.Sprintf("%s-public", args.Name), &ec2.NetworkAclArgs{
		VpcId:     vpc.ID(),
		SubnetIds: pulumi.StringArrayOutput(publicSubnetIds),
		Ingress:   args.PublicInboundNacls,
		Egress:    args.PublicOutboundNacls,
	})

	_, err = ec2.NewNetworkAcl(ctx, fmt.Sprintf("%s-private", args.Name), &ec2.NetworkAclArgs{
		VpcId:     vpc.ID(),
		SubnetIds: pulumi.StringArrayOutput(privateSubnetIds),
		Ingress:   args.PrivateInboundNacls,
		Egress:    args.PrivateOutboundNacls,
	})

	_, err = ec2.NewNetworkAcl(ctx, fmt.Sprintf("%s-intra", args.Name), &ec2.NetworkAclArgs{
		VpcId:     vpc.ID(),
		SubnetIds: pulumi.StringArrayOutput(intraSubnetIds),
		Ingress:   args.IntraInboundNacls,
		Egress:    args.IntraOutboundNacls,
	})

	// Exports
	ctx.Export("vpcId", vpc.ID())
	ctx.Export("vpcArn", vpc.Arn)
	ctx.Export("vpcCidrBlock", vpc.CidrBlock)

	ctx.Export("publicSubnets", publicSubnetIds)
	ctx.Export("privateSubnets", privateSubnetIds)
	ctx.Export("intraSubnets", intraSubnetIds)

	ctx.Export("publicSubnetsCidrBlocks", subnetArrayToCidrBlockArray(publicSubnets))
	ctx.Export("privateSubnetsCidrBlocks", subnetArrayToCidrBlockArray(privateSubnets))
	ctx.Export("intraSubnetsCidrBlocks", subnetArrayToCidrBlockArray(intraSubnets))

	return nil
}

func subnetArrayToPulumiIDArray(subnets []*ec2.Subnet) pulumi.IDArrayOutput {
	var outputs []interface{}
	for _, sn := range subnets {
		outputs = append(outputs, sn.ID())
	}
	return pulumi.All(outputs...).ApplyT(func(vs []interface{}) []pulumi.ID {
		var results []pulumi.ID
		for _, v := range vs {
			results = append(results, v.(pulumi.ID))
		}
		return results
	}).(pulumi.IDArrayOutput)
}

func subnetArrayToCidrBlockArray(subnets []*ec2.Subnet) pulumi.StringArrayOutput {
	var outputs []interface{}
	for _, sn := range subnets {
		outputs = append(outputs, sn.CidrBlock)
	}
	return pulumi.All(outputs...).ApplyT(func(vs []interface{}) []string {
		var results []string
		for _, v := range vs {
			results = append(results, v.(string))
		}
		return results
	}).(pulumi.StringArrayOutput)
}
