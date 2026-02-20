package halfonetwo

import "math"

type LogYFracHist struct {
	LnY   float64
	bins  []uint64
	peaks []bool // A peak has a higher count than both it's neighbours
}

func NewLogYFracHist(binCount int, baseY float64) *LogYFracHist {
	result := LogYFracHist{}
	result.LnY = math.Log(baseY)
	result.bins = make([]uint64, binCount)
	result.peaks = make([]bool, binCount)
	return &result
}

func (lb *LogYFracHist) BaseYLog(amount uint64) float64 {
	return math.Log(float64(amount)) / lb.LnY
}

func (lb *LogYFracHist) Wipe() {
	for i := range lb.bins {
		lb.bins[i] = 0
	}
}

func (lb *LogYFracHist) AmountToBin(amount uint64) int {
	log := lb.BaseYLog(amount)
	logInt := math.Floor(log)
	logFrac := log - logInt
	return int(logFrac * float64(len(lb.bins)))
}

func (lb *LogYFracHist) Populate(amounts []uint64) {
	for _, amount := range amounts {
		lb.bins[lb.AmountToBin(amount)]++
	}
}

func (lb *LogYFracHist) FindPeaks() {
	binCount := len(lb.bins)
	for index := range lb.bins {
		rightNeighbour := (index + 1 + binCount) % binCount
		leftNeighbour := (index - 1 + binCount) % binCount
		lb.peaks[index] = lb.bins[leftNeighbour] < lb.bins[index] && lb.bins[rightNeighbour] < lb.bins[index]
	}
}

// Recommend threshold between 1 (sensitive) and 2 (fussy)
func (lb *LogYFracHist) AssessPrimePeaks(amount uint64, threshold int) bool {
	matchingPeaks := 0
	// Note that 1,2,3,5,7,11 are all prime numbers, to avoid matching "offset" ghosts
	if lb.PeakNear(lb.AmountToBin(amount)) { // The "One" is a required peak
		if lb.PeakNear(lb.AmountToBin(amount * 2)) { // The "Two" is a required peak
			if lb.PeakNear(lb.AmountToBin(amount * 5)) { // The "Five" is a required peak
				if lb.PeakNear(lb.AmountToBin(amount * 3)) {
					matchingPeaks++
				} // "Three" is a bonus score
				if lb.PeakNear(lb.AmountToBin(amount * 7)) {
					matchingPeaks++
				} // "Seven" is a bonus score
				if lb.PeakNear(lb.AmountToBin(amount * 11)) {
					matchingPeaks++
				} // "Eleven" is a bonus score
			}
		}
	}
	return matchingPeaks >= threshold
}

func (lb *LogYFracHist) PeakNear(bin int) bool {
	bins := len(lb.bins)
	if lb.peaks[bin] {
		return true
	}
	if lb.peaks[(bin+1)%bins] {
		return true
	}
	if lb.peaks[(bin+bins-1)%bins] {
		return true
	}
	return false
}
