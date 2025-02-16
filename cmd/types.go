package cmd

import "github.com/Atish03/isolet-cli/challenge"

type DynChall struct {
	ChallName string              `json:"chall_name"`
	Custom    bool                `json:"custom"`
	ConfigMap string              `json:"config_map"`
	DepConfig challenge.DepConfig `json:"deployment_config"`
}

type ChallsJson struct {
	Challs []DynChall `json:"challs"`
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
var TRAEFIK_SVC  string = "traefik-svc"
var TRAEFIK_DEP  string = "traefik-deployment"
var TRAEFIK_CONF string = "traefik-config"