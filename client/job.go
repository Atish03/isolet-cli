package client

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Atish03/isolet-cli/logger"
)

func (challjob *ChallJob) StartJob() (*batchv1.Job, error) {
	var zeroPtr int32 = 0

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
					ServiceAccountName: "deployment-service-account",
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

	if challjob.JobPodEnv.Registry.Private {
		dockerCM := corev1.Volume{
			Name:         "docker-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: challjob.JobPodEnv.Registry.Secret,
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
					ServiceAccountName: "deployment-service-account",
					Containers: []corev1.Container{
						{
							Name:    "job-container",
							Image:   deployjob.JobImage,
							ImagePullPolicy: corev1.PullAlways,
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