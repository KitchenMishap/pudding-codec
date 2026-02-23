package chainstats

import (
	"context"
	"fmt"
	"github.com/KitchenMishap/pudding-codec/graphics"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"golang.org/x/sync/errgroup"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

// A model that captures the behaviour of humans regarding the decimal mantissa of amounts
type BehaviourPrice struct {
	mutex sync.Mutex
	Pgm   graphics.PgmHist
}

func NewBehaviourPrice(blockCount uint64) *BehaviourPrice {
	result := BehaviourPrice{}
	return &result
}

func (bp *BehaviourPrice) AnalyzeData(chain chainreadinterface.IBlockChain,
	handles chainreadinterface.IHandleCreator, behaviourModel *BehaviourModel,
	interestedBlock int64, interestedBlocks int64) error {

	fmt.Printf("Discovering price each block...\n")

	const blocksInBatch = 100
	peelColours := [][3]byte{{0, 255, 0}, {0, 0, 255}, {255, 0, 0}}

	completedBlocks := int64(0) // Atomic int for progress

	workersDivider := 1
	numWorkers := runtime.NumCPU() / workersDivider
	if numWorkers > 8 {
		numWorkers -= 4 // Save some for OS
	}

	blockBatchChan := make(chan int64)             // Work comes in here (block numbers)
	resultsChan := make(chan struct{}, numWorkers) // (Empty) work come out gere

	// Create an errgroup and a context
	g, ctx := errgroup.WithContext(context.Background())

	for w := 0; w < numWorkers; w++ {
		g.Go(func() error { // Use the errgroup instead of "go func() {"

			for blockBatch := range blockBatchChan {
				// Check if another worker already failed
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				satsArray := make([]uint64, 0, 10000)
				satsArrayLimited := make([]uint64, 0, 10000)
				satsHistogram := make(map[uint64]int, 1000)
				for blockIdx := blockBatch; blockIdx < blockBatch+blocksInBatch && blockIdx < interestedBlock+interestedBlocks; blockIdx++ {

					satsArray = satsArray[:0]

					blockHandle, err := handles.BlockHandleByHeight(int64(blockIdx))
					if err != nil {
						return err
					}
					block, err := chain.BlockInterface(blockHandle)
					if err != nil {
						return err
					}

					tCount, err := block.TransactionCount()
					if err != nil {
						return err
					}
					for t := int64(0); t < tCount; t++ {
						transHandle, err := block.NthTransaction(t)
						if err != nil {
							return err
						}
						trans, err := chain.TransInterface(transHandle)
						if err != nil {
							return err
						}
						txoAmounts, err := trans.AllTxoSatoshis()
						if err != nil {
							return err
						}
						for _, sats := range txoAmounts {
							if sats == 0 {
								// Throw it away. Messes with logarithms and we're not interested anyway
							} else if IsLessThanThreeDecimalDigits(uint64(sats)) {
								// Throw it away. Very round number in sats. Unlikely to be based on round fiat.
							} else {
								satsArray = append(satsArray, uint64(sats))
							}
						}
					} // for transactions

					//---------------------------
					// Core work of this function

					// First, only allow a particular amount to contribute 5 times per block.
					// This is to filter the effect of blockchain bots that have favourite amounts
					satsArrayLimited = satsArrayLimited[:0]
					// Reset (clear) the map so we can re-use it.
					// Don't want to create a new one as it would thrash the gc and cause a PC Freeze.
					for k := range satsHistogram { // This entire loop is optimized to a single clear in Go 1.11+
						delete(satsHistogram, k)
					}
					for _, sats := range satsArray {
						satsHistogram[sats]++
						if satsHistogram[sats] <= 5 {
							satsArrayLimited = append(satsArrayLimited, sats)
						}
					}

					// We "peel" off each currency as we find it
					numPeels := 3
					//peeledRates := make([]uint64, 0)
					for peel := 0; peel < numPeels; peel++ {
						if len(satsArrayLimited) == 0 {
							break // No more data to peel!
						}
						// I've hit the "Bayesian underflow" wall.
						// Need to use logs of probabilities
						// (In addition to the log10's of amounts and rates that I was already using)
						N := behaviourModel.BinCount
						rateScoresLog := make([]float64, N)
						// For every possible exchange rate
						for log10RateBinNumber, _ := range behaviourModel.Bins {
							probRateScoreLog := float64(0) // Log(Prob) = 0 representing Prob = 1
							// For each sats amount in the block
							for _, sats := range satsArrayLimited {
								log10Sats, celebrity := behaviourModel.SatsToBinNumber(sats)
								if celebrity {
									// Celebrities distort everything, really not interested!
								} else {
									log10Fiat := (log10Sats + uint64(log10RateBinNumber)) % N // Multiply is add for logs
									// Now we are in fiat, the human behaviour round number probability model holds
									prob := float64(behaviourModel.Bins[log10Fiat]) / float64(behaviourModel.Count)
									probLog := math.Log(prob)
									probRateScoreLog += probLog
								}
							} // for sats
							probRateScoreLog += math.Log(1 / float64(N)) // This is the (flat) P(rate)=1/N
							rateScoresLog[log10RateBinNumber] = probRateScoreLog
						} // for rate

						// Find the WINNER of this peel
						winnerBin := uint64(0)
						maxVal := rateScoresLog[0]
						for i, val := range rateScoresLog {
							if val > maxVal {
								maxVal = val
								winnerBin = uint64(i)
							}
						}

						// rateScoresLog[] is now a bunch of logs of tiny probabilities
						// We need the total in non-log space, but they're too tiny to add.
						// We find the max M of the logs, subtract M from all logs, exp, sum, then log and add M
						// 1) Find the maximum log score
						maxLog := rateScoresLog[0]
						for _, s := range rateScoresLog {
							if s > maxLog {
								maxLog = s
							}
						}
						// 2) Calculate the Log-Sum-Exp
						sumExp := float64(0)
						for _, s := range rateScoresLog {
							sumExp += math.Exp(s - maxLog)
						}
						sumRateScoresLog := maxLog + math.Log(sumExp)

						// Plot in graphics
						startPrintBlock := int64(888888 - graphics.Width)
						x := float64(blockIdx-startPrintBlock) / graphics.Width
						if x > 0 && x < 1 {
							bp.mutex.Lock()
							// FIRST we "paint" ALL the probabilities as grey
							if peel == 0 {
								numEvidenceItems := len(satsArrayLimited)
								for i := range N {
									// probLog is the natural log of probability. So one is currently at 0.
									probLog := rateScoresLog[i] - sumRateScoresLog

									// Need about 5 amounts for full brightness
									confidence := math.Min(1.0, float64(numEvidenceItems)/5)

									intensity := (255 + (probLog * 20)) * confidence
									if intensity < 0 {
										intensity = 0
									}
									if intensity > 255 {
										intensity = 255
									}
									b := byte(intensity)
									y := float64(i) / float64(N)
									bp.Pgm.SetPoint(x, y, b, b, b)
								} // for i = rate bins
							} // if peel 0

							// SECOND, we plot a single point for the winner of this peel
							y := float64(winnerBin) / float64(N)
							rd := peelColours[peel][0]
							gn := peelColours[peel][1]
							bl := peelColours[peel][2]
							bp.Pgm.SetPoint(x, y, rd, gn, bl)
							bp.mutex.Unlock()
						} // if x between 0 and 1

						// NOW that we've plotted that found fiat currency rate, evict the amounts
						// that supported it

						// Calculate the "noise floor" - what a random bin would look like
						noiseFloor := float64(behaviourModel.Count) / float64(N)

						// Define the peel sensitivity (2 means anything twice as likely as noise)
						sensitivity := 2.0

						remainingSats := make([]uint64, 0, len(satsArrayLimited))
						for _, sats := range satsArrayLimited {
							log10Sats, _ := behaviourModel.SatsToBinNumber(sats)
							winningFiatBin := (log10Sats + winnerBin) % N
							// If this sats amount looks "very human" at the winning rate,
							// it's likely part of that currency's signal.
							// We "peel" it by not adding it to the next pass
							if float64(behaviourModel.Bins[winningFiatBin]) > noiseFloor*sensitivity {
								// This amount was "signal" for the rate just found. Peel it! (don't add it)
							} else {
								// This amount was "noise" or belongs to a different currency to be found in
								// a subsequent peel. Keep it!
								remainingSats = append(remainingSats, sats)
							}
						} // for sats
						satsArrayLimited = remainingSats
					} //  for peel
					//---------------------------
				} // for block

				done := atomic.AddInt64(&completedBlocks, blocksInBatch)
				if done%1000 == 0 || done == interestedBlocks {
					fmt.Printf("\r\tProgress: %.1f%%    ", float64(100*done)/float64(interestedBlocks))
					runtime.Gosched()
				}

			} // for blockBatch from chan
			resultsChan <- struct{}{}
			runtime.Gosched()
			return nil
		}) // gofunc
	} // for workers
	go func() {
		defer close(blockBatchChan)
		for blockBatch := interestedBlock; blockBatch < interestedBlock+interestedBlocks; blockBatch += blocksInBatch {
			select { // Note: NOT a switch statement!
			case blockBatchChan <- blockBatch: // This happens if a worker is free to be fed a block number
			case <-ctx.Done(): // This happens if a worker returned an err
				return
			}
		}
	}()

	reductionDone := make(chan struct{})
	go func() {
		for range resultsChan {
		}
		close(reductionDone)
	}()

	// 3. Now Wait for the workers to finish
	err := g.Wait()
	if err != nil {
		return err
	}

	// 4. Close the results channel so the reduction loop knows to finish
	close(resultsChan)

	// 5. Wait for the reduction loop to actually finish
	<-reductionDone

	fmt.Printf("\nDone that now\n")
	return nil
}
