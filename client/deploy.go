package client

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/Atish03/isolet-cli/logger"
)

func (client *CustomClient) Deploy(namespace string, challs string, wg *sync.WaitGroup) error {
	job_name := fmt.Sprintf("deploy-job-%s", randStringRunes(5))

	publicURL, err := client.GetPublicURL()
	if err != nil {
		return fmt.Errorf("cannot get public URL: %v", err)
	}

	url, err := url.Parse(publicURL)
	if err != nil {
		return fmt.Errorf("cannot parse public URL %s", publicURL)
	}
	domain := url.Hostname()

	configMap, err := client.CreateConfigMap(job_name, namespace, challs, "challs.json")
	if err != nil {
		return fmt.Errorf("cannot create config map: %v", err)
	}

	deployJob := DeployJob {
		Namespace: namespace,
		JobName:   job_name,
		JobImage:  "b3gul4/isolet-dynamic-delpoy:v0.1.9",
		Domain:    domain,
		ClientSet: client,
	}

	job, err := deployJob.StartJob()
	if err != nil {
		return fmt.Errorf("cannot start deploy job: %v", err)
	}

	go func() {
		success, err := client.DeleteJobAndCM(namespace, job.Name, configMap.Name)
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("cannot delete job: %v", err), "Job Delete")
			wg.Done()
			return
		}

		if !success {
			logger.LogMessage("ERROR", "deployment of challenges failed", "Deploy Job")
		}

		wg.Done()
	}()

	err = client.streamLogs(namespace, job.Name)
	if err != nil {
		return fmt.Errorf("error while streaming logs: %v", err)
	}

	return nil
}