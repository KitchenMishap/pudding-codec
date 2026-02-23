package chainstats

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"golang.org/x/sync/errgroup"
	"math"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

// A model that captures the behaviour of humans regarding the decimal mantissa of amounts
type BehaviourModel struct {
	BinCount         uint64
	Bins             []uint64
	VipLounges       []BinVipLounge // For detection of annoying celebrities like 546 sats
	Count            uint64         // Total of counts in bins
	ZerosCount       uint64         // Zeros so made it into neither bins nor Count
	CelebritiesCount uint64         // Celebrities so made int into neither bins nor Count
	PollutedBins     []uint64       // These bins INCLUDE the celebrites, and are in fact USED to identify celebrities
}

type BinVipLounge struct {
	Mutex sync.Mutex        // Only need to lock if you're writing to the Seats map?
	Seats map[uint64]uint64 // Only holds 5 keys (seats) at any time
}

// According to Google Gemini, this is the Misra-Gries algorithm
func (bvl *BinVipLounge) Add(sats uint64) {
	bvl.Mutex.Lock()
	defer bvl.Mutex.Unlock()
	// 1) If already seated, just add a vote
	if _, exists := bvl.Seats[sats]; exists {
		bvl.Seats[sats]++
		return
	}
	// 2) If an empty seat exists, take it
	if len(bvl.Seats) < 5 {
		bvl.Seats[sats] = 1
		return
	}
	// 3) Musical chairs: Decrement everyone
	for s := range bvl.Seats {
		bvl.Seats[s]--
		if bvl.Seats[s] == 0 {
			delete(bvl.Seats, s)
		}
	}
}

func NewBehaviourModel(binCount uint64) *BehaviourModel {
	result := BehaviourModel{}
	result.BinCount = binCount
	result.Bins = make([]uint64, binCount)
	result.PollutedBins = make([]uint64, binCount)
	result.VipLounges = make([]BinVipLounge, binCount)
	for i := range result.VipLounges {
		result.VipLounges[i].Seats = make(map[uint64]uint64)
	}
	return &result
}

func (bm BehaviourModel) SatsToBinNumber(sats uint64) (bin uint64, warningCelebrity bool) {
	if sats == 0 {
		panic("zero sats not supported here")
	}
	bin = bm.SatsToBinNumberRegardless(sats)

	// Is it in that bin's VIP lounge as a candidate for annoying celebrities?
	vipCount, exists := bm.VipLounges[bin].Seats[sats]
	// And is it accounting for more than a third of the bin?
	warningCelebrity = false
	if exists && bm.PollutedBins[bin]/vipCount < 3 {
		warningCelebrity = true
	}

	return bin, warningCelebrity
}

// This function ignores celebrity status
func (bm BehaviourModel) SatsToBinNumberRegardless(sats uint64) (bin uint64) {
	log := math.Log10(float64(sats))
	logInt := math.Floor(log)
	logFrac := log - logInt
	bin = uint64(logFrac * float64(bm.BinCount))
	return bin
}

func (bm *BehaviourModel) GatherData(chain chainreadinterface.IBlockChain,
	handles chainreadinterface.IHandleCreator,
	interestedBlock int64, interestedBlocks int64) error {

	err := bm.gatherData(chain, handles, interestedBlock, interestedBlocks, 1)
	if err != nil {
		return err
	}
	return bm.gatherData(chain, handles, interestedBlock, interestedBlocks, 2)
}

// Call with pass=1 and then pass=2
func (bm *BehaviourModel) gatherData(chain chainreadinterface.IBlockChain,
	handles chainreadinterface.IHandleCreator,
	interestedBlock int64, interestedBlocks int64,
	pass int) error {

	if pass == 1 {
		fmt.Printf("Scanning blockchain for annoying celebrities...\n")
	} else if pass == 2 {
		fmt.Printf("Discovering non-celebrity human behaviour model...\n")
		// Clear the bins but not the VIP lounges (they are still useful)
		bm.Bins = make([]uint64, bm.BinCount)
		bm.ZerosCount = 0
		bm.CelebritiesCount = 0
		bm.Count = 0
	} else {
		panic("incorrect pass number")
	}
	const blocksInBatch = 10000

	completedBlocks := int64(0) // Atomic int for progress

	workersDivider := 1
	numWorkers := runtime.NumCPU() / workersDivider
	if numWorkers > 8 {
		numWorkers -= 4 // Save some for OS
	}

	blockBatchChan := make(chan int64) // Work comes in here (block numbers)

	type workerResult struct {
		bins             []uint64
		count            uint64
		zerosCount       uint64
		celebritiesCount uint64
	}

	resultsChan := make(chan workerResult, numWorkers) // Work come out gere

	// Create an errgroup and a context
	g, ctx := errgroup.WithContext(context.Background())

	for w := 0; w < numWorkers; w++ {
		g.Go(func() error { // Use the errgroup instead of "go func() {"
			local := workerResult{}
			local.bins = make([]uint64, bm.BinCount)

			for blockBatch := range blockBatchChan {
				// Check if another worker already failed
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

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
						trans, err := chain.TransInterface(transHandle)
						if err != nil {
							return err
						}
						txoAmounts, err := trans.AllTxoSatoshis()
						if err != nil {
							return err
						}
						for _, sats := range txoAmounts {

							//---------------------------
							// Core work of this function
							if sats == 0 {
								local.zerosCount++
							} else if pass == 1 {
								// On pass 1 celebrities have not been identified
								bin := bm.SatsToBinNumberRegardless(uint64(sats))
								bm.VipLounges[bin].Add(uint64(sats))
								local.bins[bin]++
								local.count++
							} else if pass == 2 {
								// On pass 2 we IGNORE annoying celebrities
								bin, celebrity := bm.SatsToBinNumber(uint64(sats))
								if celebrity {
									local.celebritiesCount++
								} else {
									local.bins[bin]++
									local.count++
								}
							}
							//---------------------------

						} // for txo amounts
					} // for transactions
				} // for block

				done := atomic.AddInt64(&completedBlocks, blocksInBatch)
				if done%100 == 0 || done == interestedBlocks {
					fmt.Printf("\r\tProgress: %.1f%%    ", float64(100*done)/float64(interestedBlocks))
					runtime.Gosched()
				}

			} // for blockBatch from chan
			resultsChan <- local

			// Free some mem and give the OS a chance to clear up
			// (try to alleviate PC freezes)
			local = workerResult{}
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

	// 1. Create a way to signal when the reduction is done
	reductionDone := make(chan struct{})
	// 2. Start the reduction loop in the BACKGROUND before g.Wait()
	go func() {
		for result := range resultsChan {
			bm.ZerosCount += result.zerosCount
			bm.CelebritiesCount += result.celebritiesCount
			if pass == 1 {
				for i := range result.bins {
					bm.PollutedBins[i] += result.bins[i] // Polluted bins are later used to determine celebrity status
				}
			} else if pass == 2 {
				for i := range result.bins {
					bm.Bins[i] += result.bins[i]
				}
			}
			bm.Count += result.count
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

func (bm *BehaviourModel) Save(prefix string) error {
	// 1. Save Bins as Binary
	fBins, err := os.Create(prefix + "_bins.bin")
	if err != nil {
		return err
	}
	defer fBins.Close()
	err = binary.Write(fBins, binary.LittleEndian, bm.Bins)
	if err != nil {
		return err
	}
	err = binary.Write(fBins, binary.LittleEndian, bm.Count) // Store the total count too!
	if err != nil {
		return err
	}
	fPollutedBins, err := os.Create(prefix + "_pollutedbins.bin")
	if err != nil {
		return err
	}
	defer fPollutedBins.Close()
	err = binary.Write(fPollutedBins, binary.LittleEndian, bm.PollutedBins)
	if err != nil {
		return err
	}

	// 2. Save Celebrities as JSON
	// We only need to save the Seats maps from the VipLounges
	celebData := make([]map[uint64]uint64, len(bm.VipLounges))
	for i, v := range bm.VipLounges {
		celebData[i] = v.Seats
	}
	fCeleb, err := os.Create(prefix + "_celebrities.json")
	if err != nil {
		return err
	}
	defer fCeleb.Close()
	return json.NewEncoder(fCeleb).Encode(celebData)
}

func (bm *BehaviourModel) Load(prefix string) error {
	// 1. Load Bins
	fBins, err := os.Open(prefix + "_bins.bin")
	if err != nil {
		return err
	}
	defer fBins.Close()
	err = binary.Read(fBins, binary.LittleEndian, bm.Bins)
	if err != nil {
		return err
	}
	err = binary.Read(fBins, binary.LittleEndian, &bm.Count)
	if err != nil {
		return err
	}
	fPollutedBins, err := os.Open(prefix + "_pollutedbins.bin")
	if err != nil {
		return err
	}
	defer fPollutedBins.Close()
	err = binary.Read(fPollutedBins, binary.LittleEndian, bm.PollutedBins)
	if err != nil {
		return err
	}

	// 2. Load Celebrities
	fCeleb, err := os.Open(prefix + "_celebrities.json")
	if err != nil {
		return err
	}
	defer fCeleb.Close()
	celebData := []map[uint64]uint64{}
	if err := json.NewDecoder(fCeleb).Decode(&celebData); err != nil {
		return err
	}

	for i, seats := range celebData {
		bm.VipLounges[i].Seats = seats
	}
	return nil
}
