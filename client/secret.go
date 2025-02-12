package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Registry struct {
	URL     string
	Private bool
	Secret  string
}

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

func (client *CustomClient) GetRegistry(chall_type string) (*Registry) {
	cm, err := client.CoreV1().ConfigMaps("automation").Get(context.Background(), "automation-config", metav1.GetOptions{})
	if err != nil {
		return nil
	}

	registry := Registry{}

	if chall_type == "dynamic" {
		registry.URL = cm.Data["DYNAMIC_IMAGE_REGISTRY"]
		if cm.Data["DYNAMIC_REGISTRY_PRIVATE"] == "true" {
			registry.Private = true
			registry.Secret = "dynamic-registry-secret"
		}
		
	} else if chall_type == "on-demand" {
		registry.URL = cm.Data["ISOLET_IMAGE_REGISTRY"]
		if cm.Data["ISOLET_REGISTRY_PRIVATE"] == "true" {
			registry.Private = true
			registry.Secret = "isolet-registry-secret"
		}
	}

	return &registry
}