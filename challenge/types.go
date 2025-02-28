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
	CPU          string   `yaml:"cpu,omitempty"`
	Memory       string   `yaml:"mem,omitempty"`
	Attempts	 int	  `yaml:"attempts,omitempty"`
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
	CategoryValues []string   `json:"category_values"`
	ChallValues    []string   `json:"chall_values"`
	HintsValues    [][]string `json:"hints_values"`
	DepConfig      DepConfig  `json:"deployment_config"`
	HintsChanged   bool       `json:"hints_changed"`
	ChallChanged   bool       `json:"chall_changed"`
	DockerChanged  []string   `json:"docker_changed"`
	ResChanged     []string   `json:"res_changed"`
	OldName        string     `json:"old_name"`
	NewName        string     `json:"new_name"`
}

type DepConfig struct {
	CustomDeploy CustomDeploy    `json:"custom_deploy"`
	Registry     client.Registry `json:"registry"`
	DepType      string          `json:"type"`
	DepPort      int             `json:"port"`
	Subdomain    string          `json:"subd"`
	Resources    Resources       `json:"resources"`
}

type Resources struct {
	CPULimit string `json:"cpu_limit"`
	CPUReq   string `json:"cpu_req"`
	MemLimit string `json:"mem_limit"`
	MemReq   string `json:"mem_req"`
}

type JobStatus struct {
	JobName string `json:"job_name"`
	Status  string `json:"status"`
}