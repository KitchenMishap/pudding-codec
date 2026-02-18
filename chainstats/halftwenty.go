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
		blockHeights    []uint64
		dominantAmounts []uint64
		positions       []int
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

				for blockIdx := blockBatch; blockIdx < blockBatch+blocksInBatch && blockIdx < interestedBlock+interestedBlocks; blockIdx++ {
					satsArray := make([]uint64, 0, 10000)
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
					sort.Slice(round2, func(i, j int) bool {
						return round2[i].Strength > round2[j].Strength
					})
					for position, dominant := range round2 {
						local.positions = append(local.positions, position&7)
						local.blockHeights = append(local.blockHeights, uint64(blockIdx))
						local.dominantAmounts = append(local.dominantAmounts, dominant.Amount)
					}
					//if len(round2) > 0 {
					//	local.blockHeights = append(local.blockHeights, uint64(blockIdx))
					//	local.dominantAmounts = append(local.dominantAmounts, round2[0].Amount)
					//}
				} // for block

				done := atomic.AddInt64(&completedBlocks, blocksInBatch)
				if done%1000 == 0 || done == interestedBlocks {
					fmt.Printf("\r\tProgress: %.1f%%    ", float64(100*done)/float64(interestedBlocks))
					runtime.Gosched()
				}

			} // for blockBatch from chan
			resultsChan <- local
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
			blockHeight := result.blockHeights[index]
			log10Amt := math.Log10(100000000 / float64(dominant))
			log10Frac := log10Amt - math.Floor(log10Amt)
			position := result.positions[index]
			colour3bit := 7 - (position & 7)
			hist.PlotPoint(float64(blockHeight)/888888, log10Frac, colour3bit)
		}
	}
	hist.Output("HalfTwenty.ppm")
	return nil
}
