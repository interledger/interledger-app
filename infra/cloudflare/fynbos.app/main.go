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
			Proxied: pulumi.Bool(true),
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

		// email
		_, err = cloudflare.NewRecord(ctx, "mx1", &cloudflare.RecordArgs{
			Name:     pulumi.String("fynbos.app"),
			Priority: pulumi.Int(10),
			Ttl:      pulumi.Int(1),
			Type:     pulumi.String("MX"),
			Value:    pulumi.String("alt3.aspmx.l.google.com"),
			ZoneId:   pulumi.String("1e90e3f60d09dc307fffbc638c50fc1b"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx2", &cloudflare.RecordArgs{
			Name:     pulumi.String("fynbos.app"),
			Priority: pulumi.Int(5),
			Ttl:      pulumi.Int(1),
			Type:     pulumi.String("MX"),
			Value:    pulumi.String("alt2.aspmx.l.google.com"),
			ZoneId:   pulumi.String("1e90e3f60d09dc307fffbc638c50fc1b"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx3", &cloudflare.RecordArgs{
			Name:     pulumi.String("fynbos.app"),
			Priority: pulumi.Int(1),
			Ttl:      pulumi.Int(1),
			Type:     pulumi.String("MX"),
			Value:    pulumi.String("aspmx.l.google.com"),
			ZoneId:   pulumi.String(zone.Id),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx4", &cloudflare.RecordArgs{
			Name:     pulumi.String("fynbos.app"),
			Priority: pulumi.Int(10),
			Ttl:      pulumi.Int(1),
			Type:     pulumi.String("MX"),
			Value:    pulumi.String("alt4.aspmx.l.google.com"),
			ZoneId:   pulumi.String(zone.Id),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx5", &cloudflare.RecordArgs{
			Name:     pulumi.String("fynbos.app"),
			Priority: pulumi.Int(5),
			Ttl:      pulumi.Int(1),
			Type:     pulumi.String("MX"),
			Value:    pulumi.String("alt1.aspmx.l.google.com"),
			ZoneId:   pulumi.String(zone.Id),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Current SPF record - uses TXT
		// Sending only from Google servers and Zendesk servers
		_, err = cloudflare.NewRecord(ctx, "spf_txt", &cloudflare.RecordArgs{
			Name:   pulumi.String("fynbos.app"),
			Ttl:    pulumi.Int(1),
			Type:   pulumi.String("TXT"),
			Value:  pulumi.String("v=spf1 include:_spf.google.com include:mail.zendesk.com -all"),
			ZoneId: pulumi.String(zone.Id),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// DKIM
		// Key provided by Google Workspaces - https://admin.google.com/ac/apps/gmail/authenticateemail
		_, err = cloudflare.NewRecord(ctx, "google_dkim", &cloudflare.RecordArgs{
			Name:   pulumi.String("google._domainkey"),
			Ttl:    pulumi.Int(1),
			Type:   pulumi.String("TXT"),
			Value:  pulumi.String("v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0A6qUfSTtf96OY+9/yFHIHya6Bh3u8Gdu0K1oYj99V6vVlzSocbDmf2q3Z0nyZDcEK4/Oyf/QL54wZaIVrwZFrl8UHPryyt/DPo9UJuW7GTpsBg79SKefOrsIVFGJQtixXAehRAYY54dRVJaUAdEVnCpESiPCiAWj+U72E4NY7WUY07c6MDlfsTrVJeKmxGzKsnYNluj2Ax26Ca4LE7V3eAMuMwKxn/bE8IwU8YLtKdQ/9cG+4rXRkzngwYuZ0Da3kRuWOFI2neDGDtXr1njI9L/GUKbn5TKleUryoSmGUb1INh6Z0l0foUBeVQOclFKLsnhN1cwMOAavi0LuF5u9QIDAQAB"),
			ZoneId: pulumi.String(zone.Id),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// DMARC
		// send both aggregate (rua) and forensic (ruf) reports to postmaster@fynbos.app
		// p=reject - reject all mails that fail DKIM or SPF checks
		// adkim=s; aspf=s - strict mode for DKIM and SPF checks
		// pct=100 - check all mails (100%)
		_, err = cloudflare.NewRecord(ctx, "dmarc", &cloudflare.RecordArgs{
			Name:   pulumi.String("_dmarc"),
			Ttl:    pulumi.Int(1),
			Type:   pulumi.String("TXT"),
			Value:  pulumi.String("v=DMARC1; p=reject; rua=mailto:postmaster@fynbos.app; ruf=mailto:postmaster@fynbos.app; pct=100; adkim=s; aspf=s"),
			ZoneId: pulumi.String(zone.Id),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// BIMI
		_, err = cloudflare.NewRecord(ctx, "bimi", &cloudflare.RecordArgs{
			Name:   pulumi.String("default._bimi"),
			Ttl:    pulumi.Int(1),
			Type:   pulumi.String("TXT"),
			Value:  pulumi.String("v=BIMI1; l=https://cdn.fynbos.app/logos/fynbos-icon-bimi.svg; a=;"),
			ZoneId: pulumi.String(zone.Id),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		return nil
	})
}
