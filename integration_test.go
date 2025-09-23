package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

// TestFullWorkflow tests the complete workflow from command line to file generation
func TestFullWorkflow(t *testing.T) {
	// Create a realistic test scenario
	testImg := createTestImageWithBorder(
		400,
		color.RGBA{0, 150, 255, 255}, // blue center
		color.RGBA{255, 255, 255, 255}, // white border
		50,
	)

	// Save test image
	inputPath := createTempImageFile(t, testImg)
	outputDir := t.TempDir()

	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "standard_workflow",
			config: Config{
				InputPath:     inputPath,
				OutputDir:     outputDir,
				Clean:         true,
				CropEnabled:   true,
				TrimPercent:   75,
				RadiusPercent: 25,
			},
		},
		{
			name: "no_crop_workflow",
			config: Config{
				InputPath:     inputPath,
				OutputDir:     filepath.Join(outputDir, "nocrop"),
				Clean:         false,
				CropEnabled:   false,
				TrimPercent:   100,
				RadiusPercent: 15,
			},
		},
		{
			name: "no_rounded_corners",
			config: Config{
				InputPath:     inputPath,
				OutputDir:     filepath.Join(outputDir, "norounded"),
				Clean:         false,
				CropEnabled:   true,
				TrimPercent:   90,
				RadiusPercent: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create output directory
			os.MkdirAll(tt.config.OutputDir, 0755)

			// Run the workflow
			err := generateIcons(tt.config)
			if err != nil {
				t.Fatalf("Workflow failed: %v", err)
			}

			// Verify all expected files exist and are valid
			for _, iconSize := range iconSizes {
				// Check regular icon
				regularPath := filepath.Join(tt.config.OutputDir, iconSize.Name)
				if _, err := os.Stat(regularPath); os.IsNotExist(err) {
					t.Errorf("Regular icon %s not found", iconSize.Name)
					continue
				}

				// Load and verify the icon
				img, err := loadImage(regularPath)
				if err != nil {
					t.Errorf("Failed to load %s: %v", iconSize.Name, err)
					continue
				}

				bounds := img.Bounds()
				if bounds.Dx() != iconSize.Size || bounds.Dy() != iconSize.Size {
					t.Errorf("Icon %s has wrong size: expected %dx%d, got %dx%d",
						iconSize.Name, iconSize.Size, iconSize.Size, bounds.Dx(), bounds.Dy())
				}

				// Check rounded icon if radius > 0
				if tt.config.RadiusPercent > 0 {
					roundedName := iconSize.Name[:len(iconSize.Name)-4] + "_rounded.png"
					roundedPath := filepath.Join(tt.config.OutputDir, roundedName)

					if _, err := os.Stat(roundedPath); os.IsNotExist(err) {
						t.Errorf("Rounded icon %s not found", roundedName)
						continue
					}

					roundedImg, err := loadImage(roundedPath)
					if err != nil {
						t.Errorf("Failed to load rounded %s: %v", roundedName, err)
						continue
					}

					// Verify corners are transparent for rounded icons
					if iconSize.Size >= 32 { // Only check for larger icons
						cornerColor := roundedImg.At(0, 0)
						_, _, _, a := cornerColor.RGBA()
						if a != 0 {
							t.Errorf("Rounded icon %s should have transparent corners", roundedName)
						}
					}
				}
			}
		})
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() Config
		expectError bool
	}{
		{
			name: "missing_input_file",
			setupFunc: func() Config {
				return Config{
					InputPath:     "/nonexistent/file.png",
					OutputDir:     t.TempDir(),
					TrimPercent:   80,
					RadiusPercent: 20,
				}
			},
			expectError: true,
		},
		{
			name: "readonly_output_directory",
			setupFunc: func() Config {
				testImg := createTestImage(100, color.RGBA{255, 0, 0, 255})
				inputPath := createTempImageFile(t, testImg)

				// Create a read-only directory (simulate permission error)
				readOnlyDir := filepath.Join(t.TempDir(), "readonly")
				os.MkdirAll(readOnlyDir, 0444) // read-only

				return Config{
					InputPath:     inputPath,
					OutputDir:     readOnlyDir,
					TrimPercent:   80,
					RadiusPercent: 20,
				}
			},
			expectError: true,
		},
		{
			name: "valid_config_should_succeed",
			setupFunc: func() Config {
				testImg := createTestImage(100, color.RGBA{0, 255, 0, 255})
				inputPath := createTempImageFile(t, testImg)

				return Config{
					InputPath:     inputPath,
					OutputDir:     t.TempDir(),
					TrimPercent:   80,
					RadiusPercent: 20,
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupFunc()

			// Validate config first
			err := validateConfig(config)
			if tt.expectError && err == nil {
				// If we expect an error but validation passes,
				// the error should come from generateIcons
				err = generateIcons(config)
			}

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestDifferentImageFormats tests loading various image formats
func TestDifferentImageFormats(t *testing.T) {
	// This test would be more comprehensive with actual image files,
	// but for now we test the PNG format that we generate

	formats := []struct {
		name  string
		color color.RGBA
		size  int
	}{
		{"red_square", color.RGBA{255, 0, 0, 255}, 100},
		{"green_rectangle", color.RGBA{0, 255, 0, 255}, 150},
		{"blue_large", color.RGBA{0, 0, 255, 255}, 300},
		{"transparent", color.RGBA{128, 128, 128, 128}, 80},
	}

	for _, format := range formats {
		t.Run(format.name, func(t *testing.T) {
			// Create test image
			testImg := createTestImage(format.size, format.color)
			inputPath := createTempImageFile(t, testImg)

			// Try to load it back
			loadedImg, err := loadImage(inputPath)
			if err != nil {
				t.Fatalf("Failed to load %s image: %v", format.name, err)
			}

			bounds := loadedImg.Bounds()
			if bounds.Dx() != format.size || bounds.Dy() != format.size {
				t.Errorf("Loaded image has wrong size: expected %dx%d, got %dx%d",
					format.size, format.size, bounds.Dx(), bounds.Dy())
			}

			// Test processing the loaded image
			outputDir := t.TempDir()
			config := Config{
				InputPath:     inputPath,
				OutputDir:     outputDir,
				Clean:         false,
				CropEnabled:   true,
				TrimPercent:   80,
				RadiusPercent: 20,
			}

			err = generateIcons(config)
			if err != nil {
				t.Errorf("Failed to process %s image: %v", format.name, err)
			}
		})
	}
}

// TestConcurrentGeneration tests that multiple generations can happen concurrently
func TestConcurrentGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Create test image
	testImg := createTestImage(200, color.RGBA{255, 100, 50, 255})
	inputPath := createTempImageFile(t, testImg)

	// Number of concurrent generations
	numWorkers := 5
	done := make(chan error, numWorkers)

	// Start concurrent generations
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			outputDir := filepath.Join(t.TempDir(), fmt.Sprintf("worker_%d", workerID))

			config := Config{
				InputPath:     inputPath,
				OutputDir:     outputDir,
				Clean:         false,
				CropEnabled:   true,
				TrimPercent:   80,
				RadiusPercent: 20,
			}

			err := generateIcons(config)
			done <- err
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < numWorkers; i++ {
		if err := <-done; err != nil {
			t.Errorf("Worker %d failed: %v", i, err)
		}
	}
}

// TestLargeImage tests processing of large images
func TestLargeImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large image test in short mode")
	}

	// Create a large test image (2048x2048)
	largeImg := createTestImageWithBorder(
		2048,
		color.RGBA{255, 200, 100, 255}, // orange center
		color.RGBA{50, 50, 50, 255},    // dark border
		100,
	)

	inputPath := createTempImageFile(t, largeImg)
	outputDir := t.TempDir()

	config := Config{
		InputPath:     inputPath,
		OutputDir:     outputDir,
		Clean:         false,
		CropEnabled:   true,
		TrimPercent:   85,
		RadiusPercent: 15,
	}

	// This should complete without error, even for large images
	err := generateIcons(config)
	if err != nil {
		t.Fatalf("Failed to process large image: %v", err)
	}

	// Verify that the largest icon was generated correctly
	largestIcon := filepath.Join(outputDir, "icon_1024x1024.png")
	img, err := loadImage(largestIcon)
	if err != nil {
		t.Fatalf("Failed to load largest generated icon: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 1024 || bounds.Dy() != 1024 {
		t.Errorf("Largest icon has wrong size: expected 1024x1024, got %dx%d",
			bounds.Dx(), bounds.Dy())
	}
}