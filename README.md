# Deployr 💾

Deployr helps you host your Next.js application on your own aws ec2 instance, With just a few clicks, you can set up everything, from machine provisioning to serving your next application on your own domain. 

With github actions and CI/CD, you can automate the deployment process on every push and focus on building your application. Also your deployments are secure and only you can trigger CI/CD builds. You can deploy as many applications as you want across multiple instances.

Here is a video walkthrough / how to use Deployr: [Deployr Walkthrough](https://www.loom.com/share/8c7ca17efc78416d8bec92d46bc482ae?sid=10439ae0-bdd7-4ce3-ae1e-85ea4ef9da86)

(This guide will be updated soon, Readme is sufficient to get you started)

---

# How to Deployr Your App 🚀

v1.2.0 just arrived - Deployments are now authenticated, more interactive cli and multiple deployments support are now here 🎉

How to use Deployr to host your Next.js application on your own AWS EC2 instance:

### Step 1: Get the project:

#### Linux (Intel/AMD64):
```bash
sudo curl -L -o /usr/local/bin/deployr https://github.com/adityacodes30/deployr/releases/download/v1.2.0/deployr-linux-amd64 && sudo chmod +x /usr/local/bin/deployr
```

#### macOS (ARM/M1)
```bash
sudo curl -L -o /usr/local/bin/deployr https://github.com/adityacodes30/deployr/releases/download/v1.2.0/deployr-macos-arm64 && sudo chmod +x /usr/local/bin/deployr
```

#### Linux (ARM64):
```bash
sudo curl -L -o /usr/local/bin/deployr https://github.com/adityacodes30/deployr/releases/download/v1.2.0/deployr-linux-arm64 && sudo chmod +x /usr/local/bin/deployr
```

#### macOS (Intel/AMD64):
```bash
sudo curl -L -o /usr/local/bin/deployr https://github.com/adityacodes30/deployr/releases/download/v1.2.0/deployr-macos-amd64 && sudo chmod +x /usr/local/bin/deployr
```

### Step 2: Test the installation 

```bash
deployr -v
```

You should see the version of the deployr you just installed

### Step 3: Create a config.yml file 

Run the following command to create a config.yml file in your current directory

```bash
deployr init
```

Sign into aws and make a IAM user with the following policy attached: 

```json 
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "deployr",
			"Effect": "Allow",
			"Action": [
				"ec2-instance-connect:SendSSHPublicKey",
				"ec2:DescribeInstances",
				"ec2:StartInstances",
				"ec2:CreateTags",
				"ec2:CreateSecurityGroup",
				"ec2:AuthorizeSecurityGroupIngress",
				"ec2:DescribeSecurityGroups",
				"ec2:DescribeSubnets",
				"ec2:DescribeKeyPairs",
				"ec2:StopInstances",
				"ec2:TerminateInstances",
				"ec2:RebootInstances",
				"ec2:RunInstances"
			],
			"Resource": "*"
		}
	]
}
```

Then obtain and add the credentials of the user you just created as well as the github repo of your next app to be deployr(ed) to the config.example.yml 

Here is a short [guide](https://www.loom.com/share/cf21a3c2212b45f887e46d73544dabd6?sid=00f3bd28-689c-4480-931d-bd5c4cca247b) on how to do that 

### Step 4: Run the deployr command 

```bash
deployr config.yml
```

or

```bash
deployr <path-to-config.yml>
```

if you have the config.yml in a different directory

you can also see the help menu by running 

```bash
deployr -help
```

### Step 5: Point your domain 

You will get your public IP in your terminal, go to your domain hosting provider and point your domain to that IP 

### Step 6: Your project is Deployed 

You can now access your project on your domain after a few minutes ( Depending on the buildtime of your project )

### Step 7: CI/CD

You can now automate the deployment process with CI/CD. Just copy the main.yml and add it to the .github/workflows folder in your local repository.

You can find the main.yml file here : [main.yml](https://github.com/adityacodes30/deployr/blob/main/workflow/main.sample.yml)

Create a file named main.yml in the .github/workflows folder and paste the contents of the main.sample.yml file in it. Add your domain on line 7 and the key you just received in value field after running deployr in the github repository secrets with the name [DEPLOYR_PRIVKEY] (https://docs.github.com/en/actions/security-for-github-actions/security-guides using-secrets-in-github-actions#creating-secrets-for-a-repository)

for example :

```yaml
  DEPLOY_DOMAIN: https://example.com
```


## If you want to build and contribute to the project

### Step 1: Clone the Project  
Start by forking and  cloning the Deployr repository to your local machine:

```bash
git clone <your-deployr-repo-url>
cd deployr
```

### Step 2: Install Go

Ensure that Go is installed on your system. You can download it from the official Go website. Verify the installation:

```bash
go version
```

### Step 3: Get Credentials

Sign into aws and make a IAM user with the following policy attached: 

```json 
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "deployr",
			"Effect": "Allow",
			"Action": [
				"ec2-instance-connect:SendSSHPublicKey",
				"ec2:DescribeInstances",
				"ec2:StartInstances",
				"ec2:CreateTags",
				"ec2:CreateSecurityGroup",
				"ec2:AuthorizeSecurityGroupIngress",
				"ec2:DescribeSecurityGroups",
				"ec2:DescribeSubnets",
				"ec2:DescribeKeyPairs",
				"ec2:StopInstances",
				"ec2:TerminateInstances",
				"ec2:RebootInstances",
				"ec2:RunInstances"
			],
			"Resource": "*"
		}
	]
}
```

Then obtain and add the credentials of the user you just created as well as the github repo of your next app to be deployr(ed) to the config.example.yml 

Then rename `config.example.yml` ---> `config.yml`

Here is a short [guide](https://www.loom.com/share/cf21a3c2212b45f887e46d73544dabd6?sid=00f3bd28-689c-4480-931d-bd5c4cca247b) on how to do that 


### Step 4: Build and run the project 

Run the command

```bash
go build main.go 
```

### Step 5: Point your domain 

You will get your public IP in your terminal, go to your domain hosting provider and point your domain to that IP 
 
### Step 6: Your project is Deployed 

You can now access your project on your domain after a few minutes ( Depending on the buildtime of your project )


### Development progress

This checklist tracks the current state of tasks.

V 1.0

[✅] **Gather Req**  Ensure AWS access, domain name, and repository URL are available.  

[✅] **AWS Auth** Configure AWS CLI with access keys for required services.  

[✅] **Provision Infrastructure** Create a security group and launch an EC2 instance using AWS APIs.  

[✅] **SSH into EC2 Instance** Verify SSH access and confirm instance connectivity.  

[✅] **Server Daemon** Create a daemon that automatically updates build on pushes to branch ()

[✅] **Execute Deployr Script** Run the provided Deployr script that configures the nginx , processes , ssh and server daemon on the machine

-^----Prototype Complete----^-

[✅] **CLI** Create a CLI for usage

[✅] **Platform Specific Binaries** Create platform specific binaries for easy deployment

[✅] **Documentation** Write a detailed guide on how to use Deployr

-^----v1.0.0 Release----^-

[✅] **Github Actions** Create gh actions to automate the deployment process when new code is pushed to the repository

-^----v1.1.0 Release----^-

[✅] **Build Auth** Add assymentric encryption and authentication for triggering the build process 

[✅] **Multiple Deployments** Add support for multiple deployments (on seperate instances)

[✅] **Interactive CLI** Add an interactive CLI for easy configuration


-^----v1.2.0 Release----^-

[] **Docker** Dockerise the application for windows users 

[] **GUI** Add a supporting GUI for easy accesss

  
---

