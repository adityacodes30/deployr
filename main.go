package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"golang.org/x/crypto/ssh"

	"github.com/miekg/dns"
)

func keygen() (string, string) {
	savePrivateFileTo := "./.deployr/key"
	savePublicFileTo := "./.deployr/key.pub"
	bitSize := 4096

	err := os.MkdirAll("./.deployr", 0755)
	if err != nil {
		log.Fatal("Failed to create .deployr directory: ", err)
	}

	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	err = writeKeyToFile(privateKeyBytes, savePrivateFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = writeKeyToFile([]byte(publicKeyBytes), savePublicFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}

	// fmt.Println("public key " + string(publicKeyBytes))
	// fmt.Println("private key " + string(privateKeyBytes))

	return string(publicKeyBytes), string(privateKeyBytes)
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Println("Public key generated")
	return pubKeyBytes, nil
}

// writePemToFile writes keys to a file
func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}

	log.Printf("Key saved to: %s", saveFileTo)
	return nil
}

func main() {

	DeployrConfig()

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

	cfg, erre := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("deployr"))

	if erre != nil {
		fmt.Println("error while writing loading config")
		log.Fatal(err)
	}

	fmt.Println("credentials ok")

	svc := ec2.NewFromConfig(cfg)

	// create security group

	createGroupOutput, err := svc.CreateSecurityGroup(context.TODO(), &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String("deployr-sg"),
		Description: aws.String("Security group for deployr instance with SSH, HTTP, and HTTPS access"),
	})
	if err != nil {
		log.Fatalf("Failed to create security group: %v", err)
	}

	securityGroupID := aws.ToString(createGroupOutput.GroupId)
	fmt.Printf("Created security group with ID: %s\n", securityGroupID)

	// Authorize inbound rules
	_, err = svc.AuthorizeSecurityGroupIngress(context.TODO(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(securityGroupID),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(22),
				ToPort:     aws.Int32(22),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("Allow SSH access from anywhere"),
					},
				},
			},
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(80),
				ToPort:     aws.Int32(80),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("Allow HTTP traffic from anywhere"),
					},
				},
			},
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(443),
				ToPort:     aws.Int32(443),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("Allow HTTPS traffic from anywhere"),
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to add security group rules: %v", err)
	}

	resp, instanceErr := svc.RunInstances(context.TODO(),
		&ec2.RunInstancesInput{
			ImageId:      aws.String(deployrcfg.Ami),
			InstanceType: types.InstanceTypeT2Micro,
			MinCount:     aws.Int32(1),
			MaxCount:     aws.Int32(1),
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceTypeInstance,
					Tags: []types.Tag{
						{
							Key:   aws.String("deployr"),
							Value: aws.String("deployr"),
						},
					},
				},
			},
			SecurityGroupIds: []string{securityGroupID},
		},
	)

	instanceId := resp.Instances[0].InstanceId

	fmt.Println("Instance id of provisiond response --> " + *instanceId)

	instancePubDns, instancePubIp, _ := GetPublicDNSByInstanceID(svc, *instanceId)

	fmt.Printf("\033[32m"+`
  ---------------------------
  Please point your IP to:
	%s
  ---------------------------
`+"\033[0m", instancePubIp)

	resolveConfimation(deployrcfg.Domain, instancePubIp)

	fmt.Println("the public dns is ")
	fmt.Println(instancePubDns)

	if instanceErr != nil {
		fmt.Println("error while running instance")
		log.Fatal(instanceErr.Error())
	}

	if resp == nil {
		log.Fatal("RunInstances response is nil")
	}

	time.Sleep(10 * time.Second)

	icsvc := ec2instanceconnect.NewFromConfig(cfg)

	pubkey, _ := keygen()

	respp, erroricsvc := icsvc.SendSSHPublicKey(context.TODO(), &ec2instanceconnect.SendSSHPublicKeyInput{
		InstanceId:     instanceId,
		InstanceOSUser: aws.String("ec2-user"),
		SSHPublicKey:   aws.String(pubkey),
	})

	if erroricsvc != nil {
		fmt.Println("error in sending public key ")
		log.Fatal(erroricsvc.Error())
	}

	fmt.Println(respp.Success)

	pkey, err := ioutil.ReadFile("./.deployr/key")

	signer, signError := ssh.ParsePrivateKey([]byte(pkey))

	if signError != nil {
		fmt.Println("error signing jeyt s")
		log.Fatal(signError.Error())
	}

	config := &ssh.ClientConfig{
		User: "ec2-user",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {

			fmt.Printf("Connected to host %s. Fingerprint: %s\n",
				hostname, ssh.FingerprintSHA256(key))

			if err := addHostToKnownHosts(hostname, key); err != nil {
				return fmt.Errorf("failed to add host to known_hosts: %w", err)
			}

			return nil
		},
		Timeout: 10 * time.Second,
	}

	clientresp, clienterr := ssh.Dial("tcp", instancePubDns+":22", config)

	if clienterr != nil {
		fmt.Println("error dialing to the ssh client")
		log.Fatal(clienterr.Error())
	}

	session, err := clientresp.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.

	command := fmt.Sprintf(`sudo sh -c 'if [ ! -d /.deployr ]; then mkdir /.deployr; fi && curl -o /.deployr/deployr.sh https://gist.githubusercontent.com/adityacodes30/68f9887074b0e203e0986ae03a00d842/raw/95743327ad67cfae66e979ecbf794defbd903294/gistfile1.sh && sudo chmod +x /.deployr/deployr.sh && sudo /bin/bash /.deployr/deployr.sh %s %s %s'`, deployrcfg.Target, deployrcfg.Domain, deployrcfg.Email)
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())

}

func GetPublicDNSByInstanceID(ec2Client *ec2.Client, instanceID string) (string, string, error) {
	for {
		// Describe the instance using its ID
		resp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to describe instance: %w", err)
		}

		// Check if instances were found
		if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
			return "", "", fmt.Errorf("no instances found with ID: %s", instanceID)
		}

		// Get the instance and its current state
		instance := resp.Reservations[0].Instances[0]
		state := instance.State.Name

		fmt.Printf("Instance state: %s\n", state)

		// If the instance is running and has a public DNS, return it
		if state == types.InstanceStateNameRunning && aws.ToString(instance.PublicDnsName) != "" {
			return aws.ToString(instance.PublicDnsName), aws.ToString(instance.PublicIpAddress), nil
		}

		// Wait before the next poll
		fmt.Println("waiting for public dns to be assigned")
		time.Sleep(2 * time.Second)
	}
}

func addHostToKnownHosts(host string, key ssh.PublicKey) error {
	knownHostsPath := os.Getenv("HOME") + "/.ssh/known_hosts"

	// Create the formatted host key entry
	hostKeyEntry := fmt.Sprintf("%s %s %s\n",
		host,
		key.Type(),
		base64.StdEncoding.EncodeToString(key.Marshal()))

	// Append the host key to the known_hosts file
	return ioutil.WriteFile(knownHostsPath, []byte(hostKeyEntry), 0644)
}

func PollForIpPoint(domain string, ipAddr string) bool {
	counter := 0
	dnsResolver := "8.8.8.8:53"

	for {
		ip, err := resolveWithDNS(domain, dnsResolver)
		if err != nil {
			fmt.Printf("Error looking up host: %v\n", err)
			return false
		}

		if len(ip) > 0 && ip[0] == ipAddr {
			fmt.Println("Domain resolved to the correct IP!")
			return true
		}

		counter++
		if counter >= 1000 {
			fmt.Println("Reached maximum retries. DNS propagation failed.")
			return false
		}

		fmt.Println("Waiting for public DNS to be propagated...")
		time.Sleep(20 * time.Second)
	}
}

func resolveWithDNS(domain string, resolver string) ([]string, error) {
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, _, err := c.ExchangeContext(ctx, &m, resolver)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	return ips, nil
}

func resolveConfimation(domain string, ip string) {

	fmt.Print("\n\033[32mPlease type 'confirm' if your DNS has propagated:\033[0m\n\n")
	fmt.Printf("You can check that by if the IP on this website matches --> \033[33m%s\033[0m\n \n", ip)
	fmt.Printf("Paste this in your browser --> \033[34m: https://www.nslookup.io/domains/%s/dns-records/\033[0m\n", domain)
	fmt.Printf("\033[34m\033]8;;https://www.nslookup.io/domains/%s/dns-records/\033\\Or Click here to directly go to the website\033]8;;\033\\\033[0m\n\n", domain)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ") // Prompt for input
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "confirm" {
			fmt.Println("DNS propagation confirmed. Proceeding...")
			break
		} else {
			fmt.Println("Invalid input. Please type 'confirm'.")
		}
	}

	time.Sleep(10 * time.Second)
}
