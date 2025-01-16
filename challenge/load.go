package challenge

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Atish03/isolet-cli/client"
	"github.com/Atish03/isolet-cli/random_string"
)

func (chall *Challenge) Load(cli *client.CustomClient, namespace string, waitgroup *sync.WaitGroup) {
	err := func(cli *client.CustomClient, namespace string) error {
		switch chall.Type {
		case "dynamic":
			return chall.dynLoad(cli, namespace)
		case "on-demand":
			return chall.ondemLoad(cli)
		case "static":
			return chall.statLoad(cli)
		default:
			return fmt.Errorf("invalid challenge type")
		}
	}(cli, namespace)

	if err != nil {
		fmt.Printf("cannot load challenge %s: %v", chall.ChallDir, err)
	}

	defer waitgroup.Done()
}

func (chall *Challenge) dynLoad(cli *client.CustomClient, namespace string) error {
	fmt.Println("Loading dynamic chall")
	command := []string{"python", "-u", "main.py"}
	args := []string{}

	job_name := fmt.Sprintf("%s-%s", filepath.Base(filepath.Clean(chall.ChallDir)), random_string.AlphaStringLower(5))

	job, err := cli.StartJob(namespace, job_name, "b3gul4/isolet-automation-chall:latest", &command, &args)
	if err != nil {
		return fmt.Errorf("cannot start job: %v", err)
	}

	cli.CopyAndStreamLogs(namespace, job.Name, fmt.Sprintf("%s/", chall.ChallDir), "/chall")

	err = cli.DeleteJob(namespace, job.Name)
	if err != nil {
		return fmt.Errorf("cannot delete job: %v", err)
	}

	return nil
}

func (chall *Challenge) ondemLoad(cli *client.CustomClient) error {
	fmt.Println("Loading on-demand chall")
	return nil
}

func (chall *Challenge) statLoad(cli *client.CustomClient) error {
	fmt.Println("Loading static chall")
	return nil
}