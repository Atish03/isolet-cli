package challenge

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Atish03/isolet-cli/client"
)

func (chall *Challenge) Load(cli *client.CustomClient, namespace, repository, registry string, export *string, waitgroup *sync.WaitGroup) {
	defer waitgroup.Done()

	job_name := strings.ReplaceAll(strings.ToLower(filepath.Base(filepath.Clean(chall.ChallDir))), "_", "-")

	job := client.ChallJob {
		Namespace: namespace,
		JobName:   job_name,
		JobImage:  "b3gul4/isolet-automation-chall:latest",
		JobPodEnv: client.JobPodEnv {
			ImageName: fmt.Sprintf("%s/%s-challenge-image:latest", repository, job_name),
			Export:    *export,
			ChallType: chall.Type,
			Registry:  registry,
		},
		Command:   []string{"python", "-u", "main.py"},
		Args:      []string{},
		ClientSet: cli,
	}

	jobDesc, err := job.StartJob()
	if err != nil {
		fmt.Printf("cannot start job: %v", err)
		return
	}

	err = cli.CopyAndStreamLogs(namespace, jobDesc.Name, fmt.Sprintf("%s/", chall.ChallDir), "/chall")
	if err != nil {
		fmt.Printf("error while streaming logs: %v", err)
		return
	}

	err = cli.DeleteJob(namespace, jobDesc.Name)
	if err != nil {
		fmt.Printf("cannot delete job: %v", err)
		return
	}
}