package main

import (
	"sync"

	"github.com/Atish03/isolet-cli/challenge"
	"github.com/Atish03/isolet-cli/client"
)

func main() {
	// TO BE IMPLEMENTED
	kubecli := client.GetClient()

	challs := challenge.GetChalls("./test/sample_challs/chall-4")

	var wg sync.WaitGroup
	
	for _, chall := range(challs) {
		wg.Add(1)
		go chall.Load(&kubecli, "automate", &wg)
	}

	wg.Wait()
}