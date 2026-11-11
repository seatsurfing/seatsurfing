package test

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

var testImageLimits = ImageLimits{
	MaxBytes:     1024 * 1024,
	MaxDimension: 100,
	MaxPixels:    5000,
	Formats:      []string{"png", "jpeg"},
}

func TestDecodeImageConfig(t *testing.T) {
	info, err := DecodeImageConfig(CreateTestPNG(40, 30), testImageLimits)
	CheckTestIsNil(t, err)
	CheckTestString(t, "png", info.Format)
	CheckTestUint(t, 40, info.Width)
	CheckTestUint(t, 30, info.Height)

	var buf bytes.Buffer
	CheckTestIsNil(t, jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 20, 10)), nil))
	info, err = DecodeImageConfig(buf.Bytes(), testImageLimits)
	CheckTestIsNil(t, err)
	CheckTestString(t, "jpeg", info.Format)
}

func TestDecodeImageConfigInvalid(t *testing.T) {
	_, err := DecodeImageConfig(make([]byte, testImageLimits.MaxBytes+1), testImageLimits)
	CheckTestBool(t, true, err == ErrImageTooLarge)

	_, err = DecodeImageConfig([]byte("not an image"), testImageLimits)
	CheckTestBool(t, true, err == ErrImageInvalid)

	var gifData bytes.Buffer
	CheckTestIsNil(t, gif.Encode(&gifData, image.NewPaletted(image.Rect(0, 0, 10, 10), color.Palette{color.Black}), nil))
	_, err = DecodeImageConfig(gifData.Bytes(), testImageLimits)
	CheckTestBool(t, true, err == ErrImageInvalid)

	_, err = DecodeImageConfig(CreateTestPNG(testImageLimits.MaxDimension+1, 1), testImageLimits)
	CheckTestBool(t, true, err == ErrImageInvalid)

	_, err = DecodeImageConfig(CreateTestPNG(100, 51), testImageLimits)
	CheckTestBool(t, true, err == ErrImageInvalid)
}

func TestDecodeImageConfigNoPixelLimit(t *testing.T) {
	limits := testImageLimits
	limits.MaxPixels = 0
	_, err := DecodeImageConfig(CreateTestPNG(100, 100), limits)
	CheckTestIsNil(t, err)
}

func TestValidateImage(t *testing.T) {
	data := CreateTestPNG(40, 30)
	info, err := ValidateImage(data, testImageLimits)
	CheckTestIsNil(t, err)
	CheckTestUint(t, 40, info.Width)

	// The header is intact, so only the full decode notices the truncation.
	truncated := data[:len(data)-20]
	_, err = DecodeImageConfig(truncated, testImageLimits)
	CheckTestIsNil(t, err)
	_, err = ValidateImage(truncated, testImageLimits)
	CheckTestBool(t, true, err == ErrImageInvalid)
}
