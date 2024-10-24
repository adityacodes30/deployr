package utils

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type AppCfg struct {
	Target    string `yaml:"target"`
	Domain    string `yaml:"domain"`
	Email     string `yaml:"email"`
	Region    string `yaml:"region"`
	Ami       string `yaml:"ami"`
	AwsAcess  string `yaml:"aws_access_key_id"`
	AwsSecret string `yaml:"aws_secret_access_key"`
	DeployrSh string `yaml:"deployrSH"`
}

func DeployrConfig(cfgPath string, appCfg *AppCfg) {
	f, err := os.Open(cfgPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(appCfg)
	if err != nil {
		panic(err)
	}
}

func AwsCfg(deployrcfg AppCfg) {
	err := os.MkdirAll(os.Getenv("HOME")+"/.aws", 0755)

	if err != nil {
		fmt.Println("error while making dr")
		log.Fatal("Error occored")
	}

	filePath := os.Getenv("HOME") + "/.aws/credentials"
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("error while making file")
		log.Fatal(err)
	}
	defer file.Close()

	credentialsFileContent := fmt.Sprintf(`[deployr]
aws_access_key_id = %s
aws_secret_access_key = %s`, deployrcfg.AwsAcess, deployrcfg.AwsSecret)

	_, err = file.WriteString(credentialsFileContent)
	if err != nil {
		fmt.Println("error while writing to file")
		log.Fatal(err)
	}

	// Create config file
	configPath := os.Getenv("HOME") + "/.aws/config"
	configFile, err := os.Create(configPath)
	if err != nil {
		fmt.Println("Error while creating config file:", err)
		log.Fatal(err)
	}
	defer configFile.Close()

	configfilecontent := fmt.Sprintf(`[profile deployr]
region = %s
`, deployrcfg.Region)

	// Write region to config file
	_, err = configFile.WriteString(configfilecontent)
	if err != nil {
		fmt.Println("Error while writing to config file:", err)
		log.Fatal(err)
	}
}
