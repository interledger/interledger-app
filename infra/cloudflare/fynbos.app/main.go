package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v4/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		name := "fynbos.app"
		zone, err := cloudflare.LookupZone(ctx, &cloudflare.LookupZoneArgs{
			Name: &name,
		})
		if err != nil {
			return err
		}

		// USE2 Prod Cluster
		_, err = cloudflare.NewRecord(ctx, "use2-prod-cluster", &cloudflare.RecordArgs{
			ZoneId:  pulumi.String(zone.Id),
			Name:    pulumi.String("fynbos.app"),
			Value:   pulumi.String("k8s-emissary-emissary-308d683aeb-2f6d1ec4587303c6.elb.us-east-2.amazonaws.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})

		if err != nil {
			return err
		}

		return nil
	})
}
