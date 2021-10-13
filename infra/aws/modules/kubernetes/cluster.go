package kubernetes

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EksControlPlaneArgs struct {
	Version 		string
	Name 			string
	PublicSubnets 	pulumi.StringArrayOutput
	PrivateSubnets 	pulumi.StringArrayOutput
	VpcId			pulumi.IDOutput
	AccountId       string
	IamRoles        EksIamRoles
	ExposeAdminEndpoint bool
}
func NewEksControlPlane(ctx *pulumi.Context, args EksControlPlaneArgs) (*eks.Cluster, error) {
	// This key will be used to create an encrypted envelope around K8s secrets and config.
	// Although k8s already generates its own encryption key to encrypt data it stores in etcd, we use envelope encryption
	// as it is recommended security best practice.
	// **NB** It's impossible to recover the cluster if this key is deleted. We therefore add Pulumi protection.
	key, err := kms.NewKey(ctx, "cluster-encryption-key", &kms.KeyArgs{
		Tags: pulumi.StringMap{
			"name": pulumi.String(args.Name + "-encryption-key"),
		},
	}, pulumi.Protect(true))
	if err != nil { return nil, err }

	cluster, err := eks.NewCluster(ctx, args.Name, &eks.ClusterArgs{
		Name: 						  pulumi.String(args.Name),
		Version: 					  pulumi.String(args.Version),
		VpcId:                        args.VpcId,
		PublicSubnetIds:              args.PublicSubnets, // Pulumi will manage the tagging of the subnets
		PrivateSubnetIds:             args.PrivateSubnets,
		// **NB** The role used to create the cluster (allow-full-access-from-other-accounts in this case) is mapped to the k8s 
		// cluster-admin automatically and won't show up in the role mappings. We will not override this as we want to be able to
		// manage eks provisioning using pulumi.
		RoleMappings: eks.RoleMappingArray{
			// Provides full administrator cluster access to the k8s cluster
			eks.RoleMappingArgs{
				RoleArn: args.IamRoles.Admin.Arn,
				Groups: pulumi.StringArray{pulumi.String("system:masters")},
				Username: args.IamRoles.Admin.Arn,
			},
			// Map IAM automation role arn to the k8s automation group. The role bindings are set up in `ConfigureAutomationRole`.
			eks.RoleMappingArgs{
				RoleArn: args.IamRoles.Automation.Arn,
				Groups: pulumi.StringArray{pulumi.String("automation")},
				Username: args.IamRoles.Automation.Arn,
			},
		},
		StorageClasses: eks.StorageClassArgs{ // Defaults to immediate volume binding
			Type: 					pulumi.String("gp2"),
			Encrypted: 				pulumi.Bool(true),
			Default: 				pulumi.Bool(true),
		},
		SkipDefaultNodeGroup: 		pulumi.Bool(true),	// We will manage the node group separately from the control plane
		EndpointPublicAccess: 		pulumi.Bool(args.ExposeAdminEndpoint),
		EndpointPrivateAccess: 		pulumi.Bool(true),
		NodeAssociatePublicIpAddress: pulumi.Bool(false),
		InstanceRoles: 				iam.RoleArray{      // IAM roles to register with the cluster auth.
			args.IamRoles.NodeGroup,
		},
		NodeSecurityGroupTags:     pulumi.StringMap{ 	// Run with default node security group for now.
			"Name": pulumi.String("eksNodeSecurityGroup"),
		},
		ClusterSecurityGroupTags: pulumi.StringMap{ 	// Run with default cluster security group for now. This will enable full internet egress and ingress from node groups
			"Name": pulumi.String("eksNodeSecurityGroup"),
		},
		EnabledClusterLogTypes: pulumi.StringArray{
			pulumi.String("api"),
			pulumi.String("audit"),
			pulumi.String("authenticator"),
			pulumi.String("controllerManager"),
			pulumi.String("scheduler"),
		},
		EncryptionConfigKeyArn: key.Arn, // enable envelope encyption https://aws.amazon.com/about-aws/whats-new/2020/03/amazon-eks-adds-envelope-encryption-for-secrets-with-aws-kms/
		CreateOidcProvider: pulumi.Bool(true),
	})
	if err != nil { return nil, err }	

	return cluster, nil
}
