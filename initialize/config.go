package initialize

import (
	"KeepAccount/global"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func Config() {
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) == 0 {
		configPath = "config.yaml"
	}
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &global.GvaConfig)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

}
