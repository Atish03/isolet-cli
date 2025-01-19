package main

import (
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/Atish03/isolet-cli/challenge"
	"github.com/Atish03/isolet-cli/client"
)

func main() {
	// TO BE IMPLEMENTED
	kubecli := client.GetClient()

	challs := challenge.GetChalls("./test/sample_challs")

	var wg sync.WaitGroup
	
	for _, chall := range(challs) {
		wg.Add(1)
		exp, _ := chall.GetExportStruct()
		jsonExp, _ := json.Marshal(exp)
		encExp := base64.StdEncoding.EncodeToString(jsonExp)
		go chall.Load(&kubecli, "automate", &encExp, &wg)
	}

	wg.Wait()
}