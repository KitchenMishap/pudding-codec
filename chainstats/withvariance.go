package chainstats

import (
	"context"
	"fmt"
	"github.com/KitchenMishap/pudding-codec/graphics"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/robitx/variance"
	"golang.org/x/sync/errgroup"
	"math"
	"runtime"
	"sync/atomic"
)

// A model that captures the behaviour of humans regarding the decimal mantissa of amounts
// This time, determine the most likely change amount in each transaction
// And also, take into account the variance of the amounts
func (bp *BehaviourPrice) AnalyzeDataWithVariance(chain chainreadinterface.IBlockChain,
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

				satsArray := make([]uint64, 0, 10000)             // The amounts
				satsArrayTransIndices := make([]uint64, 0, 10000) // Indices into satsArray for each transaction
				satsArrayNoChange := make([]uint64, 0, 10000)
				logMantissaStats := make([]*variance.Stats, behaviourModel.BinCount)
				sumWeights := make([]float64, behaviourModel.BinCount)

				for blockIdx := blockBatch; blockIdx < blockBatch+blocksInBatch && blockIdx < interestedBlock+interestedBlocks; blockIdx++ {

					satsArray = satsArray[:0]
					satsArrayTransIndices = satsArrayTransIndices[:0]
					satsArrayNoChange = satsArrayNoChange[:0]
					for i := range behaviourModel.BinCount {
						logMantissaStats[i] = variance.New()
						sumWeights[i] = 0.0
					}

					blockHandle, err := handles.BlockHandleByHeight(blockIdx)
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
						satsArrayTransIndices = append(satsArrayTransIndices, uint64(len(satsArray)))
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
					// One past the end
					satsArrayTransIndices = append(satsArrayTransIndices, uint64(len(satsArray)))

					//---------------------------
					// Core work of this function

					// We "peel" off each currency as we find it
					numPeels := 1
					//peeledRates := make([]uint64, 0)
					for peel := 0; peel < numPeels; peel++ {
						if len(satsArray) == 0 {
							break // No more data to peel!
						}
						// I've hit the "Bayesian underflow" wall.
						// Need to use logs of probabilities
						// (In addition to the log10's of amounts and rates that I was already using)
						N := behaviourModel.BinCount
						logOneOverN := math.Log(1 / float64(N))
						rateScoresLog := make([]float64, N)
						// For every possible exchange rate
						for log10RateBinNumber, _ := range behaviourModel.Bins {
							probRateScoreLog := float64(0) // Log(Prob) = 0 representing Prob = 1
							// For each transaction in the block
							for trans, satsIndex := range satsArrayTransIndices {
								if trans == len(satsArrayTransIndices)-1 {
									continue // Skip the last entry as it's in the next block
								}
								nextTransIndex := satsArrayTransIndices[trans+1]
								amounts := satsArray[satsIndex:nextTransIndex]
								if len(amounts) < 2 {
									continue // If zero or one amounts, not a "proper" transaction that we're interested in
								}
								// For amounts in the transaction
								// First loop... The product of all probabilities in the transaction
								probProductLog := float64(0) // Log(Prob) = 0 representing Prob = 1
								for _, sats := range amounts {
									log10Sats, _, celebrity := behaviourModel.SatsToBinNumber(sats)
									if celebrity {
										// Celebrities distort everything, really not interested!
									} else {
										log10Fiat := (log10Sats + uint64(log10RateBinNumber)) % N // Multiply is add for logs
										// Now we are in fiat, the human behaviour round number probability model holds
										prob := float64(behaviourModel.Bins[log10Fiat]) / float64(behaviourModel.Count)
										probLog := math.Log(prob)
										probProductLog += probLog
									}
								}
								// A second loop... Now we can estimate which amount of the transaction is the change
								smallestProbLogNotChange := float64(0) // Which is the probability of the BEST candidate FOR change
								bestCandidateForChange := 0
								for satsIndex1, sats := range amounts {
									log10Sats, _, celebrity := behaviourModel.SatsToBinNumber(sats)
									if celebrity {
										// Celebrities distort everything, really not interested!
									} else {
										log10Fiat := (log10Sats + uint64(log10RateBinNumber)) % N // Multiply is add for logs
										// Now we are in fiat, the human behaviour round number probability model holds
										prob := float64(behaviourModel.Bins[log10Fiat]) / float64(behaviourModel.Count)
										probLog := math.Log(prob)
										// ToDo Are you sure?
										// The probability that this is NOT change is the probability of all the others occurring.
										// That is the product of all of them divided by this one. (A subtraction as log)
										probLogNotChange := probProductLog - probLog
										if probLogNotChange < smallestProbLogNotChange {
											smallestProbLogNotChange = probLog
											bestCandidateForChange = satsIndex1
										}
									}
								}
								// A third loop... in which we exclude the change item
								for satsIndex2, sats2 := range amounts {
									log10Sats, logFrac, celebrity2 := behaviourModel.SatsToBinNumber(sats2)
									if celebrity2 {
										// Celebrities distort everything, really not interested!
									} else if satsIndex2 != bestCandidateForChange {
										log10Fiat := (log10Sats + uint64(log10RateBinNumber)) % N // Multiply is add for logs
										addToLogFrac := float64(log10RateBinNumber) / float64(N)  // Add a number between 0.0 and 1.0
										logFracFiat := logFrac + addToLogFrac
										if logFracFiat >= 1.0 {
											logFracFiat -= 1.0
										}
										// Now we are in fiat, the human behaviour round number probability model holds
										prob := float64(behaviourModel.Bins[log10Fiat]) / float64(behaviourModel.Count)
										satsArrayNoChange = append(satsArrayNoChange, sats2)
										logMantissaStats[log10Fiat].AddWeighted(logFracFiat, prob)
										sumWeights[log10Fiat] += prob
									}
								}
								probRateScoreLog += probProductLog - smallestProbLogNotChange // Divide by the one that was change
							} // for transaction

							// Now we are in a position to take account of the variances in each bin
							// Go through each of the bins
							ProbRateScoreLog := float64(0)
							for bin := range N {
								// weighted variance of the mantissas of the amounts in this bin
								weightedVariance := logMantissaStats[bin].Variance()
								// Sum of weights that contributed
								sumOfWeights := sumWeights[bin]
								// human behaviour variance for this bin as captured across whole blockchain
								humanVariance := behaviourModel.Stats[bin].Variance()
								// Noise variance based on uniform density
								noiseVariance := float64(1) / float64(12) / float64(N) / float64(N)
								// The Log Likelihood Ratio
								LLR := LogLikelihoodRatio(weightedVariance, humanVariance, noiseVariance, sumOfWeights-1)
								// What the behaviour model says is the probability of finding an amount in this bin for this rate
								prob := float64(behaviourModel.Bins[bin]) / float64(behaviourModel.Count)
								// The combined log probability P(data|Rate) taking into account variance
								probLogData := math.Log(prob) + LLR
								// Amalgamate to get towards P(rate)
								ProbRateScoreLog += probLogData
							}

							probRateScoreLog += logOneOverN // This is the (flat) P(rate)=1/N
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
						//sumRateScoresLog := maxLog + math.Log(sumExp)

						// Plot in graphics
						startPrintBlock := int64(888888 - graphics.Width)
						x := float64(blockIdx-startPrintBlock) / graphics.Width
						//x := float64(blockIdx) / 888888
						if x > 0 && x < 1 {
							bp.mutex.Lock()
							// FIRST we "paint" ALL the probabilities as grey
							if peel == 0 {
								/*
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
										//b := byte(255 - intensity)
										//y := float64(i) / float64(N)
										//bp.Pgm.SetPoint(x, y, b, b, b)
									} // for i = rate bins
								*/
							} // if peel 0

							// SECOND, we plot a single point for the winner of this peel
							y := float64(winnerBin) / float64(N)
							rd := peelColours[peel][0]
							gn := peelColours[peel][1]
							bl := peelColours[peel][2]
							bp.Pgm.SetPoint(x, y, rd, gn, bl)
							bp.mutex.Unlock()
						} // if x between 0 and 1

						/*
							// NOW that we've plotted that found fiat currency rate, evict the amounts
							// that supported it

							// Calculate the "noise floor" - what a random bin would look like
							noiseFloor := float64(behaviourModel.Count) / float64(N)

							remainingSats = remainingSats[:0]
							for _, sats := range satsArrayLimited {
								log10Sats, _ := behaviourModel.SatsToBinNumber(sats)
								winningFiatBin := (log10Sats + winnerBin) % N

								binValue := float64(behaviourModel.Bins[winningFiatBin])

								// Aggressive peel:
								if binValue > noiseFloor {
									// This amount was "signal" for the rate just found. Peel it! (don't add it)
								} else {
									// This amount was "noise" or belongs to a different currency to be found in
									// a subsequent peel. Keep it!
									remainingSats = append(remainingSats, sats)
								}
							} // for sats
							satsArrayLimited = remainingSats
						*/
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

func Lvar(weightedVariance float64, modelVariance float64, v float64) float64 {
	topLeft := math.Pow(modelVariance*v/2, v/2)
	bottomLeft := math.Gamma(v / 2)
	topRight := math.Exp(-v * modelVariance / (2 * weightedVariance))
	bottomRight := math.Pow(weightedVariance, (1 + v/2))
	return topLeft * topRight / bottomLeft / bottomRight
}

func LogLikelihoodRatio(weightedVariance float64, humanVariance float64, noiseVariance float64, v float64) float64 {
	return math.Log(Lvar(weightedVariance, humanVariance, v)) - math.Log(Lvar(weightedVariance, noiseVariance, v))
}
