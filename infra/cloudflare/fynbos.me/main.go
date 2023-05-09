package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		zone, err := cloudflare.NewZone(ctx, "fynbos.me", &cloudflare.ZoneArgs{
			Type: pulumi.String("full"),
			Zone: pulumi.String("fynbos.me"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}
		ctx.Export("dnsZoneId", zone.ID())

		// EU1 Dev Cluster
		_, err = cloudflare.NewRecord(ctx, "eu1-dev-cluster", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("eu1.fynbos.me"),
			Value:   pulumi.String("k8s-emissary-emissary-7be6ccefa6-ef78c501a78a76be.elb.eu-west-1.amazonaws.com"),
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
