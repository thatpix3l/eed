#include <math.h>
#include <stdio.h> /* Sobel.c */

int image_input[256][256];
int image_output_x[256][256];
int image_output_y[256][256];
int mask_x[3][3] = {{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}};
int mask_y[3][3] = {{1, 2, 1}, {0, 0, 0}, {-1, -2, -1}};
double ival[256][256], maxival;

int pixel_input_mask = 255;

double apply_threshold(double threshold, double pixel) {
    return (pixel >= (threshold * 255)) * 255;
}

int main(int argc, char **argv) {
    char *input_file_name = argv[1];
    FILE *input_file = fopen(input_file_name, "rb");

    char *output_file_name = argv[2];
    FILE *output_file = fopen(output_file_name, "wb");

    int use_threshold = 0;
    double threshold = 0;

    if (argc >= 4) {
        use_threshold = 1;
        char *threshold_ascii = argv[3];
        threshold = atof(threshold_ascii);
    }

    int inputColumns, inputRows, maxPixelVal;
    fscanf(input_file, "P5\n%d %d\n%d\n", &inputColumns, &inputRows,
           &maxPixelVal);

    // Make in-memory copy of input image.
    for (int i = 0; i < 256; i++) {
        for (int j = 0; j < 256; j++) {
            image_input[i][j] = getc(input_file);
            image_input[i][j] &= pixel_input_mask;
        }
    }

    // Math stuff.
    int mask_radius = 1;
    for (int i = mask_radius; i < 256 - mask_radius; i++) {
        for (int j = mask_radius; j < 256 - mask_radius; j++) {
            int weighted_sum_x = 0;
            int weighted_sum_y = 0;
            for (int mask_row = -mask_radius; mask_row <= mask_radius;
                 mask_row++) {
                for (int mask_col = -mask_radius; mask_col <= mask_radius;
                     mask_col++) {
                    int pixel = image_input[i + mask_row][j + mask_col];

                    int weight_x =
                        mask_x[mask_row + mask_radius][mask_col + mask_radius];

                    int weight_y =
                        mask_y[mask_row + mask_radius][mask_col + mask_radius];

                    weighted_sum_x += pixel * weight_x;
                    weighted_sum_y += pixel * weight_y;
                }
            }
            image_output_x[i][j] = weighted_sum_x;
            image_output_y[i][j] = weighted_sum_y;
        }
    }

    // More math stuff.
    maxival = 0;
    for (int i = mask_radius; i < 256 - mask_radius; i++) {
        for (int j = mask_radius; j < 256 - mask_radius; j++) {
            ival[i][j] =
                sqrt((double)((image_output_x[i][j] * image_output_x[i][j]) +
                              (image_output_y[i][j] * image_output_y[i][j])));
            if (ival[i][j] > maxival) maxival = ival[i][j];
        }
    }

    int rows = 256;
    int cols = rows;

    // Scale each pixel to ratio relative to brightest pixel.
    for (int i = 0; i < 256; i++) {
        for (int j = 0; j < 256; j++) {
            ival[i][j] = (ival[i][j] / maxival) * 255;
        }
    }

    // Apply threshold to each pixel value
    if (use_threshold) {
        for (int i = 0; i < 256; i++) {
            for (int j = 0; j < 256; j++) {
                ival[i][j] = apply_threshold(threshold, ival[i][j]);
            }
        }
    }

    // Write image header to output file.
    fprintf(output_file, "P5\n%d %d\n255\n", rows, cols);

    // Write image to output file..
    for (int i = 0; i < 256; i++) {
        for (int j = 0; j < 256; j++) {
            fprintf(output_file, "%c", (char)((int)(ival[i][j])));
        }
    }
}
