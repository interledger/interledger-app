package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		zone, err := cloudflare.NewZone(ctx, "ilp.link", &cloudflare.ZoneArgs{
			Type: pulumi.String("full"),
			Zone: pulumi.String("ilp.link"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}
		ctx.Export("dnsZoneId", zone.ID())

		// EU1 Dev Cluster
		_, err = cloudflare.NewRecord(ctx, "eu1-dev-cluster", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("eu1.ilp.link"),
			Value:   pulumi.String("k8s-emissary-emissary-7be6ccefa6-ef78c501a78a76be.elb.eu-west-1.amazonaws.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})

		// Prod Cluster
		_, err = cloudflare.NewRecord(ctx, "use2-prod-cluster", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("ilp.link"),
			Value:   pulumi.String("k8s-emissary-emissary-308d683aeb-2f6d1ec4587303c6.elb.us-east-2.amazonaws.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})

		// DNSSEC
		_, err = cloudflare.NewZoneDnssec(ctx, "dnssec", &cloudflare.ZoneDnssecArgs{
			ZoneId: zone.ID(),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		return nil
	})
}
