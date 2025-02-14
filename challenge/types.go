package challenge

import (
	"time"

	"github.com/Atish03/isolet-cli/client"
)

type ChallCache struct {
	ChallHash   string            `json:"chall_hash"`
	ChallName   string            `json:"chall_name"`
	HintsHash   string            `json:"hints_hash"`
	DockerHashs map[string]string `json:"docker_hash"`
	ResHashs    map[string]string `json:"resources_hash"` 
	TimeStamp   time.Time         `json:"push_time"`
}

type DirHash struct {
	DirName string `json:"dir_name"`
	Hash    string `json:"hash"`
}

type CustomDeploy struct {
	Custom     bool   `json:"custom"`
	Deployment string `json:"deployment"`
}

type Challenge struct {
    ChallName    string   `yaml:"chall_name"`
    Type         string   `yaml:"type"`
    CategoryName string   `yaml:"category_name"`
    Prompt       string   `yaml:"prompt"`
    Points       int      `yaml:"points"`
    Requirements []string `yaml:"requirements"`
    Files        []string `yaml:"files"`
    Flag         string   `yaml:"flag"`
    Hints        []Hint   `yaml:"hints"`
    Author       string   `yaml:"author"`
    Visible      bool     `yaml:"visible,omitempty"`
    Tags         []string `yaml:"tags"`
    Links        []string `yaml:"links"`
	DepType      string   `yaml:"deployment_type,omitempty"`
	DepPort      int      `yaml:"deployment_port,omitempty"`
	CPU          int      `yaml:"cpu,omitempty"`
	Memory       int      `yaml:"mem,omitempty"`
	ChallDir     string
	CustomDeploy CustomDeploy
	Registry     *client.Registry
	ChallCache   ChallCache
	PrevCache    ChallCache
}

type Hint struct {
    Hint    string `yaml:"hint"`
    Cost    int    `yaml:"cost"`
	Visible bool   `yaml:"visible"`
}

type ExportStruct struct {
	CategoryQuery string   `json:"category_query"`
	ChallQuery    string   `json:"chall_query"`
	HintsQuery    string   `json:"hints_query"`
	DepMeta       DepMeta  `json:"deployment_metadata"`
	HintsChanged  bool     `json:"hints_changed"`
	ChallChanged  bool     `json:"chall_changed"`
	DockerChanged []string `json:"docker_changed"`
	ResChanged    []string `json:"res_changed"`
	OldName       string   `json:"old_name"`
	NewName       string   `json:"new_name"`
}

type DepMeta struct {
	DepType   string `json:"deployment_type"`
	DepPort   int    `json:"deployment_port"`
	Subdomain string `json:"subdomain"`
	CPU       int    `json:"cpu"`
	Memory    int    `json:"mem"`
}