package main

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"os"
	"text/template"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		gitlabToken := os.Getenv("GITLAB_TOKEN")
		if gitlabToken == "" {
			return fmt.Errorf("gitlab token was not provided")
		}
		baselineStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-baseline/main", nil)
		if err != nil {
			return err
		}
		ebsKmsKeyArn := baselineStack.GetStringOutput(pulumi.String("ebsEncryptionKeyArn"))
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		privateSubnets := vpcStack.GetOutput(pulumi.String("privateSubnets"))
		privateSubnet := privateSubnets.ApplyT(func(arg interface{}) string {
			stringArray := arg.([]interface{})
			return stringArray[0].(string)
		}).(pulumi.StringOutput)
		vpcId := vpcStack.GetStringOutput(pulumi.String("vpcId"))
		privateSubnetCidrBlocks := vpcStack.GetOutput(pulumi.String("privateSubnetsCidrBlocks")).ApplyT(func(arg interface{}) []string {
			stringArray := arg.([]interface{})
			var outputs []string
			for _, s := range stringArray {
				outputs = append(outputs, s.(string))
			}
			return outputs
		}).(pulumi.StringArrayOutput)

		// Create a runner role and instance profile
		runnerRole, err := CreateRunnerRole(ctx)
		if err != nil {
			return err
		}
		// export so that we can configure eksAutomationRole to be assumeable by the gl-runner role.
		ctx.Export("glRunnerRoleArn", runnerRole.Arn)
		runnerInstanceProfile, err := iam.NewInstanceProfile(ctx, "gl-runner-instance", &iam.InstanceProfileArgs{
			Role: runnerRole.Name,
		})
		if err != nil {
			return err
		}

		// Create a manager role and instance profile
		managerRole, err := CreateManagerRole(ctx, ebsKmsKeyArn, runnerRole)
		if err != nil {
			return err
		}
		managerInstanceProfile, err := iam.NewInstanceProfile(ctx, "gl-runner-manager", &iam.InstanceProfileArgs{
			Role: managerRole.Name,
		})
		if err != nil {
			return err
		}

		// Create an EC2 Instance for the runner manager
		opt0 := true
		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			MostRecent: &opt0,
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

		managerSg, err := ec2.NewSecurityGroup(ctx, "gl-runner-manager", &ec2.SecurityGroupArgs{
			Name:        pulumi.String("gl-runner-manager"),
			VpcId:       vpcId,
			Description: pulumi.String("GL Runner Manager"),
			Ingress:     ec2.SecurityGroupIngressArray{},
			Egress: ec2.SecurityGroupEgressArray{
				// Allow HTTPS outbound communication. Possible limit to Cloudflare only?
				&ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(443),
					ToPort:   pulumi.Int(443),
					Protocol: pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
				// Allow docker outbound communication to docker-machines
				&ec2.SecurityGroupEgressArgs{
					FromPort:   pulumi.Int(2376),
					ToPort:     pulumi.Int(2376),
					Protocol:   pulumi.String("tcp"),
					CidrBlocks: privateSubnetCidrBlocks,
				},
				&ec2.SecurityGroupEgressArgs{
					// Allow ssh outbound communication to docker-machines
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					Protocol:   pulumi.String("tcp"),
					CidrBlocks: privateSubnetCidrBlocks,
				},
			},
		})
		if err != nil {
			return err
		}

		sg, err := ec2.NewSecurityGroup(ctx, "gl-runner-machine", &ec2.SecurityGroupArgs{
			Name:        pulumi.String("gl-runner-machine"),
			VpcId:       vpcId,
			Description: pulumi.String("Docker machine for GL Runner"),
			Ingress: ec2.SecurityGroupIngressArray{
				// Allow docker inbound communication from manager
				&ec2.SecurityGroupIngressArgs{
					FromPort:   pulumi.Int(2376),
					ToPort:     pulumi.Int(2376),
					Protocol:   pulumi.String("tcp"),
					CidrBlocks: privateSubnetCidrBlocks,
				},
				// Allow SSH inbound communication from manager
				&ec2.SecurityGroupIngressArgs{
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					Protocol:   pulumi.String("tcp"),
					CidrBlocks: privateSubnetCidrBlocks,
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				// Allow all outbound communication.
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
		// export runner sg so that the k8s cluster sg can be configured to allow inbound tcp from it.
		ctx.Export("glRunnerSecurityGroupID", sg.ID())

		tmplt, err := ec2.NewLaunchTemplate(ctx, "gl-runner-manager", &ec2.LaunchTemplateArgs{
			NamePrefix:   pulumi.String("gl-runner-manager"),
			ImageId:      pulumi.String(ami.ImageId),
			InstanceType: pulumi.String("t2.nano"),
			IamInstanceProfile: &ec2.LaunchTemplateIamInstanceProfileArgs{
				Arn: managerInstanceProfile.Arn,
			},
			UserData: cloudInitData(cloudInitDataArgs{
				Token:             gitlabToken,
				VpcId:             vpcId,
				SubnetId:          privateSubnet,
				SgName:            sg.Name,
				RunnerProfileName: runnerInstanceProfile.Name,
			}),
			VpcSecurityGroupIds: pulumi.StringArray{
				managerSg.ID(),
			},
			Tags: pulumi.StringMap{"Name": pulumi.String("gl-manager")},
		})
		if err != nil {
			return err
		}
		_, err = autoscaling.NewGroup(ctx, "gl-runner-manager", &autoscaling.GroupArgs{
			DesiredCapacity: pulumi.Int(1),
			MaxSize:         pulumi.Int(1),
			MinSize:         pulumi.Int(1),
			LaunchTemplate: &autoscaling.GroupLaunchTemplateArgs{
				Id:      tmplt.ID(),
				Version: pulumi.String(fmt.Sprintf("%v%v", "$", "Latest")),
			},
			VpcZoneIdentifiers: pulumi.StringArrayOutput(privateSubnets),
		})
		if err != nil {
			return err
		}
		return nil
	})
}

type cloudInitDataArgs struct {
	Token             string
	VpcId             pulumi.StringOutput
	SubnetId          pulumi.StringOutput
	SgName            pulumi.StringOutput
	RunnerProfileName pulumi.StringOutput
}

func cloudInitData(
	opts cloudInitDataArgs,
) pulumi.StringOutput {
	type Data struct {
		GitlabToken       string
		AwsRegion         string
		AwsVpcId          string
		AwsSubnetId       string
		SgName            string
		RunnerProfileName string
	}
	return pulumi.All(
		opts.VpcId,
		opts.SubnetId,
		opts.SgName,
		opts.RunnerProfileName,
	).ApplyT(
		func(args []interface{}) (string, error) {
			data := Data{
				GitlabToken:       opts.Token,
				AwsRegion:         "eu-west-1",
				AwsVpcId:          args[0].(string),
				AwsSubnetId:       args[1].(string),
				SgName:            args[2].(string),
				RunnerProfileName: args[3].(string),
			}

			tmp, err := template.ParseFiles("./cloudinit.yaml")
			if err != nil {
				return "", err
			}
			document := &bytes.Buffer{}
			err = tmp.Execute(document, data)
			if err != nil {
				return "", err
			}

			return b64.StdEncoding.EncodeToString(document.Bytes()), nil
		}).(pulumi.StringOutput)
}
