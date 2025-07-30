package publisher

import (
	"github.com/mzki/erago/attribute"
	"github.com/mzki/erago/view/exp/text/pubdata"
)

// Type alias so that user do not need to import pubdata explicityly
type ImageFetchType = pubdata.ImageFetchType

// Type alias so that user do not need to import pubdata explicityly
var (
	ImageFetchNone       = pubdata.ImageFetchType_IMAGE_FETCH_TYPE_NONE
	ImageFetchRawRGBA    = pubdata.ImageFetchType_IMAGE_FETCH_TYPE_RAW_RGBA
	ImageFetchEncodedPNG = pubdata.ImageFetchType_IMAGE_FETCH_TYPE_ENCODED_PNG
)

var attrAlignmentToPdAlignmentMap = map[attribute.Alignment]pubdata.Alignment{
	attribute.AlignmentLeft:   pubdata.Alignment_ALIGNMENT_LEFT,
	attribute.AlignmentCenter: pubdata.Alignment_ALIGNMENT_CENTER,
	attribute.AlignmentRight:  pubdata.Alignment_ALIGNMENT_RIGHT,
}

func PdAlignment(align attribute.Alignment) pubdata.Alignment {
	pdalign, ok := attrAlignmentToPdAlignmentMap[align]
	if ok {
		return pdalign
	} else {
		return pubdata.Alignment_ALIGNMENT_UNKNOWN
	}
}
