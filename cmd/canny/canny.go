package canny

import (
	"fmt"
	"math"

	"github.com/spf13/cobra"
	"github.com/thatpix3l/edd/util"
)

const (
	thresholdLow  float64 = float64(35) / 255
	thresholdHigh float64 = float64(100) / 255
)

var Command = &cobra.Command{
	Use:   "canny",
	Short: "Use the Canny filter.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("in canny thing")
		return nil
	},
}

var (
	sigma int
)

func init() {
	Command.Flags().IntVar(&sigma, "sigma", 0, "canny edge detection sigma")
}

func gaussian1stDerivMultiplePixels(sigma float64, nums ...float64) float64 {
	exponentialPower := -(1 / (2 * math.Pow(sigma, 2))) * util.SquaredSum(nums...)
	exponentialPowerApplied := math.Pow(math.E, exponentialPower)
	return nums[0] * exponentialPowerApplied
}

func cannyEdgeHighLow(image [][]float64) (high float64, low float64) {
	imagePgmScaled := util.ApplyScaleImage(image, 255)
	imagePgmScaledInt := util.CastNestedSlice[float64, byte](imagePgmScaled)
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
