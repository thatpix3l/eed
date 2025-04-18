VERSION 0.8

FROM golang:1.23.4-alpine
WORKDIR /go-workdir

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
    FROM +deps
    COPY . .
    RUN go mod download
    RUN go build -o build/ .

build-and-save:
    FROM +build
    SAVE ARTIFACT build/* AS LOCAL ./

edd-run-deps:
    FROM +build
    WORKDIR ./build
    COPY images/input input
    RUN mkdir -p output

sobel:
    FROM +edd-run-deps
    RUN ./edd sobel --input input/garb34.pgm --output output/garb34_sobel_mag.pgm
    RUN ./edd sobel --input input/garb34.pgm --output output/garb34_sobel_low.pgm --threshold low
    RUN ./edd sobel --input input/garb34.pgm --output output/garb34_sobel_high.pgm --threshold high
    SAVE ARTIFACT output/* AS LOCAL images/output/

canny:
    COPY cmd/ao ./ao
    RUN mkdir output
    WORKDIR ao
    RUN for f in *; do cat "$f" > ../output/"$f"; done
    RUN ls -alh ../output
    SAVE ARTIFACT ../output/cannyfinal.pgm AS LOCAL images/output/garb34_canny_final.pgm
    SAVE ARTIFACT ../output/cannymag.pgm AS LOCAL images/output/garb34_canny_mag.pgm
    SAVE ARTIFACT ../output/cannypeaks.pgm AS LOCAL images/output/garb34_canny_peaks.pgm