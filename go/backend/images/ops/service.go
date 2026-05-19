package ops

import (
	"bytes"
	"context"
	"fmt"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"gitlab.com/fynbos/backend/images"
)

func GenerateTwitterImage(_ context.Context, a images.Assets, _ Backends, walletUrl, identifier string) ([]byte, error) {
	var W = 620
	var H = 400
	dc := gg.NewContext(W, H)

	dc.DrawImage(a.Twitter, 0, 0)
	dc.SetHexColor("#0F172A")
	dc.SetFontFace(truetype.NewFace(a.InterMedium, &truetype.Options{
		Size: 40,
	}))
	dc.DrawStringAnchored(fmt.Sprintf("@%s", identifier), 48, 171, 0, 0.5)

	dc.SetFontFace(truetype.NewFace(a.InterRegular, &truetype.Options{
		Size: 24,
	}))
	dc.SetHexColor("#334155")
	dc.DrawStringAnchored(walletUrl, 48, 224, 0, 0.5)

	// encode png to buffer
	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateTwitterOGImage(_ context.Context, a images.Assets, _ Backends, walletUrl, identifier string) ([]byte, error) {
	var W = 1600
	var H = 800
	dc := gg.NewContext(W, H)

	dc.DrawImage(a.TwitterOG, 0, 0)
	dc.SetHexColor("#0F172A")
	dc.SetFontFace(truetype.NewFace(a.InterMedium, &truetype.Options{
		Size: 62.85,
	}))
	dc.DrawStringAnchored(fmt.Sprintf("@%s", identifier), 388.34, 352.3, 0, 0.5)
	dc.SetFontFace(truetype.NewFace(a.InterRegular, &truetype.Options{
		Size: 37.71,
	}))
	dc.SetHexColor("#334155")
	dc.DrawStringAnchored(walletUrl, 388.34, 437.7, 0, 0.5)

	// encode png to buffer
	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateDomainImage(_ context.Context, a images.Assets, _ Backends, walletUrl, identifier string) ([]byte, error) {
	var W = 620
	var H = 400
	dc := gg.NewContext(W, H)

	dc.DrawImage(a.Domain, 0, 0)
	dc.SetHexColor("#0F172A")
	dc.SetFontFace(truetype.NewFace(a.InterMedium, &truetype.Options{
		Size: 40,
	}))
	dc.DrawStringAnchored(identifier, 48, 171, 0, 0.5)

	dc.SetFontFace(truetype.NewFace(a.InterRegular, &truetype.Options{
		Size: 24,
	}))
	dc.SetHexColor("#334155")
	dc.DrawStringAnchored(walletUrl, 48, 224, 0, 0.5)

	// encode png to buffer
	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateDomainOGImage(_ context.Context, a images.Assets, _ Backends, walletUrl, identifier string) ([]byte, error) {
	var W = 1600
	var H = 800
	dc := gg.NewContext(W, H)

	dc.DrawImage(a.DomainOG, 0, 0)
	dc.SetHexColor("#0F172A")
	dc.SetFontFace(truetype.NewFace(a.InterMedium, &truetype.Options{
		Size: 62.85,
	}))
	dc.DrawStringAnchored(identifier, 388.34, 352.3, 0, 0.5)
	dc.SetFontFace(truetype.NewFace(a.InterRegular, &truetype.Options{
		Size: 37.71,
	}))
	dc.SetHexColor("#334155")
	dc.DrawStringAnchored(walletUrl, 388.34, 437.7, 0, 0.5)

	// encode png to buffer
	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateSlackImage(_ context.Context, a images.Assets, _ Backends, walletUrl, identifier string) ([]byte, error) {
	var W = 620
	var H = 400
	dc := gg.NewContext(W, H)

	dc.DrawImage(a.Slack, 0, 0)
	dc.SetHexColor("#0F172A")
	dc.SetFontFace(truetype.NewFace(a.InterMedium, &truetype.Options{
		Size: 40,
	}))
	dc.DrawStringAnchored(identifier, 48, 171, 0, 0.5)

	dc.SetFontFace(truetype.NewFace(a.InterRegular, &truetype.Options{
		Size: 24,
	}))
	dc.SetHexColor("#334155")
	dc.DrawStringAnchored(walletUrl, 48, 224, 0, 0.5)

	// encode png to buffer
	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateSlackOGImage(_ context.Context, a images.Assets, _ Backends, walletUrl, identifier string) ([]byte, error) {
	var W = 1600
	var H = 800
	dc := gg.NewContext(W, H)

	dc.DrawImage(a.SlackOG, 0, 0)
	dc.SetHexColor("#0F172A")
	dc.SetFontFace(truetype.NewFace(a.InterMedium, &truetype.Options{
		Size: 62.85,
	}))
	dc.DrawStringAnchored(identifier, 388.34, 352.3, 0, 0.5)
	dc.SetFontFace(truetype.NewFace(a.InterRegular, &truetype.Options{
		Size: 37.71,
	}))
	dc.SetHexColor("#334155")
	dc.DrawStringAnchored(walletUrl, 388.34, 437.7, 0, 0.5)

	// encode png to buffer
	var buf bytes.Buffer
	err := dc.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
