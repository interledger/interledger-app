package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcPeerStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-vpc-peering-dev/main", nil)
		if err != nil {
			return err
		}
		vpcPeeringConnectionId := vpcPeerStack.GetStringOutput(pulumi.String("VpcPeeringConnectionId"))

		_, err = ec2.NewVpcPeeringConnectionAccepter(ctx, "vpc-peer-shared", &ec2.VpcPeeringConnectionAccepterArgs{
			VpcPeeringConnectionId: vpcPeeringConnectionId,
			AutoAccept:             pulumi.BoolPtr(true),
		})

		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-dev-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		publicRouteTableId := vpcStack.GetStringOutput(pulumi.String("publicRoutingTableId"))

		_, err = ec2.NewRoute(ctx, "shared-peer-route", &ec2.RouteArgs{
			DestinationCidrBlock:   pulumi.String("10.100.0.0/16"),
			RouteTableId:           publicRouteTableId,
			VpcPeeringConnectionId: vpcPeeringConnectionId,
		})
		if err != nil {
			return err
		}

		privateRouteTableIds := vpcStack.GetOutput(pulumi.String("privateRoutingTableIds"))
		privateRouteTableIds.ApplyT(func(arg interface{}) error {
			ids := arg.([]interface{})
			for i, s := range ids {
				_, err = ec2.NewRoute(ctx, fmt.Sprintf("dev-private-peer-route-%d", i), &ec2.RouteArgs{
					DestinationCidrBlock:   pulumi.String("10.100.0.0/16"),
					RouteTableId:           pulumi.String(s.(string)),
					VpcPeeringConnectionId: vpcPeeringConnectionId,
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

		return nil
	})
}
