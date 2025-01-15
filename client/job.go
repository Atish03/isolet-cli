package client

import (
	"bufio"
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (clientset *CustomClient) StartJob(namespace, jobName, image string, command *[]string, args *[]string) (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"job": jobName},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					ServiceAccountName: "api-service-account",
					Containers: []corev1.Container{
						{
							Name:    "job-container",
							Image:   image,
							Command: *command,
							Args:    *args,
						},
					},
				},
			},
		},
	}

	jobsClient := clientset.BatchV1().Jobs(namespace)
	job, err := jobsClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	fmt.Printf("Job %s created successfully in namespace %s\n", jobName, namespace)
	return job, nil
}

func (clientset *CustomClient) StreamLogsForJob(namespace, jobName string) {
	labelSelector := fmt.Sprintf("job=%s", jobName)

	for {
		podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			fmt.Printf("cannot get list of pods for job %s: %v", jobName, err)
			return
		}
		if len(podList.Items) != 0 {
			for _, pod := range(podList.Items) {
				if pod.Status.Phase != corev1.PodRunning {
					continue
				}
				err := clientset.getPodLog(namespace, pod.Name)
				if err != nil {
					fmt.Printf("error streaming logs for pod %s: %v", pod.Name, err)
					return
				}
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (clientset *CustomClient) getPodLog(namespace, podName string) error {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})

	logStream, err := req.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("error opening log stream for pod %s: %w", podName, err)
	}

	defer logStream.Close()

	scanner := bufio.NewScanner(logStream)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log stream: %w", err)
	}

	return nil
}