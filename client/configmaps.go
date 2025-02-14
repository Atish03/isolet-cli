package client

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (clientset *CustomClient) CreateConfigMap(configName, namespace, config, key string) (*v1.ConfigMap, error) {
	clientset.DeleteConfigMap(namespace, configName)

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: namespace,
		},
		Data: map[string] string {
			key: config,
		},
	}

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot create configmap %v", err)
	}

	return configMap, nil
}

func (clientset *CustomClient) DeleteConfigMap(namespace, configName string) error {
	err := clientset.CoreV1().ConfigMaps(namespace).Delete(context.Background(), configName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cannot delete configmap %s: %v", configName, err)
	}

	return nil
}