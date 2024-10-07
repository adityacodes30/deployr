package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

const cofgPath = "config.yml"

type AppCfg struct {
	Target    string `yaml:"target"`
	Domain    string `yaml:"domain"`
	Email     string `yaml:"email"`
	Region    string `yaml:"region"`
	Ami       string `yaml:"ami"`
	AwsAcess  string `yaml:"aws_access_key_id"`
	AwsSecret string `yaml:"aws_secret_access_key"`
}

var deployrcfg AppCfg

func DeployrConfig() {
	f, err := os.Open(cofgPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&deployrcfg)

	if err != nil {
		panic(err)
	}
}
