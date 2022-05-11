package kubernetes

import (
	eks2 "github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ManagedNodeGroupArgs struct {
	Cluster       *eks.Cluster
	MinSize       int
	MaxSize       int
	DesiredSize   int
	Name          string
	InstanceTypes pulumi.StringArray
	NodeRole      *iam.Role
	SubnetIds     pulumi.StringArrayOutput // This **MUST** be the private subnets for the VPC
}

func NewManagedNodeGroup(ctx *pulumi.Context, args ManagedNodeGroupArgs) (*eks.ManagedNodeGroup, error) {
	ng, err := eks.NewManagedNodeGroup(ctx, args.Name, &eks.ManagedNodeGroupArgs{
		NodeGroupName: pulumi.String(args.Name),
		ScalingConfig: eks2.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(args.DesiredSize),
			MinSize:     pulumi.Int(args.MinSize),
			MaxSize:     pulumi.Int(args.MaxSize),
		},
		InstanceTypes: args.InstanceTypes,
		Cluster:       args.Cluster.Core,
		NodeRole:      args.NodeRole,
		SubnetIds:     args.SubnetIds,
		AmiType:       pulumi.String("BOTTLEROCKET_x86_64"),
	})
	if err != nil {
		return nil, err
	}

	return ng, nil
}
