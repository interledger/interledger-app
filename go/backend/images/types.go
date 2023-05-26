package images

import (
	"github.com/golang/freetype/truetype"
	"image"
)

type Assets struct {
	Twitter      image.Image
	TwitterOG    image.Image
	InterMedium  *truetype.Font
	InterRegular *truetype.Font
}
