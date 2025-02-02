package client

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *CustomClient) GetAdminSecret() (secret string, err error) {
	item, err := client.CoreV1().Secrets("platform").Get(context.Background(), "platform-secrets", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	secret = string(item.Data["ADMIN_SECRET"])

	return
}

func (client *CustomClient) GetPublicURL() (url string, err error) {
	item, err := client.CoreV1().ConfigMaps("platform").Get(context.Background(), "api-config", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	url = string(item.Data["PUBLIC_URL"])

	return
}

func (client *CustomClient) GetRegistrySecretName(namespace, chall_type string) (*string) {
	var cm *v1.Secret;
	var err error;

	if chall_type == "dynamic" {
		cm, err = client.CoreV1().Secrets(namespace).Get(context.Background(), "dynamic-registry-secret", metav1.GetOptions{})
		if err != nil {
			return nil
		}
	} else if chall_type == "on-demand" {
		cm, err = client.CoreV1().Secrets(namespace).Get(context.Background(), "isolet-registry-secret", metav1.GetOptions{})
		if err != nil {
			return nil
		}
	} else {
		return nil
	}

	return &cm.Name
}