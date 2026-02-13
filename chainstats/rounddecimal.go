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
	"os"
	"reflect"
	"runtime"
	"sync/atomic"
)

const blocksInBatch = 1000

func PriceDiscovery(chain chainreadinterface.IBlockChain,
	handles chainreadinterface.IHandleCreator,
	interestedBlock int64, interestedBlocks int64) (string, error) {

	completedBlocks := int64(0) // Atomic int

	workersDivider := 1
	numWorkers := runtime.NumCPU() / workersDivider
	if numWorkers > 8 {
		numWorkers -= 4 // Save some for OS
	}

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
				} // for block

				littleReport, err := DoMyThing(satsArray, blockBatch)
				if err != nil {
					return err
				}
				local.localReport = local.localReport + littleReport

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

func DoMyThing(satsArray []uint64, blockHeight int64) (string, error) {
	data := satsArray
	report := ""

	for rate := uint64(11); rate <= 19; rate++ {

		fileName := fmt.Sprintf("TestBlock%d_%d.bin", blockHeight, rate)

		fmt.Printf("Initializing engine\n")
		metaDataRootDemo := compositeroots.NewRawScribe()
		dataRootDemo := compositeroots.NewDemographerPairTrainee(1, rate)
		engineDemo := engine2.NewNextGenEngine(
			scribenode.WrapScribeAsBidderScribe(metaDataRootDemo),
			dataRootDemo)

		fmt.Printf("Engine observing\n")
		err := engineDemo.DataNode.Observe(data)
		if err != nil {
			return "", err
		}
		fmt.Printf("Engine improving\n")
		err = engineDemo.DataNode.Improve()
		if err != nil {
			return "", err
		}
		file, err := os.Create(fileName)
		if err != nil {
			return "", err
		}
		fmt.Printf("Engine encoding\n")
		bw := bitstream.NewBitWriter(file)
		refused, err := engineDemo.Encode(data, bw)
		if err != nil {
			return "", err
		}
		if refused {
			return "", errors.New("trained engine refused")
		}
		err = bw.FlushBits()
		if err != nil {
			return "", err
		}
		err = file.Close()
		if err != nil {
			return "", err
		}

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
	}
	return report, nil
}
