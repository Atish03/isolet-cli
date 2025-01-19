package challenge

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Atish03/isolet-cli/client"
)

func (chall *Challenge) Load(cli *client.CustomClient, namespace string, export *string, waitgroup *sync.WaitGroup) {
	command := []string{"python", "-u", "main.py"}
	args := []string{}
	defer waitgroup.Done()

	job_name := strings.ReplaceAll(strings.ToLower(filepath.Base(filepath.Clean(chall.ChallDir))), "_", "-")

	job, err := cli.StartJob(namespace, job_name, "b3gul4/isolet-automation-chall:latest", export, chall.Type, &command, &args)
	if err != nil {
		fmt.Printf("cannot start job: %v", err)
		return
	}

	err = cli.CopyAndStreamLogs(namespace, job.Name, fmt.Sprintf("%s/", chall.ChallDir), "/chall")
	if err != nil {
		fmt.Printf("error while streaming logs: %v", err)
		return
	}

	err = cli.DeleteJob(namespace, job.Name)
	if err != nil {
		fmt.Printf("cannot delete job: %v", err)
		return
	}
}