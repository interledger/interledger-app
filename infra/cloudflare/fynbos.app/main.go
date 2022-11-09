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

		_, err = cloudflare.NewRecord(ctx, "sendgrid-cname-1", &cloudflare.RecordArgs{
			ZoneId:  pulumi.String(zone.Id),
			Name:    pulumi.String("url9532.fynbos.app"),
			Value:   pulumi.String("sendgrid.net"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(false),
		})

		_, err = cloudflare.NewRecord(ctx, "sendgrid-cname-2", &cloudflare.RecordArgs{
			ZoneId:  pulumi.String(zone.Id),
			Name:    pulumi.String("26945468.fynbos.app"),
			Value:   pulumi.String("sendgrid.net"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(false),
		})

		_, err = cloudflare.NewRecord(ctx, "sendgrid-cname-3", &cloudflare.RecordArgs{
			ZoneId:  pulumi.String(zone.Id),
			Name:    pulumi.String("em5194.fynbos.app"),
			Value:   pulumi.String("u26945468.wl219.sendgrid.net"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(false),
		})

		_, err = cloudflare.NewRecord(ctx, "sendgrid-cname-4", &cloudflare.RecordArgs{
			ZoneId:  pulumi.String(zone.Id),
			Name:    pulumi.String("s1._domainkey.fynbos.app"),
			Value:   pulumi.String("s1.domainkey.u26945468.wl219.sendgrid.net"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(false),
		})

		_, err = cloudflare.NewRecord(ctx, "sendgrid-cname-5", &cloudflare.RecordArgs{
			ZoneId:  pulumi.String(zone.Id),
			Name:    pulumi.String("s2._domainkey.fynbos.app"),
			Value:   pulumi.String("s2.domainkey.u26945468.wl219.sendgrid.net"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(false),
		})

		if err != nil {
			return err
		}

		return nil
	})
}
