package main

import "github.com/gnolang/tajarin/tajarin"

func main() {
	tajarin.Subscribe(tajarin.JsonTajarinRequest{
		Name:    "sergio",
		PubKey:  "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqjqy9gk8qcd2j6h5wqyj6fu5elup0f60funv258ayyc80yez3ravjj2ttf",
		Address: "g1gdtzrkkgt52efdhfs6tl8d7laag3u3fmcszuet",
	})
}
