package challenge

import (
	"fmt"
)

func (chall *Challenge) Load() error {
	switch chall.Type {
	case "dynamic":
		return chall.dynLoad()
	case "on-demand":
		return chall.ondemLoad()
	case "static":
		return chall.statLoad()
	default:
		return fmt.Errorf("invalid challenge type")
	}
}

func (chall *Challenge) dynLoad() error {
	fmt.Println("Loading dynamic chall")
	return nil
}

func (chall *Challenge) ondemLoad() error {
	fmt.Println("Loading on-demand chall")
	return nil
}

func (chall *Challenge) statLoad() error {
	fmt.Println("Loading static chall")
	return nil
}