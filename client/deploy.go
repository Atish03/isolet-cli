package client

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/Atish03/isolet-cli/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func (clientset *CustomClient) streamLogs(namespace, jobName string) error {
	labelSelector := fmt.Sprintf("job=%s", jobName)

	for {
		podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return fmt.Errorf("cannot get list of pods for job %s: %v", jobName, err)
		}
		if len(podList.Items) != 0 {
			for _, pod := range(podList.Items) {
				if pod.Status.Phase != corev1.PodRunning {
					continue
				}
				
				err = clientset.getPodLog(namespace, pod.Name, jobName)
				if err != nil {
					return fmt.Errorf("error streaming logs for pod %s: %v", pod.Name, err)
				}
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
}

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
		JobImage:  "b3gul4/isolet-dynamic-delpoy:v0.1.8",
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