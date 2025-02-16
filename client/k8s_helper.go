package client

import (
	"context"

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
	item, err := client.CoreV1().ConfigMaps("automation").Get(context.Background(), "automation-config", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	url = string(item.Data["PUBLIC_URL"])

	return
}

func (client *CustomClient) GetRegistry() (*Registry) {
	cm, err := client.CoreV1().ConfigMaps("automation").Get(context.Background(), "automation-config", metav1.GetOptions{})
	if err != nil {
		return nil
	}

	registry := Registry{}

	if cm.Data["CHALLENGE_REGISTRY_PRIVATE"] == "true" {
		registry.Private = true
		registry.Secret = "challenge-registry-secret"
	}

	registry.URL = cm.Data["CHALLENGE_REGISTRY"]

	return &registry
}