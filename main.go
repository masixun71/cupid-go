package main

import (
	"./process"
	"sync"
)

func main() {


	var wg sync.WaitGroup
	wg.Add(1)

	process := process.NewProcess("/apps/conf/cupid/myConfigGo.json")

	process.Run()

	wg.Wait()
}
