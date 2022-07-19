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
		ctx.Export("dnsZoneId", zone.ID())

		// Vercel landing page
		_, err = cloudflare.NewRecord(ctx, "vercel_A_record", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("fynbos.dev"),
			Value:   pulumi.String("76.76.21.21"),
			Type:    pulumi.String("A"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "vercel_CNAME", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("www.fynbos.dev"),
			Value:   pulumi.String("cname.vercel-dns.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		devLBDNS := "a72f11483d4ee4297b7348b4297aae9e-75814212bbc0fc08.elb.eu-west-1.amazonaws.com"
		// Dev Cluster
		_, err = cloudflare.NewRecord(ctx, "dev-cluster", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("dev.fynbos.dev"),
			Value:   pulumi.String(devLBDNS),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "dev-mail", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("mail.fynbos.dev"),
			Value:   pulumi.String(devLBDNS),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "dev-retool", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("retool.fynbos.dev"),
			Value:   pulumi.String(devLBDNS),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "dev-pay", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("pay.fynbos.dev"),
			Value:   pulumi.String(devLBDNS),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "dev-support", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("support.fynbos.dev"),
			Value:   pulumi.String("fynbos.zendesk.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
		})
		if err != nil {
			return err
		}

		// google
		_, err = cloudflare.NewRecord(ctx, "_domainconnect", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("_domainconnect"),
			Value:   pulumi.String("connect.domains.google.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Zendesk

		// Domain Verification Key: c786ca99dc1dfd77
		// https://fynbos.zendesk.com/admin/channels/talk_and_email/email
		_, err = cloudflare.NewRecord(ctx, "zendesk-verification", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("zendeskverification.fynbos.dev."),
			Value:  pulumi.String("c786ca99dc1dfd77"),
			Type:   pulumi.String("TXT"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		// email
		// send both aggregate (rua) and forensic (ruf) reports to postmaster@fynbos.dev
		// p=reject - reject all mails that fail DKIM or SPF checks
		// adkim=s; aspf=s - strict mode for DKIM and SPF checks
		// pct=100 - check all mails (100%)
		_, err = cloudflare.NewRecord(ctx, "dmarc", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("_dmarc.fynbos.dev."),
			Value:  pulumi.String("v=DMARC1; p=reject; rua=mailto:postmaster@fynbos.dev; ruf=mailto:postmaster@fynbos.dev; pct=100; adkim=s; aspf=s"),
			Type:   pulumi.String("TXT"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx1", &cloudflare.RecordArgs{
			ZoneId:   zone.ID().ToStringOutput(),
			Name:     pulumi.String("fynbos.dev."),
			Value:    pulumi.String("aspmx.l.google.com"),
			Priority: pulumi.Int(1),
			Type:     pulumi.String("MX"),
			Ttl:      pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx2", &cloudflare.RecordArgs{
			ZoneId:   zone.ID().ToStringOutput(),
			Name:     pulumi.String("fynbos.dev."),
			Value:    pulumi.String("alt1.aspmx.l.google.com"),
			Priority: pulumi.Int(5),
			Type:     pulumi.String("MX"),
			Ttl:      pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx3", &cloudflare.RecordArgs{
			ZoneId:   zone.ID().ToStringOutput(),
			Name:     pulumi.String("fynbos.dev."),
			Value:    pulumi.String("alt2.aspmx.l.google.com"),
			Priority: pulumi.Int(5),
			Type:     pulumi.String("MX"),
			Ttl:      pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx4", &cloudflare.RecordArgs{
			ZoneId:   zone.ID().ToStringOutput(),
			Name:     pulumi.String("fynbos.dev."),
			Value:    pulumi.String("alt3.aspmx.l.google.com"),
			Priority: pulumi.Int(10),
			Type:     pulumi.String("MX"),
			Ttl:      pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "mx5", &cloudflare.RecordArgs{
			ZoneId:   zone.ID().ToStringOutput(),
			Name:     pulumi.String("fynbos.dev."),
			Value:    pulumi.String("alt4.aspmx.l.google.com"),
			Priority: pulumi.Int(10),
			Type:     pulumi.String("MX"),
			Ttl:      pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		// Current SPF record - uses TXT
		// Sending only from Google servers and Zendesk servers
		_, err = cloudflare.NewRecord(ctx, "spf_txt", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("fynbos.dev."),
			Value:  pulumi.String("v=spf1 include:_spf.google.com include:mail.zendesk.com -all"),
			Type:   pulumi.String("TXT"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		// DKIM

		// Key provided by Google Workspaces - https://admin.google.com/ac/apps/gmail/authenticateemail
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

		// Additional keys hosted by Zendesk - https://support.zendesk.com/hc/en-us/articles/4408822303386
		_, err = cloudflare.NewRecord(ctx, "zendesk-dkim1", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("zendesk1._domainkey"),
			Value:  pulumi.String("zendesk1._domainkey.zendesk.com"),
			Type:   pulumi.String("CNAME"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "zendesk-dkim2", &cloudflare.RecordArgs{
			ZoneId: zone.ID().ToStringOutput(),
			Name:   pulumi.String("zendesk2._domainkey"),
			Value:  pulumi.String("zendesk2._domainkey.zendesk.com"),
			Type:   pulumi.String("CNAME"),
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		// EU1 Dev Cluster
		_, err = cloudflare.NewRecord(ctx, "eu1-dev-cluster", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("eu1.fynbos.dev"),
			Value:   pulumi.String("k8s-emissary-emissary-ffb342d685-679be1c9b02820eb.elb.eu-west-1.amazonaws.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})

		if err != nil {
			return err
		}
		_, err = cloudflare.NewRecord(ctx, "eu1-dev-cluster-wild", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("*.eu1.fynbos.dev"),
			Value:   pulumi.String("k8s-emissary-emissary-ffb342d685-679be1c9b02820eb.elb.eu-west-1.amazonaws.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Shared Cluster
		_, err = cloudflare.NewRecord(ctx, "eu1-shared-cluster", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("mgnt.fynbos.dev"),
			Value:   pulumi.String("k8s-emissary-emissary-a896348535-702945cedff89f60.elb.eu-west-1.amazonaws.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
		})

		if err != nil {
			return err
		}
		_, err = cloudflare.NewRecord(ctx, "eu1-shared-cluster-wild", &cloudflare.RecordArgs{
			ZoneId:  zone.ID().ToStringOutput(),
			Name:    pulumi.String("*.mgnt.fynbos.dev"),
			Value:   pulumi.String("k8s-emissary-emissary-a896348535-702945cedff89f60.elb.eu-west-1.amazonaws.com"),
			Type:    pulumi.String("CNAME"),
			Ttl:     pulumi.Int(1),
			Proxied: pulumi.Bool(true),
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

		// Main advanced cert
		_, err = cloudflare.NewCertificatePack(ctx, "advanced-cert", &cloudflare.CertificatePackArgs{
			CertificateAuthority: pulumi.String("digicert"),
			CloudflareBranding:   pulumi.Bool(false),
			Hosts: pulumi.StringArray{
				pulumi.String("fynbos.dev"),
				pulumi.String("*.fynbos.dev"),
				pulumi.String("*.eu1.fynbos.dev"),
				pulumi.String("*.mgnt.fynbos.dev"),
			},
			Type:             pulumi.String("advanced"),
			ValidationMethod: pulumi.String("txt"),
			ValidityDays:     pulumi.Int(30),
			ZoneId:           zone.ID(),
		})
		if err != nil {
			return err
		}

		boundaryPrivateKey, boundaryCert, err := newBoundaryCertificate(ctx)
		if err != nil {
			return nil
		}

		ctx.Export("boundaryPrivateKey", boundaryPrivateKey.PrivateKeyPem)
		ctx.Export("boundaryCert", boundaryCert.Certificate)

		devClusterPK, devClusterCert, err := newDevClusterCertificate(ctx)
		if err != nil {
			return nil
		}
		ctx.Export("devClusterPrivateKey", devClusterPK.PrivateKeyPem)
		ctx.Export("devClusterCert", devClusterCert.Certificate)

		// Create CF Applications
		err = CreateDevClusterAccess(ctx, zone.ID())
		if err != nil {
			return err
		}

		err = CreateEu1DevClusterAccess(ctx, zone.ID())
		if err != nil {
			return err
		}

		err = CreateEu1SharedClusterAccess(ctx, zone.ID())
		if err != nil {
			return err
		}

		err = CreateDevMailAccess(ctx, zone.ID())
		if err != nil {
			return err
		}

		err = CreateDevRetoolAccess(ctx, zone.ID())
		if err != nil {
			return err
		}

		if err = CreateDevPayAccess(ctx, zone.ID()); err != nil {
			return err
		}

		return nil
	})
}
