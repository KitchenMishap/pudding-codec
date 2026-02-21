package halfonetwo

import "math"

const primeNumberCount = 6

type LogYFracHist struct {
	LnY  float64
	bins []uint64
	//primeVotes		[][primeNumberCount]bool
	peaks           []bool // A peak has a higher count than both it's neighbours
	populationCount uint64
	primeNumbers    []uint64
}

func NewLogYFracHist(binCount int, baseY float64) *LogYFracHist {
	result := LogYFracHist{}
	result.LnY = math.Log(baseY)
	result.bins = make([]uint64, binCount)
	result.peaks = make([]bool, binCount)
	//result.primeVotes = make([][primeNumberCount]bool, binCount)
	result.primeNumbers = ([]uint64{1, 2, 3, 5, 7, 11, 13, 17, 19, 23, 29})[0:primeNumberCount]
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
func (lb *LogYFracHist) AssessPrimePeaksStrength(amount uint64) float64 {
	matchingPopulation := uint64(0)

	// We use prime multipliers, to avoid matching "offset" ghosts
	primeBins := [primeNumberCount]int{}
	for p, prime := range lb.primeNumbers {
		bin := lb.AmountToBin(amount * prime)
		primeBins[p] = bin
	}

	for p := range lb.primeNumbers {
		if lb.PeakNear(primeBins[p]) {
			matchingPopulation += lb.bins[primeBins[p]]
		} else {
			return 0
		}
	}
	return float64(matchingPopulation) / float64(lb.populationCount)
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
