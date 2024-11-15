package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// proprietary flag handling for CLI

func Cli(osArgs []string) (bool, string) {

	if len(osArgs) < 2 {
		fmt.Println("Usage: deployr <command>")
		os.Exit(1)
	}

	arg1 := osArgs[1]

	switch arg1 {
	case "-v":
		fmt.Println("Deployr on v1.1")
		return true, ""

	case "config.yml":
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
			os.Exit(1)
		}
		var configFilePath string
		if !filepath.IsAbs(arg1) {
			configFilePath = filepath.Join(cwd, arg1)
		}
		printWelcome()
		fmt.Println("Using config file at:", configFilePath)
		return false, configFilePath

	case "cred":

		res := ShowAuthCreds(osArgs[2])

		if !res {
			return true, "Credentials are not yet generated, please run the app at least once"
		}

		return true, ""

	case "init":
		err := writeConfigFile()
		if err != nil {
			fmt.Println("Error writing config file:", err)
			return true, "Failed to write config file"
		}
		fmt.Println("Config file created successfully")
		return true, "Config file initialized"

	case "-help":
		printHelp()
		return true, "Help Command executed"

	default:
		return true, "Please add a valid CLI argument \n Run deployr -help to see available commands"
	}
}

func writeConfigFile() error {

	url := "https://raw.githubusercontent.com/adityacodes30/deployr/refs/heads/main/config.example.yml"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch config file: " + resp.Status)
	}

	// Read the content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	filePath := "config.yml"
	err = os.WriteFile(filePath, content, 0644)
	if err != nil {
		return err
	}

	return nil
}

func printHelp() {
	green := "\033[32m"
	reset := "\033[0m"

	fmt.Println(green + "Deployr CLI - Available Commands: \n" + reset)
	fmt.Println(green + "  -v           : Display the current version of deployr" + reset)
	fmt.Println(green + "  init         : Initialize a config.yml file in the current directory with sample content" + reset)
	fmt.Println(green + "  config.yml   : Specify a custom config file path (defaults to current directory) and run the deployment process" + reset)
	fmt.Println(green + "  cred <domain.com>       : Display private and public keys for the domain specicifc daemon auth Replace <domain.com> with your domain " + reset)
	fmt.Println(green + "  -help        : Display this help message" + reset)
}

func printWelcome() {
	asciiArt := `
 _____       ______       ______     __           ______       __  __       ______    
/\  __-.    /\  ___\     /\  == \   /\ \         /\  __ \     /\ \_\ \     /\  == \   
\ \ \/\ \   \ \  __\     \ \  _-/   \ \ \____    \ \ \/\ \    \ \____ \    \ \  __<   
 \ \____-    \ \_____\    \ \_\      \ \_____\    \ \_____\    \/\_____\    \ \_\ \_\ 
  \/____/     \/_____/     \/_/       \/_____/     \/_____/     \/_____/     \/_/ /_/ 
                                                                                      
`
	fmt.Println(asciiArt)
}

func ShowAuthCreds(domain string) bool {

	domainDirPrefix := strings.Replace(domain, ".com", "", 1)

	deployrdir := os.Getenv("HOME") + "/.deployr"
	tempKeyDir := deployrdir + "/" + domainDirPrefix + "/auth"
	privateKeyPath := tempKeyDir + "/key"
	publicKeyPath := tempKeyDir + "/key.pub"

	privateKeyContent := readFile(privateKeyPath)
	publicKeyContent := readFile(publicKeyPath)

	if privateKeyContent == "" || publicKeyContent == "" {
		return false
	}

	fmt.Println("Private Key: \n \n ")
	fmt.Println(privateKeyContent)
	fmt.Println("\n \n Public Key: \n \n ")
	fmt.Println(publicKeyContent)

	return true

}
