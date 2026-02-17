package chainstats

import (
	"context"
	"errors"
	"fmt"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/compositeroots"
	engine2 "github.com/KitchenMishap/pudding-codec/engine"
	"github.com/KitchenMishap/pudding-codec/scribenode"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"golang.org/x/sync/errgroup"
	"math"
	"runtime"
	"sync/atomic"
)

const blocksInBatch = 100

func PriceDiscovery(chain chainreadinterface.IBlockChain,
	handles chainreadinterface.IHandleCreator,
	interestedBlock int64, interestedBlocks int64) (string, error) {

	completedBlocks := int64(0) // Atomic int

	workersDivider := 1
	numWorkers := runtime.NumCPU() / workersDivider
	if numWorkers > 8 {
		numWorkers -= 4 // Save some for OS
	}
	numWorkers = 1

	blockBatchChan := make(chan int64)

	type workerResult struct {
		localReport string
	}
	resultsChan := make(chan workerResult, numWorkers)

	// Create an errgroup and a context
	g, ctx := errgroup.WithContext(context.Background())

	for w := 0; w < numWorkers; w++ {
		g.Go(func() error { // Use the errgroup instead of "go func() {"
			local := workerResult{}

			for blockBatch := range blockBatchChan {
				// Check if another worker already failed
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				satsArray := make([]uint64, 0, 10000)
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
					if len(satsArray) >= 50000 {
						copyForWorkers := make([]uint64, len(satsArray))
						copy(copyForWorkers, satsArray)
						littleReport, err := DoMyThing(copyForWorkers, blockIdx)
						if err != nil {
							return err
						}
						local.localReport = local.localReport + littleReport
						satsArray = satsArray[:0]
					}
				} // for block

				if len(satsArray) > 0 {
					// Flush any remaining sats so they don't pollute the next batch
					copyForWorkers := make([]uint64, len(satsArray))
					copy(copyForWorkers, satsArray)
					littleReport, err := DoMyThing(copyForWorkers, blockBatch+blocksInBatch-1)
					if err != nil {
						return err
					}
					local.localReport = local.localReport + littleReport
					satsArray = satsArray[:0]
				}

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
		return "", err
	}
	close(resultsChan)
	fmt.Printf("\nDone that now\n")

	// Final Reduction
	report := "This is the start of the report!\n"
	for result := range resultsChan {
		report += result.localReport
	}

	return report, nil
}

func DoMyThingSub(satsArray []uint64, log10rate float64) (uint64, error) {
	data := satsArray

	mantissa := math.Pow(10, log10rate)
	baseRate := 10000
	rate := int(float64(baseRate) * mantissa)

	//fmt.Printf("Initializing engine\n")
	metaDataRootDemo := compositeroots.NewRawScribe()
	dataRootDemo := compositeroots.NewDemographerPairTrainee(uint64(baseRate), uint64(rate))
	engineDemo := engine2.NewNextGenEngine(
		scribenode.WrapScribeAsBidderScribe(metaDataRootDemo),
		dataRootDemo)

	//fmt.Printf("Engine observing\n")
	err := engineDemo.DataNode.Observe(data)
	if err != nil {
		return 0, err
	}
	//fmt.Printf("Engine improving\n")
	err = engineDemo.DataNode.Improve()
	if err != nil {
		return 0, err
	}
	//fmt.Printf("Engine encoding\n")
	bw := bitstream.NewBitCounter()
	refused, err := engineDemo.Encode(data, bw)
	if err != nil {
		return 0, err
	}
	if refused {
		return 0, errors.New("trained engine refused")
	}
	err = bw.FlushBits()
	if err != nil {
		return 0, err
	}
	bitCount := bw.CountBits()

	/*
		file2, err := os.Open(fileName)
		if err != nil {
			return "", err
		}
		br := bitstream.NewBitReader(file2)
		metaDataRootDemo2 := compositeroots.NewRawScribe()
		dataRootDemo2 := compositeroots.NewDemographerPairTrainee(1, 2)
		engineDemo2 := engine2.NewNextGenEngine(
			scribenode.WrapScribeAsBidderScribe(metaDataRootDemo2),
			dataRootDemo2)

		fmt.Printf("Engine decoding\n")
		dataOut, err := engineDemo2.Decode(br)
		if err != nil {
			return "", err
		}
		err = file2.Close()
		if err != nil {
			return "", err
		}

		if !reflect.DeepEqual(data, dataOut) {
			return "", errors.New("encode error, expected")
		}
	*/
	return uint64(bitCount), nil
}

func DoMyThing(satsArray []uint64, blockHeight int64) (string, error) {
	fmt.Printf("Starting on Block %d...\n", blockHeight)
	startLog10 := 0.0
	endLog10 := 1.0
	const steps = 75
	index, err := Refine(satsArray, startLog10, endLog10, steps)
	if err != nil {
		return "", err
	}

	log10 := StepValExclusive(startLog10, endLog10, steps, index)
	fmt.Printf("blockHeight %d, first estimate: $%d\n", blockHeight, int(10000*math.Pow(10, log10)))

	startLog10b := StepValExclusive(startLog10, endLog10, steps, index-1)
	endLog10b := StepValExclusive(startLog10, endLog10, steps, index+1)
	index, err = Refine(satsArray, startLog10b, endLog10b, steps)
	if err != nil {
		return "", err
	}

	log10 = StepValExclusive(startLog10b, endLog10b, steps, index)
	fmt.Printf("blockHeight %d, second estimate: $%d\n", blockHeight, int(10000*math.Pow(10, log10)))

	return fmt.Sprintf("%d,\t%d\n", blockHeight, int(10000*math.Pow(10, log10))), nil
}

// There are a number of steps distributed between left and right, but not
// including left and right
func StepValExclusive(left float64, right float64, steps int, index int) float64 {
	stepSize := (right - left) / float64(steps)
	return left + (0.5+float64(index))*stepSize
}

func Refine(satsArray []uint64, firstLog10 float64, lastLog10 float64, steps int) (int, error) {
	// We use the context from the caller if possible, or a new one
	g, ctx := errgroup.WithContext(context.Background())

	// Use a thread-safe way to track the winner
	// Or just an array we process at the end (cleaner)
	scores := make([]uint64, steps)

	// Instead of a channel, we just launch the steps.
	// Go's scheduler will handle the 36-core distribution.
	for i := 0; i < steps; i++ {
		index := i // Essential: capture the loop variable
		g.Go(func() error {
			// Respect cancellation if a sibling worker fails
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			log10rate := StepValExclusive(firstLog10, lastLog10, steps, index)

			bitCount, err := DoMyThingSub(satsArray, log10rate)
			if err != nil {
				return err
			}

			scores[index] = bitCount
			return nil
		})
	}

	// This is where it was sticking.
	// By launching all Gs upfront, we remove the "Sender" deadlock.
	if err := g.Wait(); err != nil {
		return 0, err
	}

	// Post-processing: Find the winner
	winningScore := uint64(math.MaxUint64)
	winningIndex := -1
	for i, s := range scores {
		if s < winningScore {
			winningScore = s
			winningIndex = i
		}
	}

	return winningIndex, nil
}
