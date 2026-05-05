package images

import (
	"image"

	"github.com/golang/freetype/truetype"
)

type Assets struct {
	Twitter      image.Image
	TwitterOG    image.Image
	Domain       image.Image
	DomainOG     image.Image
	Slack        image.Image
	SlackOG      image.Image
	InterMedium  *truetype.Font
	InterRegular *truetype.Font
}
