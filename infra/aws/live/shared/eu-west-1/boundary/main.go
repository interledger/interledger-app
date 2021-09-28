package main

import (
	"bytes"
	b64 "encoding/base64"
	"text/template"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/rds"
	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		cfStack, err := pulumi.NewStackReference(ctx, "fynbos/cf-fynbos.dev/main", nil)
		if err != nil {
			return err
		}
		dnsZoneId := cfStack.GetIDOutput(pulumi.String("dnsZoneId"))
		tlsCert := cfStack.GetStringOutput(pulumi.String("boundaryCert"))
		tlsPrivateKey := cfStack.GetStringOutput(pulumi.String("boundaryPrivateKey"))
		vpcId := vpcStack.GetStringOutput(pulumi.String("vpcId"))
		intraSubnets := AnyOutputToStringArrayOutput(vpcStack.GetOutput(pulumi.String("intraSubnets")))
		privateSubnetsCidrBlocks := AnyOutputToStringArrayOutput(vpcStack.GetOutput(pulumi.String("privateSubnetsCidrBlocks")))
		privateSubnet := ValueFromStringArrayOutput(vpcStack.GetOutput(pulumi.String("privateSubnets")), 1)
		publicSubnets := AnyOutputToStringArrayOutput(vpcStack.GetOutput(pulumi.String("publicSubnets")))

		// Create Postgres
		pgSg, err := createPostgresSecurityGroup(ctx, vpcId, privateSubnetsCidrBlocks)
		if err != nil {
			return err
		}
		postgres, dbPassword, err := createPostgres(ctx, pgSg, intraSubnets)
		if err != nil {
			return err
		}
		ctx.Export("postgresEndpoint", postgres.Endpoint)

		recoveryKey, rootKey, workerKey, err := createKeys(ctx)
		if err != nil {
			return err
		}

		// Create Controller
		instanceSg, err := createControllerSecurityGroup(ctx, vpcId, privateSubnetsCidrBlocks)
		if err != nil {
			return err
		}
		controllerProfile, err := createControllerEc2Profile(ctx, recoveryKey.Arn, rootKey.Arn, workerKey.Arn)
		if err != nil {
			return err
		}

		// encrypt private key as it is used in instance user data which can be queried via instance meta data endpoint
		encryptedPrivateKey, err := kms.NewCiphertext(ctx, "encrypted-tls-private-key", &kms.CiphertextArgs{
			KeyId: workerKey.ID(), // using worker key as controller and worker are allowed to decrypt with it.
			Plaintext: tlsPrivateKey,
		})

		// TODO: provision with Vault certificate
		controller, err := createControllerInstance(CreateControllerArgs{
			ctx: ctx,
			subnetId: privateSubnet,
			sg: instanceSg,
			recoveryKey: recoveryKey,
			rootKey: rootKey,
			workerKey: workerKey,
			postgres: postgres,
			ec2Profile: controllerProfile,
			tlsCert: tlsCert,
			encryptedTlsPrivateKey: encryptedPrivateKey.CiphertextBlob,
			dbPassword: dbPassword.Result,
		})
		if err != nil {
			return err
		}

		// Create Load Balancer
		loadBalancer, controllerTg, workerTg, err := createLB(ctx, vpcId, publicSubnets)
		if err != nil {
			return err
		}

		// Create worker
		workerProfile, err := createWorkerEc2Profile(ctx, workerKey.Arn)
		if err != nil {
			return err
		}
		workerSg, err := createWorkerSecurityGroup(ctx, vpcId)
		if err != nil {
			return err
		}
		worker, err := createWorkerInstance(CreateWorkerArgs{
			ctx: ctx,
			sg: workerSg,
			subnetId: privateSubnet,
			workerKey: workerKey,
			controllerIps: pulumi.StringArray{
				controller.PrivateIp,
			},
			ec2Profile: workerProfile,
			publicHost: loadBalancer.DnsName,
			tlsCert: tlsCert,
			encryptedTlsPrivateKey: tlsPrivateKey,
		})
		if err != nil {
			return err
		}

		// Add instances to target groups for forwarding
		err = addToTargetGroup(ctx, "controller", controllerTg, controller, 9200)
		if err != nil {
			return err
		}
		err = addToTargetGroup(ctx, "worker", workerTg, worker, 9202)
		if err != nil {
			return err
		}

		// update DNS record for network load balancer on CF.
		_, err = cloudflare.NewRecord(ctx, "boundary_CNAME", &cloudflare.RecordArgs{
			ZoneId: dnsZoneId,
			Name:   pulumi.String("boundary.fynbos.dev"),
			Value:  loadBalancer.DnsName,
			Type:   pulumi.String("CNAME"),
			Ttl:    pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func parseTemplate (data interface{}, filePath string) string {
	tmp, err := template.ParseFiles(filePath)
	if err != nil {
		return ""
	}
	document := &bytes.Buffer{}
	err = tmp.Execute(document, data)
	if err != nil {
		return ""
	}

	return b64.StdEncoding.EncodeToString(document.Bytes())
}

func createPostgresSecurityGroup(ctx *pulumi.Context, vpcId pulumi.StringOutput, privateSubnetsCidrBlocks pulumi.StringArrayOutput) (*ec2.SecurityGroup, error) {
	return ec2.NewSecurityGroup(ctx, "boundary", &ec2.SecurityGroupArgs{
		Name:        pulumi.String("boundary-db"),
		Description: pulumi.String("SG for boundary postgres database"),
		VpcId:       vpcId,
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Description: pulumi.String("Incoming postgres from ec2"),
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(5432),
				ToPort:      pulumi.Int(5432),
				CidrBlocks:  privateSubnetsCidrBlocks,
			},
		},
	})
}

func createPostgres(ctx *pulumi.Context, sg *ec2.SecurityGroup, intraSubnets pulumi.StringArrayOutput) (*rds.Instance, *random.RandomPassword, error) {

	snGroup, err := rds.NewSubnetGroup(ctx, "boundary", &rds.SubnetGroupArgs{
		Description: pulumi.String("Subnet group for Boundary RDS Postgres DB"),
		Name:        pulumi.String("boundary"),
		SubnetIds:   intraSubnets,
	})
	if err != nil {
		return nil, nil, err
	}

	password, err := random.NewRandomPassword(ctx, "db-password", &random.RandomPasswordArgs{
		Length:  pulumi.Int(32),
		Special: pulumi.Bool(true),
		Number:  pulumi.Bool(true),
		Lower:   pulumi.Bool(true),
		Upper:   pulumi.Bool(true),
		OverrideSpecial: pulumi.String("!#$&*()-_=+[]{}<>?"),
	}, pulumi.AdditionalSecretOutputs([]string{"Result"}))
	if err != nil {
		return nil, nil, err
	}

	instance, err := rds.NewInstance(ctx, "boundary", &rds.InstanceArgs{
		AllocatedStorage:    pulumi.Int(20),
		MaxAllocatedStorage: pulumi.Int(100),
		Engine:              pulumi.String("postgres"),
		EngineVersion:       pulumi.String("13.3"),
		Name:                pulumi.String("boundary"),
		Username:            pulumi.String("boundary"),
		Password:            password.Result,
		InstanceClass:       pulumi.String("db.t3.micro"),
		DbSubnetGroupName:   snGroup.Name,
		SkipFinalSnapshot:   pulumi.Bool(true), // TODO: NB!!!! remove when ready for production
		VpcSecurityGroupIds: pulumi.StringArray{
			sg.ID(),
		},
		MultiAz: pulumi.Bool(false),
	})
	if err != nil {
		return nil, nil, err
	}

	return instance, password, nil
}

func ValueFromStringArrayOutput(output pulumi.AnyOutput, index int) pulumi.StringOutput {
	return output.ApplyT(func(arg interface{}) string {
		stringArray := arg.([]interface{})
		return stringArray[index].(string)
	}).(pulumi.StringOutput)
}

func AnyOutputToStringArrayOutput(output pulumi.AnyOutput) pulumi.StringArrayOutput {
	return output.ApplyT(func(arg interface{}) []string {
		stringArray := arg.([]interface{})
		var outputs []string
		for _, s := range stringArray {
			outputs = append(outputs, s.(string))
		}
		return outputs
	}).(pulumi.StringArrayOutput)
}
