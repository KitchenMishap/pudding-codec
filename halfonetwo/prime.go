package halfonetwo

import "math"

type LogYFracHist struct {
	LnY             float64
	bins            []uint64
	peaks           []bool // A peak has a higher count than both it's neighbours
	populationCount uint64
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
	lb.populationCount = 0
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
		lb.populationCount++
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
// Returns a strength as a proportion of population captured
func (lb *LogYFracHist) AssessPrimePeaks(amount uint64, threshold int) float64 {
	matchingPeaks := 0
	matchingPopulation := uint64(0)
	// Note that 1,2,3,5,7,11 are all prime numbers, to avoid matching "offset" ghosts
	bin1 := lb.AmountToBin(amount)
	bin2 := lb.AmountToBin(amount * 2)
	bin3 := lb.AmountToBin(amount * 3)
	bin5 := lb.AmountToBin(amount * 5)
	bin7 := lb.AmountToBin(amount * 7)
	bin11 := lb.AmountToBin(amount * 11)

	if lb.PeakNear(bin1) { // The "One" is a required peak
		matchingPopulation += lb.bins[bin1]
		if lb.PeakNear(bin2) { // The "Two" is a required peak
			matchingPopulation += lb.bins[bin2]
			if lb.PeakNear(bin5) { // The "Five" is a required peak
				matchingPopulation += lb.bins[bin5]
				if lb.PeakNear(bin3) {
					matchingPeaks++
					matchingPopulation += lb.bins[bin3]
				} // "Three" is a bonus score
				if lb.PeakNear(bin7) {
					matchingPeaks++
					matchingPopulation += lb.bins[bin7]
				} // "Seven" is a bonus score
				if lb.PeakNear(bin11) {
					matchingPeaks++
					matchingPopulation += lb.bins[bin11]
				} // "Eleven" is a bonus score
			}
		}
	}
	if matchingPeaks >= threshold {
		return float64(matchingPopulation) / float64(lb.populationCount)
	} else {
		return 0
	}
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
