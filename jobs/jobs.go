package jobs

import (
	"fmt"
	"github.com/KitchenMishap/pudding-codec/blockchain"
	"github.com/KitchenMishap/pudding-codec/chainstats"
)

func PriceDiscovery500000(folder string) error {
	chain, err := blockchain.NewChainReader(folder)
	if err != nil {
		return err
	}

	report, err := chainstats.PriceDiscovery(chain.Blockchain(), chain.HandleCreator(), 0, 888888)
	if err != nil {
		return err
	}
	fmt.Printf("report: %s\n", report)

	fmt.Printf("Done price discovery\n")
	return nil
}

func PriceDiscoveryHalfTwenty(folder string) error {
	chain, err := blockchain.NewChainReader(folder)
	if err != nil {
		return err
	}

	err = chainstats.PriceDiscoveryHalfTwenty(chain.Blockchain(), chain.HandleCreator(), 000_000, 888_888)
	if err != nil {
		return err
	}

	fmt.Printf("Done price discovery\n")
	return nil
}

func PriceDiscoveryHumanBehaviour(folder string) error {
	chain, err := blockchain.NewChainReader(folder)
	if err != nil {
		return err
	}

	binCount := 1024
	filenamePrefix := fmt.Sprintf("model_%d", binCount)
	bm := chainstats.NewBehaviourModel(uint64(binCount))
	err = bm.Load(filenamePrefix)
	if err != nil {
		fmt.Printf("Model files not found (%s), Running full blockchain scan...\n", err.Error())
		err := bm.GatherData(chain.Blockchain(), chain.HandleCreator(), 0, 888_888)
		if err != nil {
			return err
		}
		fmt.Printf("Saving new model files...\n")
		err = bm.Save(filenamePrefix)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Loaded model files from disk, skipping scan.\n")
	}

	bp := chainstats.NewBehaviourPrice(888_888)
	err = bp.AnalyzeData(chain.Blockchain(), chain.HandleCreator(), bm, 880_000, 8_888)
	if err != nil {
		return err
	}
	bp.Pgm.Output("Probabilities.ppm")

	return nil
}
