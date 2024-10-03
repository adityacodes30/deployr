# Deployr 💾

Deployr helps you host your Next.js application on your own aws ec2 instance, With just a few clicks, you can set up everything, from machine provisioning to serving your next application on your own domain.

---

### Development progress

This checklist tracks the current state of tasks. There will be a hosted build for all platforms soon 

[✅] **Gather Req**  Ensure AWS access, domain name, and repository URL are available.  

[] **AWS Auth** Configure AWS CLI with access keys for required services.  

[] **Provision Infrastructure** Create a security group and launch an EC2 instance using AWS APIs.  

[] **SSH into EC2 Instance** Verify SSH access and confirm instance connectivity.  

[] **Server Daemon** Create a daemon that automatically updates build on pushes to branch

[] **Execute Deployr Script** Run the provided Deployr script that configures the nginx , processes , ssh and server daemon on the machine
  
---

