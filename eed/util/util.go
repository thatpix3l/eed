package util

import (
	"flag"
	"fmt"
	"io"
	"math"
)

type Number interface {
	float32 | float64 | int | byte
}

func WriteImage(writer io.Writer, image [][]byte) error {
	for _, row := range image {
		if _, err := writer.Write(row); err != nil {
			return fmt.Errorf("unable to write image: %v", err)
		}
	}

	return nil
}

func CastNestedSlice[T, U Number](slice [][]T) [][]U {
	output := New2dSlice[U](len(slice), len(slice))
	for i, row := range slice {
		for j, v := range row {
			output[i][j] = U(v)
		}
	}
	return output
}

func ApplyEuclideanDistance[T Number](nums ...T) float64 {
	var summedSquares float64 = 0
	for _, num := range nums {
		summedSquares += math.Pow(float64(num), 2)
	}
	return math.Sqrt(summedSquares)
}

func New2dSlice[T any](rows int, columns int) [][]T {
	slice := make([][]T, rows)
	for i := range slice {
		slice[i] = make([]T, columns)
	}
	return slice
}

func ApplyMask[T, U Number](image [][]T, mask [][]U) [][]T {
	maskRadius := (len(mask) - 1) / 2

	output := New2dSlice[T](len(image), len(image))

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

func ApplyEuclideanDistanceImage[T Number](maskRadius int, maxPixelValue *float64, componentMagnitudes ...[][]T) [][]float64 {

	if maxPixelValue == nil {
		var dummy float64
		maxPixelValue = &dummy
	}

	imageDimension := len(componentMagnitudes[0])
	imageDimensionSafe := imageDimension - maskRadius

	output := New2dSlice[float64](imageDimension, imageDimension)

	for i := maskRadius; i < imageDimensionSafe; i++ {
		for j := maskRadius; j < imageDimensionSafe; j++ {

			pixelComponents := []T{}
			for _, magnitudes := range componentMagnitudes {
				pixelComponents = append(pixelComponents, magnitudes[i][j])
			}

			output[i][j] = ApplyEuclideanDistance(pixelComponents...)

			if output[i][j] > *maxPixelValue {
				*maxPixelValue = output[i][j]
			}

		}
	}

	return output
}

func ApplyThreshold(threshold float64, max float64, pixel float64) float64 {

	if pixel >= threshold*max {
		return max
	}

	return 0
}

func ApplyThresholdImage(threshold float64, max float64, image [][]float64) [][]float64 {

	output := New2dSlice[float64](len(image), len(image))

	for i, row := range image {
		for j, pixel := range row {
			output[i][j] = ApplyThreshold(threshold, max, pixel)
		}
	}

	return output
}

// Scale each pixel by factor.
func ApplyScaleImage(image [][]float64, factor float64) [][]float64 {

	output := New2dSlice[float64](len(image), len(image))

	for i, row := range image {
		for j, pixel := range row {
			output[i][j] = pixel * factor
		}
	}

	return output

}

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func SquaredSum(nums ...float64) float64 {
	var sum float64 = 0

	for _, num := range nums {
		sum += math.Pow(num, 2)
	}

	return sum

}

func MaskRadius[T any](mask [][]T) int {
	return (len(mask) - 1) / 2
}
