package client

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client CustomClient) GetDockerConfig(namespace string, name string) (configjson string, err error) {
	item, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	configjson = string(item.Data[".dockerconfigjson"])

	return
}