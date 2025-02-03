package client

import (
	"fmt"

	"github.com/Atish03/isolet-cli/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"path/filepath"
)

type CustomClient struct {
	*kubernetes.Clientset
	Config *rest.Config
}

func GetClient() (CustomClient) {
	var kubeconfig string
	home := homedir.HomeDir()

	if home == "" {
		logger.LogMessage("ERROR", "path for home is not found", "Main")
	}

	kubeconfig = filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("cannot get config for cluster: %v", err), "Main")
		return CustomClient{}
	}

	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	config.APIPath = "/api"
	config.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion()

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("cannot get client for config: %v", err), "Main")
		return CustomClient{}
	}
	
	clientset := CustomClient{client, config}

	return clientset
}