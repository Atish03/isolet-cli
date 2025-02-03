package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/Atish03/isolet-cli/challenge"
	"github.com/Atish03/isolet-cli/logger"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ChallsJson struct {
	Challs []string `json:"challs"`
}

type TraefikConfig struct {
	API struct {
		Insecure bool `yaml:"insecure"`
	} `yaml:"api"`
	EntryPoints map[string]struct {
		Address string `yaml:"address"`
	} `yaml:"entryPoints"`
	Providers struct {
		KubernetesCRD struct {
			Namespaces []string `yaml:"namespaces"`
		} `yaml:"kubernetesCRD"`
	} `yaml:"providers"`
}

var TRAEFIK_NS   string = "traefik"
var TRAEFIK_SVC  string = "traefik-lb"
var TRAEFIK_DEP  string = "traefik"
var TRAEFIK_CONF string = "traefik-config"

func compareMaps(map1, map2 map[string]string) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, value := range map1 {
		if map2[key] != value {
			return false
		}
	}

	return true
}

func isChallChanged(chall challenge.Challenge) bool {
	if chall.ChallCache.ChallHash != chall.PrevCache.ChallHash {
		logger.LogMessage("WARN", fmt.Sprintf("%s, you are trying to deploy a challenge without loading the latest version first, please load or use --force", chall.ChallDir), "Main")
		return true
	}

	if !compareMaps(chall.ChallCache.DockerHashs, chall.PrevCache.DockerHashs) {
		logger.LogMessage("WARN", fmt.Sprintf("%s: you are trying to deploy a challenge with docker files changed, please load or use --force", chall.ChallDir), "Main")
		return true
	}

	return false
}

func deployChalls(challs []challenge.Challenge, force bool) {
	chall_names := []string {}
	exp := ChallsJson{}

	for _, chall := range(challs) {
		if chall.Type == "dynamic" {
			if !isChallChanged(chall) || force {
				chall_names = append(chall_names, chall.ChallName)
			}
		}
	}

	if len(chall_names) == 0 {
		logger.LogMessage("WARN", "no dynamic challenges found to deploy", "Main")
		return
	}

	exp.Challs = chall_names

	jsonBytes, err := json.Marshal(exp)
	if err != nil {
		logger.LogMessage("error", fmt.Sprintf("cannot marshal json: %v", err), "Main")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	kubecli.Deploy("automation", string(jsonBytes), &wg)

	wg.Wait()
}

func patchService(portsToRemove []int32) error {
	service, err := kubecli.CoreV1().Services(TRAEFIK_NS).Get(context.Background(), TRAEFIK_SVC, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get service: %v", err)
	}

	var newPorts []corev1.ServicePort
	for _, port := range service.Spec.Ports {
		if !slices.Contains(portsToRemove, port.Port) {
			newPorts = append(newPorts, port)
		}
	}

	if len(newPorts) == len(service.Spec.Ports) {
		logger.LogMessage("WARN", "port not found in service, no changes made.", "Main")
		return nil
	}

	service.Spec.Ports = newPorts

	_, err = kubecli.CoreV1().Services(TRAEFIK_NS).Update(context.Background(), service, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update service: %v", err)
	}

	return nil
}

func updateCM(subd string) error {
	configMap, err := kubecli.CoreV1().ConfigMaps(TRAEFIK_NS).Get(context.Background(), TRAEFIK_CONF, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error fetching ConfigMap: %v", err)
	}

	traefikYaml := configMap.Data["traefik.yaml"]
	if traefikYaml == "" {
		return fmt.Errorf("traefik.yaml key not found in ConfigMap")
	}

	var config TraefikConfig
	err = yaml.Unmarshal([]byte(traefikYaml), &config)
	if err != nil {
		return fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	delete(config.EntryPoints, subd)

	updatedYaml, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("error marshalling updated content back to YAML: %v", err)
	}

	configMap.Data["traefik.yaml"] = string(updatedYaml)

	_, err = kubecli.CoreV1().ConfigMaps(TRAEFIK_NS).Update(context.Background(), configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating ConfigMap: %v", err)
	}

	return nil
}

func restartDeployment() error {
	deployment, err := kubecli.AppsV1().Deployments(TRAEFIK_NS).Get(context.Background(), TRAEFIK_DEP, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Error fetching deployment: %v", err)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = kubecli.AppsV1().Deployments(TRAEFIK_NS).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Error updating deployment: %v", err)
	}

	return nil
}

func deleteChalls(challs []challenge.Challenge) {
	ports := []int32{}

	for _, chall := range(challs) {
		if chall.Type == "dynamic" {
			subd := challenge.ConvertToSubdomain(chall.ChallName)

			err := kubecli.AppsV1().Deployments("dynamic").Delete(context.Background(), subd, metav1.DeleteOptions{})
			if err != nil {
				logger.LogMessage("ERROR", fmt.Sprintf("Failed to delete Deployment: %v", err), "Main")
			}

			err = kubecli.CoreV1().Services("dynamic").Delete(context.Background(), fmt.Sprintf("%s-svc", subd), metav1.DeleteOptions{})
			if err != nil {
				logger.LogMessage("ERROR", fmt.Sprintf("Failed to delete Service: %v", err), "Main")
			}

			gvr := schema.GroupVersionResource{
				Group:    "traefik.io",
				Version:  "v1alpha1",
				Resource: "ingressroutetcps",
			}
		
			dynamicClient, err := dynamic.NewForConfig(kubecli.Config)
			if err != nil {
				logger.LogMessage("ERROR", fmt.Sprintf("Failed to create dynamic client: %v", err), "Main")
			}

			err = dynamicClient.Resource(gvr).Namespace("dynamic").Delete(context.Background(), fmt.Sprintf("%s-ingress", subd), metav1.DeleteOptions{})
			if err != nil {
				logger.LogMessage("ERROR", fmt.Sprintf("Failed to delete ingressroutetcp: %v", err), "Main")
			}

			if chall.DepPort != 80 {
				ports = append(ports, int32(chall.DepPort))
			}

			err = updateCM(subd)
			if err != nil {
				logger.LogMessage("ERROR", fmt.Sprintf("cannot remove from entrypoint: %v", err), "Main")
			}

			logger.LogMessage("INFO", fmt.Sprintf("undeployed challenge \"%s\"", chall.ChallName), "Main")
		}
	}

	err := patchService(ports)
	if err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("cannot patch service: %v", err), "Main")
	}

	err = restartDeployment()
	if err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("cannot restart traefik deployment: %v", err), "Main")
	}
}