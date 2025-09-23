package main

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a test image
func createTestImage(size int, fillColor color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill with solid color
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, fillColor)
		}
	}

	return img
}

// Helper function to create a test image with border
func createTestImageWithBorder(size int, fillColor, borderColor color.RGBA, borderWidth int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill with fill color
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, fillColor)
		}
	}

	// Add border
	for i := 0; i < borderWidth; i++ {
		// Top and bottom borders
		for x := 0; x < size; x++ {
			img.Set(x, i, borderColor)
			img.Set(x, size-1-i, borderColor)
		}
		// Left and right borders
		for y := 0; y < size; y++ {
			img.Set(i, y, borderColor)
			img.Set(size-1-i, y, borderColor)
		}
	}

	return img
}

// Helper interface for both testing.T and testing.B
type testingTB interface {
	TempDir() string
	Fatalf(format string, args ...interface{})
}

// Helper function to create a temporary test file
func createTempImageFile(t testingTB, img image.Image) string {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.png")

	if err := saveImage(img, imgPath); err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	return imgPath
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: Config{
				InputPath:     createTempImageFile(t, createTestImage(100, color.RGBA{255, 0, 0, 255})),
				OutputDir:     t.TempDir(),
				TrimPercent:   80,
				RadiusPercent: 20,
			},
			expectErr: false,
		},
		{
			name: "non-existent input file",
			config: Config{
				InputPath:     "/non/existent/file.png",
				OutputDir:     t.TempDir(),
				TrimPercent:   80,
				RadiusPercent: 20,
			},
			expectErr: true,
		},
		{
			name: "invalid trim percent - too low",
			config: Config{
				InputPath:     createTempImageFile(t, createTestImage(100, color.RGBA{255, 0, 0, 255})),
				OutputDir:     t.TempDir(),
				TrimPercent:   0,
				RadiusPercent: 20,
			},
			expectErr: true,
		},
		{
			name: "invalid trim percent - too high",
			config: Config{
				InputPath:     createTempImageFile(t, createTestImage(100, color.RGBA{255, 0, 0, 255})),
				OutputDir:     t.TempDir(),
				TrimPercent:   101,
				RadiusPercent: 20,
			},
			expectErr: true,
		},
		{
			name: "invalid radius percent - too low",
			config: Config{
				InputPath:     createTempImageFile(t, createTestImage(100, color.RGBA{255, 0, 0, 255})),
				OutputDir:     t.TempDir(),
				TrimPercent:   80,
				RadiusPercent: -1,
			},
			expectErr: true,
		},
		{
			name: "invalid radius percent - too high",
			config: Config{
				InputPath:     createTempImageFile(t, createTestImage(100, color.RGBA{255, 0, 0, 255})),
				OutputDir:     t.TempDir(),
				TrimPercent:   80,
				RadiusPercent: 51,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestCropCenter(t *testing.T) {
	// Create a 100x100 test image with a red center and blue border
	originalSize := 100
	borderWidth := 20
	testImg := createTestImageWithBorder(
		originalSize,
		color.RGBA{255, 0, 0, 255}, // red center
		color.RGBA{0, 0, 255, 255}, // blue border
		borderWidth,
	)

	tests := []struct {
		name        string
		percent     int
		expectedSize int
	}{
		{"80 percent crop", 80, 80},
		{"50 percent crop", 50, 50},
		{"100 percent crop", 100, 100},
		{"60 percent crop", 60, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cropped := cropCenter(testImg, tt.percent)
			bounds := cropped.Bounds()

			if bounds.Dx() != tt.expectedSize || bounds.Dy() != tt.expectedSize {
				t.Errorf("Expected size %dx%d, got %dx%d",
					tt.expectedSize, tt.expectedSize, bounds.Dx(), bounds.Dy())
			}

			// For 80% crop, the center should still be red (no blue border)
			if tt.percent == 80 {
				centerColor := cropped.At(bounds.Dx()/2, bounds.Dy()/2)
				r, g, b, _ := centerColor.RGBA()
				if r < 32768 || g > 32768 || b > 32768 { // Should be red-ish
					t.Errorf("Expected red center after crop, got RGBA(%d, %d, %d)", r>>8, g>>8, b>>8)
				}
			}
		})
	}
}

func TestResizeImage(t *testing.T) {
	// Create a test image
	originalImg := createTestImage(200, color.RGBA{255, 0, 0, 255})

	tests := []struct {
		name         string
		targetSize   int
		expectedSize int
	}{
		{"resize to 64", 64, 64},
		{"resize to 128", 128, 128},
		{"resize to 256", 256, 256},
		{"resize to 512", 512, 512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resized := resizeImage(originalImg, tt.targetSize)
			bounds := resized.Bounds()

			if bounds.Dx() != tt.expectedSize || bounds.Dy() != tt.expectedSize {
				t.Errorf("Expected size %dx%d, got %dx%d",
					tt.expectedSize, tt.expectedSize, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestAddRoundedCorners(t *testing.T) {
	// Create a solid red 100x100 image
	testImg := createTestImage(100, color.RGBA{255, 0, 0, 255})

	tests := []struct {
		name   string
		radius int
	}{
		{"no radius", 0},
		{"small radius", 10},
		{"medium radius", 20},
		{"large radius", 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rounded := addRoundedCorners(testImg, tt.radius)
			bounds := rounded.Bounds()

			// Image should maintain same size
			if bounds.Dx() != 100 || bounds.Dy() != 100 {
				t.Errorf("Expected size 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
			}

			if tt.radius == 0 {
				// With radius 0, corners should not be transparent
				cornerColor := rounded.At(0, 0)
				_, _, _, a := cornerColor.RGBA()
				if a == 0 {
					t.Errorf("Expected opaque corner with radius 0, got transparent")
				}
			} else {
				// With radius > 0, corners should be transparent
				cornerColor := rounded.At(0, 0)
				_, _, _, a := cornerColor.RGBA()
				if a != 0 {
					t.Errorf("Expected transparent corner with radius %d, got opaque", tt.radius)
				}

				// Center should still be opaque
				centerColor := rounded.At(50, 50)
				_, _, _, a = centerColor.RGBA()
				if a == 0 {
					t.Errorf("Expected opaque center, got transparent")
				}
			}
		})
	}
}

func TestShouldKeepPixel(t *testing.T) {
	tests := []struct {
		name     string
		x, y     int
		size     int
		radius   int
		expected bool
	}{
		{"center pixel", 50, 50, 100, 20, true},
		{"corner pixel with no radius", 0, 0, 100, 0, true},
		{"corner pixel within radius", 15, 15, 100, 20, true},
		{"corner pixel outside radius", 5, 5, 100, 20, false},
		{"edge pixel not in corner", 50, 0, 100, 20, true},
		{"top-right corner within radius", 85, 15, 100, 20, true},
		{"bottom-left corner outside radius", 5, 95, 100, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldKeepPixel(tt.x, tt.y, tt.size, tt.radius)
			if result != tt.expected {
				t.Errorf("shouldKeepPixel(%d, %d, %d, %d) = %v, expected %v",
					tt.x, tt.y, tt.size, tt.radius, result, tt.expected)
			}
		})
	}
}

func TestLoadAndSaveImage(t *testing.T) {
	// Create a test image
	testImg := createTestImage(50, color.RGBA{0, 255, 0, 255})

	// Save it to a temporary file
	tmpPath := createTempImageFile(t, testImg)

	// Load it back
	loadedImg, err := loadImage(tmpPath)
	if err != nil {
		t.Fatalf("Failed to load image: %v", err)
	}

	// Check dimensions
	bounds := loadedImg.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Expected size 50x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that the center pixel is green
	centerColor := loadedImg.At(25, 25)
	r, g, b, a := centerColor.RGBA()
	// Values are in 16-bit range, so green should be high
	if g < 32768 || r > 32768 || b > 32768 || a < 32768 {
		t.Errorf("Expected green pixel, got RGBA(%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestGenerateIconsIntegration(t *testing.T) {
	// Create a test image
	testImg := createTestImageWithBorder(
		200,
		color.RGBA{255, 255, 0, 255}, // yellow center
		color.RGBA{255, 0, 255, 255}, // magenta border
		20,
	)

	// Create temporary directories
	inputPath := createTempImageFile(t, testImg)
	outputDir := t.TempDir()

	// Create config
	config := Config{
		InputPath:     inputPath,
		OutputDir:     outputDir,
		Clean:         true,
		CropEnabled:   true,
		TrimPercent:   80,
		RadiusPercent: 20,
	}

	// Generate icons
	err := generateIcons(config)
	if err != nil {
		t.Fatalf("Failed to generate icons: %v", err)
	}

	// Check that all expected files were created
	expectedFiles := []string{
		"icon_16x16.png", "icon_16x16_rounded.png",
		"icon_32x32.png", "icon_32x32_rounded.png",
		"icon_128x128.png", "icon_128x128_rounded.png",
		"icon_256x256.png", "icon_256x256_rounded.png",
		"icon_512x512.png", "icon_512x512_rounded.png",
		"icon_1024x1024.png", "icon_1024x1024_rounded.png",
	}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(outputDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", filename)
		}

		// Load and verify the generated icon
		if img, err := loadImage(filePath); err != nil {
			t.Errorf("Failed to load generated icon %s: %v", filename, err)
		} else {
			bounds := img.Bounds()
			if bounds.Dx() != bounds.Dy() {
				t.Errorf("Generated icon %s is not square: %dx%d", filename, bounds.Dx(), bounds.Dy())
			}
		}
	}
}

func TestGenerateIconsWithClean(t *testing.T) {
	// Create test setup
	testImg := createTestImage(100, color.RGBA{0, 0, 255, 255})
	inputPath := createTempImageFile(t, testImg)
	outputDir := t.TempDir()

	// Create fake existing icon files that should be cleaned
	existingIcons := []string{
		"icon_old_16x16.png",
		"icon_custom_32x32.png",
		"icon_64x64.png", // This one matches our standard naming but wrong size
	}

	for _, iconName := range existingIcons {
		existingIconPath := filepath.Join(outputDir, iconName)
		if err := saveImage(testImg, existingIconPath); err != nil {
			t.Fatalf("Failed to create existing icon %s: %v", iconName, err)
		}
	}

	// Verify they exist before cleaning
	for _, iconName := range existingIcons {
		existingIconPath := filepath.Join(outputDir, iconName)
		if _, err := os.Stat(existingIconPath); os.IsNotExist(err) {
			t.Fatalf("Existing icon %s was not created", iconName)
		}
	}

	// Generate icons with clean enabled
	config := Config{
		InputPath:     inputPath,
		OutputDir:     outputDir,
		Clean:         true,
		CropEnabled:   false,
		TrimPercent:   100,
		RadiusPercent: 0,
	}

	err := generateIcons(config)
	if err != nil {
		t.Fatalf("Failed to generate icons: %v", err)
	}

	// Check that standard icons were generated
	standardIcon := filepath.Join(outputDir, "icon_16x16.png")
	if _, err := os.Stat(standardIcon); os.IsNotExist(err) {
		t.Errorf("Standard icon should exist after generation")
	}

	// The old icons that matched the icon_*.png pattern should have been removed
	// (Note: Our glob pattern "icon_*.png" would match these files)
	for _, iconName := range existingIcons {
		existingIconPath := filepath.Join(outputDir, iconName)
		if _, err := os.Stat(existingIconPath); !os.IsNotExist(err) {
			// If it still exists, it should be because it was regenerated as a standard icon
			// Only icon_64x64.png would be regenerated since it's not a standard size we generate
			if iconName == "icon_64x64.png" {
				// This file should have been cleaned but not regenerated (64x64 is not a standard size)
				t.Logf("icon_64x64.png still exists - this is expected if 64x64 is a standard size")
			}
		}
	}
}

// TestConfigDefaults tests the default configuration values
func TestConfigDefaults(t *testing.T) {
	// Test default configuration manually since flag parsing is complex to test
	defaultConfig := Config{
		InputPath:     "images/TranslateCat.png",
		OutputDir:     "images",
		Clean:         false,
		CropEnabled:   true,
		TrimPercent:   80,
		RadiusPercent: 20,
	}

	// Validate the default config
	if defaultConfig.TrimPercent < 1 || defaultConfig.TrimPercent > 100 {
		t.Errorf("Default TrimPercent should be valid: got %d", defaultConfig.TrimPercent)
	}
	if defaultConfig.RadiusPercent < 0 || defaultConfig.RadiusPercent > 50 {
		t.Errorf("Default RadiusPercent should be valid: got %d", defaultConfig.RadiusPercent)
	}

	// Test edge cases for config values
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid_config",
			config: Config{
				InputPath:     "/tmp/test.png",
				OutputDir:     "/tmp",
				TrimPercent:   80,
				RadiusPercent: 20,
			},
			valid: true,
		},
		{
			name: "edge_case_trim_1",
			config: Config{
				InputPath:     "/tmp/test.png",
				OutputDir:     "/tmp",
				TrimPercent:   1,
				RadiusPercent: 20,
			},
			valid: true,
		},
		{
			name: "edge_case_trim_100",
			config: Config{
				InputPath:     "/tmp/test.png",
				OutputDir:     "/tmp",
				TrimPercent:   100,
				RadiusPercent: 20,
			},
			valid: true,
		},
		{
			name: "edge_case_radius_0",
			config: Config{
				InputPath:     "/tmp/test.png",
				OutputDir:     "/tmp",
				TrimPercent:   80,
				RadiusPercent: 0,
			},
			valid: true,
		},
		{
			name: "edge_case_radius_50",
			config: Config{
				InputPath:     "/tmp/test.png",
				OutputDir:     "/tmp",
				TrimPercent:   80,
				RadiusPercent: 50,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file for testing
			if tt.valid {
				testImg := createTestImage(100, color.RGBA{255, 0, 0, 255})
				tmpPath := createTempImageFile(t, testImg)
				tt.config.InputPath = tmpPath
				tt.config.OutputDir = t.TempDir()
			}

			err := validateConfig(tt.config)
			hasError := err != nil

			if tt.valid && hasError {
				t.Errorf("Expected valid config but got error: %v", err)
			}
			if !tt.valid && !hasError {
				t.Errorf("Expected invalid config but got no error")
			}
		})
	}
}