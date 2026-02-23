package main

import (
	"fmt"
	"github.com/KitchenMishap/pudding-codec/jobs"
)

func main() {
	//err := jobs.PriceDiscovery500000("E:\\Data\\FleeSwallowImmune888888CswHashesDeleted")
	//err := jobs.PriceDiscoveryHalfTwenty("E:\\Data\\FleeSwallowImmune888888CswHashesDeleted")
	err := jobs.PriceDiscoveryHumanBehaviour("E:\\Data\\FleeSwallowImmune888888CswHashesDeleted")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	fmt.Printf("main() completed\n")
}
