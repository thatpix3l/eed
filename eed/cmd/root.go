package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thatpix3l/eed/eed/cmd/canny"
	"github.com/thatpix3l/eed/eed/cmd/shared"
	"github.com/thatpix3l/eed/eed/cmd/sobel"
	"github.com/thatpix3l/eed/eed/util"
)

var Root = &cobra.Command{
	Use:                "edd",
	Short:              "Elementary-Edge-Detection is an edge detector and visualizer for PGM images.",
	PersistentPreRunE:  readInputImage,
	PersistentPostRunE: writeOutputImage,
}

var (
	inputPath  string
	outputPath string
)

func init() {
	cobra.EnableTraverseRunHooks = true

	Root.PersistentFlags().StringVar(&inputPath, "input", "", "path to read input image file")
	Root.PersistentFlags().StringVar(&outputPath, "output", "", "path to write output image file")

	if err := Root.MarkPersistentFlagRequired("input"); err != nil {
		panic(`unable to mark persistent flag "input" as required`)
	}

	if err := Root.MarkPersistentFlagRequired("output"); err != nil {
		panic(`unable to mark persistent flag "output" as required`)
	}

	Root.AddCommand(sobel.Command, canny.Command)
}

var imageByte [][]byte
var outputFile *os.File

func readInputImage(cmd *cobra.Command, args []string) error {

	// Open input file.
	var inputFile *os.File
	{
		f, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("could not open input file \"%s\": %v", inputPath, err)
		}
		inputFile = f
		defer inputFile.Close()
	}

	// Open output file
	{
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("unable to create output file \"%s\": %v", outputPath, err)
		}
		outputFile = f
	}

	// Read header data, pre-allocate image size.
	imageInputMaxPixelVal := 0
	{
		columns := 0
		rows := 0

		if _, err := fmt.Fscanf(inputFile, "P5\n%d %d\n%d\n", &columns, &rows, &imageInputMaxPixelVal); err != nil {
			return fmt.Errorf("unable to read header: %v", err)
		}

		imageByte = util.New2dSlice[byte](rows, columns)
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
	imageInputFloat := util.CastNestedSlice[byte, float64](imageByte)

	// Map image pixel values from the PGM format range to the unit interval range.
	shared.ImageBeforeFilter = util.ApplyScaleImage(imageInputFloat, 1/float64(imageInputMaxPixelVal))

	return nil
}

func writeOutputImage(cmd *cobra.Command, args []string) error {
	defer outputFile.Close()

	// Scale pixels from unit interval range to PGM pixel range.
	imagePgmScaled := util.ApplyScaleImage(shared.ImageAfterFilter, 255)

	// Cast image pixels from floats into bytes.
	outputImage := util.CastNestedSlice[float64, byte](imagePgmScaled)

	// Write header to output file.
	fmt.Fprintf(outputFile, "P5\n%d %d\n255\n", len(imageByte), len(imageByte[0]))

	// Write image to output file.
	if err := util.WriteImage(outputFile, outputImage); err != nil {
		return fmt.Errorf("unable to write output image: %v", err)
	}

	return nil
}
