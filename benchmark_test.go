package main

import (
	"fmt"
	"image/color"
	"testing"
)

// Benchmark functions to measure performance

func BenchmarkCropCenter(b *testing.B) {
	// Create a large test image
	testImg := createTestImage(1024, color.RGBA{255, 128, 64, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cropCenter(testImg, 80)
	}
}

func BenchmarkResizeImage(b *testing.B) {
	testImg := createTestImage(1024, color.RGBA{255, 128, 64, 255})

	benchmarks := []struct {
		name       string
		targetSize int
	}{
		{"resize_to_16", 16},
		{"resize_to_64", 64},
		{"resize_to_256", 256},
		{"resize_to_512", 512},
		{"resize_to_1024", 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = resizeImage(testImg, bm.targetSize)
			}
		})
	}
}

func BenchmarkAddRoundedCorners(b *testing.B) {
	testImg := createTestImage(512, color.RGBA{255, 128, 64, 255})

	benchmarks := []struct {
		name   string
		radius int
	}{
		{"radius_0", 0},
		{"radius_10", 10},
		{"radius_50", 50},
		{"radius_100", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = addRoundedCorners(testImg, bm.radius)
			}
		})
	}
}

func BenchmarkShouldKeepPixel(b *testing.B) {
	size := 512
	radius := 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				_ = shouldKeepPixel(x, y, size, radius)
			}
		}
	}
}

func BenchmarkGenerateAllIcons(b *testing.B) {
	// Create a realistic test image
	testImg := createTestImageWithBorder(
		512,
		color.RGBA{255, 200, 100, 255}, // orange center
		color.RGBA{100, 100, 100, 255}, // gray border
		40,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup for each iteration
		inputPath := createTempImageFile(b, testImg)
		outputDir := b.TempDir()

		config := Config{
			InputPath:     inputPath,
			OutputDir:     outputDir,
			Clean:         false,
			CropEnabled:   true,
			TrimPercent:   80,
			RadiusPercent: 20,
		}
		b.StartTimer()

		// Measure the actual icon generation
		err := generateIcons(config)
		if err != nil {
			b.Fatalf("Failed to generate icons: %v", err)
		}
	}
}

func BenchmarkImageFormats(b *testing.B) {
	// Test different image sizes to see performance characteristics
	sizes := []int{64, 128, 256, 512, 1024}

	for _, size := range sizes {
		testImg := createTestImage(size, color.RGBA{255, 0, 128, 255})

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				tmpPath := createTempImageFile(b, testImg)
				b.StartTimer()

				// Measure load -> process -> save cycle
				img, err := loadImage(tmpPath)
				if err != nil {
					b.Fatalf("Failed to load image: %v", err)
				}

				cropped := cropCenter(img, 80)
				resized := resizeImage(cropped, 128)
				rounded := addRoundedCorners(resized, 26)

				b.StopTimer()
				saveImage(rounded, tmpPath+"_out.png")
				b.StartTimer()
			}
		})
	}
}

// Benchmark memory allocations
func BenchmarkMemoryUsage(b *testing.B) {
	testImg := createTestImage(1024, color.RGBA{255, 128, 64, 255})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cropped := cropCenter(testImg, 80)
		resized := resizeImage(cropped, 512)
		_ = addRoundedCorners(resized, 100)
	}
}
