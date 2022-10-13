package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return nil
		}
		peerVpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-dev-euwest1-networking/main", nil)
		if err != nil {
			return nil
		}

		vpcId := vpcStack.GetStringOutput(pulumi.String("vpcId"))
		peerVpcId := peerVpcStack.GetStringOutput(pulumi.String("vpcId"))

		peerConnection, err := ec2.NewVpcPeeringConnection(ctx, "vpc-peer-dev", &ec2.VpcPeeringConnectionArgs{
			VpcId:       vpcId,
			PeerOwnerId: pulumi.String("634848879735"),
			PeerRegion:  pulumi.String("eu-west-1"),
			PeerVpcId:   peerVpcId,
		})

		ctx.Export("VpcPeeringConnectionId", peerConnection.ID())

		publicRouteTableId := vpcStack.GetStringOutput(pulumi.String("publicRoutingTableId"))

		_, err = ec2.NewRoute(ctx, "dev-public-peer-route", &ec2.RouteArgs{
			DestinationCidrBlock:   pulumi.String("10.10.0.0/16"),
			RouteTableId:           publicRouteTableId,
			VpcPeeringConnectionId: peerConnection.ID(),
		})
		if err != nil {
			return err
		}

		privateRouteTableIds := vpcStack.GetOutput(pulumi.String("privateRoutingTableIds"))
		privateRouteTableIds.ApplyT(func(arg interface{}) error {
			ids := arg.([]interface{})
			for i, s := range ids {
				_, err = ec2.NewRoute(ctx, fmt.Sprintf("dev-private-peer-route-%d", i), &ec2.RouteArgs{
					DestinationCidrBlock:   pulumi.String("10.10.0.0/16"),
					RouteTableId:           pulumi.String(s.(string)),
					VpcPeeringConnectionId: peerConnection.ID(),
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

		// Recommended to disable after association
		// https://aws.amazon.com/premiumsupport/knowledge-center/route53-private-hosted-zone/
		//_, err = route53.NewVpcAssociationAuthorization(ctx, "dev-dns-association", &route53.VpcAssociationAuthorizationArgs{
		//	VpcId:  peerVpcId,
		//	ZoneId: vpcStack.GetStringOutput(pulumi.String("dnsZoneId")),
		//})
		//if err != nil {
		//	return err
		//}

		return nil
	})
}
