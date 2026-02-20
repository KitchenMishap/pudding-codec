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

	err = chainstats.PriceDiscoveryHalfTwenty(chain.Blockchain(), chain.HandleCreator(), 880_000, 8_888)
	if err != nil {
		return err
	}

	fmt.Printf("Done price discovery\n")
	return nil
}
