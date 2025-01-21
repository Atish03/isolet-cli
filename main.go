package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Atish03/isolet-cli/challenge"
	"github.com/Atish03/isolet-cli/client"
	"github.com/Atish03/isolet-cli/logger"
)

func main() {
	// TO BE IMPLEMENTED
	kubecli, err := client.GetClient()
	if err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("cannot get client: %v", err), "Main")
	}

	challs := challenge.GetChalls("./test/sample_challs/chall-1")

	var wg sync.WaitGroup
	
	for _, chall := range(challs) {
		wg.Add(1)
		exp, _ := chall.GetExportStruct()
		jsonExp, _ := json.Marshal(exp)
		encExp := base64.StdEncoding.EncodeToString(jsonExp)
		go chall.Load(&kubecli, "automate", "docker.io/b3gul4", "https://index.docker.io/v1/", &encExp, &wg)
	}

	wg.Wait()
}