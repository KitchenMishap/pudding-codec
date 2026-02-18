package chainstats

import (
	"context"
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-codec/graphics"
	"github.com/KitchenMishap/pudding-codec/halfonetwo"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"golang.org/x/sync/errgroup"
	"math"
	"runtime"
	"sort"
	"sync/atomic"
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
		dominantAmounts  []uint64
		positions        []int
	}
	resultsChan := make(chan workerResult, numWorkers)

	// Create an errgroup and a context
	g, ctx := errgroup.WithContext(context.Background())

	for w := 0; w < numWorkers; w++ {
		g.Go(func() error { // Use the errgroup instead of "go func() {"
			local := workerResult{}
			met := halfonetwo.NewMantissaExponentTallies(16)

			for blockBatch := range blockBatchChan {
				// Check if another worker already failed
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				satsArray := make([]uint64, 0, 10000)
				firstBlock := blockBatch

				for blockIdx := blockBatch; blockIdx < blockBatch+blocksInBatch && blockIdx < interestedBlock+interestedBlocks; blockIdx++ {
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
							satsArray = append(satsArray, uint64(sats))
						} // for txo amounts
					} // for transactions

					// Here is the main work!
					met.Wipe()
					met.Populate(satsArray)
					round1 := met.AnalyzeHalfOneTwo()
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
							local.dominantAmounts = append(local.dominantAmounts, dominant.Amount)
						}
						// Reset for next group of blocks
						satsArray = satsArray[:0]
						firstBlock = blockIdx + 1
					} else {
						// Let it roll, more blocks for more data until we get enough hits
						// Where we wait for multiple blocks like this, the data point will be rendered wider
					}
				} // for block

				done := atomic.AddInt64(&completedBlocks, blocksInBatch)
				if done%1000 == 0 || done == interestedBlocks {
					fmt.Printf("\r\tProgress: %.1f%%    ", float64(100*done)/float64(interestedBlocks))
					runtime.Gosched()
				}

			} // for blockBatch from chan
			resultsChan <- local

			// Free some mem and give the OS a chanc to clear up
			// (try to alleviate PC freezes)
			met = nil
			local = workerResult{}
			runtime.Gosched()

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

	err := g.Wait()
	if err != nil {
		return err
	}
	close(resultsChan)
	fmt.Printf("\nDone that now\n")

	// Final Reduction
	for result := range resultsChan {
		for index, dominant := range result.dominantAmounts {
			blockHeightStart := result.blockHeightStart[index]
			blockHeightEnd := result.blockHeightEnd[index]
			log10Amt := math.Log10(100000000 / float64(dominant))
			log10Frac := log10Amt - math.Floor(log10Amt)
			position := result.positions[index]
			colour3bit := 7 - (position & 7)
			hist.PlotWidePoint(float64(blockHeightStart)/888888, float64(blockHeightEnd)/888888, log10Frac, colour3bit)
		}
	}
	hist.Output("HalfTwenty.ppm")
	return nil
}
