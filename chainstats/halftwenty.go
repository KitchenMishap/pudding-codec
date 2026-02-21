package chainstats

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-codec/graphics"
	"github.com/KitchenMishap/pudding-codec/halfonetwo"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"golang.org/x/sync/errgroup"
	"math"
	"runtime"
	"slices"
	"sync/atomic"
	"time"
)

const blocksInBatch = 1000

func PriceDiscoveryHalfTwenty(chain chainreadinterface.IBlockChain,
	handles chainreadinterface.IHandleCreator,
	interestedBlock int64, interestedBlocks int64) error {

	hist := graphics.PgmHist{}

	completedBlocks := int64(0) // Atomic int

	workersDivider := 1
	numWorkers := runtime.NumCPU() / workersDivider
	if numWorkers > 8 {
		numWorkers -= 4 // Save some for OS
	}

	blockBatchChan := make(chan int64)

	type workerResult struct {
		blockHeightStart []uint64
		blockHeightEnd   []uint64
		blockHourCol     []int
		dominantAmounts  []uint64
		positions        []int
	}
	resultsChan := make(chan workerResult, numWorkers)

	// Create an errgroup and a context
	g, ctx := errgroup.WithContext(context.Background())

	fmt.Printf("numWorkers: %d\n", numWorkers)
	for w := 0; w < numWorkers; w++ {
		fmt.Printf("Starting worker %d\n", w)
		g.Go(func() error { // Use the errgroup instead of "go func() {"
			local := workerResult{}
			met := halfonetwo.NewMantissaExponentTallies(12)
			logBase := 5 * math.Pi // 5 Pi as a log base is bigger than 10 and transcendental (no repeats)
			logY := halfonetwo.NewLogYFracHist(10000, logBase)

			for blockBatch := range blockBatchChan {
				// Check if another worker already failed
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				satsArray := make([]uint64, 0, 10000)
				dominantAmounts := make([]halfonetwo.DominantAmount, 1000)
				firstBlock := blockBatch
				//currentBlock := blockBatch

				for blockIdx := blockBatch; blockIdx < blockBatch+blocksInBatch && blockIdx < interestedBlock+interestedBlocks; blockIdx++ {
					//currentBlock = blockIdx
					blockHandle, err := handles.BlockHandleByHeight(int64(blockIdx))
					if err != nil {
						return err
					}
					block, err := chain.BlockInterface(blockHandle)
					if err != nil {
						return err
					}

					nei, err := block.NonEssentialInts()
					if err != nil {
						continue // Bodge ?!
					}
					blockTime, ok := (*nei)["time"]
					if !ok {
						panic("could not get block time")
					}
					bTime := time.Unix(blockTime, 0)
					hourCol := bTime.Hour()/4 + 1 // A number between 1 and 6

					tCount, err := block.TransactionCount()
					if err != nil {
						return err
					}
					for t := int64(0); t < tCount; t++ {
						transHandle, err := block.NthTransaction(t)
						if err != nil {
							return err
						}
						if !transHandle.HeightSpecified() {
							return errors.New("transaction height not specified")
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
							if IsLessThanThreeDecimalDigits(uint64(sats)) {
								continue // Round number of sats, not interested
							}
							satsArray = append(satsArray, uint64(sats))
						} // for txo amounts
					} // for transactions

					// Here is the main work!
					met.Wipe()
					met.Populate(satsArray)
					logY.Wipe()
					logY.Populate(satsArray)
					logY.FindPeaks()
					//roundA := met.AnalyzeHalfOneTwo()
					//roundB := met.AnalyzeEighthQuarterHalf()
					//roundC := halfonetwo.Filter50_100_125_200_250_500(roundA, roundB)
					dominantAmounts = dominantAmounts[:0]
					for _, amount := range satsArray {
						logY.AssessPrimePeaksStrength(amount, false)
					}
					for _, amount := range satsArray {
						strength := logY.AssessPrimePeaksStrength(amount, true)
						if strength > 0 {
							dominantAmounts = append(dominantAmounts, halfonetwo.DominantAmount{amount, strength})
						}
					}
					// Reverse sort by strength
					slices.SortFunc(dominantAmounts, func(a, b halfonetwo.DominantAmount) int {
						return -cmp.Compare(a.Strength, b.Strength) // -ve to reverse sort
					})
					// Hit parade
					for _, dominant := range dominantAmounts[0:50] {
						local.positions = append(local.positions, 0)
						local.blockHeightStart = append(local.blockHeightStart, uint64(firstBlock))
						local.blockHeightEnd = append(local.blockHeightEnd, uint64(blockIdx))
						local.blockHourCol = append(local.blockHourCol, hourCol)
						local.dominantAmounts = append(local.dominantAmounts, dominant.Amount)
					}
					// Reset for next group of blocks
					satsArray = satsArray[:0]
					firstBlock = blockIdx + 1
					/*
						round2 := halfonetwo.FilterHalfOneTwoFiveTenTwenty(round1)
						if len(round2) >= 10 {
							// This is the fussy, filtered method for data rich blocks (half-one-two-ten-twenty)
							sort.Slice(round2, func(i, j int) bool {
								return round2[i].Strength > round2[j].Strength
							})
							for position, dominant := range round2 {
								local.positions = append(local.positions, position&7)
								local.blockHeightStart = append(local.blockHeightStart, uint64(firstBlock))
								local.blockHeightEnd = append(local.blockHeightEnd, uint64(blockIdx))
								local.blockHourCol = append(local.blockHourCol, hourCol)
								local.dominantAmounts = append(local.dominantAmounts, dominant.Amount)
							}
							// Reset for next group of blocks
							satsArray = satsArray[:0]
							firstBlock = blockIdx + 1
						} else if len(round1) >= 5 {
							// This is the less fussy method for data-poor blocks (half-one-two)
							sort.Slice(round1, func(i, j int) bool {
								return round1[i].Strength > round1[j].Strength
							})
							for position, dominant := range round1 {
								local.positions = append(local.positions, position&7)
								local.blockHeightStart = append(local.blockHeightStart, uint64(firstBlock))
								local.blockHeightEnd = append(local.blockHeightEnd, uint64(blockIdx))
								local.blockHourCol = append(local.blockHourCol, hourCol)
								local.dominantAmounts = append(local.dominantAmounts, dominant.Amount)
							}
							// Reset for next group of blocks
							satsArray = satsArray[:0]
							firstBlock = blockIdx + 1
						} else {
							// Let it roll, more blocks for more data until we get enough hits
							// Where we wait for multiple blocks like this, the data point will be rendered wider
						}*/
				} // for block
				/*
					// If we have unprocessed sats at the end of a batch we will have to process
					// them here rather than erroneously carrying them over to the next (unrelated) batch
					// This is the less fussy method for data-poor blocks (half-one-two)
					met.Wipe()
					met.Populate(satsArray)
					round1b := met.AnalyzeHalfOneTwo()
					sort.Slice(round1b, func(i, j int) bool {
						return round1b[i].Strength > round1b[j].Strength
					})
					for position, dominant := range round1b {
						local.positions = append(local.positions, position&7)
						local.blockHeightStart = append(local.blockHeightStart, uint64(firstBlock))
						local.blockHeightEnd = append(local.blockHeightEnd, uint64(currentBlock))
						local.blockHourCol = append(local.blockHourCol, 6)
						local.dominantAmounts = append(local.dominantAmounts, dominant.Amount)
					}
					// Reset for next group of blocks
					satsArray = satsArray[:0]
					firstBlock = currentBlock + 1
				*/
				done := atomic.AddInt64(&completedBlocks, blocksInBatch)
				if done%1000 == 0 || done == interestedBlocks {
					fmt.Printf("\r\tProgress: %.1f%%    ", float64(100*done)/float64(interestedBlocks))
					runtime.Gosched()
				}

			} // for blockBatch from chan
			resultsChan <- local

			fmt.Printf("Worker done and cleaning up\n")
			// Free some mem and give the OS a chance to clear up
			// (try to alleviate PC freezes)
			met = nil
			logY = nil
			local = workerResult{}
			runtime.Gosched()

			fmt.Printf("Worker cleaned up and returning")
			return nil
		}) // gofunc
	} // for workers
	go func() {
		defer close(blockBatchChan)
		for blockBatch := interestedBlock; blockBatch < interestedBlock+interestedBlocks; blockBatch += blocksInBatch {
			select { // Note: NOT a switch statement!
			case blockBatchChan <- blockBatch: // This happens if a worker is free to be fed an epoch ID
			case <-ctx.Done(): // This happens if a worker returned an err
				return
			}
		}
	}()

	// ... after the goroutine that feeds blockBatchChan ...

	// 1. Create a way to signal when the reduction is done
	reductionDone := make(chan struct{})

	// 2. Start the reduction loop in the BACKGROUND before g.Wait()
	go func() {
		worker := 0
		for result := range resultsChan {
			fmt.Printf("Plotting data from a finished worker, %d of %d\n", worker, numWorkers)
			for index, dominant := range result.dominantAmounts {
				blockHeightStart := result.blockHeightStart[index]
				blockHeightEnd := result.blockHeightEnd[index]
				log10Amt := math.Log10(100000000 / float64(dominant))
				log10Frac := log10Amt - math.Floor(log10Amt)
				colour3bit := result.blockHourCol[index]
				hist.PlotWidePoint(float64(blockHeightStart-uint64(interestedBlock))/float64(interestedBlocks), float64(blockHeightEnd-uint64(interestedBlock))/float64(interestedBlocks), log10Frac, colour3bit)
			}
			fmt.Printf("Finished plotting data from a finished worker\n")
			worker++
		}
		close(reductionDone)
	}()

	// 3. Now Wait for the workers to finish
	err := g.Wait()

	// 4. Close the results channel so the reduction loop knows to finish
	close(resultsChan)

	// 5. Wait for the reduction loop to actually finish plotting
	<-reductionDone

	if err != nil {
		return err
	}
	fmt.Printf("\nDone that now\n")
	hist.Output("HalfTwenty.ppm")
	hist.NormalizeColumns()
	hist.Output("HalfTwentyNormalized.ppm")
	return nil
}

func IsLessThanThreeDecimalDigits(val uint64) bool {
	if val == 0 {
		return true
	}
	for val%10 == 0 {
		val /= 10
	}
	return val <= 99
}
