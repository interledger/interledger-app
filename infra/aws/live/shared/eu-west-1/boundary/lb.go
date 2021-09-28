package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createLBSecurityGroup(ctx *pulumi.Context, vpcId pulumi.StringOutput) (*ec2.SecurityGroup, error) {
	return ec2.NewSecurityGroup(ctx, "lb-boundary-controller", &ec2.SecurityGroupArgs{
		Name:        pulumi.String("lb-boundary-controller"),
		Description: pulumi.String("SG for ALB for boundary controller"),
		VpcId:       vpcId,
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Description: pulumi.String("All incoming traffic for controller port 443"),
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(443),
				ToPort:      pulumi.Int(443),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			&ec2.SecurityGroupIngressArgs{
				Description: pulumi.String("All incoming traffic for worker port 9202"),
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(9202),
				ToPort:      pulumi.Int(9202),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
	})
}

func createLB(ctx *pulumi.Context, vpcId pulumi.StringOutput, publicSubnets pulumi.StringArrayOutput) (*lb.LoadBalancer, *lb.TargetGroup, *lb.TargetGroup, error) {

	networkLb, err := lb.NewLoadBalancer(ctx, "boundary-controller", &lb.LoadBalancerArgs{
		Name:             pulumi.String("boundary-controller"),
		LoadBalancerType: pulumi.String("network"),
		Internal:         pulumi.Bool(false),
		Subnets: publicSubnets,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	controllerTg, err := lb.NewTargetGroup(ctx, "controller-tg", &lb.TargetGroupArgs{
		Port:       pulumi.Int(9200),
		Protocol:   pulumi.String("TCP"),
		VpcId:      vpcId,
		TargetType: pulumi.String("instance"),
		HealthCheck: lb.TargetGroupHealthCheckArgs{
			Enabled:  pulumi.Bool(true),
			Protocol: pulumi.String("TCP"),
			Port:     pulumi.String("9200"),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	workerTg, err := lb.NewTargetGroup(ctx, "worker-tg", &lb.TargetGroupArgs{
		Port:       pulumi.Int(9202),
		Protocol:   pulumi.String("TCP"),
		VpcId:      vpcId,
		TargetType: pulumi.String("instance"),
		HealthCheck: lb.TargetGroupHealthCheckArgs{
			Enabled:  pulumi.Bool(true),
			Protocol: pulumi.String("TCP"),
			Port:     pulumi.String("9202"),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	_, err = lb.NewListener(ctx, "controller-tcp-listener", &lb.ListenerArgs{
		LoadBalancerArn: networkLb.Arn,
		Port:            pulumi.Int(443),
		Protocol:        pulumi.String("TCP"),
		DefaultActions: &lb.ListenerDefaultActionArray{
			lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: controllerTg.Arn,
			},
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	_, err = lb.NewListener(ctx, "worker-tcp-listener", &lb.ListenerArgs{
		LoadBalancerArn: networkLb.Arn,
		Port:            pulumi.Int(9202),
		Protocol:        pulumi.String("TCP"),
		DefaultActions: &lb.ListenerDefaultActionArray{
			lb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: workerTg.Arn,
			},
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return networkLb, controllerTg, workerTg, nil
}

func addToTargetGroup(ctx *pulumi.Context, name string, targetGroup *lb.TargetGroup, instance *ec2.Instance, port int) error {
	_, err := lb.NewTargetGroupAttachment(ctx, name + "tg-attachment", &lb.TargetGroupAttachmentArgs{
		TargetGroupArn: targetGroup.Arn,
		TargetId:       instance.ID(),
		Port:           pulumi.Int(port),
	})
	if err != nil {
		return err
	}

	return nil
}
