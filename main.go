package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"golang.org/x/crypto/ssh"

	"github.com/adityacodes30/deployr/utils"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: deployr <config.yml>")
		os.Exit(1)
	}

	configFilePath := os.Args[1]

	if configFilePath == "-v" {
		fmt.Println("Deployr on v1.0")
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		os.Exit(1)
	}

	if !filepath.IsAbs(configFilePath) {
		configFilePath = filepath.Join(cwd, configFilePath)
	}

	fmt.Println("Using config file at:", configFilePath)

	var deployrcfg utils.AppCfg

	utils.DeployrConfig(configFilePath, &deployrcfg)

	utils.AwsCfg(deployrcfg)

	cfg, cfgErr := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("deployr"))

	if cfgErr != nil {
		fmt.Println("Error while loading aws config")
		log.Fatal(cfgErr)
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

	// Inbound rules
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

	runInstanceResp, instanceErr := svc.RunInstances(context.TODO(),
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

	if instanceErr != nil {
		fmt.Println("error while running instance")
		log.Fatal(instanceErr.Error())
	}

	if runInstanceResp == nil {
		log.Fatal("RunInstances response is nil")
	}

	instanceId := runInstanceResp.Instances[0].InstanceId

	fmt.Println("Instance was provisioned sucessfully\nInstance id of provisiond response --> " + *instanceId)

	instancePubDns, instancePubIp, _ := GetPublicDNSByInstanceID(svc, *instanceId)

	fmt.Printf("\033[32m"+`
  ---------------------------
  Please point your IP to:
	%s
  ---------------------------
`+"\033[0m", instancePubIp)

	utils.ResolveConfimation(deployrcfg.Domain, instancePubIp)

	time.Sleep(10 * time.Second)

	icsvc := ec2instanceconnect.NewFromConfig(cfg)

	pubkey, _, privKeyPath := utils.Keygen()

	respp, erroricsvc := icsvc.SendSSHPublicKey(context.TODO(), &ec2instanceconnect.SendSSHPublicKeyInput{
		InstanceId:     instanceId,
		InstanceOSUser: aws.String("ec2-user"),
		SSHPublicKey:   aws.String(pubkey),
	})

	if erroricsvc != nil {
		fmt.Println("error in sending public key ")
		log.Fatal(erroricsvc.Error())
	}

	fmt.Println("Shh key sent: ", respp.Success)

	pkey, err := ioutil.ReadFile(privKeyPath)

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

			if err := utils.AddHostToKnownHosts(hostname, key); err != nil {
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

	command := fmt.Sprintf(`sudo sh -c 'if [ ! -d /.deployr ]; then mkdir /.deployr; fi && curl -o /.deployr/deployr.sh %s && sudo chmod +x /.deployr/deployr.sh && sudo /bin/bash /.deployr/deployr.sh %s %s %s'`, deployrcfg.DeployrSh, deployrcfg.Target, deployrcfg.Domain, deployrcfg.Email)
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())

	utils.PrintSucesss(deployrcfg.Domain)
}

func GetPublicDNSByInstanceID(ec2Client *ec2.Client, instanceID string) (string, string, error) {
	for {

		resp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to describe instance: %w", err)
		}

		if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
			return "", "", fmt.Errorf("no instances found with ID: %s", instanceID)
		}

		instance := resp.Reservations[0].Instances[0]
		state := instance.State.Name

		fmt.Printf("Instance state: %s\n", state)

		if state == types.InstanceStateNameRunning && aws.ToString(instance.PublicDnsName) != "" {
			return aws.ToString(instance.PublicDnsName), aws.ToString(instance.PublicIpAddress), nil
		}

		fmt.Println("waiting for public dns to be assigned")
		time.Sleep(2 * time.Second)
	}
}
