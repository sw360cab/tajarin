package main

import "github.com/gnolang/tajarin/tajarin"

func main() {
	// launch producer
	tajarinProducer := tajarin.NewTajarinProducer(1)
	tajarinProducer.ListenAndWait()
}
