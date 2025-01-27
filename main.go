package main

import (
	// "fmt"
	// "sync"

	// "github.com/Atish03/isolet-cli/challenge"
	// "github.com/Atish03/isolet-cli/client"
	// "github.com/Atish03/isolet-cli/logger"
	"github.com/Atish03/isolet-cli/cmd"
)

func main() {
	// TO BE IMPLEMENTED
	// kubecli, err := client.GetClient()
	// if err != nil {
	// 	logger.LogMessage("ERROR", fmt.Sprintf("cannot get client: %v", err), "Main")
	// }

	// challs := challenge.GetChalls("./test/sample_challs/")

	// var wg sync.WaitGroup
	
	// for _, chall := range(challs) {
	// 	wg.Add(1)
		
	// 	go func(){
	// 		err := chall.Load(&kubecli, "automation", "docker.io/b3gul4", "https://index.docker.io/v1/", &wg)
	// 		if err != nil {
	// 			logger.LogMessage("ERROR", fmt.Sprintf("error loading challenge: %v", err), "Main")
	// 		}
	// 	}()
	// }

	// wg.Wait()

	cmd.Execute()
}