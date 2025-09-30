package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	InputPath        string
	OutputDir        string
	Clean            bool
	CropEnabled      bool
	TrimPercent      int
	RadiusPercent    int
	PaddingPercent   int
	PaddingIOSMode   bool
}

type IconSize struct {
	Name string
	Size int
}

var iconSizes = []IconSize{
	{"icon_16x16.png", 16},
	{"icon_16x16@2x.png", 32},
	{"icon_32x32.png", 32},
	{"icon_32x32@2x.png", 64},
	{"icon_128x128.png", 128},
	{"icon_128x128@2x.png", 256},
	{"icon_256x256.png", 256},
	{"icon_256x256@2x.png", 512},
	{"icon_512x512.png", 512},
	{"icon_512x512@2x.png", 1024},
	{"icon_1024x1024.png", 1024},
}

func main() {
	config := parseFlags()

	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := generateIcons(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating icons: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Done. Generated icon_* PNGs alongside source.")
}

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.InputPath, "input", "images/TranslateCat.png", "Input image path")
	flag.StringVar(&config.OutputDir, "output", "", "Output directory (defaults to input image directory)")
	flag.BoolVar(&config.Clean, "clean", false, "Remove existing icon_*.png files before generating")
	flag.BoolVar(&config.CropEnabled, "crop", true, "Enable center cropping")
	flag.IntVar(&config.TrimPercent, "trim-percent", 80, "Percentage of image to keep when cropping (1-100)")
	flag.IntVar(&config.RadiusPercent, "radius-percent", 20, "Corner radius as percentage of size for rounded variants")
	flag.IntVar(&config.PaddingPercent, "padding-percent", 0, "Padding as percentage of image size (0-50)")
	flag.BoolVar(&config.PaddingIOSMode, "padding-ios-mode", false, "iOS-compliant padding: exclude base icon_1024x1024.png from padding")

	// Handle --no-crop flag
	noCrop := flag.Bool("no-crop", false, "Disable center cropping")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [input-image] [output-dir]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate app icon PNGs from a single source image.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s app-icon.png\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --clean --trim-percent=75 source.png icons/\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --no-crop logo.png\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --padding-percent=15 --padding-ios-mode source.png  # iOS: base 1024x1024 stays full size\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --padding-percent=10 source.png  # All sizes get padding\n", os.Args[0])
	}

	flag.Parse()

	// Handle positional arguments
	args := flag.Args()
	if len(args) > 0 {
		config.InputPath = args[0]
	}
	if len(args) > 1 {
		config.OutputDir = args[1]
	}

	// Handle special flags
	if *noCrop {
		config.CropEnabled = false
	}

	// Set default output directory
	if config.OutputDir == "" {
		config.OutputDir = filepath.Dir(config.InputPath)
	}

	return config
}

func validateConfig(config Config) error {
	if _, err := os.Stat(config.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input image not found: %s", config.InputPath)
	}

	if config.TrimPercent < 1 || config.TrimPercent > 100 {
		return fmt.Errorf("trim percent must be between 1 and 100 (got %d)", config.TrimPercent)
	}

	if config.RadiusPercent < 0 || config.RadiusPercent > 50 {
		return fmt.Errorf("radius percent must be between 0 and 50 (got %d)", config.RadiusPercent)
	}

	if config.PaddingPercent < 0 || config.PaddingPercent > 50 {
		return fmt.Errorf("padding percent must be between 0 and 50 (got %d)", config.PaddingPercent)
	}

	return nil
}

func generateIcons(config Config) error {
	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Clean existing icons if requested
	if config.Clean {
		fmt.Printf("Cleaning existing icon_*.png in: %s\n", config.OutputDir)
		pattern := filepath.Join(config.OutputDir, "icon_*.png")
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			os.Remove(match)
		}
	}

	// Load source image
	sourceImg, err := loadImage(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to load source image: %w", err)
	}

	// Apply cropping if enabled
	if config.CropEnabled {
		fmt.Printf("Pre-trimming input to centered %d%% area, then generating PNGs in: %s\n",
			config.TrimPercent, config.OutputDir)
		sourceImg = cropCenter(sourceImg, config.TrimPercent)
	} else {
		fmt.Printf("Cropping disabled; generating PNGs from full image in: %s\n", config.OutputDir)
	}

	// Generate all icon sizes
	for _, iconSize := range iconSizes {
		fmt.Printf(" - %s (%dx%d)\n", iconSize.Name, iconSize.Size, iconSize.Size)

		// Resize image
		resized := resizeImage(sourceImg, iconSize.Size)

		// Apply padding if specified
		processed := resized
		shouldApplyPadding := config.PaddingPercent > 0
		if config.PaddingIOSMode && iconSize.Name == "icon_1024x1024.png" {
			shouldApplyPadding = false // iOS mode: exclude base 1024x1024 icon only
		}
		if shouldApplyPadding {
			processed = addPadding(resized, config.PaddingPercent, iconSize.Size)
		}

		// Save regular version
		outputPath := filepath.Join(config.OutputDir, iconSize.Name)
		if err := saveImage(processed, outputPath); err != nil {
			return fmt.Errorf("failed to save %s: %w", iconSize.Name, err)
		}

		// Generate rounded version
		if config.RadiusPercent > 0 {
			roundedName := strings.TrimSuffix(iconSize.Name, ".png") + "_rounded.png"
			radius := iconSize.Size * config.RadiusPercent / 100
			fmt.Printf(" - %s (%dx%d, r=%d)\n", roundedName, iconSize.Size, iconSize.Size, radius)

			rounded := addRoundedCorners(resized, radius)

			// Apply padding to rounded version if specified
			processedRounded := rounded
			shouldApplyPaddingRounded := config.PaddingPercent > 0
			if config.PaddingIOSMode && iconSize.Name == "icon_1024x1024.png" {
				shouldApplyPaddingRounded = false // iOS mode: exclude base 1024x1024 icon only
			}
			if shouldApplyPaddingRounded {
				processedRounded = addPadding(rounded, config.PaddingPercent, iconSize.Size)
			}

			roundedPath := filepath.Join(config.OutputDir, roundedName)
			if err := saveImage(processedRounded, roundedPath); err != nil {
				return fmt.Errorf("failed to save %s: %w", roundedName, err)
			}
		}
	}

	return nil
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

func saveImage(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func cropCenter(img image.Image, percent int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate crop dimensions
	cropWidth := width * percent / 100
	cropHeight := height * percent / 100

	// Calculate offset to center the crop
	offsetX := (width - cropWidth) / 2
	offsetY := (height - cropHeight) / 2

	// Create cropped rectangle
	cropRect := image.Rect(
		bounds.Min.X+offsetX,
		bounds.Min.Y+offsetY,
		bounds.Min.X+offsetX+cropWidth,
		bounds.Min.Y+offsetY+cropHeight,
	)

	// Create new image with cropped content
	cropped := image.NewRGBA(image.Rect(0, 0, cropWidth, cropHeight))
	draw.Draw(cropped, cropped.Bounds(), img, cropRect.Min, draw.Src)

	return cropped
}

func resizeImage(img image.Image, size int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scaling to fit within square while maintaining aspect ratio
	scale := float64(size) / math.Max(float64(width), float64(height))
	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Create new image
	resized := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill with transparent background
	transparent := &image.Uniform{color.RGBA{0, 0, 0, 0}}
	draw.Draw(resized, resized.Bounds(), transparent, image.Point{}, draw.Src)

	// Calculate centering offset
	offsetX := (size - newWidth) / 2
	offsetY := (size - newHeight) / 2

	// Bilinear interpolation scaling for smoother results
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Calculate source coordinates with sub-pixel precision
			srcXf := float64(x) / scale
			srcYf := float64(y) / scale

			// Get integer and fractional parts
			srcX := int(srcXf)
			srcY := int(srcYf)
			fracX := srcXf - float64(srcX)
			fracY := srcYf - float64(srcY)

			// Adjust for bounds offset
			srcX += bounds.Min.X
			srcY += bounds.Min.Y

			// Ensure we don't go out of bounds
			if srcX >= bounds.Max.X-1 {
				srcX = bounds.Max.X - 2
				fracX = 1.0
			}
			if srcY >= bounds.Max.Y-1 {
				srcY = bounds.Max.Y - 2
				fracY = 1.0
			}

			if srcX >= bounds.Min.X && srcY >= bounds.Min.Y {
				// Get the four surrounding pixels
				c00 := img.At(srcX, srcY)
				c10 := img.At(srcX+1, srcY)
				c01 := img.At(srcX, srcY+1)
				c11 := img.At(srcX+1, srcY+1)

				// Convert to RGBA for interpolation
				r00, g00, b00, a00 := c00.RGBA()
				r10, g10, b10, a10 := c10.RGBA()
				r01, g01, b01, a01 := c01.RGBA()
				r11, g11, b11, a11 := c11.RGBA()

				// Bilinear interpolation
				r := bilinearInterpolate(float64(r00), float64(r10), float64(r01), float64(r11), fracX, fracY)
				g := bilinearInterpolate(float64(g00), float64(g10), float64(g01), float64(g11), fracX, fracY)
				b := bilinearInterpolate(float64(b00), float64(b10), float64(b01), float64(b11), fracX, fracY)
				a := bilinearInterpolate(float64(a00), float64(a10), float64(a01), float64(a11), fracX, fracY)

				// Convert back to 8-bit and set pixel
				interpolated := color.RGBA64{
					R: uint16(r),
					G: uint16(g),
					B: uint16(b),
					A: uint16(a),
				}
				resized.Set(offsetX+x, offsetY+y, interpolated)
			}
		}
	}

	return resized
}

func addRoundedCorners(img image.Image, radius int) image.Image {
	bounds := img.Bounds()
	size := bounds.Dx() // Assuming square image

	// Create new RGBA image
	rounded := image.NewRGBA(bounds)

	// Create mask for rounded corners
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if shouldKeepPixel(x, y, size, radius) {
				rounded.Set(x, y, img.At(x, y))
			} else {
				// Transparent pixel
				rounded.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return rounded
}

func bilinearInterpolate(c00, c10, c01, c11, fracX, fracY float64) float64 {
	// Interpolate along X axis
	top := c00*(1-fracX) + c10*fracX
	bottom := c01*(1-fracX) + c11*fracX

	// Interpolate along Y axis
	return top*(1-fracY) + bottom*fracY
}

func addPadding(img image.Image, paddingPercent int, targetSize int) image.Image {
	if paddingPercent <= 0 {
		return img
	}

	bounds := img.Bounds()
	currentSize := bounds.Dx() // Assuming square image

	// Calculate padding size
	paddingSize := currentSize * paddingPercent / 100
	paddedSize := currentSize + (paddingSize * 2)

	// Create new image with padding
	padded := image.NewRGBA(image.Rect(0, 0, paddedSize, paddedSize))

	// Fill with transparent background
	transparent := &image.Uniform{color.RGBA{0, 0, 0, 0}}
	draw.Draw(padded, padded.Bounds(), transparent, image.Point{}, draw.Src)

	// Center the original image in the padded canvas
	offsetX := paddingSize
	offsetY := paddingSize
	dstRect := image.Rect(offsetX, offsetY, offsetX+currentSize, offsetY+currentSize)
	draw.Draw(padded, dstRect, img, bounds.Min, draw.Src)

	// Resize the padded image back to target size
	resizedPadded := resizeImage(padded, targetSize)

	return resizedPadded
}

func shouldKeepPixel(x, y, size, radius int) bool {
	// If radius is 0, keep all pixels
	if radius == 0 {
		return true
	}

	// Check if pixel is in corner regions
	inTopLeft := x < radius && y < radius
	inTopRight := x >= size-radius && y < radius
	inBottomLeft := x < radius && y >= size-radius
	inBottomRight := x >= size-radius && y >= size-radius

	if !inTopLeft && !inTopRight && !inBottomLeft && !inBottomRight {
		// Not in corner, keep pixel
		return true
	}

	// Calculate distance from corner center
	var centerX, centerY int

	if inTopLeft {
		centerX, centerY = radius, radius
	} else if inTopRight {
		centerX, centerY = size-radius, radius
	} else if inBottomLeft {
		centerX, centerY = radius, size-radius
	} else { // inBottomRight
		centerX, centerY = size-radius, size-radius
	}

	// Calculate distance
	dx := float64(x - centerX)
	dy := float64(y - centerY)
	distance := math.Sqrt(dx*dx + dy*dy)

	// Keep pixel if within radius
	return distance <= float64(radius)
}
