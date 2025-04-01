package sobel

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thatpix3l/edge_detection/cmd/shared"
	"github.com/thatpix3l/edge_detection/util"
)

const (
	thresholdLow  float64 = 0.1568
	thresholdHigh float64 = 0.4314
)

var thresholdMap = map[string]float64{
	"low":  thresholdLow,
	"high": thresholdHigh,
}

var (
	maskX = [][]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	maskY = [][]int{
		{1, 2, 1},
		{0, 0, 0},
		{-1, -2, -1},
	}
)

var threshold sobelThreshold

var (
	Command = &cobra.Command{
		Use:   "sobel",
		Short: "Use the Sobel filter.",
		RunE: func(cmd *cobra.Command, args []string) error {

			shared.ImageAfterFilter = sobelEdgeDetection(shared.ImageBeforeFilter, cmd.Flag("threshold").Changed, float64(threshold))

			return nil
		},
	}
)

type sobelThreshold float64

func (st *sobelThreshold) String() string {
	return fmt.Sprint(*st)
}

func (st *sobelThreshold) Set(v string) error {
	acceptableThreshold, ok := thresholdMap[v]
	if !ok {
		if parsedThreshold, err := strconv.ParseFloat(v, 64); err != nil {
			return fmt.Errorf("could not parse as \"high\", \"low\", or a float: %v", err)
		} else {
			acceptableThreshold = parsedThreshold
		}
	}

	if err := thresholdInRange(acceptableThreshold); err != nil {
		return err
	}

	*st = sobelThreshold(acceptableThreshold)

	return nil

}

func (t *sobelThreshold) Type() string {
	return "sobel_threshold"
}

func thresholdInRange(value float64) error {
	if value < 0 || value > 1 {
		return errors.New("not a float value in between 0 and 1")
	}

	return nil
}

func init() {
	Command.Flags().Var(&threshold, "threshold", "minimum threshold for a pixel to be considered 'on'")
}

func sobelEdgeDetection(image [][]float64, useThreshold bool, threshold float64) [][]float64 {

	// Calculate weighted sums for x and y direction for each pixel.
	imageWeightedSumsX := util.ApplyMask(image, maskX)
	imageWeightedSumsY := util.ApplyMask(image, maskY)

	// Calculate euclidean distance and highest pixel value, based on image pixel direction differences.
	var maxPixelValue float64
	imageMagnitudeAbsolute := util.ApplyEuclideanDistanceImage(util.MaskRadius(maskX), &maxPixelValue, imageWeightedSumsX, imageWeightedSumsY)

	// Map image pixel values from [0-maxPixelValue] range to unit interval range.
	imageMagnitudeUnitInterval := util.ApplyScaleImage(imageMagnitudeAbsolute, 1/maxPixelValue)

	// Apply threshold to each pixel.
	imageThresholded := imageMagnitudeUnitInterval
	if useThreshold {
		imageThresholded = util.ApplyThresholdImage(threshold, 1, imageThresholded)
	}

	return imageThresholded
}
