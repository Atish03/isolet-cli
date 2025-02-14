package cmd

type DynChall struct {
	ChallName string `json:"chall_name"`
	Custom    bool   `json:"custom"`
	ConfigMap string `json:"config_map"`
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
var TRAEFIK_SVC  string = "traefik-lb"
var TRAEFIK_DEP  string = "traefik"
var TRAEFIK_CONF string = "traefik-config"