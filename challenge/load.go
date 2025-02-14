package challenge

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Atish03/isolet-cli/client"
	"github.com/Atish03/isolet-cli/logger"
)

func (chall *Challenge) Load(cli *client.CustomClient, namespace string, wg *sync.WaitGroup) error {
	job_name := filepath.Base(filepath.Clean(chall.ChallDir))

	adminSecret, err := cli.GetAdminSecret()
	if err != nil {
		return fmt.Errorf("cannot get admin secret: %v", err)
	}

	publicURL, err := cli.GetPublicURL()
	if err != nil {
		return fmt.Errorf("cannot get public URL: %v", err)
	}

	job := client.ChallJob {
		Namespace: namespace,
		JobName:   job_name,
		JobImage:  "b3gul4/isolet-challenge-load:v0.1.5",
		JobPodEnv: client.JobPodEnv {
			ChallType:   chall.Type,
			Registry:    chall.Registry,
			AdminSecret: adminSecret,
			Public_URL:  publicURL,
		},
		Command:   []string{"python", "-u", "main.py"},
		Args:      []string{},
		ClientSet: cli,
	}

	export, err := chall.GetExportStruct()
	if err != nil {
		return fmt.Errorf("cannot get export data: %v", err)
	}

	configMap, err := cli.CreateConfigMap(job_name, namespace, export, "config.json")
	if err != nil {
		return fmt.Errorf("cannot create config map: %v", err)
	}

	if chall.CustomDeploy.Custom {
		_, err := cli.CreateConfigMap(job_name, "deployments", chall.CustomDeploy.Deployment, "deployment.yaml")
		if err != nil {
			return fmt.Errorf("cannot create deployment config map for custom deployment %s", chall.ChallDir)
		}
	}

	jobDesc, err := job.StartJob()
	if err != nil {
		return fmt.Errorf("cannot start job: %v", err)
	}

	go func() {
		success, err := cli.DeleteJobAndCM(namespace, jobDesc.Name, configMap.Name)
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("cannot delete job: %v", err), "Job Delete")
			wg.Done()
			return
		}
		if success {
			chall.SaveCache()
		}

		wg.Done()
	}()

	err = cli.CopyAndStreamLogs(namespace, jobDesc.Name, fmt.Sprintf("%s/", chall.ChallDir), "/chall")
	if err != nil {
		return fmt.Errorf("error while streaming logs: %v", err)
	}

	return nil
}