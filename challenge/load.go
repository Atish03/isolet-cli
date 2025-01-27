package challenge

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Atish03/isolet-cli/client"
)

func (chall *Challenge) Load(cli *client.CustomClient, namespace, registry, registry_url string, waitgroup *sync.WaitGroup) error {
	defer waitgroup.Done()

	job_name := filepath.Base(filepath.Clean(chall.ChallDir))

	job := client.ChallJob {
		Namespace: namespace,
		JobName:   job_name,
		JobImage:  "b3gul4/isolet-automation-chall:latest",
		JobPodEnv: client.JobPodEnv {
			ChallType:   chall.Type,
			RegistryURL: registry_url,
			Registry:    registry,
		},
		Command:   []string{"python", "-u", "main.py"},
		Args:      []string{},
		ClientSet: cli,
	}

	export, err := chall.GetExportStruct()
	if err != nil {
		return fmt.Errorf("cannot get export data: %v", err)
	}

	configMap, err := cli.CreateConfigMap(job_name, namespace, export)
	if err != nil {
		return fmt.Errorf("cannot create config map: %v", err)
	}

	jobDesc, err := job.StartJob()
	if err != nil {
		return fmt.Errorf("cannot start job: %v", err)
	}

	err = cli.CopyAndStreamLogs(namespace, jobDesc.Name, fmt.Sprintf("%s/", chall.ChallDir), "/chall")
	if err != nil {
		return fmt.Errorf("error while streaming logs: %v", err)
	}

	success, err := cli.DeleteJob(namespace, jobDesc.Name)
	if err != nil {
		return fmt.Errorf("cannot delete job: %v", err)
	}
	if success {
		chall.SaveCache()
	}

	err = cli.DeleteConfigMap(namespace, configMap.Name)
	if err != nil {
		return err
	}

	return nil
}