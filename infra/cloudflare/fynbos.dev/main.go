package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		zone, err := cloudflare.NewZone(ctx, "fynbos.dev", &cloudflare.ZoneArgs{
			Type: pulumi.String("full"),
			Zone: pulumi.String("fynbos.dev"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Vercel landing page
		_, err = cloudflare.NewRecord(ctx, "vercel_A_record", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev"),
			Value:  pulumi.String("76.76.21.21"),
			Type:   pulumi.String("A"),
			Ttl:    pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "vercel_CNAME", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("www.fynbos.dev"),
			Value:  pulumi.String("cname.vercel-dns.com"),
			Type:   pulumi.String("CNAME"),
			Ttl:    pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// google
		_, err = cloudflare.NewRecord(ctx, "_domainconnect", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("_domainconnect"),
			Value:  pulumi.String("connect.domains.google.com"),
			Type:   pulumi.String("CNAME"),
			Ttl:    pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// email
		_, err = cloudflare.NewRecord(ctx, "dmarc", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("_dmarc.fynbos.dev."),
			Value:  pulumi.String("v=DMARC1; p=reject; rua=mailto:postmaster@fynbos.dev; pct=100; adkim=s; aspf=s"),
			Type:   pulumi.String("TXT"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx1", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("aspmx.l.google.com"),
			Priority: pulumi.Int(1),
			Type:   pulumi.String("MX"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx2", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("alt1.aspmx.l.google.com"),
			Priority: pulumi.Int(5),
			Type:   pulumi.String("MX"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx3", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("alt2.aspmx.l.google.com"),
			Priority: pulumi.Int(5),
			Type:   pulumi.String("MX"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx4", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("alt3.aspmx.l.google.com"),
			Priority: pulumi.Int(10),
			Type:   pulumi.String("MX"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx5", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("alt4.aspmx.l.google.com"),
			Priority: pulumi.Int(10),
			Type:   pulumi.String("MX"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "spf", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("v=spf1 include:_spf.google.com -all"),
			Type:   pulumi.String("SPF"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "spf_txt", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("v=spf1 include:_spf.google.com -all"),
			Type:   pulumi.String("TXT"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "domain_key", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("google._domainkey.fynbos.dev."),
			Value:  pulumi.String("v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnjoOga7xnV9J9GsunNAa97b8pJ2QV6x7IF+7h2rDftSjg9JgpOuu1ntryS1SWssuVzNcNJkDfeFnrBFbDDbMcviM+4UcZ4GqX/eBCxKYOTHWEpfEZFUKFZ5RxCOE7y1yDDP3Cp/5ne4JVM9Bq6EDSQmfYewLTRz5wabvknCq1h1aZQXd3dun4wNvcocyOXI8LUGFzkmEZz22vpQV1TxzXoP6AxYSslsINh+CtjLXHLIGZbkLzW/N9UWOJBb07oXMslFuInyAQuajeFORknO4uPhJt415cajcpHXeVAqOmphT6jl1+hJNJll09pAnXDtQ23YrxLUJn2Gp/Y4415Jl6QIDAQAB"),
			Type:   pulumi.String("TXT"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		// DNSSEC
		_, err = cloudflare.NewZoneDnssec(ctx, "dnssec", &cloudflare.ZoneDnssecArgs{
			ZoneId: zone.ID(),
		})
		if err != nil {
			return err
		}

		return nil
	})
}
