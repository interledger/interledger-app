package client

import (
	"bytes"
	"context"
	"errors"
	"image"
	"io"
	"net/http"

	"github.com/golang/freetype/truetype"
	"github.com/interledger/interledger-app/go/backend/images"
	"github.com/interledger/interledger-app/go/backend/images/ops"
	"github.com/interledger/interledger-app/go/log"
)

var _ images.Client = client{}
var errAssetsUnavailable = errors.New("images: assets unavailable")

type client struct {
	b  ops.Backends
	a  images.Assets
	ok bool
}

func New(b ops.Backends) images.Client {

	// Load the files
	assets, err := loadAssets()
	if err != nil {
		log.Error("images: failed to load assets; identity image generation disabled")
		return &client{b: b}
	}

	return &client{
		a:  *assets,
		b:  b,
		ok: true,
	}
}

func (c client) GenerateTwitterIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error) {
	if !c.ok {
		return nil, errAssetsUnavailable
	}
	return ops.GenerateTwitterImage(ctx, c.a, c.b, walletUrl, identifier)
}

func (c client) GenerateTwitterIdentityOG(ctx context.Context, walletUrl, identifier string) ([]byte, error) {
	if !c.ok {
		return nil, errAssetsUnavailable
	}
	return ops.GenerateTwitterOGImage(ctx, c.a, c.b, walletUrl, identifier)
}

func (c client) GenerateDomainIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error) {
	if !c.ok {
		return nil, errAssetsUnavailable
	}
	return ops.GenerateDomainImage(ctx, c.a, c.b, walletUrl, identifier)
}

func (c client) GenerateDomainIdentityOG(ctx context.Context, walletUrl, identifier string) ([]byte, error) {
	if !c.ok {
		return nil, errAssetsUnavailable
	}
	return ops.GenerateDomainOGImage(ctx, c.a, c.b, walletUrl, identifier)
}

func (c client) GenerateSlackIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error) {
	if !c.ok {
		return nil, errAssetsUnavailable
	}
	return ops.GenerateSlackImage(ctx, c.a, c.b, walletUrl, identifier)
}

func (c client) GenerateSlackIdentityOG(ctx context.Context, walletUrl, identifier string) ([]byte, error) {
	if !c.ok {
		return nil, errAssetsUnavailable
	}
	return ops.GenerateSlackOGImage(ctx, c.a, c.b, walletUrl, identifier)
}

func loadAssets() (*images.Assets, error) {
	twitterImg, err := loadImageFromURL("https://mockcdn.interledger.test/identities/twitter/template.png")
	if err != nil {
		return nil, err
	}
	twitterImgOG, err := loadImageFromURL("https://mockcdn.interledger.test/identities/twitter/og-template.png")
	if err != nil {
		return nil, err
	}
	domainImg, err := loadImageFromURL("https://mockcdn.interledger.test/identities/domain/v3/template.png")
	if err != nil {
		return nil, err
	}
	domainImgOG, err := loadImageFromURL("https://mockcdn.interledger.test/identities/domain/v2/og-template.png")
	if err != nil {
		return nil, err
	}

	slackImg, err := loadImageFromURL("https://mockcdn.interledger.test/identities/slack/template.png")
	if err != nil {
		return nil, err
	}
	slackImgOG, err := loadImageFromURL("https://mockcdn.interledger.test/identities/slack/og-template.png")
	if err != nil {
		return nil, err
	}

	fontMedium, err := loadFontFromURL("https://mockcdn.interledger.test/fonts/inter/v12/Medium-Desktop.ttf")
	if err != nil {
		return nil, err
	}

	fontRegular, err := loadFontFromURL("https://mockcdn.interledger.test/fonts/inter/v12/Regular-Desktop.ttf")
	if err != nil {
		return nil, err
	}

	return &images.Assets{
		Twitter:      twitterImg,
		TwitterOG:    twitterImgOG,
		Domain:       domainImg,
		DomainOG:     domainImgOG,
		InterMedium:  fontMedium,
		InterRegular: fontRegular,
		Slack:        slackImg,
		SlackOG:      slackImgOG,
	}, nil
}

func loadImageFromURL(url string) (img image.Image, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	pix, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	img, _, err = image.Decode(bytes.NewReader(pix))
	return
}

func loadFontFromURL(url string) (*truetype.Font, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fontBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return truetype.Parse(fontBytes)
}
