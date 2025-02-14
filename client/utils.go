package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd/exec"

	"github.com/Atish03/isolet-cli/cp"
	"github.com/Atish03/isolet-cli/logger"
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

func (clientset *CustomClient) startCopying(namespace, podName, container, src, dest string) error {
	err := clientset.runCmdInPod(namespace, podName, []string{"touch", "/tmp/resources.lock"})
	if err != nil {
		return fmt.Errorf("error creating lock: %v", err)
	}
	err = clientset.copyToPod(namespace, container, src, dest)
	if err != nil {
		return fmt.Errorf("error copying: %v", err)
	}
	err = clientset.runCmdInPod(namespace, podName, []string{"rm", "/tmp/resources.lock"})
	if err != nil {
		return fmt.Errorf("error removing lock: %v", err)
	}

	return nil
}

func (clientset *CustomClient) getPodLog(namespace, podName, jobName string) error {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})

	logStream, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("error opening log stream for pod %s: %w", podName, err)
	}

	defer logStream.Close()

	scanner := bufio.NewScanner(logStream)
	for scanner.Scan() {
		logger.LogMessage("DEBUG", scanner.Text(), jobName, podName)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log stream: %w", err)
	}

	return nil
}

func (clientset *CustomClient) copyToPod(namespace, containerName, src, dest string) error {
	copyOpts := cp.CopyOptions {
		Namespace: namespace,
		Args: []string{src, dest},
		Clientset: clientset,
		ClientConfig: clientset.Config,
		NoPreserve: true,
		Container: containerName,
		IOStreams: genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}

	err := copyOpts.Run()
	if err != nil {
		return fmt.Errorf("error when running the copy command: %v", err)
	}
	
	return nil
}

func (clientset *CustomClient) runCmdInPod(namespace, podName string, cmd []string) error {
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericiooptions.IOStreams{
				Out:    io.Discard,
				ErrOut: io.Discard,
			},

			Namespace: namespace,
			PodName:   podName,
		},

		Command:  cmd,
		Executor: &exec.DefaultRemoteExecutor{},
	}

	options.Config = clientset.Config
	options.Namespace = namespace
	options.PodClient = clientset.CoreV1()

	return options.Run()
}

func (clientset *CustomClient) CopyAndStreamLogs(namespace, jobName, src, dest string) error {
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
				dest = fmt.Sprintf("%s/%s:%s", namespace, pod.Name, dest)
				container := pod.Spec.Containers[0].Name

				go func() {
					err := clientset.startCopying(namespace, pod.Name, container, src, dest)
					if err != nil {
						logger.LogMessage("ERROR", fmt.Sprintf("error when copying: %v", err), jobName)
					}
				}()
				
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