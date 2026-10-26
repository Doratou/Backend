package bootstrap

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	Constant    *Constant
	Environment *Environment
}

func Run() (*Config, error) {
	environment, err := NewEnvironment()
	if err != nil {
		return nil, err
	}
	constant := NewConstant()

	printBootstrapConfig(constant, environment)

	return &Config{Constant: constant, Environment: environment}, nil
}

func printBootstrapConfig(c *Constant, e *Environment) {
	fmt.Println("================================================================================")
	fmt.Println("                         BOOTSTRAP CONFIGURATION                                ")
	fmt.Println("================================================================================")

	if cBytes, err := json.MarshalIndent(c, "", "  "); err == nil {
		fmt.Printf("[BOOTSTRAP] Constants:\n%s\n\n", string(cBytes))
	} else {
		fmt.Printf("[BOOTSTRAP] Constants: %+v\n\n", c)
	}

	if eBytes, err := json.MarshalIndent(e, "", "  "); err == nil {
		fmt.Printf("[BOOTSTRAP] Environment:\n%s\n", string(eBytes))
	} else {
		fmt.Printf("[BOOTSTRAP] Environment: %+v\n", e)
	}
	fmt.Println("================================================================================")
}
