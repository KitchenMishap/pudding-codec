package graphics

import (
	"fmt"
	"os"
)

// Quad FullHD
const width = 1920 * 17
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

func (ph *PgmHist) Output(filename string) {
	fp, _ := os.Create(filename)
	fmt.Printf("Writing pgm file")
	fmt.Fprintf(fp, "P6 %d %d 255\n", width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			//lsb := byte(ph.data[x][y] & 0xFF)
			//msb := byte((ph.data[x][y] & 0xFF00) >> 8)
			fp.Write([]byte{ph.data[x][y][0]})
			fp.Write([]byte{ph.data[x][y][1]})
			fp.Write([]byte{ph.data[x][y][2]})
		}
	}
	fp.Close()
	fmt.Printf("File written")
}
