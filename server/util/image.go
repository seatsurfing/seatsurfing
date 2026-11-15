package util

import (
	"bytes"
	"errors"
	"image"
	"slices"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type ImageLimits struct {
	MaxBytes     int
	MaxDimension int
	// MaxPixels limits width * height; 0 means no limit.
	MaxPixels int
	Formats   []string
}

type ImageInfo struct {
	Width  uint
	Height uint
	Format string
}

var (
	ErrImageTooLarge = errors.New("image too large")
	ErrImageInvalid  = errors.New("invalid or unsupported image")
)

// DecodeImageConfig checks data against limits by reading only the image header.
func DecodeImageConfig(data []byte, limits ImageLimits) (*ImageInfo, error) {
	if len(data) > limits.MaxBytes {
		return nil, ErrImageTooLarge
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || !slices.Contains(limits.Formats, format) {
		return nil, ErrImageInvalid
	}
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > limits.MaxDimension || cfg.Height > limits.MaxDimension {
		return nil, ErrImageInvalid
	}
	if limits.MaxPixels > 0 && cfg.Width*cfg.Height > limits.MaxPixels {
		return nil, ErrImageInvalid
	}
	return &ImageInfo{Width: uint(cfg.Width), Height: uint(cfg.Height), Format: format}, nil
}

// ValidateImage is DecodeImageConfig plus a full decode to reject corrupt image data.
func ValidateImage(data []byte, limits ImageLimits) (*ImageInfo, error) {
	info, err := DecodeImageConfig(data, limits)
	if err != nil {
		return nil, err
	}
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		return nil, ErrImageInvalid
	}
	return info, nil
}
