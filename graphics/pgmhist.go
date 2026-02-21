package graphics

import (
	"fmt"
	"os"
)

// Seventeen years of FHD
const width = 1920 * 17

// const width = 8888
const height = 1080

func TestPgmHist() {
	pgm := PgmHist{}
	pgm.PlotPoint(0.75, 0.75, 7)
	pgm.PlotVertical(0.5)
	pgm.PlotHorizontal(0.5)
	pgm.Output("PgmTest.pgm")
}

type PgmHist struct {
	data [width][height][3]byte
}

func (ph *PgmHist) PlotPoint(x float64, y float64, colour3bit int) {
	pixelX := int(width * x)
	pixelY := int(height * (1 - y))
	ph.PlotPixel(pixelX, pixelY, colour3bit)
	ph.PlotPixel(pixelX, pixelY, colour3bit)
	ph.PlotPixel(pixelX-1, pixelY, colour3bit)
	ph.PlotPixel(pixelX+1, pixelY, colour3bit)
	ph.PlotPixel(pixelX, pixelY+1, colour3bit)
	ph.PlotPixel(pixelX, pixelY-1, colour3bit)
}

func (ph *PgmHist) PlotWidePoint(xLeft float64, xRight float64, y float64, colour3bit int) {
	pixelXLeft := int(width * xLeft)
	pixelXRight := int(width * xRight)
	pixelXMid := int(width * (xLeft + xRight) / 2)
	pixelY := int(height * (1 - y))
	ph.PlotPixel(pixelXMid, pixelY, colour3bit)
	ph.PlotPixel(pixelXMid, pixelY, colour3bit)
	for pixelXScan := pixelXLeft - 1; pixelXScan <= pixelXRight+1; pixelXScan++ {
		ph.PlotPixel(pixelXScan, pixelY, colour3bit)
	}
	ph.PlotPixel(pixelXMid, pixelY+1, colour3bit)
	ph.PlotPixel(pixelXMid, pixelY-1, colour3bit)
}

func (ph *PgmHist) PlotVertical(x float64) {
	pixelX := int(width * x)
	for y := range height {
		ph.PlotPixel(pixelX, y, 7)
	}
}

func (ph *PgmHist) PlotHorizontal(y float64) {
	pixelY := int(height * y)
	for x := range width {
		ph.PlotPixel(x, pixelY, 7)
	}
}

func (ph *PgmHist) PlotPixel(x int, y int, colour3bit int) {
	if x >= 0 && x < width && y >= 0 && y < height {
		red := ((colour3bit & 4) == 4)
		green := ((colour3bit & 2) == 2)
		blue := ((colour3bit & 1) == 1)
		if red {
			ph.data[x][y][0]++
		}
		if green {
			ph.data[x][y][1]++
		}
		if blue {
			ph.data[x][y][2]++
		}
	}
}

func (ph *PgmHist) NormalizeColumns() {
	for x := 0; x < width; x++ {
		maxByte := byte(0)
		for y := 0; y < height; y++ {
			if ph.data[x][y][0] > maxByte {
				maxByte = ph.data[x][y][0]
			}
			if ph.data[x][y][1] > maxByte {
				maxByte = ph.data[x][y][1]
			}
			if ph.data[x][y][2] > maxByte {
				maxByte = ph.data[x][y][2]
			}
		}
		if maxByte > 0 && maxByte < 255 {
			for y := 0; y < height; y++ {
				ph.data[x][y][0] = byte((255 * int(ph.data[x][y][0])) / int(maxByte))
				ph.data[x][y][1] = byte((255 * int(ph.data[x][y][1])) / int(maxByte))
				ph.data[x][y][2] = byte((255 * int(ph.data[x][y][2])) / int(maxByte))
			}
		}
	}
}

func (ph *PgmHist) Output(filename string) {
	fp, _ := os.Create(filename)
	fmt.Printf("Writing ppm file\n")
	fmt.Fprintf(fp, "P6 %d %d 255\n", width, height)
	data := [width * height * 3]byte{}
	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			//lsb := byte(ph.data[x][y] & 0xFF)
			//msb := byte((ph.data[x][y] & 0xFF00) >> 8)
			//fp.Write([]byte{ph.data[x][y][0]})
			//fp.Write([]byte{ph.data[x][y][1]})
			//fp.Write([]byte{ph.data[x][y][2]})
			data[index] = ph.data[x][y][0]
			index++
			data[index] = ph.data[x][y][1]
			index++
			data[index] = ph.data[x][y][2]
			index++
		}
	}
	fp.Write(data[:])
	fp.Close()
	fmt.Printf("File written\n")
}
