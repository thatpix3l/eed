package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

func maskRadius[T any](mask [][]T) int {
	return (len(mask) - 1) / 2
}

var (
	sobelMaskX = [][]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	sobelMaskY = [][]int{
		{1, 2, 1},
		{0, 0, 0},
		{-1, -2, -1},
	}
)

const (
	thresholdPa1High float64 = 0.4314
	thresholdPa1Low  float64 = 0.1568

	thresholdPa2Low  float64 = float64(35) / 255
	thresholdPa2High float64 = float64(100) / 255
)

type number interface {
	float32 | float64 | int | byte
}

func writeImage(writer io.Writer, image [][]byte) error {
	for _, row := range image {
		if _, err := writer.Write(row); err != nil {
			return fmt.Errorf("unable to write image: %v", err)
		}
	}

	return nil
}

func castNestedSlice[T, U number](slice [][]T) [][]U {
	output := new2dSlice[U](len(slice), len(slice))
	for i, row := range slice {
		for j, v := range row {
			output[i][j] = U(v)
		}
	}
	return output
}

func applyEuclideanDistance[T number](nums ...T) float64 {
	var summedSquares float64 = 0
	for _, num := range nums {
		summedSquares += math.Pow(float64(num), 2)
	}
	return math.Sqrt(summedSquares)
}

func new2dSlice[T any](rows int, columns int) [][]T {
	slice := make([][]T, rows)
	for i := range slice {
		slice[i] = make([]T, columns)
	}
	return slice
}

func applyMask[T, U number](image [][]T, mask [][]U) [][]T {
	maskRadius := (len(mask) - 1) / 2

	output := new2dSlice[T](len(image), len(image))

	for maskCenterRowIdx := 0 + maskRadius; maskCenterRowIdx < len(image)-maskRadius; maskCenterRowIdx++ {
		for maskCenterColIdx := 0 + maskRadius; maskCenterColIdx < len(image)-maskRadius; maskCenterColIdx++ {

			for maskRowIdx, maskRow := range mask {
				for maskColIdx, maskPixel := range maskRow {

					imagePixelRowIdx := maskCenterRowIdx - maskRadius + maskRowIdx
					imagePixelColIdx := maskCenterColIdx - maskRadius + maskColIdx

					imagePixel := image[imagePixelRowIdx][imagePixelColIdx]

					output[maskCenterRowIdx][maskCenterColIdx] += imagePixel * T(maskPixel)

				}
			}

		}
	}

	return output
}

func Map[Input, Output any](slice []Input, mapper func(int, Input) Output) []Output {
	output := make([]Output, len(slice))
	for i, v := range slice {
		output[i] = mapper(i, v)
	}
	return output
}

func applyEuclideanDistanceImage[T number](maskRadius int, maxPixelValue *float64, componentMagnitudes ...[][]T) [][]float64 {

	if maxPixelValue == nil {
		var dummy float64
		maxPixelValue = &dummy
	}

	imageDimension := len(componentMagnitudes[0])
	imageDimensionSafe := imageDimension - maskRadius

	output := new2dSlice[float64](imageDimension, imageDimension)

	for i := maskRadius; i < imageDimensionSafe; i++ {
		for j := maskRadius; j < imageDimensionSafe; j++ {

			pixelComponents := []T{}
			for _, magnitudes := range componentMagnitudes {
				pixelComponents = append(pixelComponents, magnitudes[i][j])
			}

			output[i][j] = applyEuclideanDistance(pixelComponents...)

			if output[i][j] > *maxPixelValue {
				*maxPixelValue = output[i][j]
			}

		}
	}

	return output
}

func applyThreshold(threshold float64, max float64, pixel float64) float64 {

	if pixel >= threshold*max {
		return max
	}

	return 0
}

func applyThresholdImage(threshold float64, max float64, image [][]float64) [][]float64 {

	output := new2dSlice[float64](len(image), len(image))

	for i, row := range image {
		for j, pixel := range row {
			output[i][j] = applyThreshold(threshold, max, pixel)
		}
	}

	return output
}

// Scale each pixel by factor.
func applyScaleImage(image [][]float64, factor float64) [][]float64 {

	output := new2dSlice[float64](len(image), len(image))

	for i, row := range image {
		for j, pixel := range row {
			output[i][j] = pixel * factor
		}
	}

	return output

}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func squaredSum(nums ...float64) float64 {
	var sum float64 = 0

	for _, num := range nums {
		sum += math.Pow(num, 2)
	}

	return sum

}

func gaussian1stDerivMultiplePixels(sigma float64, nums ...float64) float64 {
	exponentialPower := -(1 / (2 * math.Pow(sigma, 2))) * squaredSum(nums...)
	exponentialPowerApplied := math.Pow(math.E, exponentialPower)
	return nums[0] * exponentialPowerApplied
}

func sobelEdgeDetection(image [][]float64, useThreshold bool, threshold float64) [][]float64 {

	// Calculate weighted sums for x and y direction for each pixel.
	imageWeightedSumsX := applyMask(image, sobelMaskX)
	imageWeightedSumsY := applyMask(image, sobelMaskY)

	// Calculate euclidean distance and highest pixel value, based on image pixel direction differences.
	var maxPixelValue float64
	imageMagnitudeAbsolute := applyEuclideanDistanceImage(maskRadius(sobelMaskX), &maxPixelValue, imageWeightedSumsX, imageWeightedSumsY)

	// Map image pixel values from [0-maxPixelValue] range to unit interval range.
	imageMagnitudeUnitInterval := applyScaleImage(imageMagnitudeAbsolute, 1/maxPixelValue)

	// Apply threshold to each pixel.
	imageThresholded := imageMagnitudeUnitInterval
	if useThreshold {
		imageThresholded = applyThresholdImage(threshold, 1, imageThresholded)
	}

	return imageThresholded
}

func cannyEdgeHighLow(image [][]float64) (high float64, low float64) {
	imagePgmScaled := applyScaleImage(image, 255)
	imagePgmScaledInt := castNestedSlice[float64, int](imagePgmScaled)
	histogram := [255]int{}

	for _, row := range imagePgmScaledInt {
		for _, pixel := range row {
			histogram[pixel]++
		}
	}

	var mostOccurredPixel int
	for pixel, occurrence := range histogram {
		if occurrence > histogram[mostOccurredPixel] {
			mostOccurredPixel = pixel
		}
	}

	high = float64(mostOccurredPixel) / 255
	low = high * 0.35

	return high, low
}

func cannyEdgeDetection(image [][]float64, sigma int) [][]float64 {
	return [][]float64{}
}

func run() error {

	var inputPath string
	flag.StringVar(&inputPath, "input", "input.pgm", "path to input file")

	var outputPath string
	flag.StringVar(&outputPath, "output", "output.pgm", "path to output file")

	thresholdFlag := flag.String("threshold", "0", "minimum threshold for a pixel to be considered 'on'")

	var sigma int
	flag.IntVar(&sigma, "sigma", 0, "canny edge detection sigma")

	flag.Parse()

	useCannyEdgeDetection := isFlagPassed("sigma")

	var threshold float64 = 0
	useThreshold := isFlagPassed("threshold")

	if *thresholdFlag == "pa1high" {
		threshold = thresholdPa1High

	} else if *thresholdFlag == "pa1low" {
		threshold = thresholdPa1Low

	} else if *thresholdFlag == "pa2high" {
		threshold = thresholdPa2High

	} else if *thresholdFlag == "pa2low" {
		threshold = thresholdPa2Low

	} else if thresholdParsed, err := strconv.ParseFloat(*thresholdFlag, 64); err != nil {
		return fmt.Errorf("unable to parse threshold: %v", err)

	} else {
		threshold = thresholdParsed
	}

	// Open input file.
	var inputFile *os.File
	{
		f, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("unable to open input file: %v", err)
		}
		inputFile = f
		defer inputFile.Close()
	}

	// Open output file
	var outputFile *os.File
	{
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("unable to create output file: %v", err)
		}
		outputFile = f
		defer outputFile.Close()
	}

	// Read header data, pre-allocate image size.
	var imageByte [][]byte
	imageInputMaxPixelVal := 0
	{
		columns := 0
		rows := 0

		if _, err := fmt.Fscanf(inputFile, "P5\n%d %d\n%d\n", &columns, &rows, &imageInputMaxPixelVal); err != nil {
			return fmt.Errorf("unable to read header: %v", err)
		}

		imageByte = new2dSlice[byte](rows, columns)
	}

	// Make in-memory copy of image.
	b := make([]byte, 1)
	for i := range imageByte {
		for j := range imageByte {

			// Read single byte
			if _, err := inputFile.Read(b); err != nil {
				return fmt.Errorf("unable to read pixel byte from input image: %v", err)
			}

			b[0] &= ^byte(0)

			// Store as int for later use.
			imageByte[i][j] = b[0]

		}
	}

	// Cast image from bytes to floats.
	imageInputFloat := castNestedSlice[byte, float64](imageByte)

	// Map image pixel values from the PGM format range to the unit interval range.
	imageUnitInterval := applyScaleImage(imageInputFloat, 1/float64(imageInputMaxPixelVal))

	var imageDetectedEdges [][]float64

	if useCannyEdgeDetection {

	} else {
		imageDetectedEdges = sobelEdgeDetection(imageUnitInterval, useThreshold, threshold)
	}

	// Scale pixels from unit interval range to PGM pixel range.
	imagePgmScaled := applyScaleImage(imageDetectedEdges, 255)

	// Cast image pixels from floats into bytes.
	outputImage := castNestedSlice[float64, byte](imagePgmScaled)

	// Write header to output file.
	fmt.Fprintf(outputFile, "P5\n%d %d\n255\n", len(imageByte), len(imageByte[0]))

	// Write image to output file.
	if err := writeImage(outputFile, outputImage); err != nil {
		return fmt.Errorf("unable to write output image: %v", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("pgm processing error: %v", err)
	}
}
