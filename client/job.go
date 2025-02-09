package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd/exec"

	"github.com/Atish03/isolet-cli/cp"
	"github.com/Atish03/isolet-cli/logger"
)

type JobPodEnv struct {
	ChallType   string;
	Registry    string;
	AdminSecret string;
	Public_URL  string;
}

type ChallJob struct {
	Namespace  string;
	JobName    string;
	JobImage   string;
	JobPodEnv  JobPodEnv;
	Command    []string;
	Args       []string;
	ClientSet  *CustomClient;
}

type DeployJob struct {
	Namespace string;
	JobName   string;
	JobImage  string;
	Domain    string;
	ClientSet *CustomClient;
}

func (challjob *ChallJob) StartJob() (*batchv1.Job, error) {
	var zeroPtr int32 = 0
	dockerConfigMap := challjob.ClientSet.GetRegistrySecretName(challjob.Namespace, challjob.JobPodEnv.ChallType)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: challjob.JobName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &zeroPtr,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"job": challjob.JobName},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					ServiceAccountName: "cli-service-account",
					Containers: []corev1.Container{
						{
							Name:    "job-container",
							Image:   challjob.JobImage,
							Command: challjob.Command,
							Args:    challjob.Args,
							ImagePullPolicy: corev1.PullAlways,
							SecurityContext: &corev1.SecurityContext{
								Privileged: func(b bool) *bool { return &b }(true),
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeUnconfined,
								},
							},
							Env: 	 []corev1.EnvVar{
								{
									Name: "CHALL_TYPE",
									Value: challjob.JobPodEnv.ChallType,
								},
								{
									Name: "IMAGE_REGISTRY",
									Value: challjob.JobPodEnv.Registry,
								},
								{
									Name: "ADMIN_SECRET",
									Value: challjob.JobPodEnv.AdminSecret,
								},
								{
									Name: "PUBLIC_URL",
									Value: challjob.JobPodEnv.Public_URL,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-vol",
									MountPath: "/config",
									ReadOnly:  true,
								},
							},	
						},
					},
					Volumes: []corev1.Volume{
						{
							Name:         "config-vol",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: challjob.JobName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if dockerConfigMap != nil {
		dockerCM := corev1.Volume{
			Name:         "docker-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: *dockerConfigMap,
				},
			},
		}

		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, dockerCM)

		dockerVolMount := corev1.VolumeMount{
			Name:      "docker-config",
			MountPath: "/docker",
			ReadOnly:  true,
		}

		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, dockerVolMount)
	}

	jobsClient := challjob.ClientSet.BatchV1().Jobs(challjob.Namespace)
	job, err := jobsClient.Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	logger.LogMessage("INFO", fmt.Sprintf("Job created successfully in namespace %s", challjob.Namespace), challjob.JobName)
	return job, nil
}

func (deployjob *DeployJob) StartJob() (*batchv1.Job, error) {
	var zeroPtr int32 = 0

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployjob.JobName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &zeroPtr,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"job": deployjob.JobName},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					ServiceAccountName: "deploy-sa",
					Containers: []corev1.Container{
						{
							Name:    "job-container",
							Image:   deployjob.JobImage,
							ImagePullPolicy: corev1.PullAlways,
							Env: 	 []corev1.EnvVar{
								{
									Name: "DOMAIN_NAME",
									Value: deployjob.Domain,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "challs-vol",
									MountPath: "/config",
									ReadOnly:  true,
								},
							},	
						},
					},
					Volumes: []corev1.Volume{
						{
							Name:         "challs-vol",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: deployjob.JobName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	jobsClient := deployjob.ClientSet.BatchV1().Jobs(deployjob.Namespace)
	job, err := jobsClient.Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	logger.LogMessage("INFO", fmt.Sprintf("Job created successfully in namespace %s", deployjob.Namespace), deployjob.JobName)
	return job, nil
}

func (clientset *CustomClient) DeleteJobAndCM(namespace, jobName, configMapName string) (bool, error) {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	for {
		job, err := clientset.BatchV1().Jobs(namespace).Get(context.Background(), jobName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("cannot get job %s: %v", jobName, err)
		}
		
		if job.Status.Succeeded == 1 || job.Status.Failed == 1 {
			err := clientset.BatchV1().Jobs(namespace).Delete(context.Background(), jobName, deleteOptions)
			if err != nil {
				return false, fmt.Errorf("cannot delete job %s: %v", jobName, err)
			}

			err = clientset.DeleteConfigMap(namespace, configMapName)
			if err != nil {
				return false, err
			}

			logger.LogMessage("INFO", "Job completed and deleted", jobName)

			if job.Status.Succeeded == 1 {
				return true, nil
			} else {
				return false, nil
			}
		}

		time.Sleep(2 * time.Second)
	}
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