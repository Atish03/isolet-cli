package client

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
)

type CustomClient struct {
	*kubernetes.Clientset
	Config *rest.Config
}

type JobPodEnv struct {
	ChallType   string;
	Registry    *Registry;
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

type Registry struct {
	URL     string `json:"url"`
	Private bool   `json:"private"` 
	Secret  string `json:"secret"`
}