package tests

import (
	"github.com/KitchenMishap/pudding-codec/jobs"
	"testing"
)

func Test_main(t *testing.T) {
	err := jobs.PriceDiscoveryHalfTwenty("E:\\Data\\FleeSwallowImmune888888CswHashesDeleted")
	if err != nil {
		t.Error(err)
	}
}
