# Deployr 💾

Deployr helps you host your Next.js application on your own aws ec2 instance, With just a few clicks, you can set up everything, from machine provisioning to serving your next application on your own domain.

Here is a video walkthrough / how to use Deployr: [Deployr Walkthrough](https://www.loom.com/share/8c7ca17efc78416d8bec92d46bc482ae?sid=10439ae0-bdd7-4ce3-ae1e-85ea4ef9da86)

---

# How to Deployr Your App 🚀

Application is in active development. Upon stable release a cli and platform specific binaries will be released. Meanwhile use these steps to contribute /setup the projects. 


### Step 1: Clone the Project  
Start by cloning the Deployr repository to your local machine:

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
go build main.go deployrconfig.go
```

### Step 5: Point your domain 

You will get your public IP in your terminal, go to your domain hosting provider and point your domain to that IP 

### Step 6: Your project is Deployed 

You can now access your project on your domain after a few minutes ( Depending on the buildtime of your project )


### Development progress

This checklist tracks the current state of tasks. There will be a hosted build for all platforms soon. Currently you can build from source and use it 

V 1.0

[✅] **Gather Req**  Ensure AWS access, domain name, and repository URL are available.  

[✅] **AWS Auth** Configure AWS CLI with access keys for required services.  

[✅] **Provision Infrastructure** Create a security group and launch an EC2 instance using AWS APIs.  

[✅] **SSH into EC2 Instance** Verify SSH access and confirm instance connectivity.  

[✅] **Server Daemon** Create a daemon that automatically updates build on pushes to branch ()

[✅] **Execute Deployr Script** Run the provided Deployr script that configures the nginx , processes , ssh and server daemon on the machine

-^----Prototype Complete----^-

[] **Github Actions** Create gh actions to automate the deployment process when new code is pushed to the repository

[] **CLI** Create a CLI for usage

[] **Platform Specific Binaries** Create platform specific binaries for easy deployment

[] **Documentation** Write a detailed guide on how to use Deployr

-^----V1 Release----^-

[] **Docker** Dockerise the application for easy deployment
  
---

