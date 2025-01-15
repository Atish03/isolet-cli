package main

import (
	"fmt"

	"github.com/Atish03/isolet-cli/challenge"
	"github.com/Atish03/isolet-cli/client"
)

func main() {
	// TO BE IMPLEMENTED
	kubecli := client.GetClient()

	challs := challenge.GetChalls("./test/sample_challs/chall-2")
	
	for _, chall := range(challs) {
		err := chall.Load(&kubecli)
		if err != nil {
			fmt.Println(err)
		}
	}
}